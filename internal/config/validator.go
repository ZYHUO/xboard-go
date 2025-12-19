package config

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// ValidationEngine 配置验证引擎
type ValidationEngine struct {
	rules map[string][]ValidationRule
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid   bool                     `json:"valid"`
	Errors  []ConfigValidationError  `json:"errors,omitempty"`
	Warnings []ConfigValidationError `json:"warnings,omitempty"`
	Suggestions []string             `json:"suggestions,omitempty"`
}

// ValidationSeverity 验证严重程度
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
	SeverityInfo    ValidationSeverity = "info"
)

// EnhancedValidationError 增强的验证错误
type EnhancedValidationError struct {
	ConfigValidationError
	Severity    ValidationSeverity `json:"severity"`
	Code        string            `json:"code"`
	Suggestion  string            `json:"suggestion,omitempty"`
	Path        string            `json:"path"`
}

// NewValidationEngine 创建验证引擎
func NewValidationEngine() *ValidationEngine {
	engine := &ValidationEngine{
		rules: make(map[string][]ValidationRule),
	}
	
	engine.loadDefaultValidationRules()
	return engine
}

// ValidateConfig 验证完整配置
func (ve *ValidationEngine) ValidateConfig(config *SingBoxConfig) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		Errors:      []ConfigValidationError{},
		Warnings:    []ConfigValidationError{},
		Suggestions: []string{},
	}

	// JSON 格式验证
	if err := ve.validateJSONFormat(config); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, *err)
	}

	// 配置完整性验证
	if errors := ve.validateConfigIntegrity(config); len(errors) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors, errors...)
	}

	// 出站配置验证
	for i, outbound := range config.Outbounds {
		if errors := ve.validateOutbound(&outbound, fmt.Sprintf("outbounds[%d]", i)); len(errors) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors, errors...)
		}
	}

	// 入站配置验证
	for i, inbound := range config.Inbounds {
		if errors := ve.validateInbound(&inbound, fmt.Sprintf("inbounds[%d]", i)); len(errors) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors, errors...)
		}
	}

	// DNS 配置验证
	if config.DNS != nil {
		if errors := ve.validateDNSConfig(config.DNS); len(errors) > 0 {
			result.Warnings = append(result.Warnings, errors...)
		}
	}

	// 路由配置验证
	if config.Route != nil {
		if errors := ve.validateRouteConfig(config.Route); len(errors) > 0 {
			result.Warnings = append(result.Warnings, errors...)
		}
	}

	// 生成建议
	result.Suggestions = ve.generateSuggestions(config, result)

	return result
}

// validateJSONFormat 验证 JSON 格式
func (ve *ValidationEngine) validateJSONFormat(config *SingBoxConfig) *ConfigValidationError {
	_, err := json.Marshal(config)
	if err != nil {
		return &ConfigValidationError{
			Field:   "config",
			Value:   config,
			Rule:    "json_format",
			Message: fmt.Sprintf("invalid JSON format: %v", err),
		}
	}
	return nil
}

// validateConfigIntegrity 验证配置完整性
func (ve *ValidationEngine) validateConfigIntegrity(config *SingBoxConfig) []ConfigValidationError {
	var errors []ConfigValidationError

	// 检查必需的出站配置
	if len(config.Outbounds) == 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "outbounds",
			Value:   config.Outbounds,
			Rule:    "required",
			Message: "at least one outbound is required",
		})
	}

	// 检查出站标签唯一性
	tags := make(map[string]bool)
	for i, outbound := range config.Outbounds {
		if outbound.Tag == "" {
			errors = append(errors, ConfigValidationError{
				Field:   fmt.Sprintf("outbounds[%d].tag", i),
				Value:   outbound.Tag,
				Rule:    "required",
				Message: "outbound tag is required",
			})
			continue
		}

		if tags[outbound.Tag] {
			errors = append(errors, ConfigValidationError{
				Field:   fmt.Sprintf("outbounds[%d].tag", i),
				Value:   outbound.Tag,
				Rule:    "unique",
				Message: fmt.Sprintf("duplicate outbound tag: %s", outbound.Tag),
			})
		}
		tags[outbound.Tag] = true
	}

	// 检查入站标签唯一性
	inboundTags := make(map[string]bool)
	for i, inbound := range config.Inbounds {
		if inbound.Tag != "" {
			if inboundTags[inbound.Tag] {
				errors = append(errors, ConfigValidationError{
					Field:   fmt.Sprintf("inbounds[%d].tag", i),
					Value:   inbound.Tag,
					Rule:    "unique",
					Message: fmt.Sprintf("duplicate inbound tag: %s", inbound.Tag),
				})
			}
			inboundTags[inbound.Tag] = true
		}
	}

	return errors
}

