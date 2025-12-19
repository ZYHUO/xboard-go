package protocol

import (
	"fmt"
	"strings"
)

// ProtocolAdapter 协议适配器接口
type ProtocolAdapter interface {
	GetProtocolName() string
	ConvertParams(params map[string]interface{}) (map[string]interface{}, error)
	ValidateParams(params map[string]interface{}) error
	GetRequiredParams() []string
	GetOptionalParams() []string
}

// ProtocolParamError 协议参数错误
type ProtocolParamError struct {
	Protocol string
	Param    string
	Expected string
	Actual   interface{}
	Message  string
}

func (e *ProtocolParamError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("protocol %s parameter error for '%s': %s", e.Protocol, e.Param, e.Message)
	}
	return fmt.Sprintf("protocol %s parameter error for '%s': expected %s, got %v", 
		e.Protocol, e.Param, e.Expected, e.Actual)
}

// VMessAdapter VMess 协议适配器
type VMessAdapter struct{}

func (va *VMessAdapter) GetProtocolName() string {
	return "vmess"
}

func (va *VMessAdapter) GetRequiredParams() []string {
	return []string{"uuid"}
}

func (va *VMessAdapter) GetOptionalParams() []string {
	return []string{"alter_id", "security", "network", "tls", "network_settings"}
}

func (va *VMessAdapter) ConvertParams(params map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 必需参数：UUID
	uuid, ok := params["uuid"].(string)
	if !ok || uuid == "" {
		return nil, &ProtocolParamError{
			Protocol: "vmess",
			Param:    "uuid",
			Expected: "non-empty string",
			Actual:   params["uuid"],
		}
	}
	result["uuid"] = uuid

	// 可选参数：alter_id
	if alterId, ok := params["alter_id"].(float64); ok {
		result["alter_id"] = int(alterId)
	} else {
		result["alter_id"] = 0
	}

	// 可选参数：security
	if security, ok := params["security"].(string); ok {
		result["security"] = security
	} else {
		result["security"] = "auto"
	}

	// TLS 配置
	if tls, ok := params["tls"].(map[string]interface{}); ok {
		tlsConfig := make(map[string]interface{})
		tlsConfig["enabled"] = true

		if serverName, ok := tls["server_name"].(string); ok {
			tlsConfig["server_name"] = serverName
		}

		if insecure, ok := tls["allow_insecure"].(bool); ok {
			tlsConfig["insecure"] = insecure
		}

		result["tls"] = tlsConfig
	}

	// 传输配置
	if network, ok := params["network"].(string); ok {
		result["network"] = network
		if networkSettings, ok := params["network_settings"].(map[string]interface{}); ok {
			result["network_settings"] = networkSettings
		}
	}

	return result, nil
}

func (va *VMessAdapter) ValidateParams(params map[string]interface{}) error {
	// 验证 UUID
	uuid, ok := params["uuid"].(string)
	if !ok || uuid == "" {
		return &ProtocolParamError{
			Protocol: "vmess",
			Param:    "uuid",
			Message:  "UUID is required and cannot be empty",
		}
	}

	// 验证 UUID 格式
	if len(uuid) != 36 || strings.Count(uuid, "-") != 4 {
		return &ProtocolParamError{
			Protocol: "vmess",
			Param:    "uuid",
			Message:  "invalid UUID format",
		}
	}

	// 验证 alter_id
	if alterId, ok := params["alter_id"].(float64); ok {
		if alterId < 0 || alterId > 65535 {
			return &ProtocolParamError{
				Protocol: "vmess",
				Param:    "alter_id",
				Message:  "alter_id must be between 0 and 65535",
			}
		}
	}

	// 验证 security
	if security, ok := params["security"].(string); ok {
		validSecurity := []string{"auto", "aes-128-gcm", "chacha20-poly1305", "none"}
		valid := false
		for _, v := range validSecurity {
			if security == v {
				valid = true
				break
			}
		}
		if !valid {
			return &ProtocolParamError{
				Protocol: "vmess",
				Param:    "security",
				Message:  fmt.Sprintf("invalid security method, must be one of: %s", strings.Join(validSecurity, ", ")),
			}
		}
	}

	return nil
}

// VLessAdapter VLESS 协议适配器
type VLessAdapter struct{}

func (vla *VLessAdapter) GetProtocolName() string {
	return "vless"
}

func (vla *VLessAdapter) GetRequiredParams() []string {
	return []string{"uuid"}
}

func (vla *VLessAdapter) GetOptionalParams() []string {
	return []string{"flow", "encryption", "network", "tls", "reality", "network_settings"}
}

