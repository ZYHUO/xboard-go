package protocol

import (
	"fmt"
	"regexp"
	"strings"
)

// ProtocolHandler 协议处理器
type ProtocolHandler struct {
	registry *AdapterRegistry
}

// NewProtocolHandler 创建协议处理器
func NewProtocolHandler() *ProtocolHandler {
	return &ProtocolHandler{
		registry: NewAdapterRegistry(),
	}
}

// StandardizeVMessParams 标准化 VMess 参数
func (ph *ProtocolHandler) StandardizeVMessParams(params map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// UUID 处理
	uuid, ok := params["uuid"].(string)
	if !ok || uuid == "" {
		return nil, fmt.Errorf("VMess UUID is required")
	}
	
	// 标准化 UUID 格式
	uuid = strings.ToLower(uuid)
	if !isValidUUID(uuid) {
		return nil, fmt.Errorf("invalid VMess UUID format")
	}
	result["uuid"] = uuid

	// alterID 标准化
	alterID := 0
	if aid, ok := params["alter_id"].(float64); ok {
		alterID = int(aid)
	} else if aid, ok := params["alterId"].(float64); ok {
		alterID = int(aid)
	} else if aid, ok := params["aid"].(float64); ok {
		alterID = int(aid)
	}
	
	if alterID < 0 || alterID > 65535 {
		return nil, fmt.Errorf("VMess alterID must be between 0 and 65535")
	}
	result["alter_id"] = alterID

	// security 标准化
	security := "auto"
	if sec, ok := params["security"].(string); ok && sec != "" {
		security = strings.ToLower(sec)
		validSecurity := []string{"auto", "aes-128-gcm", "chacha20-poly1305", "none"}
		if !contains(validSecurity, security) {
			return nil, fmt.Errorf("invalid VMess security method: %s", security)
		}
	}
	result["security"] = security

	// network 标准化
	network := "tcp"
	if net, ok := params["network"].(string); ok && net != "" {
		network = strings.ToLower(net)
		validNetworks := []string{"tcp", "ws", "grpc", "h2", "quic"}
		if !contains(validNetworks, network) {
			return nil, fmt.Errorf("invalid VMess network type: %s", network)
		}
	}
	result["network"] = network

	// TLS 配置标准化
	if tls, ok := params["tls"].(map[string]interface{}); ok {
		tlsConfig, err := ph.standardizeTLSConfig(tls)
		if err != nil {
			return nil, fmt.Errorf("invalid VMess TLS config: %w", err)
		}
		result["tls"] = tlsConfig
	}

	// 网络设置标准化
	if networkSettings, ok := params["network_settings"].(map[string]interface{}); ok {
		netConfig, err := ph.standardizeNetworkSettings(network, networkSettings)
		if err != nil {
			return nil, fmt.Errorf("invalid VMess network settings: %w", err)
		}
		result["network_settings"] = netConfig
	}

	return result, nil
}

// StandardizeVLessParams 标准化 VLESS 参数
func (ph *ProtocolHandler) StandardizeVLessParams(params map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// UUID 处理
	uuid, ok := params["uuid"].(string)
	if !ok || uuid == "" {
		return nil, fmt.Errorf("VLESS UUID is required")
	}
	
	uuid = strings.ToLower(uuid)
	if !isValidUUID(uuid) {
		return nil, fmt.Errorf("invalid VLESS UUID format")
	}
	result["uuid"] = uuid

	// flow 标准化
	flow := ""
	if f, ok := params["flow"].(string); ok {
		flow = strings.ToLower(f)
		validFlows := []string{"", "xtls-rprx-vision", "xtls-rprx-vision-udp443"}
		if !contains(validFlows, flow) {
			return nil, fmt.Errorf("invalid VLESS flow: %s", flow)
		}
	}
	result["flow"] = flow

	// encryption 标准化（VLESS 只支持 none）
	encryption := "none"
	if enc, ok := params["encryption"].(string); ok && enc != "" {
		if strings.ToLower(enc) != "none" {
			return nil, fmt.Errorf("VLESS encryption must be 'none', got: %s", enc)
		}
	}
	result["encryption"] = encryption

	// 数据包编码
	result["packet_encoding"] = "xudp"

	// network 标准化
	network := "tcp"
	if net, ok := params["network"].(string); ok && net != "" {
		network = strings.ToLower(net)
		validNetworks := []string{"tcp", "ws", "grpc", "h2", "quic"}
		if !contains(validNetworks, network) {
			return nil, fmt.Errorf("invalid VLESS network type: %s", network)
		}
	}
	result["network"] = network

	// TLS 配置标准化
	if tls, ok := params["tls"].(map[string]interface{}); ok {
		tlsConfig, err := ph.standardizeVLessTLSConfig(tls)
		if err != nil {
			return nil, fmt.Errorf("invalid VLESS TLS config: %w", err)
		}
		result["tls"] = tlsConfig
	}

	// Reality 配置标准化
	if reality, ok := params["reality"].(map[string]interface{}); ok {
		realityConfig, err := ph.standardizeRealityConfig(reality)
		if err != nil {
			return nil, fmt.Errorf("invalid VLESS Reality config: %w", err)
		}
		result["reality"] = realityConfig
	}

	// 网络设置标准化
	if networkSettings, ok := params["network_settings"].(map[string]interface{}); ok {
		netConfig, err := ph.standardizeNetworkSettings(network, networkSettings)
		if err != nil {
			return nil, fmt.Errorf("invalid VLESS network settings: %w", err)
		}
		result["network_settings"] = netConfig
	}

	return result, nil
}