// validateOutbound 验证出站配置
func (ve *ValidationEngine) validateOutbound(outbound *Outbound, path string) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证类型
	if outbound.Type == "" {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".type",
			Value:   outbound.Type,
			Rule:    "required",
			Message: "outbound type is required",
		})
	} else {
		validTypes := []string{"direct", "block", "dns", "selector", "urltest", "vmess", "vless", "trojan", "shadowsocks", "hysteria", "hysteria2", "tuic", "socks", "http"}
		if !contains(validTypes, outbound.Type) {
			errors = append(errors, ConfigValidationError{
				Field:   path + ".type",
				Value:   outbound.Type,
				Rule:    "enum",
				Message: fmt.Sprintf("invalid outbound type: %s", outbound.Type),
			})
		}
	}

	// 验证标签
	if outbound.Tag == "" {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".tag",
			Value:   outbound.Tag,
			Rule:    "required",
			Message: "outbound tag is required",
		})
	}

	// 验证服务器地址（对于代理类型）
	proxyTypes := []string{"vmess", "vless", "trojan", "shadowsocks", "hysteria", "hysteria2", "tuic", "socks", "http"}
	if contains(proxyTypes, outbound.Type) {
		if outbound.Server == "" {
			errors = append(errors, ConfigValidationError{
				Field:   path + ".server",
				Value:   outbound.Server,
				Rule:    "required",
				Message: "server address is required for proxy outbound",
			})
		} else if !ve.isValidHost(outbound.Server) {
			errors = append(errors, ConfigValidationError{
				Field:   path + ".server",
				Value:   outbound.Server,
				Rule:    "format",
				Message: "invalid server address format",
			})
		}

		// 验证端口
		if outbound.ServerPort <= 0 || outbound.ServerPort > 65535 {
			errors = append(errors, ConfigValidationError{
				Field:   path + ".server_port",
				Value:   outbound.ServerPort,
				Rule:    "range",
				Message: "server port must be between 1 and 65535",
			})
		}
	}

	// 协议特定验证
	switch outbound.Type {
	case "vmess":
		errors = append(errors, ve.validateVMessOutbound(outbound, path)...)
	case "vless":
		errors = append(errors, ve.validateVLessOutbound(outbound, path)...)
	case "trojan":
		errors = append(errors, ve.validateTrojanOutbound(outbound, path)...)
	case "shadowsocks":
		errors = append(errors, ve.validateShadowsocksOutbound(outbound, path)...)
	}

	return errors
}

// validateInbound 验证入站配置
func (ve *ValidationEngine) validateInbound(inbound *Inbound, path string) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证类型
	if inbound.Type == "" {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".type",
			Value:   inbound.Type,
			Rule:    "required",
			Message: "inbound type is required",
		})
	} else {
		validTypes := []string{"mixed", "socks", "http", "tun", "redirect", "tproxy"}
		if !contains(validTypes, inbound.Type) {
			errors = append(errors, ConfigValidationError{
				Field:   path + ".type",
				Value:   inbound.Type,
				Rule:    "enum",
				Message: fmt.Sprintf("invalid inbound type: %s", inbound.Type),
			})
		}
	}

	// 验证监听地址
	if inbound.Listen != "" && !ve.isValidIP(inbound.Listen) {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".listen",
			Value:   inbound.Listen,
			Rule:    "format",
			Message: "invalid listen address format",
		})
	}

	// 验证端口
	if inbound.Port <= 0 || inbound.Port > 65535 {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".listen_port",
			Value:   inbound.Port,
			Rule:    "range",
			Message: "listen port must be between 1 and 65535",
		})
	}

	return errors
}

// validateVMessOutbound 验证 VMess 出站配置
func (ve *ValidationEngine) validateVMessOutbound(outbound *Outbound, path string) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证 UUID
	if outbound.UUID == "" {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".uuid",
			Value:   outbound.UUID,
			Rule:    "required",
			Message: "VMess UUID is required",
		})
	} else if !ve.isValidUUID(outbound.UUID) {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".uuid",
			Value:   outbound.UUID,
			Rule:    "format",
			Message: "invalid VMess UUID format",
		})
	}

	return errors
}

// validateVLessOutbound 验证 VLESS 出站配置
func (ve *ValidationEngine) validateVLessOutbound(outbound *Outbound, path string) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证 UUID
	if outbound.UUID == "" {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".uuid",
			Value:   outbound.UUID,
			Rule:    "required",
			Message: "VLESS UUID is required",
		})
	} else if !ve.isValidUUID(outbound.UUID) {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".uuid",
			Value:   outbound.UUID,
			Rule:    "format",
			Message: "invalid VLESS UUID format",
		})
	}

	// 验证 flow
	if outbound.Flow != "" {
		validFlows := []string{"xtls-rprx-vision", "xtls-rprx-vision-udp443"}
		if !contains(validFlows, outbound.Flow) {
			errors = append(errors, ConfigValidationError{
				Field:   path + ".flow",
				Value:   outbound.Flow,
				Rule:    "enum",
				Message: fmt.Sprintf("invalid VLESS flow: %s", outbound.Flow),
			})
		}
	}

	return errors
}

