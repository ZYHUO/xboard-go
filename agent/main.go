package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var (
	panelURL   string
	token      string
	configPath string
	singboxBin string
)

func init() {
	flag.StringVar(&panelURL, "panel", "", "面板地址 (如: https://your-panel.com)")
	flag.StringVar(&token, "token", "", "主机 Token")
	flag.StringVar(&configPath, "config", "/etc/sing-box/config.json", "sing-box 配置文件路径")
	flag.StringVar(&singboxBin, "singbox", "sing-box", "sing-box 可执行文件路径")
}

type AgentConfig struct {
	SingBoxConfig map[string]interface{} `json:"singbox_config"`
	Nodes         []NodeConfig           `json:"nodes"`
}

type NodeConfig struct {
	ID    int64                    `json:"id"`
	Type  string                   `json:"type"`
	Port  int                      `json:"port"`
	Tag   string                   `json:"tag"`
	Users []map[string]interface{} `json:"users"`
}

type Agent struct {
	panelURL      string
	token         string
	configPath    string
	singboxBin    string
	singboxCmd    *exec.Cmd
	lastConfig    string
	httpClient    *http.Client
	userVersions  map[int64]int64        // 节点用户版本缓存
	userHashes    map[int64]string       // 节点用户哈希缓存
	lastTraffic   map[string]TrafficData // 上次流量数据，用于计算增量
	nodeConfigs   []NodeConfig           // 当前节点配置
	clashAPIPort  int                    // Clash API 端口
	portUserMap   map[int][]string       // 端口到用户的映射（用于单端口多用户场景）
}

// TrafficData 流量数据
type TrafficData struct {
	Upload   int64
	Download int64
}

func NewAgent() *Agent {
	return &Agent{
		panelURL:     panelURL,
		token:        token,
		configPath:   configPath,
		singboxBin:   singboxBin,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		userVersions: make(map[int64]int64),
		userHashes:   make(map[int64]string),
		lastTraffic:  make(map[string]TrafficData),
		portUserMap:  make(map[int][]string),
		clashAPIPort: 9090,
	}
}

// getNodeUsers 获取节点用户（支持增量同步）
// nodeType: "server" 或 "node"
func (a *Agent) getNodeUsers(nodeID int64, nodeType string) ([]map[string]interface{}, bool, error) {
	hash := a.userHashes[nodeID]

	url := fmt.Sprintf("/users?node_id=%d&type=%s&hash=%s", nodeID, nodeType, hash)
	result, err := a.apiRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("invalid response")
	}

	hasChange, _ := data["has_change"].(bool)
	if !hasChange {
		return nil, false, nil
	}

	// 更新哈希
	if h, ok := data["hash"].(string); ok {
		a.userHashes[nodeID] = h
	}

	users, ok := data["users"].([]interface{})
	if !ok {
		return nil, true, nil
	}

	result_users := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		if user, ok := u.(map[string]interface{}); ok {
			result_users = append(result_users, user)
		}
	}

	return result_users, true, nil
}

func (a *Agent) apiRequest(method, path string, body interface{}) (map[string]interface{}, error) {
	url := a.panelURL + "/api/v1/agent" + path
	
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", a.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		if errMsg, ok := result["error"].(string); ok {
			return nil, fmt.Errorf(errMsg)
		}
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return result, nil
}

func (a *Agent) sendHeartbeat() error {
	systemInfo := map[string]interface{}{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"cpus":    runtime.NumCPU(),
		"version": "1.0.0",
	}

	_, err := a.apiRequest("POST", "/heartbeat", map[string]interface{}{
		"system_info": systemInfo,
	})
	return err
}

