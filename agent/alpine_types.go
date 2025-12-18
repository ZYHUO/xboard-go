package main

import (
	"os"
	"time"
)

// SystemInfo 系统信息结构
type SystemInfo struct {
	// 基本系统信息
	OS           string    `json:"os"`
	Kernel       string    `json:"kernel"`
	Architecture string    `json:"architecture"`
	AlpineVersion string   `json:"alpine_version"`
	
	// 资源信息
	MemoryTotal  int64     `json:"memory_total"`
	MemoryFree   int64     `json:"memory_free"`
	DiskFree     int64     `json:"disk_free"`
	CPUCores     int       `json:"cpu_cores"`
	
	// 环境信息
	IsContainer  bool      `json:"is_container"`
	HasSystemd   bool      `json:"has_systemd"`
	HasOpenRC    bool      `json:"has_openrc"`
	
	// 网络信息
	NetworkInterfaces []NetworkInterface `json:"network_interfaces"`
	DNSServers       []string           `json:"dns_servers"`
	
	// 时间戳
	Timestamp    time.Time `json:"timestamp"`
}

// NetworkInterface 网络接口信息
type NetworkInterface struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	IsUp      bool     `json:"is_up"`
}

// DependencyInfo 依赖项信息
type DependencyInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Version   string `json:"version"`
	Available bool   `json:"available"`
	Error     error  `json:"error,omitempty"`
}

// NetworkTest 网络测试结果
type NetworkTest struct {
	Target    string        `json:"target"`
	Type      string        `json:"type"` // "ping", "http", "https", "dns"
	Success   bool          `json:"success"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
}

// FileCheck 文件检查结果
type FileCheck struct {
	Path        string      `json:"path"`
	Exists      bool        `json:"exists"`
	Permissions os.FileMode `json:"permissions"`
	Size        int64       `json:"size"`
	Error       string      `json:"error,omitempty"`
}

// ErrorCategory 错误类别
type ErrorCategory string

const (
	// ErrorCategorySystem 系统环境错误
	ErrorCategorySystem ErrorCategory = "system"
	// ErrorCategoryDependency 依赖项错误
	ErrorCategoryDependency ErrorCategory = "dependency"
	// ErrorCategoryResource 资源错误
	ErrorCategoryResource ErrorCategory = "resource"
	// ErrorCategoryConfiguration 配置错误
	ErrorCategoryConfiguration ErrorCategory = "configuration"
	// ErrorCategoryPermission 权限错误
	ErrorCategoryPermission ErrorCategory = "permission"
	// ErrorCategoryCompatibility 兼容性错误
	ErrorCategoryCompatibility ErrorCategory = "compatibility"
)

// AlpineError Alpine 特定错误类型
type AlpineError struct {
	Category    ErrorCategory `json:"category"`
	Message     string        `json:"message"`
	Suggestion  string        `json:"suggestion"`
	SystemInfo  *SystemInfo   `json:"system_info,omitempty"`
	StackTrace  string        `json:"stack_trace,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// Error 实现 error 接口
func (e *AlpineError) Error() string {
	return e.Message
}

// DiagnosticResult 诊断结果
type DiagnosticResult struct {
	SystemInfo    *SystemInfo      `json:"system_info"`
	Dependencies  []DependencyInfo `json:"dependencies"`
	NetworkTests  []NetworkTest    `json:"network_tests"`
	FileChecks    []FileCheck      `json:"file_checks"`
	Errors        []AlpineError    `json:"errors"`
	Warnings      []string         `json:"warnings"`
	Suggestions   []string         `json:"suggestions"`
	Timestamp     time.Time        `json:"timestamp"`
}

// DiagnosticReport 诊断报告
type DiagnosticReport struct {
	// 报告元数据
	ReportID     string    `json:"report_id"`
	Timestamp    time.Time `json:"timestamp"`
	AgentVersion string    `json:"agent_version"`
	
	// 系统信息
	SystemInfo   SystemInfo `json:"system_info"`
	
	// 检查结果
	Dependencies []DependencyInfo `json:"dependencies"`
	NetworkTests []NetworkTest    `json:"network_tests"`
	FileChecks   []FileCheck      `json:"file_checks"`
	
	// 错误和建议
	Errors      []AlpineError `json:"errors"`
	Warnings    []string      `json:"warnings"`
	Suggestions []string      `json:"suggestions"`
	
	// 状态摘要
	OverallStatus string `json:"overall_status"` // "healthy", "warning", "error"
}

// NetworkStatus 网络状态
type NetworkStatus struct {
	DNSResolution bool          `json:"dns_resolution"`
	HTTPSAccess   bool          `json:"https_access"`
	PanelAccess   bool          `json:"panel_access"`
	Tests         []NetworkTest `json:"tests"`
}

// ContainerInfo 容器环境信息
type ContainerInfo struct {
	IsContainer     bool   `json:"is_container"`
	Runtime         string `json:"runtime"`         // docker, podman, etc.
	ImageName       string `json:"image_name"`
	ContainerID     string `json:"container_id"`
	HasPrivileged   bool   `json:"has_privileged"`
	HasNetworkHost  bool   `json:"has_network_host"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
	LoadAverage   float64 `json:"load_average"`
}

// ToolAvailability 工具可用性
type ToolAvailability struct {
	Name        string `json:"name"`
	Available   bool   `json:"available"`
	Path        string `json:"path"`
	Version     string `json:"version"`
	Alternative string `json:"alternative"` // 替代工具或方法
}

// 注意: UpdateInfo 类型已在 update_checker.go 中定义
// 这里不再重复定义以避免编译冲突