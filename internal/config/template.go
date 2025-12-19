package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateManager 模板管理器
type TemplateManager struct {
	templates map[string]*ConfigTemplate
	basePath  string
}

// TemplateVersion 模板版本信息
type TemplateVersion struct {
	Version     string `json:"version"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	Author      string `json:"author"`
}

// EnhancedConfigTemplate 增强的配置模板
type EnhancedConfigTemplate struct {
	*ConfigTemplate
	Version     *TemplateVersion           `json:"version"`
	Variables   map[string]TemplateVar     `json:"variables"`
	Conditions  map[string]TemplateCondition `json:"conditions"`
	Inheritance *TemplateInheritance       `json:"inheritance,omitempty"`
}

// TemplateVar 模板变量
type TemplateVar struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`         // "string", "int", "bool", "array", "object"
	Default      interface{} `json:"default"`
	Required     bool        `json:"required"`
	Description  string      `json:"description"`
	Validation   *VarValidation `json:"validation,omitempty"`
}

// VarValidation 变量验证规则
type VarValidation struct {
	Min    *int     `json:"min,omitempty"`
	Max    *int     `json:"max,omitempty"`
	Enum   []string `json:"enum,omitempty"`
	Regex  string   `json:"regex,omitempty"`
}

// TemplateCondition 模板条件
type TemplateCondition struct {
	Name      string                 `json:"name"`
	Condition string                 `json:"condition"` // 条件表达式
	Template  map[string]interface{} `json:"template"`  // 条件满足时的模板片段
}

// TemplateInheritance 模板继承
type TemplateInheritance struct {
	Parent   string   `json:"parent"`
	Override []string `json:"override"` // 要覆盖的字段
	Extend   []string `json:"extend"`   // 要扩展的字段
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager(basePath string) *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*ConfigTemplate),
		basePath:  basePath,
	}

	// 加载默认模板
	tm.loadDefaultTemplates()

	return tm
}

// LoadTemplate 从文件加载模板
func (tm *TemplateManager) LoadTemplate(filename string) error {
	fullPath := filepath.Join(tm.basePath, filename)
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", filename, err)
	}

	var template EnhancedConfigTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return fmt.Errorf("failed to parse template file %s: %w", filename, err)
	}

	// 处理模板继承
	if template.Inheritance != nil {
		if err := tm.processInheritance(&template); err != nil {
			return fmt.Errorf("failed to process template inheritance: %w", err)
		}
	}

	tm.templates[template.Protocol] = template.ConfigTemplate
	return nil
}

// SaveTemplate 保存模板到文件
func (tm *TemplateManager) SaveTemplate(protocol string, filename string) error {
	template, exists := tm.templates[protocol]
	if !exists {
		return fmt.Errorf("template for protocol %s not found", protocol)
	}

	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	fullPath := filepath.Join(tm.basePath, filename)
	if err := ioutil.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file %s: %w", filename, err)
	}

	return nil
}

// RegisterTemplate 注册模板
func (tm *TemplateManager) RegisterTemplate(template *ConfigTemplate) {
	tm.templates[template.Protocol] = template
}

// GetTemplate 获取模板
func (tm *TemplateManager) GetTemplate(protocol string) (*ConfigTemplate, bool) {
	template, exists := tm.templates[protocol]
	return template, exists
}

// ListTemplates 列出所有模板
func (tm *TemplateManager) ListTemplates() map[string]*ConfigTemplate {
	result := make(map[string]*ConfigTemplate)
	for k, v := range tm.templates {
		result[k] = v
	}
	return result
}

