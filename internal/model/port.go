package model

import (
	"fmt"
	"time"
)

// PortAllocation 端口分配记录数据模型
type PortAllocation struct {
	ID          int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Port        int        `json:"port" gorm:"not null;index"`
	NodeID      int64      `json:"node_id" gorm:"not null;index"`
	Purpose     string     `json:"purpose" gorm:"not null"`     // "inbound", "outbound", "system"
	Status      string     `json:"status" gorm:"not null"`      // "allocated", "released"
	AllocatedAt time.Time  `json:"allocated_at" gorm:"not null"`
	ReleasedAt  *time.Time `json:"released_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (PortAllocation) TableName() string {
	return "port_allocations"
}

// PortRange 端口范围配置
type PortRange struct {
	Start int `json:"start" validate:"min=1,max=65535"`
	End   int `json:"end" validate:"min=1,max=65535"`
}

// Validate 验证端口范围
func (pr *PortRange) Validate() error {
	if pr.Start < 1 || pr.Start > 65535 {
		return NewValidationError("start", pr.Start, "must be between 1 and 65535")
	}
	if pr.End < 1 || pr.End > 65535 {
		return NewValidationError("end", pr.End, "must be between 1 and 65535")
	}
	if pr.Start > pr.End {
		return NewValidationError("range", pr, "start port cannot be greater than end port")
	}
	return nil
}

// Contains 检查端口是否在范围内
func (pr *PortRange) Contains(port int) bool {
	return port >= pr.Start && port <= pr.End
}

// Size 获取端口范围大小
func (pr *PortRange) Size() int {
	return pr.End - pr.Start + 1
}
// ValidationError 验证错误
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// NewValidationError 创建验证错误
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}