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
	panelURL    string
	token       string
	configPath  string
	singboxBin  string
	singboxCmd  *exec.Cmd
	lastConfig  string
	httpClient  *http.Client
}

func NewAgent() *Agent {
	return &Agent{
		panelURL:   panelURL,
		token:      token,
		configPath: configPath,
		singboxBin: singboxBin,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
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
	// 注入用户到 inbounds
	singboxConfig := config.SingBoxConfig
	if inbounds, ok := singboxConfig["inbounds"].([]interface{}); ok {
		for i, inbound := range inbounds {
			if ib, ok := inbound.(map[string]interface{}); ok {
				tag := ib["tag"].(string)
				// 找到对应的节点配置
				for _, node := range config.Nodes {
					if node.Tag == tag && len(node.Users) > 0 {
						ib["users"] = node.Users
						inbounds[i] = ib
						break
					}
				}
			}
		}
		singboxConfig["inbounds"] = inbounds
	}

	configJSON, _ := json.MarshalIndent(singboxConfig, "", "  ")
	configStr := string(configJSON)

	if configStr == a.lastConfig {
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

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-heartbeatTicker.C:
			if err := a.sendHeartbeat(); err != nil {
				fmt.Printf("⚠ 心跳失败: %v\n", err)
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