// RenderTemplate 渲染模板
func (tm *TemplateManager) RenderTemplate(protocol string, params map[string]interface{}) (map[string]interface{}, error) {
	template, exists := tm.templates[protocol]
	if !exists {
		return nil, fmt.Errorf("template for protocol %s not found", protocol)
	}

	// 深拷贝模板
	templateData, err := json.Marshal(template.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(templateData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	// 参数替换
	if err := tm.replaceTemplateParams(result, params); err != nil {
		return nil, fmt.Errorf("failed to replace template parameters: %w", err)
	}

	return result, nil
}

// ValidateTemplateParams 验证模板参数
func (tm *TemplateManager) ValidateTemplateParams(protocol string, params map[string]interface{}) error {
	template, exists := tm.templates[protocol]
	if !exists {
		return fmt.Errorf("template for protocol %s not found", protocol)
	}

	// 检查必需参数
	for _, required := range template.Required {
		if _, exists := params[required]; !exists {
			return fmt.Errorf("required parameter '%s' is missing", required)
		}
	}

	return nil
}

// UpdateTemplate 更新模板
func (tm *TemplateManager) UpdateTemplate(protocol string, updates map[string]interface{}) error {
	template, exists := tm.templates[protocol]
	if !exists {
		return fmt.Errorf("template for protocol %s not found", protocol)
	}

	// 更新模板字段
	for key, value := range updates {
		switch key {
		case "version":
			if version, ok := value.(string); ok {
				template.Version = version
			}
		case "required":
			if required, ok := value.([]string); ok {
				template.Required = required
			}
		case "template":
			if templateData, ok := value.(map[string]interface{}); ok {
				template.Template = templateData
			}
		}
	}

	return nil
}

// DeleteTemplate 删除模板
func (tm *TemplateManager) DeleteTemplate(protocol string) error {
	if _, exists := tm.templates[protocol]; !exists {
		return fmt.Errorf("template for protocol %s not found", protocol)
	}

	delete(tm.templates, protocol)
	return nil
}

// CloneTemplate 克隆模板
func (tm *TemplateManager) CloneTemplate(sourceProtocol, targetProtocol string) error {
	source, exists := tm.templates[sourceProtocol]
	if !exists {
		return fmt.Errorf("source template for protocol %s not found", sourceProtocol)
	}

	// 深拷贝模板
	data, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("failed to marshal source template: %w", err)
	}

	var cloned ConfigTemplate
	if err := json.Unmarshal(data, &cloned); err != nil {
		return fmt.Errorf("failed to unmarshal cloned template: %w", err)
	}

	cloned.Protocol = targetProtocol
	tm.templates[targetProtocol] = &cloned

	return nil
}

// replaceTemplateParams 替换模板参数
func (tm *TemplateManager) replaceTemplateParams(template map[string]interface{}, params map[string]interface{}) error {
	for key, value := range template {
		switch v := value.(type) {
		case string:
			// 处理字符串模板变量
			if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
				varName := strings.Trim(v[2:len(v)-2], " ")
				if paramValue, exists := params[varName]; exists {
					template[key] = paramValue
				}
			}
		case map[string]interface{}:
			// 递归处理嵌套对象
			if err := tm.replaceTemplateParams(v, params); err != nil {
				return err
			}
		case []interface{}:
			// 处理数组
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if err := tm.replaceTemplateParams(itemMap, params); err != nil {
						return err
					}
				} else if itemStr, ok := item.(string); ok {
					if strings.HasPrefix(itemStr, "{{") && strings.HasSuffix(itemStr, "}}") {
						varName := strings.Trim(itemStr[2:len(itemStr)-2], " ")
						if paramValue, exists := params[varName]; exists {
							v[i] = paramValue
						}
					}
				}
			}
		}
	}
	return nil
}