func (a *Agent) getConfig() (*AgentConfig, error) {
	result, err := a.apiRequest("GET", "/config", nil)
	if err != nil {
		return nil, err
	}

	data, ok := result["data"]
	if !ok {
		return nil, fmt.Errorf("invalid response")
	}

	configData, _ := json.Marshal(data)
	var config AgentConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (a *Agent) updateConfig(config *AgentConfig) (bool, error) {
	// 保存节点配置用于流量上报
	a.nodeConfigs = config.Nodes

	// 构建端口到用户的映射
	a.portUserMap = make(map[int][]string)
	for _, node := range config.Nodes {
		users := make([]string, 0, len(node.Users))
		for _, user := range node.Users {
			if name, ok := user["name"].(string); ok {
				users = append(users, name)
			}
		}
		a.portUserMap[node.Port] = users
	}

	// 注入用户到 inbounds
	singboxConfig := config.SingBoxConfig
	hasUserChange := false

	if inbounds, ok := singboxConfig["inbounds"].([]interface{}); ok {
		for i, inbound := range inbounds {
			if ib, ok := inbound.(map[string]interface{}); ok {
				tag, _ := ib["tag"].(string)
				// 找到对应的节点配置
				for _, node := range config.Nodes {
					if node.Tag == tag {
						// 直接使用配置中的用户（已经是正确格式）
						// 不再单独调用用户接口，因为 GetAgentConfig 已经返回了正确格式的用户
						if len(node.Users) > 0 {
							ib["users"] = node.Users
							hasUserChange = true
						}
						inbounds[i] = ib
						break
					}
				}
			}
		}
		singboxConfig["inbounds"] = inbounds
	}

	// 添加 experimental 配置用于流量统计
	if _, ok := singboxConfig["experimental"]; !ok {
		singboxConfig["experimental"] = map[string]interface{}{}
	}
	experimental := singboxConfig["experimental"].(map[string]interface{})
	
	// 添加 Clash API 用于获取连接信息
	if _, ok := experimental["clash_api"]; !ok {
		experimental["clash_api"] = map[string]interface{}{
			"external_controller": fmt.Sprintf("127.0.0.1:%d", a.clashAPIPort),
		}
	}
	singboxConfig["experimental"] = experimental

	configJSON, _ := json.MarshalIndent(singboxConfig, "", "  ")
	configStr := string(configJSON)

	if configStr == a.lastConfig && !hasUserChange {
		return false, nil
	}

	// 写入配置文件
	if err := os.WriteFile(a.configPath, configJSON, 0644); err != nil {
		return false, err
	}

	a.lastConfig = configStr
	return true, nil
}

func (a *Agent) startSingbox() error {
	a.stopSingbox()

	a.singboxCmd = exec.Command(a.singboxBin, "run", "-c", a.configPath)
	a.singboxCmd.Stdout = os.Stdout
	a.singboxCmd.Stderr = os.Stderr

	if err := a.singboxCmd.Start(); err != nil {
		return err
	}

	fmt.Println("✓ sing-box 已启动")
	return nil
}

func (a *Agent) stopSingbox() {
	if a.singboxCmd != nil && a.singboxCmd.Process != nil {
		a.singboxCmd.Process.Signal(syscall.SIGTERM)
		a.singboxCmd.Wait()
		fmt.Println("✓ sing-box 已停止")
	}
}

// ConnectionTraffic 连接流量记录
type ConnectionTraffic struct {
	Upload   int64
	Download int64
}

// getTrafficFromClashAPI 从 Clash API 获取流量统计
// 通过跟踪每个连接的流量变化来计算用户流量
func (a *Agent) getTrafficFromClashAPI() (map[string]TrafficData, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/connections", a.clashAPIPort)
	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 使用 map 解析以支持不同版本的 sing-box
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 按用户聚合当前连接的流量
	traffic := make(map[string]TrafficData)
	
	connections, ok := result["connections"].([]interface{})
	if !ok {
		return traffic, nil
	}

	for _, c := range connections {
		conn, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		upload, _ := conn["upload"].(float64)
		download, _ := conn["download"].(float64)

		// 获取用户名，尝试多种字段
		var user string
		if metadata, ok := conn["metadata"].(map[string]interface{}); ok {
			// 尝试不同的字段名
			if u, ok := metadata["inboundUser"].(string); ok && u != "" {
				user = u
			} else if u, ok := metadata["user"].(string); ok && u != "" {
				user = u
			} else if u, ok := metadata["inbound_user"].(string); ok && u != "" {
				user = u
			}
		}

		if user == "" {
			continue
		}

		data := traffic[user]
		data.Upload += int64(upload)
		data.Download += int64(download)
		traffic[user] = data
	}

	return traffic, nil
}

// reportTraffic 上报流量到面板
func (a *Agent) reportTraffic() error {
	// 尝试从 Clash API 获取用户流量
	traffic, err := a.getTrafficFromClashAPI()
	if err != nil {
		// Clash API 不可用，使用端口流量平均分配方案
		return a.reportTrafficByPort()
	}

	// 调试：打印获取到的流量数据
	if len(traffic) > 0 {
		fmt.Printf("📊 获取到 %d 个用户的流量数据\n", len(traffic))
	}

	// 计算增量流量
	trafficReport := make([]map[string]interface{}, 0)
	for user, data := range traffic {
		last := a.lastTraffic[user]
		uploadDelta := data.Upload - last.Upload
		downloadDelta := data.Download - last.Download

		// 只上报有增量的用户
		if uploadDelta > 0 || downloadDelta > 0 {
			trafficReport = append(trafficReport, map[string]interface{}{
				"username": user,
				"upload":   uploadDelta,
				"download": downloadDelta,
			})
			fmt.Printf("  用户 %s: ↑%.2f MB ↓%.2f MB\n", user, float64(uploadDelta)/1024/1024, float64(downloadDelta)/1024/1024)
		}
		a.lastTraffic[user] = data
	}

	if len(trafficReport) == 0 {
		// 没有用户流量，尝试端口流量方案
		return a.reportTrafficByPort()
	}

	// 构建上报数据
	nodes := make([]map[string]interface{}, 0)
	for _, node := range a.nodeConfigs {
		nodes = append(nodes, map[string]interface{}{
			"id":    node.ID,
			"users": trafficReport,
		})
	}

	_, err = a.apiRequest("POST", "/traffic", map[string]interface{}{
		"nodes": nodes,
	})
	if err != nil {
		fmt.Printf("⚠ 流量上报失败: %v\n", err)
	} else {
		fmt.Printf("✓ 已上报 %d 个用户的流量\n", len(trafficReport))
	}
	return err
}

// reportTrafficByPort 通过端口流量平均分配给用户（备用方案）
func (a *Agent) reportTrafficByPort() error {
	// 获取总流量
	url := fmt.Sprintf("http://127.0.0.1:%d/traffic", a.clashAPIPort)
	resp, err := a.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Up   int64 `json:"up"`
		Down int64 `json:"down"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// 如果没有流量，直接返回
	if result.Up == 0 && result.Down == 0 {
		return nil
	}

	// 计算增量
	lastTotal := a.lastTraffic["__total__"]
	uploadDelta := result.Up - lastTotal.Upload
	downloadDelta := result.Down - lastTotal.Download

	if uploadDelta <= 0 && downloadDelta <= 0 {
		return nil
	}

	a.lastTraffic["__total__"] = TrafficData{
		Upload:   result.Up,
		Download: result.Down,
	}

	fmt.Printf("📊 总流量: ↑%.2f MB ↓%.2f MB\n", float64(uploadDelta)/1024/1024, float64(downloadDelta)/1024/1024)

	// 为每个节点的所有用户平均分配流量
	nodes := make([]map[string]interface{}, 0)
	for _, node := range a.nodeConfigs {
		users := a.portUserMap[node.Port]
		if len(users) == 0 {
			continue
		}

		// 平均分配流量
		avgUpload := uploadDelta / int64(len(users))
		avgDownload := downloadDelta / int64(len(users))

		trafficReport := make([]map[string]interface{}, 0, len(users))
		for _, user := range users {
			trafficReport = append(trafficReport, map[string]interface{}{
				"username": user,
				"upload":   avgUpload,
				"download": avgDownload,
			})
		}

		nodes = append(nodes, map[string]interface{}{
			"id":    node.ID,
			"users": trafficReport,
		})

		fmt.Printf("  节点 %d: 为 %d 个用户平均分配流量\n", node.ID, len(users))
	}

	if len(nodes) == 0 {
		return nil
	}

	_, err = a.apiRequest("POST", "/traffic", map[string]interface{}{
		"nodes": nodes,
	})
	if err != nil {
		fmt.Printf("⚠ 流量上报失败: %v\n", err)
	} else {
		fmt.Printf("✓ 已上报流量（平均分配模式）\n")
	}
	return err
}

func (a *Agent) Run() {
	fmt.Println("XBoard Agent v1.0.0")
	fmt.Printf("面板: %s\n", a.panelURL)
	fmt.Println("正在连接...")

	// 首次获取配置并启动
	config, err := a.getConfig()
	if err != nil {
		fmt.Printf("✗ 获取配置失败: %v\n", err)
		os.Exit(1)
	}

	if _, err := a.updateConfig(config); err != nil {
		fmt.Printf("✗ 更新配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := a.startSingbox(); err != nil {
		fmt.Printf("✗ 启动 sing-box 失败: %v\n", err)
		os.Exit(1)
	}

	// 发送首次心跳
	if err := a.sendHeartbeat(); err != nil {
		fmt.Printf("⚠ 心跳发送失败: %v\n", err)
	} else {
		fmt.Println("✓ 已连接到面板")
	}

	// 启动定时任务
	heartbeatTicker := time.NewTicker(30 * time.Second)
	configTicker := time.NewTicker(60 * time.Second)
	trafficTicker := time.NewTicker(60 * time.Second) // 每分钟上报流量

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-heartbeatTicker.C:
			if err := a.sendHeartbeat(); err != nil {
				fmt.Printf("⚠ 心跳失败: %v\n", err)
			}

		case <-trafficTicker.C:
			if err := a.reportTraffic(); err != nil {
				// 流量上报失败不打印错误，可能是 sing-box 还没启动完成
			}

		case <-configTicker.C:
			config, err := a.getConfig()
			if err != nil {
				fmt.Printf("⚠ 获取配置失败: %v\n", err)
				continue
			}

			updated, err := a.updateConfig(config)
			if err != nil {
				fmt.Printf("⚠ 更新配置失败: %v\n", err)
				continue
			}

			if updated {
				fmt.Println("配置已更新，重启 sing-box...")
				if err := a.startSingbox(); err != nil {
					fmt.Printf("✗ 重启失败: %v\n", err)
				}
			}

		case sig := <-sigChan:
			fmt.Printf("\n收到信号 %v，正在退出...\n", sig)
			heartbeatTicker.Stop()
			configTicker.Stop()
			trafficTicker.Stop()
			a.stopSingbox()
			return
		}
	}
}

func main() {
	flag.Parse()

	if panelURL == "" || token == "" {
		fmt.Println("用法: xboard-agent -panel <面板地址> -token <主机Token>")
		fmt.Println()
		fmt.Println("参数:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	agent := NewAgent()
	agent.Run()
}