func (vla *VLessAdapter) ConvertParams(params map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 必需参数：UUID
	uuid, ok := params["uuid"].(string)
	if !ok || uuid == "" {
		return nil, &ProtocolParamError{
			Protocol: "vless",
			Param:    "uuid",
			Expected: "non-empty string",
			Actual:   params["uuid"],
		}
	}
	result["uuid"] = uuid

	// 可选参数：flow
	if flow, ok := params["flow"].(string); ok {
		result["flow"] = flow
	}

	// 可选参数：encryption
	if encryption, ok := params["encryption"].(string); ok {
		result["encryption"] = encryption
	} else {
		result["encryption"] = "none"
	}

	// 数据包编码
	result["packet_encoding"] = "xudp"

	// TLS 配置
	if tls, ok := params["tls"].(map[string]interface{}); ok {
		tlsConfig := make(map[string]interface{})
		tlsConfig["enabled"] = true

		if serverName, ok := tls["server_name"].(string); ok {
			tlsConfig["server_name"] = serverName
		}

		// UTLS 配置
		tlsConfig["utls"] = map[string]interface{}{
			"enabled":     true,
			"fingerprint": "chrome",
		}

		// Reality 配置
		if reality, ok := params["reality"].(map[string]interface{}); ok {
			realityConfig := map[string]interface{}{
				"enabled": true,
			}

			if publicKey, ok := reality["public_key"].(string); ok {
				realityConfig["public_key"] = publicKey
			}

			if shortID, ok := reality["short_id"].(string); ok {
				realityConfig["short_id"] = shortID
			}

			tlsConfig["reality"] = realityConfig
		}

		result["tls"] = tlsConfig
	}

	// 传输配置
	if network, ok := params["network"].(string); ok {
		result["network"] = network
		if networkSettings, ok := params["network_settings"].(map[string]interface{}); ok {
			result["network_settings"] = networkSettings
		}
	}

	return result, nil
}

func (vla *VLessAdapter) ValidateParams(params map[string]interface{}) error {
	// 验证 UUID
	uuid, ok := params["uuid"].(string)
	if !ok || uuid == "" {
		return &ProtocolParamError{
			Protocol: "vless",
			Param:    "uuid",
			Message:  "UUID is required and cannot be empty",
		}
	}

	// 验证 UUID 格式
	if len(uuid) != 36 || strings.Count(uuid, "-") != 4 {
		return &ProtocolParamError{
			Protocol: "vless",
			Param:    "uuid",
			Message:  "invalid UUID format",
		}
	}

	// 验证 flow
	if flow, ok := params["flow"].(string); ok {
		validFlows := []string{"", "xtls-rprx-vision", "xtls-rprx-vision-udp443"}
		valid := false
		for _, v := range validFlows {
			if flow == v {
				valid = true
				break
			}
		}
		if !valid {
			return &ProtocolParamError{
				Protocol: "vless",
				Param:    "flow",
				Message:  fmt.Sprintf("invalid flow, must be one of: %s", strings.Join(validFlows, ", ")),
			}
		}
	}

	// 验证 encryption
	if encryption, ok := params["encryption"].(string); ok {
		if encryption != "none" {
			return &ProtocolParamError{
				Protocol: "vless",
				Param:    "encryption",
				Message:  "VLESS encryption must be 'none'",
			}
		}
	}

	return nil
}

// TrojanAdapter Trojan 协议适配器
type TrojanAdapter struct{}

func (ta *TrojanAdapter) GetProtocolName() string {
	return "trojan"
}

func (ta *TrojanAdapter) GetRequiredParams() []string {
	return []string{"password"}
}

func (ta *TrojanAdapter) GetOptionalParams() []string {
	return []string{"server_name", "sni", "alpn", "allow_insecure", "network", "network_settings"}
}

func (ta *TrojanAdapter) ConvertParams(params map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 必需参数：password
	password, ok := params["password"].(string)
	if !ok || password == "" {
		return nil, &ProtocolParamError{
			Protocol: "trojan",
			Param:    "password",
			Expected: "non-empty string",
			Actual:   params["password"],
		}
	}
	result["password"] = password

	// TLS 配置（Trojan 默认需要 TLS）
	tlsConfig := map[string]interface{}{
		"enabled": true,
	}

	// 服务器名称
	if serverName, ok := params["server_name"].(string); ok {
		tlsConfig["server_name"] = serverName
	} else if sni, ok := params["sni"].(string); ok {
		tlsConfig["server_name"] = sni
	}

	// 不安全连接
	if insecure, ok := params["allow_insecure"].(bool); ok {
		tlsConfig["insecure"] = insecure
	}

	// ALPN
	if alpn, ok := params["alpn"].([]string); ok {
		tlsConfig["alpn"] = alpn
	}

	result["tls"] = tlsConfig

	// 传输配置
	if network, ok := params["network"].(string); ok {
		result["network"] = network
		if networkSettings, ok := params["network_settings"].(map[string]interface{}); ok {
			result["network_settings"] = networkSettings
		}
	}

	return result, nil
}