// processInheritance 处理模板继承
func (tm *TemplateManager) processInheritance(template *EnhancedConfigTemplate) error {
	if template.Inheritance == nil {
		return nil
	}

	parent, exists := tm.templates[template.Inheritance.Parent]
	if !exists {
		return fmt.Errorf("parent template %s not found", template.Inheritance.Parent)
	}

	// 继承父模板
	if template.ConfigTemplate == nil {
		template.ConfigTemplate = &ConfigTemplate{}
	}

	// 继承基本属性
	if template.Version == "" {
		template.Version = parent.Version
	}

	// 合并必需参数
	requiredMap := make(map[string]bool)
	for _, req := range parent.Required {
		requiredMap[req] = true
	}
	for _, req := range template.Required {
		requiredMap[req] = true
	}

	template.Required = make([]string, 0, len(requiredMap))
	for req := range requiredMap {
		template.Required = append(template.Required, req)
	}

	// 合并模板内容
	if template.Template == nil {
		template.Template = make(map[string]interface{})
	}

	// 复制父模板内容
	for key, value := range parent.Template {
		if _, exists := template.Template[key]; !exists {
			template.Template[key] = value
		}
	}

	return nil
}

// loadDefaultTemplates 加载默认模板
func (tm *TemplateManager) loadDefaultTemplates() {
	// VMess 模板
	tm.templates["vmess"] = &ConfigTemplate{
		Protocol: "vmess",
		Version:  "1.0",
		Required: []string{"uuid"},
		Template: map[string]interface{}{
			"type":     "vmess",
			"uuid":     "{{uuid}}",
			"security": "{{security|auto}}",
			"alter_id": "{{alter_id|0}}",
		},
	}

	// VLESS 模板
	tm.templates["vless"] = &ConfigTemplate{
		Protocol: "vless",
		Version:  "1.0",
		Required: []string{"uuid"},
		Template: map[string]interface{}{
			"type":            "vless",
			"uuid":            "{{uuid}}",
			"packet_encoding": "xudp",
			"flow":            "{{flow}}",
		},
	}

	// Trojan 模板
	tm.templates["trojan"] = &ConfigTemplate{
		Protocol: "trojan",
		Version:  "1.0",
		Required: []string{"password"},
		Template: map[string]interface{}{
			"type":     "trojan",
			"password": "{{password}}",
			"tls": map[string]interface{}{
				"enabled": true,
			},
		},
	}

	// Shadowsocks 模板
	tm.templates["shadowsocks"] = &ConfigTemplate{
		Protocol: "shadowsocks",
		Version:  "1.0",
		Required: []string{"method", "password"},
		Template: map[string]interface{}{
			"type":     "shadowsocks",
			"method":   "{{method}}",
			"password": "{{password}}",
		},
	}
}

// GetTemplateVersions 获取模板版本信息
func (tm *TemplateManager) GetTemplateVersions() map[string]string {
	versions := make(map[string]string)
	for protocol, template := range tm.templates {
		versions[protocol] = template.Version
	}
	return versions
}

// CompareTemplates 比较两个模板的差异
func (tm *TemplateManager) CompareTemplates(protocol1, protocol2 string) (map[string]interface{}, error) {
	template1, exists1 := tm.templates[protocol1]
	template2, exists2 := tm.templates[protocol2]

	if !exists1 {
		return nil, fmt.Errorf("template for protocol %s not found", protocol1)
	}
	if !exists2 {
		return nil, fmt.Errorf("template for protocol %s not found", protocol2)
	}

	diff := make(map[string]interface{})
	
	// 比较版本
	if template1.Version != template2.Version {
		diff["version"] = map[string]string{
			protocol1: template1.Version,
			protocol2: template2.Version,
		}
	}

	// 比较必需参数
	req1Map := make(map[string]bool)
	req2Map := make(map[string]bool)
	
	for _, req := range template1.Required {
		req1Map[req] = true
	}
	for _, req := range template2.Required {
		req2Map[req] = true
	}

	var onlyIn1, onlyIn2 []string
	for req := range req1Map {
		if !req2Map[req] {
			onlyIn1 = append(onlyIn1, req)
		}
	}
	for req := range req2Map {
		if !req1Map[req] {
			onlyIn2 = append(onlyIn2, req)
		}
	}

	if len(onlyIn1) > 0 || len(onlyIn2) > 0 {
		diff["required"] = map[string][]string{
			"only_in_" + protocol1: onlyIn1,
			"only_in_" + protocol2: onlyIn2,
		}
	}

	return diff, nil
}