// StandardizeTrojanParams 标准化 Trojan 参数
func (ph *ProtocolHandler) StandardizeTrojanParams(params map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// password 处理
	password, ok := params["password"].(string)
	if !ok || password == "" {
		return nil, fmt.Errorf("Trojan password is required")
	}
	result["password"] = password

	// network 标准化
	network := "tcp"
	if net, ok := params["network"].(string); ok && net != "" {
		network = strings.ToLower(net)
		validNetworks := []string{"tcp", "ws", "grpc"}
		if !contains(validNetworks, network) {
			return nil, fmt.Errorf("invalid Trojan network type: %s", network)
		}
	}
	result["network"] = network

	// TLS 配置标准化（Trojan 默认需要 TLS）
	tlsConfig := map[string]interface{}{
		"enabled": true,
	}

	// server_name/sni 标准化
	if serverName, ok := params["server_name"].(string); ok && serverName != "" {
		tlsConfig["server_name"] = serverName
	} else if sni, ok := params["sni"].(string); ok && sni != "" {
		tlsConfig["server_name"] = sni
	}

	// allow_insecure 标准化
	if insecure, ok := params["allow_insecure"].(bool); ok {
		tlsConfig["insecure"] = insecure
	}

	// ALPN 标准化
	if alpn, ok := params["alpn"].([]interface{}); ok {
		alpnStrings := make([]string, 0, len(alpn))
		for _, a := range alpn {
			if alpnStr, ok := a.(string); ok {
				alpnStrings = append(alpnStrings, alpnStr)
			}
		}
		if len(alpnStrings) > 0 {
			tlsConfig["alpn"] = alpnStrings
		}
	} else if alpnStr, ok := params["alpn"].(string); ok && alpnStr != "" {
		// 处理逗号分隔的 ALPN 字符串
		alpnList := strings.Split(alpnStr, ",")
		for i, a := range alpnList {
			alpnList[i] = strings.TrimSpace(a)
		}
		tlsConfig["alpn"] = alpnList
	}

	result["tls"] = tlsConfig

	// 网络设置标准化
	if networkSettings, ok := params["network_settings"].(map[string]interface{}); ok {
		netConfig, err := ph.standardizeNetworkSettings(network, networkSettings)
		if err != nil {
			return nil, fmt.Errorf("invalid Trojan network settings: %w", err)
		}
		result["network_settings"] = netConfig
	}

	return result, nil
}

// StandardizeShadowsocksParams 标准化 Shadowsocks 参数
func (ph *ProtocolHandler) StandardizeShadowsocksParams(params map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// method 标准化
	method, ok := params["method"].(string)
	if !ok || method == "" {
		return nil, fmt.Errorf("Shadowsocks method is required")
	}
	
	method = strings.ToLower(method)
	validMethods := []string{
		"aes-128-gcm", "aes-192-gcm", "aes-256-gcm",
		"chacha20-ietf-poly1305", "xchacha20-ietf-poly1305",
		"2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305",
	}
	
	if !contains(validMethods, method) {
		return nil, fmt.Errorf("invalid Shadowsocks method: %s", method)
	}
	result["method"] = method

	// password 处理
	password, ok := params["password"].(string)
	if !ok || password == "" {
		return nil, fmt.Errorf("Shadowsocks password is required")
	}
	result["password"] = password

	// plugin 标准化
	if plugin, ok := params["plugin"].(string); ok && plugin != "" {
		validPlugins := []string{"obfs-local", "v2ray-plugin", "kcptun", "simple-obfs"}
		if !contains(validPlugins, plugin) {
			return nil, fmt.Errorf("unsupported Shadowsocks plugin: %s", plugin)
		}
		result["plugin"] = plugin

		// plugin_opts 处理
		if pluginOpts, ok := params["plugin_opts"].(string); ok {
			result["plugin_opts"] = pluginOpts
		}
	}

	return result, nil
}

