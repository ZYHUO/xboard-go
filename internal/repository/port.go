package repository

import (
	"dashgo/internal/model"
	"time"

	"gorm.io/gorm"
)

// PortRepository 端口分配数据访问层
type PortRepository struct {
	db *gorm.DB
}

// NewPortRepository 创建端口仓库实例
func NewPortRepository(db *gorm.DB) *PortRepository {
	return &PortRepository{db: db}
}

// CreateAllocation 创建端口分配记录
func (r *PortRepository) CreateAllocation(allocation *model.PortAllocation) error {
	return r.db.Create(allocation).Error
}

// GetAllocation 根据端口获取分配记录
func (r *PortRepository) GetAllocation(port int) (*model.PortAllocation, error) {
	var allocation model.PortAllocation
	err := r.db.Where("port = ? AND status = ?", port, "allocated").First(&allocation).Error
	if err != nil {
		return nil, err
	}
	return &allocation, nil
}

// GetAllocationsByNode 获取节点的所有端口分配
func (r *PortRepository) GetAllocationsByNode(nodeID int64) ([]model.PortAllocation, error) {
	var allocations []model.PortAllocation
	err := r.db.Where("node_id = ? AND status = ?", nodeID, "allocated").Find(&allocations).Error
	return allocations, err
}

// GetAllocatedPorts 获取所有已分配的端口
func (r *PortRepository) GetAllocatedPorts() ([]int, error) {
	var ports []int
	err := r.db.Model(&model.PortAllocation{}).
		Where("status = ?", "allocated").
		Pluck("port", &ports).Error
	return ports, err
}

// ReleasePort 释放端口
func (r *PortRepository) ReleasePort(port int) error {
	now := time.Now()
	return r.db.Model(&model.PortAllocation{}).
		Where("port = ? AND status = ?", port, "allocated").
		Updates(map[string]interface{}{
			"status":      "released",
			"released_at": now,
			"updated_at":  now,
		}).Error
}

// ReleasePortsByNode 释放节点的所有端口
func (r *PortRepository) ReleasePortsByNode(nodeID int64) error {
	now := time.Now()
	return r.db.Model(&model.PortAllocation{}).
		Where("node_id = ? AND status = ?", nodeID, "allocated").
		Updates(map[string]interface{}{
			"status":      "released",
			"released_at": now,
			"updated_at":  now,
		}).Error
}

// GetPortUsageStats 获取端口使用统计
func (r *PortRepository) GetPortUsageStats() (map[string]int64, error) {
	var stats []struct {
		Status string
		Count  int64
	}

	err := r.db.Model(&model.PortAllocation{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, stat := range stats {
		result[stat.Status] = stat.Count
	}

	return result, nil
}

// GetPortHistory 获取端口分配历史
func (r *PortRepository) GetPortHistory(port int, limit int) ([]model.PortAllocation, error) {
	var allocations []model.PortAllocation
	query := r.db.Where("port = ?", port).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&allocations).Error
	return allocations, err
}

// GetNodePortHistory 获取节点端口分配历史
func (r *PortRepository) GetNodePortHistory(nodeID int64, limit int) ([]model.PortAllocation, error) {
	var allocations []model.PortAllocation
	query := r.db.Where("node_id = ?", nodeID).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&allocations).Error
	return allocations, err
}

// CleanupOldRecords 清理旧的端口分配记录
func (r *PortRepository) CleanupOldRecords(olderThan time.Time) error {
	return r.db.Where("status = ? AND released_at < ?", "released", olderThan).
		Delete(&model.PortAllocation{}).Error
}

// IsPortAllocated 检查端口是否已分配
func (r *PortRepository) IsPortAllocated(port int) (bool, error) {
	var count int64
	err := r.db.Model(&model.PortAllocation{}).
		Where("port = ? AND status = ?", port, "allocated").
		Count(&count).Error
	return count > 0, err
}

// GetConflictCount 获取端口冲突次数统计
func (r *PortRepository) GetConflictCount(since time.Time) (int64, error) {
	// 这里可以通过日志表或其他方式统计冲突次数
	// 暂时返回0，后续可以扩展
	return 0, nil
}