// validateTrojanOutbound 验证 Trojan 出站配置
func (ve *ValidationEngine) validateTrojanOutbound(outbound *Outbound, path string) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证密码
	if outbound.Password == "" {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".password",
			Value:   outbound.Password,
			Rule:    "required",
			Message: "Trojan password is required",
		})
	}

	// Trojan 必须使用 TLS
	if outbound.TLS == nil || !outbound.TLS.Enabled {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".tls",
			Value:   outbound.TLS,
			Rule:    "required",
			Message: "Trojan requires TLS to be enabled",
		})
	}

	return errors
}

// validateShadowsocksOutbound 验证 Shadowsocks 出站配置
func (ve *ValidationEngine) validateShadowsocksOutbound(outbound *Outbound, path string) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证加密方法
	if outbound.Method == "" {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".method",
			Value:   outbound.Method,
			Rule:    "required",
			Message: "Shadowsocks method is required",
		})
	} else {
		validMethods := []string{
			"aes-128-gcm", "aes-192-gcm", "aes-256-gcm",
			"chacha20-ietf-poly1305", "xchacha20-ietf-poly1305",
			"2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305",
		}
		if !contains(validMethods, outbound.Method) {
			errors = append(errors, ConfigValidationError{
				Field:   path + ".method",
				Value:   outbound.Method,
				Rule:    "enum",
				Message: fmt.Sprintf("invalid Shadowsocks method: %s", outbound.Method),
			})
		}
	}

	// 验证密码
	if outbound.Password == "" {
		errors = append(errors, ConfigValidationError{
			Field:   path + ".password",
			Value:   outbound.Password,
			Rule:    "required",
			Message: "Shadowsocks password is required",
		})
	}

	return errors
}

// validateDNSConfig 验证 DNS 配置
func (ve *ValidationEngine) validateDNSConfig(dns *DNSConfig) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证 DNS 服务器
	for i, server := range dns.Servers {
		if server.Address == "" {
			errors = append(errors, ConfigValidationError{
				Field:   fmt.Sprintf("dns.servers[%d].address", i),
				Value:   server.Address,
				Rule:    "required",
				Message: "DNS server address is required",
			})
		}
	}

	return errors
}

// validateRouteConfig 验证路由配置
func (ve *ValidationEngine) validateRouteConfig(route *RouteConfig) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证路由规则
	for i, rule := range route.Rules {
		if rule.Outbound == "" {
			errors = append(errors, ConfigValidationError{
				Field:   fmt.Sprintf("route.rules[%d].outbound", i),
				Value:   rule.Outbound,
				Rule:    "required",
				Message: "route rule outbound is required",
			})
		}
	}

	return errors
}

// generateSuggestions 生成配置建议
func (ve *ValidationEngine) generateSuggestions(config *SingBoxConfig, result *ValidationResult) []string {
	var suggestions []string

	// 如果没有入站配置，建议添加
	if len(config.Inbounds) == 0 {
		suggestions = append(suggestions, "Consider adding an inbound configuration for local proxy")
	}

	// 如果没有 DNS 配置，建议添加
	if config.DNS == nil {
		suggestions = append(suggestions, "Consider adding DNS configuration for better domain resolution")
	}

	// 如果没有路由配置，建议添加
	if config.Route == nil {
		suggestions = append(suggestions, "Consider adding route configuration for traffic control")
	}

	// 检查是否有代理出站但没有选择器
	hasProxy := false
	hasSelector := false
	for _, outbound := range config.Outbounds {
		if contains([]string{"vmess", "vless", "trojan", "shadowsocks"}, outbound.Type) {
			hasProxy = true
		}
		if outbound.Type == "selector" {
			hasSelector = true
		}
	}

	if hasProxy && !hasSelector {
		suggestions = append(suggestions, "Consider adding a selector outbound for better proxy management")
	}

	return suggestions
}

// 辅助验证函数
func (ve *ValidationEngine) isValidHost(host string) bool {
	// 检查是否为有效的 IP 地址
	if net.ParseIP(host) != nil {
		return true
	}

	// 检查是否为有效的域名
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	return domainRegex.MatchString(host)
}

func (ve *ValidationEngine) isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func (ve *ValidationEngine) isValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

func (ve *ValidationEngine) isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p > 0 && p <= 65535
}

// loadDefaultValidationRules 加载默认验证规则
func (ve *ValidationEngine) loadDefaultValidationRules() {
	// 端口验证规则
	ve.rules["port"] = []ValidationRule{
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
	ve.rules["uuid"] = []ValidationRule{
		{
			Name: "format",
			Validator: func(value interface{}) bool {
				if uuid, ok := value.(string); ok {
					return ve.isValidUUID(uuid)
				}
				return false
			},
			Message: "invalid UUID format",
		},
	}

	// 主机名验证规则
	ve.rules["hostname"] = []ValidationRule{
		{
			Name: "format",
			Validator: func(value interface{}) bool {
				if host, ok := value.(string); ok {
					return ve.isValidHost(host)
				}
				return false
			},
			Message: "invalid hostname format",
		},
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}