// standardizeTLSConfig 标准化 TLS 配置
func (ph *ProtocolHandler) standardizeTLSConfig(tls map[string]interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"enabled": true,
	}

	if serverName, ok := tls["server_name"].(string); ok && serverName != "" {
		result["server_name"] = serverName
	}

	if insecure, ok := tls["allow_insecure"].(bool); ok {
		result["insecure"] = insecure
	} else if insecure, ok := tls["insecure"].(bool); ok {
		result["insecure"] = insecure
	}

	if alpn, ok := tls["alpn"].([]interface{}); ok {
		alpnStrings := make([]string, 0, len(alpn))
		for _, a := range alpn {
			if alpnStr, ok := a.(string); ok {
				alpnStrings = append(alpnStrings, alpnStr)
			}
		}
		if len(alpnStrings) > 0 {
			result["alpn"] = alpnStrings
		}
	}

	return result, nil
}

// standardizeVLessTLSConfig 标准化 VLESS TLS 配置
func (ph *ProtocolHandler) standardizeVLessTLSConfig(tls map[string]interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"enabled": true,
		"utls": map[string]interface{}{
			"enabled":     true,
			"fingerprint": "chrome",
		},
	}

	if serverName, ok := tls["server_name"].(string); ok && serverName != "" {
		result["server_name"] = serverName
	}

	if insecure, ok := tls["allow_insecure"].(bool); ok {
		result["insecure"] = insecure
	}

	// UTLS 配置
	if utls, ok := tls["utls"].(map[string]interface{}); ok {
		utlsConfig := result["utls"].(map[string]interface{})
		if fingerprint, ok := utls["fingerprint"].(string); ok {
			validFingerprints := []string{"chrome", "firefox", "safari", "ios", "android", "edge", "360", "qq", "random", "randomized"}
			if contains(validFingerprints, fingerprint) {
				utlsConfig["fingerprint"] = fingerprint
			}
		}
	}

	return result, nil
}

// standardizeRealityConfig 标准化 Reality 配置
func (ph *ProtocolHandler) standardizeRealityConfig(reality map[string]interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"enabled": true,
	}

	if publicKey, ok := reality["public_key"].(string); ok && publicKey != "" {
		result["public_key"] = publicKey
	} else {
		return nil, fmt.Errorf("Reality public_key is required")
	}

	if shortID, ok := reality["short_id"].(string); ok && shortID != "" {
		result["short_id"] = shortID
	}

	return result, nil
}

// standardizeNetworkSettings 标准化网络设置
func (ph *ProtocolHandler) standardizeNetworkSettings(network string, settings map[string]interface{}) (map[string]interface{}, error) {
	switch network {
	case "ws":
		return ph.standardizeWSSettings(settings)
	case "grpc":
		return ph.standardizeGRPCSettings(settings)
	case "h2":
		return ph.standardizeH2Settings(settings)
	case "quic":
		return ph.standardizeQUICSettings(settings)
	default:
		return settings, nil
	}
}

// standardizeWSSettings 标准化 WebSocket 设置
func (ph *ProtocolHandler) standardizeWSSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if path, ok := settings["path"].(string); ok {
		result["path"] = path
	} else {
		result["path"] = "/"
	}

	if headers, ok := settings["headers"].(map[string]interface{}); ok {
		result["headers"] = headers
	}

	return result, nil
}

// standardizeGRPCSettings 标准化 gRPC 设置
func (ph *ProtocolHandler) standardizeGRPCSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if serviceName, ok := settings["serviceName"].(string); ok {
		result["service_name"] = serviceName
	} else if serviceName, ok := settings["service_name"].(string); ok {
		result["service_name"] = serviceName
	} else {
		result["service_name"] = "TunService"
	}

	return result, nil
}

// standardizeH2Settings 标准化 HTTP/2 设置
func (ph *ProtocolHandler) standardizeH2Settings(settings map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if path, ok := settings["path"].(string); ok {
		result["path"] = path
	} else {
		result["path"] = "/"
	}

	if host, ok := settings["host"].([]interface{}); ok {
		hostStrings := make([]string, 0, len(host))
		for _, h := range host {
			if hostStr, ok := h.(string); ok {
				hostStrings = append(hostStrings, hostStr)
			}
		}
		result["host"] = hostStrings
	}

	return result, nil
}

// standardizeQUICSettings 标准化 QUIC 设置
func (ph *ProtocolHandler) standardizeQUICSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if security, ok := settings["security"].(string); ok {
		result["security"] = security
	} else {
		result["security"] = "none"
	}

	if key, ok := settings["key"].(string); ok {
		result["key"] = key
	}

	if headerType, ok := settings["header"].(map[string]interface{}); ok {
		result["header"] = headerType
	}

	return result, nil
}

// 辅助函数
func isValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(uuid)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}