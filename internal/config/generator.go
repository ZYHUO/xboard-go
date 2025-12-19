package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"dashgo/internal/model"
)

// ConfigGenerator 配置生成器
type ConfigGenerator struct {
	templates map[string]*ConfigTemplate
	validator *ConfigValidator
}

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	Protocol string                 `json:"protocol"`
	Version  string                 `json:"version"`
	Template map[string]interface{} `json:"template"`
	Required []string               `json:"required"` // 必需参数列表
}

// ConfigValidationError 配置验证错误
type ConfigValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Rule    string      `json:"rule"`
	Message string      `json:"message"`
}

func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// ConfigValidator 配置验证器
type ConfigValidator struct {
	rules map[string][]ValidationRule
}

// ValidationRule 验证规则
type ValidationRule struct {
	Name      string                 `json:"name"`
	Validator func(interface{}) bool `json:"-"`
	Message   string                 `json:"message"`
}

// SingBoxConfig sing-box 配置结构
type SingBoxConfig struct {
	Log       *LogConfig    `json:"log,omitempty"`
	DNS       *DNSConfig    `json:"dns,omitempty"`
	Inbounds  []Inbound     `json:"inbounds,omitempty"`
	Outbounds []Outbound    `json:"outbounds"`
	Route     *RouteConfig  `json:"route,omitempty"`
}

type LogConfig struct {
	Level     string `json:"level,omitempty"`
	Timestamp bool   `json:"timestamp,omitempty"`
}

type DNSConfig struct {
	Servers []DNSServer `json:"servers,omitempty"`
	Rules   []DNSRule   `json:"rules,omitempty"`
	Final   string      `json:"final,omitempty"`
}

type DNSServer struct {
	Tag     string `json:"tag,omitempty"`
	Address string `json:"address"`
	Detour  string `json:"detour,omitempty"`
}

type DNSRule struct {
	DomainSuffix []string `json:"domain_suffix,omitempty"`
	Geosite      string   `json:"geosite,omitempty"`
	Server       string   `json:"server"`
}

type Inbound struct {
	Type   string `json:"type"`
	Tag    string `json:"tag,omitempty"`
	Listen string `json:"listen,omitempty"`
	Port   int    `json:"listen_port,omitempty"`
}

type Outbound struct {
	Type       string                 `json:"type"`
	Tag        string                 `json:"tag"`
	Server     string                 `json:"server,omitempty"`
	ServerPort int                    `json:"server_port,omitempty"`
	UUID       string                 `json:"uuid,omitempty"`
	Password   string                 `json:"password,omitempty"`
	Method     string                 `json:"method,omitempty"`
	TLS        *TLSConfig             `json:"tls,omitempty"`
	Transport  map[string]interface{} `json:"transport,omitempty"`
	Flow       string                 `json:"flow,omitempty"`
	Outbounds  []string               `json:"outbounds,omitempty"`
}

type TLSConfig struct {
	Enabled    bool        `json:"enabled"`
	ServerName string      `json:"server_name,omitempty"`
	Insecure   bool        `json:"insecure,omitempty"`
	ALPN       []string    `json:"alpn,omitempty"`
	UTLS       *UTLSConfig `json:"utls,omitempty"`
	Reality    *Reality    `json:"reality,omitempty"`
}

type UTLSConfig struct {
	Enabled     bool   `json:"enabled"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

type Reality struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key,omitempty"`
	ShortID   string `json:"short_id,omitempty"`
}

type RouteConfig struct {
	Rules              []RouteRule `json:"rules,omitempty"`
	Final              string      `json:"final,omitempty"`
	AutoDetectInterface bool       `json:"auto_detect_interface,omitempty"`
}

type RouteRule struct {
	Protocol     []string `json:"protocol,omitempty"`
	DomainSuffix []string `json:"domain_suffix,omitempty"`
	DomainKeyword []string `json:"domain_keyword,omitempty"`
	IPCidr       []string `json:"ip_cidr,omitempty"`
	Geosite      string   `json:"geosite,omitempty"`
	Geoip        string   `json:"geoip,omitempty"`
	IPIsPrivate  bool     `json:"ip_is_private,omitempty"`
	Outbound     string   `json:"outbound"`
}

// NewConfigGenerator 创建配置生成器
func NewConfigGenerator() *ConfigGenerator {
	generator := &ConfigGenerator{
		templates: make(map[string]*ConfigTemplate),
		validator: NewConfigValidator(),
	}

	// 加载默认模板
	generator.loadDefaultTemplates()
	
	return generator
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	validator := &ConfigValidator{
		rules: make(map[string][]ValidationRule),
	}

	// 加载默认验证规则
	validator.loadDefaultRules()
	
	return validator
}