func (ta *TrojanAdapter) ValidateParams(params map[string]interface{}) error {
	// 验证 password
	password, ok := params["password"].(string)
	if !ok || password == "" {
		return &ProtocolParamError{
			Protocol: "trojan",
			Param:    "password",
			Message:  "password is required and cannot be empty",
		}
	}

	// 验证 ALPN
	if alpn, ok := params["alpn"].([]interface{}); ok {
		for i, v := range alpn {
			if _, ok := v.(string); !ok {
				return &ProtocolParamError{
					Protocol: "trojan",
					Param:    fmt.Sprintf("alpn[%d]", i),
					Message:  "ALPN values must be strings",
				}
			}
		}
	}

	return nil
}

// ShadowsocksAdapter Shadowsocks 协议适配器
type ShadowsocksAdapter struct{}

func (ssa *ShadowsocksAdapter) GetProtocolName() string {
	return "shadowsocks"
}

func (ssa *ShadowsocksAdapter) GetRequiredParams() []string {
	return []string{"method", "password"}
}

func (ssa *ShadowsocksAdapter) GetOptionalParams() []string {
	return []string{"plugin", "plugin_opts"}
}

func (ssa *ShadowsocksAdapter) ConvertParams(params map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 必需参数：method
	method, ok := params["method"].(string)
	if !ok || method == "" {
		return nil, &ProtocolParamError{
			Protocol: "shadowsocks",
			Param:    "method",
			Expected: "non-empty string",
			Actual:   params["method"],
		}
	}
	result["method"] = method

	// 必需参数：password
	password, ok := params["password"].(string)
	if !ok || password == "" {
		return nil, &ProtocolParamError{
			Protocol: "shadowsocks",
			Param:    "password",
			Expected: "non-empty string",
			Actual:   params["password"],
		}
	}
	result["password"] = password

	// 可选参数：plugin
	if plugin, ok := params["plugin"].(string); ok && plugin != "" {
		result["plugin"] = plugin

		if pluginOpts, ok := params["plugin_opts"].(string); ok {
			result["plugin_opts"] = pluginOpts
		}
	}

	return result, nil
}

func (ssa *ShadowsocksAdapter) ValidateParams(params map[string]interface{}) error {
	// 验证 method
	method, ok := params["method"].(string)
	if !ok || method == "" {
		return &ProtocolParamError{
			Protocol: "shadowsocks",
			Param:    "method",
			Message:  "method is required and cannot be empty",
		}
	}

	// 验证加密方法
	validMethods := []string{
		"aes-128-gcm", "aes-192-gcm", "aes-256-gcm",
		"chacha20-ietf-poly1305", "xchacha20-ietf-poly1305",
		"2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305",
	}
	valid := false
	for _, v := range validMethods {
		if method == v {
			valid = true
			break
		}
	}
	if !valid {
		return &ProtocolParamError{
			Protocol: "shadowsocks",
			Param:    "method",
			Message:  fmt.Sprintf("invalid encryption method, must be one of: %s", strings.Join(validMethods, ", ")),
		}
	}

	// 验证 password
	password, ok := params["password"].(string)
	if !ok || password == "" {
		return &ProtocolParamError{
			Protocol: "shadowsocks",
			Param:    "password",
			Message:  "password is required and cannot be empty",
		}
	}

	return nil
}

// AdapterRegistry 适配器注册表
type AdapterRegistry struct {
	adapters map[string]ProtocolAdapter
}

// NewAdapterRegistry 创建适配器注册表
func NewAdapterRegistry() *AdapterRegistry {
	registry := &AdapterRegistry{
		adapters: make(map[string]ProtocolAdapter),
	}

	// 注册默认适配器
	registry.Register(&VMessAdapter{})
	registry.Register(&VLessAdapter{})
	registry.Register(&TrojanAdapter{})
	registry.Register(&ShadowsocksAdapter{})

	return registry
}

// Register 注册适配器
func (ar *AdapterRegistry) Register(adapter ProtocolAdapter) {
	ar.adapters[adapter.GetProtocolName()] = adapter
}

// GetAdapter 获取适配器
func (ar *AdapterRegistry) GetAdapter(protocol string) (ProtocolAdapter, bool) {
	adapter, exists := ar.adapters[protocol]
	return adapter, exists
}

// GetSupportedProtocols 获取支持的协议列表
func (ar *AdapterRegistry) GetSupportedProtocols() []string {
	protocols := make([]string, 0, len(ar.adapters))
	for protocol := range ar.adapters {
		protocols = append(protocols, protocol)
	}
	return protocols
}

// ValidateProtocolParams 验证协议参数
func (ar *AdapterRegistry) ValidateProtocolParams(protocol string, params map[string]interface{}) error {
	adapter, exists := ar.adapters[protocol]
	if !exists {
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}

	return adapter.ValidateParams(params)
}

// ConvertProtocolParams 转换协议参数
func (ar *AdapterRegistry) ConvertProtocolParams(protocol string, params map[string]interface{}) (map[string]interface{}, error) {
	adapter, exists := ar.adapters[protocol]
	if !exists {
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	return adapter.ConvertParams(params)
}