// GenerateConfig 生成 sing-box 配置
func (cg *ConfigGenerator) GenerateConfig(nodes []NodeConfig, options *GenerateOptions) (*SingBoxConfig, error) {
	if options == nil {
		options = &GenerateOptions{}
	}

	config := &SingBoxConfig{
		Log: &LogConfig{
			Level:     "info",
			Timestamp: true,
		},
		DNS:       cg.generateDNSConfig(options),
		Inbounds:  cg.generateInbounds(options),
		Outbounds: []Outbound{},
		Route:     cg.generateRouteConfig(options),
	}

	// 生成节点配置
	for _, node := range nodes {
		outbound, err := cg.generateOutbound(node)
		if err != nil {
			return nil, fmt.Errorf("failed to generate outbound for node %s: %w", node.Name, err)
		}
		config.Outbounds = append(config.Outbounds, *outbound)
	}

	// 添加默认出站
	config.Outbounds = append(config.Outbounds, cg.getDefaultOutbounds()...)

	// 验证配置
	if err := cg.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// NodeConfig 节点配置
type NodeConfig struct {
	Name     string                 `json:"name"`
	Protocol string                 `json:"protocol"`
	Host     string                 `json:"host"`
	Port     int                    `json:"port"`
	Params   map[string]interface{} `json:"params"`
}

// GenerateOptions 生成选项
type GenerateOptions struct {
	LocalPort    int    `json:"local_port"`
	LogLevel     string `json:"log_level"`
	EnableDNS    bool   `json:"enable_dns"`
	EnableRoute  bool   `json:"enable_route"`
}

// ValidateConfig 验证配置
func (cg *ConfigGenerator) ValidateConfig(config *SingBoxConfig) error {
	// 验证 JSON 格式
	if _, err := json.Marshal(config); err != nil {
		return &ConfigValidationError{
			Field:   "config",
			Value:   config,
			Rule:    "json_format",
			Message: fmt.Sprintf("invalid JSON format: %v", err),
		}
	}

	// 验证出站配置
	for i, outbound := range config.Outbounds {
		if err := cg.validateOutbound(&outbound); err != nil {
			return fmt.Errorf("outbound[%d] validation failed: %w", i, err)
		}
	}

	// 验证入站配置
	for i, inbound := range config.Inbounds {
		if err := cg.validateInbound(&inbound); err != nil {
			return fmt.Errorf("inbound[%d] validation failed: %w", i, err)
		}
	}

	return nil
}

// RegisterTemplate 注册配置模板
func (cg *ConfigGenerator) RegisterTemplate(protocol string, template *ConfigTemplate) {
	cg.templates[protocol] = template
}

// GetTemplate 获取配置模板
func (cg *ConfigGenerator) GetTemplate(protocol string) (*ConfigTemplate, bool) {
	template, exists := cg.templates[protocol]
	return template, exists
}

// generateOutbound 生成出站配置
func (cg *ConfigGenerator) generateOutbound(node NodeConfig) (*Outbound, error) {
	template, exists := cg.templates[node.Protocol]
	if !exists {
		return nil, fmt.Errorf("unsupported protocol: %s", node.Protocol)
	}

	// 验证必需参数
	for _, required := range template.Required {
		if _, exists := node.Params[required]; !exists {
			return nil, &ConfigValidationError{
				Field:   required,
				Value:   nil,
				Rule:    "required",
				Message: fmt.Sprintf("required parameter '%s' is missing", required),
			}
		}
	}

	outbound := &Outbound{
		Type:       node.Protocol,
		Tag:        node.Name,
		Server:     node.Host,
		ServerPort: node.Port,
	}

	// 应用协议特定参数
	if err := cg.applyProtocolParams(outbound, node.Protocol, node.Params); err != nil {
		return nil, err
	}

	return outbound, nil
}

// applyProtocolParams 应用协议特定参数
func (cg *ConfigGenerator) applyProtocolParams(outbound *Outbound, protocol string, params map[string]interface{}) error {
	switch protocol {
	case "vmess":
		return cg.applyVMessParams(outbound, params)
	case "vless":
		return cg.applyVLessParams(outbound, params)
	case "trojan":
		return cg.applyTrojanParams(outbound, params)
	case "shadowsocks":
		return cg.applyShadowsocksParams(outbound, params)
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// validateOutbound 验证出站配置
func (cg *ConfigGenerator) validateOutbound(outbound *Outbound) error {
	if outbound.Type == "" {
		return &ConfigValidationError{
			Field:   "type",
			Value:   outbound.Type,
			Rule:    "required",
			Message: "outbound type is required",
		}
	}

	if outbound.Tag == "" {
		return &ConfigValidationError{
			Field:   "tag",
			Value:   outbound.Tag,
			Rule:    "required",
			Message: "outbound tag is required",
		}
	}

	return nil
}

// validateInbound 验证入站配置
func (cg *ConfigGenerator) validateInbound(inbound *Inbound) error {
	if inbound.Type == "" {
		return &ConfigValidationError{
			Field:   "type",
			Value:   inbound.Type,
			Rule:    "required",
			Message: "inbound type is required",
		}
	}

	return nil
}

// ToJSON 转换为 JSON 字符串
func (config *SingBoxConfig) ToJSON() (string, error) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSONCompact 转换为紧凑 JSON 字符串
func (config *SingBoxConfig) ToJSONCompact() (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
// applyVMessParams 应用 VMess 协议参数
func (cg *ConfigGenerator) applyVMessParams(outbound *Outbound, params map[string]interface{}) error {
	if uuid, ok := params["uuid"].(string); ok {
		outbound.UUID = uuid
	}

	if security, ok := params["security"].(string); ok {
		// VMess 的 security 字段在 sing-box 中对应不同的处理
		// 这里需要根据实际情况调整
	}

	if alterID, ok := params["alter_id"].(float64); ok {
		// sing-box 中 VMess 的 alter_id 处理
		_ = alterID // 根据需要处理
	}

	// TLS 配置
	if tls, ok := params["tls"].(map[string]interface{}); ok {
		tlsConfig := &TLSConfig{
			Enabled: true,
		}
		
		if serverName, ok := tls["server_name"].(string); ok {
			tlsConfig.ServerName = serverName
		}
		
		if insecure, ok := tls["allow_insecure"].(bool); ok {
			tlsConfig.Insecure = insecure
		}
		
		outbound.TLS = tlsConfig
	}

	// Transport 配置
	if network, ok := params["network"].(string); ok {
		transport := cg.buildTransport(network, params)
		if transport != nil {
			outbound.Transport = transport
		}
	}

	return nil
}

// applyVLessParams 应用 VLESS 协议参数
func (cg *ConfigGenerator) applyVLessParams(outbound *Outbound, params map[string]interface{}) error {
	if uuid, ok := params["uuid"].(string); ok {
		outbound.UUID = uuid
	}

	if flow, ok := params["flow"].(string); ok {
		outbound.Flow = flow
	}

	// TLS 配置
	if tls, ok := params["tls"].(map[string]interface{}); ok {
		tlsConfig := &TLSConfig{
			Enabled: true,
		}
		
		if serverName, ok := tls["server_name"].(string); ok {
			tlsConfig.ServerName = serverName
		}
		
		// Reality 配置
		if reality, ok := tls["reality"].(map[string]interface{}); ok {
			realityConfig := &Reality{
				Enabled: true,
			}
			
			if publicKey, ok := reality["public_key"].(string); ok {
				realityConfig.PublicKey = publicKey
			}
			
			if shortID, ok := reality["short_id"].(string); ok {
				realityConfig.ShortID = shortID
			}
			
			tlsConfig.Reality = realityConfig
		}
		
		outbound.TLS = tlsConfig
	}

	// Transport 配置
	if network, ok := params["network"].(string); ok {
		transport := cg.buildTransport(network, params)
		if transport != nil {
			outbound.Transport = transport
		}
	}

	return nil
}

// applyTrojanParams 应用 Trojan 协议参数
func (cg *ConfigGenerator) applyTrojanParams(outbound *Outbound, params map[string]interface{}) error {
	if password, ok := params["password"].(string); ok {
		outbound.Password = password
	}

	// TLS 配置（Trojan 默认需要 TLS）
	tlsConfig := &TLSConfig{
		Enabled: true,
	}

	if serverName, ok := params["server_name"].(string); ok {
		tlsConfig.ServerName = serverName
	}

	if insecure, ok := params["allow_insecure"].(bool); ok {
		tlsConfig.Insecure = insecure
	}

	if alpn, ok := params["alpn"].([]string); ok {
		tlsConfig.ALPN = alpn
	}

	outbound.TLS = tlsConfig

	// Transport 配置
	if network, ok := params["network"].(string); ok {
		transport := cg.buildTransport(network, params)
		if transport != nil {
			outbound.Transport = transport
		}
	}

	return nil
}

// applyShadowsocksParams 应用 Shadowsocks 协议参数
func (cg *ConfigGenerator) applyShadowsocksParams(outbound *Outbound, params map[string]interface{}) error {
	if method, ok := params["method"].(string); ok {
		outbound.Method = method
	}

	if password, ok := params["password"].(string); ok {
		outbound.Password = password
	}

	// 插件配置
	if plugin, ok := params["plugin"].(string); ok {
		if outbound.Transport == nil {
			outbound.Transport = make(map[string]interface{})
		}
		outbound.Transport["plugin"] = plugin

		if pluginOpts, ok := params["plugin_opts"].(string); ok {
			outbound.Transport["plugin_opts"] = pluginOpts
		}
	}

	return nil
}

// buildTransport 构建传输配置
func (cg *ConfigGenerator) buildTransport(network string, params map[string]interface{}) map[string]interface{} {
	networkSettings, ok := params["network_settings"].(map[string]interface{})
	if !ok {
		return nil
	}

	switch network {
	case "ws":
		transport := map[string]interface{}{
			"type":                   "ws",
			"max_early_data":         2048,
			"early_data_header_name": "Sec-WebSocket-Protocol",
		}
		
		if path, ok := networkSettings["path"].(string); ok {
			transport["path"] = path
		}
		
		if headers, ok := networkSettings["headers"].(map[string]interface{}); ok {
			if host, ok := headers["Host"].(string); ok {
				transport["headers"] = map[string]string{"Host": host}
			}
		}
		
		return transport

	case "grpc":
		transport := map[string]interface{}{
			"type": "grpc",
		}
		
		if serviceName, ok := networkSettings["serviceName"].(string); ok {
			transport["service_name"] = serviceName
		}
		
		return transport

	case "tcp":
		if header, ok := networkSettings["header"].(map[string]interface{}); ok {
			if headerType, ok := header["type"].(string); ok && headerType == "http" {
				return map[string]interface{}{
					"type": "http",
					"path": "/",
				}
			}
		}
	}

	return nil
}

// generateDNSConfig 生成 DNS 配置
func (cg *ConfigGenerator) generateDNSConfig(options *GenerateOptions) *DNSConfig {
	if !options.EnableDNS {
		return nil
	}

	return &DNSConfig{
		Servers: []DNSServer{
			{Tag: "google", Address: "https://dns.google/dns-query", Detour: "🚀 节点选择"},
			{Tag: "cloudflare", Address: "https://cloudflare-dns.com/dns-query", Detour: "🚀 节点选择"},
			{Tag: "alidns", Address: "https://dns.alidns.com/dns-query", Detour: "direct"},
			{Tag: "local", Address: "223.5.5.5", Detour: "direct"},
		},
		Rules: []DNSRule{
			{DomainSuffix: []string{".cn"}, Server: "local"},
			{Geosite: "cn", Server: "local"},
		},
		Final: "google",
	}
}

// generateInbounds 生成入站配置
func (cg *ConfigGenerator) generateInbounds(options *GenerateOptions) []Inbound {
	localPort := 7890
	if options.LocalPort > 0 {
		localPort = options.LocalPort
	}

	return []Inbound{
		{
			Type:   "mixed",
			Tag:    "mixed-in",
			Listen: "127.0.0.1",
			Port:   localPort,
		},
	}
}

// generateRouteConfig 生成路由配置
func (cg *ConfigGenerator) generateRouteConfig(options *GenerateOptions) *RouteConfig {
	if !options.EnableRoute {
		return nil
	}

	return &RouteConfig{
		Rules: []RouteRule{
			{Protocol: []string{"dns"}, Outbound: "dns-out"},
			{IPIsPrivate: true, Outbound: "direct"},
			// OpenAI
			{DomainSuffix: []string{"openai.com", "ai.com", "anthropic.com", "claude.ai"}, Outbound: "🤖 OpenAI"},
			{DomainKeyword: []string{"openai"}, Outbound: "🤖 OpenAI"},
			// Telegram
			{DomainSuffix: []string{"telegram.org", "t.me", "tg.dev"}, Outbound: "📲 电报消息"},
			{IPCidr: []string{"91.108.0.0/16", "109.239.140.0/24", "149.154.160.0/20"}, Outbound: "📲 电报消息"},
			// YouTube
			{DomainSuffix: []string{"youtube.com", "googlevideo.com", "ytimg.com", "yt.be"}, Outbound: "📹 YouTube"},
			// Netflix
			{DomainSuffix: []string{"netflix.com", "netflix.net", "nflximg.com", "nflximg.net", "nflxvideo.net"}, Outbound: "🎬 Netflix"},
			// Apple
			{DomainSuffix: []string{"apple.com", "icloud.com", "icloud-content.com", "mzstatic.com"}, Outbound: "🍎 苹果服务"},
			// 国内直连
			{Geosite: "cn", Outbound: "direct"},
			{Geoip: "cn", Outbound: "direct"},
		},
		Final:               "🐟 漏网之鱼",
		AutoDetectInterface: true,
	}
}

// getDefaultOutbounds 获取默认出站配置
func (cg *ConfigGenerator) getDefaultOutbounds() []Outbound {
	return []Outbound{
		{Type: "selector", Tag: "🚀 节点选择", Outbounds: []string{"♻️ 自动选择", "🔯 故障转移", "direct"}},
		{Type: "urltest", Tag: "♻️ 自动选择", Outbounds: []string{}},
		{Type: "urltest", Tag: "🔯 故障转移", Outbounds: []string{}},
		{Type: "selector", Tag: "📲 电报消息", Outbounds: []string{"🚀 节点选择", "♻️ 自动选择", "direct"}},
		{Type: "selector", Tag: "🤖 OpenAI", Outbounds: []string{"🚀 节点选择", "♻️ 自动选择"}},
		{Type: "selector", Tag: "📹 YouTube", Outbounds: []string{"🚀 节点选择", "♻️ 自动选择", "direct"}},
		{Type: "selector", Tag: "🎬 Netflix", Outbounds: []string{"🚀 节点选择", "♻️ 自动选择", "direct"}},
		{Type: "selector", Tag: "🍎 苹果服务", Outbounds: []string{"direct", "🚀 节点选择"}},
		{Type: "selector", Tag: "🐟 漏网之鱼", Outbounds: []string{"🚀 节点选择", "♻️ 自动选择", "direct"}},
		{Type: "direct", Tag: "direct"},
		{Type: "block", Tag: "block"},
		{Type: "dns", Tag: "dns-out"},
	}
}

// loadDefaultTemplates 加载默认模板
func (cg *ConfigGenerator) loadDefaultTemplates() {
	// VMess 模板
	cg.templates["vmess"] = &ConfigTemplate{
		Protocol: "vmess",
		Version:  "1.0",
		Required: []string{"uuid"},
		Template: map[string]interface{}{
			"type":     "vmess",
			"security": "auto",
			"alter_id": 0,
		},
	}

	// VLESS 模板
	cg.templates["vless"] = &ConfigTemplate{
		Protocol: "vless",
		Version:  "1.0",
		Required: []string{"uuid"},
		Template: map[string]interface{}{
			"type":            "vless",
			"packet_encoding": "xudp",
		},
	}

	// Trojan 模板
	cg.templates["trojan"] = &ConfigTemplate{
		Protocol: "trojan",
		Version:  "1.0",
		Required: []string{"password"},
		Template: map[string]interface{}{
			"type": "trojan",
			"tls": map[string]interface{}{
				"enabled": true,
			},
		},
	}

	// Shadowsocks 模板
	cg.templates["shadowsocks"] = &ConfigTemplate{
		Protocol: "shadowsocks",
		Version:  "1.0",
		Required: []string{"method", "password"},
		Template: map[string]interface{}{
			"type": "shadowsocks",
		},
	}
}

// loadDefaultRules 加载默认验证规则
func (cv *ConfigValidator) loadDefaultRules() {
	// 端口验证规则
	cv.rules["port"] = []ValidationRule{
		{
			Name: "range",
			Validator: func(value interface{}) bool {
				if port, ok := value.(int); ok {
					return port > 0 && port <= 65535
				}
				return false
			},
			Message: "port must be between 1 and 65535",
		},
	}

	// UUID 验证规则
	cv.rules["uuid"] = []ValidationRule{
		{
			Name: "format",
			Validator: func(value interface{}) bool {
				if uuid, ok := value.(string); ok {
					return len(uuid) == 36 && strings.Count(uuid, "-") == 4
				}
				return false
			},
			Message: "invalid UUID format",
		},
	}

	// 主机名验证规则
	cv.rules["hostname"] = []ValidationRule{
		{
			Name: "not_empty",
			Validator: func(value interface{}) bool {
				if host, ok := value.(string); ok {
					return strings.TrimSpace(host) != ""
				}
				return false
			},
			Message: "hostname cannot be empty",
		},
	}
}