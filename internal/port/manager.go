package port

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// PortRange 端口范围配置
type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// PortAllocation 端口分配记录
type PortAllocation struct {
	ID          int64      `json:"id"`
	Port        int        `json:"port"`
	NodeID      int64      `json:"node_id"`
	Purpose     string     `json:"purpose"`     // "inbound" or "outbound"
	Status      string     `json:"status"`      // "allocated", "released"
	AllocatedAt time.Time  `json:"allocated_at"`
	ReleasedAt  *time.Time `json:"released_at,omitempty"`
}

// PortManager 端口管理器
type PortManager struct {
	allocatedPorts map[int]*PortAllocation
	portRange      PortRange
	mutex          sync.RWMutex
	conflictCount  int
}

// PortConflictError 端口冲突错误
type PortConflictError struct {
	Port    int
	NodeID  int64
	Message string
}

func (e *PortConflictError) Error() string {
	return fmt.Sprintf("port %d is already allocated to node %d: %s",
		e.Port, e.NodeID, e.Message)
}

// PortExhaustionError 端口耗尽错误
type PortExhaustionError struct {
	Range PortRange
	Used  int
}

func (e *PortExhaustionError) Error() string {
	return fmt.Sprintf("no available ports in range %d-%d, %d ports already used",
		e.Range.Start, e.Range.End, e.Used)
}

// NewPortManager 创建新的端口管理器
func NewPortManager(portRange PortRange) *PortManager {
	return &PortManager{
		allocatedPorts: make(map[int]*PortAllocation),
		portRange:      portRange,
		conflictCount:  0,
	}
}

// AllocatePort 分配一个可用端口
func (pm *PortManager) AllocatePort(nodeID int64, purpose string) (int, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for port := pm.portRange.Start; port <= pm.portRange.End; port++ {
		if !pm.isPortAllocated(port) && pm.isSystemPortAvailable(port) {
			allocation := &PortAllocation{
				Port:        port,
				NodeID:      nodeID,
				Purpose:     purpose,
				Status:      "allocated",
				AllocatedAt: time.Now(),
			}
			pm.allocatedPorts[port] = allocation
			return port, nil
		}
	}

	return 0, &PortExhaustionError{
		Range: pm.portRange,
		Used:  len(pm.allocatedPorts),
	}
}

// AllocateSpecificPort 分配指定端口
func (pm *PortManager) AllocateSpecificPort(port int, nodeID int64, purpose string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if port < pm.portRange.Start || port > pm.portRange.End {
		return fmt.Errorf("port %d is outside allowed range %d-%d",
			port, pm.portRange.Start, pm.portRange.End)
	}

	if existing := pm.allocatedPorts[port]; existing != nil {
		pm.conflictCount++
		return &PortConflictError{
			Port:    port,
			NodeID:  existing.NodeID,
			Message: fmt.Sprintf("port already allocated for %s", existing.Purpose),
		}
	}

	if !pm.isSystemPortAvailable(port) {
		return fmt.Errorf("port %d is not available on system", port)
	}

	allocation := &PortAllocation{
		Port:        port,
		NodeID:      nodeID,
		Purpose:     purpose,
		Status:      "allocated",
		AllocatedAt: time.Now(),
	}
	pm.allocatedPorts[port] = allocation
	return nil
}

// ReleasePort 释放端口
func (pm *PortManager) ReleasePort(port int) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	allocation := pm.allocatedPorts[port]
	if allocation == nil {
		return fmt.Errorf("port %d is not allocated", port)
	}

	now := time.Now()
	allocation.Status = "released"
	allocation.ReleasedAt = &now

	delete(pm.allocatedPorts, port)
	return nil
}

// ReleasePortsByNode 释放节点的所有端口
func (pm *PortManager) ReleasePortsByNode(nodeID int64) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	var releasedPorts []int
	for port, allocation := range pm.allocatedPorts {
		if allocation.NodeID == nodeID {
			now := time.Now()
			allocation.Status = "released"
			allocation.ReleasedAt = &now
			releasedPorts = append(releasedPorts, port)
		}
	}

	for _, port := range releasedPorts {
		delete(pm.allocatedPorts, port)
	}

	return nil
}

// IsPortAvailable 检查端口是否可用
func (pm *PortManager) IsPortAvailable(port int) bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return !pm.isPortAllocated(port) && pm.isSystemPortAvailable(port)
}

// GetUsedPorts 获取已使用的端口列表
func (pm *PortManager) GetUsedPorts() []int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	ports := make([]int, 0, len(pm.allocatedPorts))
	for port := range pm.allocatedPorts {
		ports = append(ports, port)
	}
	return ports
}

// GetPortAllocation 获取端口分配信息
func (pm *PortManager) GetPortAllocation(port int) *PortAllocation {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if allocation := pm.allocatedPorts[port]; allocation != nil {
		// 返回副本以避免并发修改
		copy := *allocation
		return &copy
	}
	return nil
}

// GetNodePorts 获取节点的所有端口
func (pm *PortManager) GetNodePorts(nodeID int64) []int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var ports []int
	for port, allocation := range pm.allocatedPorts {
		if allocation.NodeID == nodeID {
			ports = append(ports, port)
		}
	}
	return ports
}

// isPortAllocated 检查端口是否已分配（内部方法，不加锁）
func (pm *PortManager) isPortAllocated(port int) bool {
	return pm.allocatedPorts[port] != nil
}

// isSystemPortAvailable 检查系统端口是否可用
func (pm *PortManager) isSystemPortAvailable(port int) bool {
	// 尝试监听端口来检查可用性
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// PortUsageMetrics 端口使用统计
type PortUsageMetrics struct {
	TotalPorts     int `json:"total_ports"`
	AllocatedPorts int `json:"allocated_ports"`
	AvailablePorts int `json:"available_ports"`
	ConflictCount  int `json:"conflict_count"`
}

// GetMetrics 获取端口使用统计
func (pm *PortManager) GetMetrics() *PortUsageMetrics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	totalPorts := pm.portRange.End - pm.portRange.Start + 1
	allocatedPorts := len(pm.allocatedPorts)

	return &PortUsageMetrics{
		TotalPorts:     totalPorts,
		AllocatedPorts: allocatedPorts,
		AvailablePorts: totalPorts - allocatedPorts,
		ConflictCount:  pm.conflictCount,
	}
}

// ScanAndRecordUsedPorts 扫描并记录当前系统使用的端口
func (pm *PortManager) ScanAndRecordUsedPorts() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for port := pm.portRange.Start; port <= pm.portRange.End; port++ {
		if !pm.isSystemPortAvailable(port) && !pm.isPortAllocated(port) {
			// 发现系统占用但未记录的端口，记录为系统占用
			allocation := &PortAllocation{
				Port:        port,
				NodeID:      -1, // 使用 -1 表示系统占用
				Purpose:     "system",
				Status:      "allocated",
				AllocatedAt: time.Now(),
			}
			pm.allocatedPorts[port] = allocation
		}
	}

	return nil
}

// GetAvailablePortCount 获取可用端口数量
func (pm *PortManager) GetAvailablePortCount() int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	totalPorts := pm.portRange.End - pm.portRange.Start + 1
	return totalPorts - len(pm.allocatedPorts)
}

// ValidatePortRange 验证端口范围配置
func (pm *PortManager) ValidatePortRange() error {
	if pm.portRange.Start < 1 || pm.portRange.Start > 65535 {
		return fmt.Errorf("invalid start port: %d, must be between 1 and 65535", pm.portRange.Start)
	}
	if pm.portRange.End < 1 || pm.portRange.End > 65535 {
		return fmt.Errorf("invalid end port: %d, must be between 1 and 65535", pm.portRange.End)
	}
	if pm.portRange.Start > pm.portRange.End {
		return fmt.Errorf("start port %d cannot be greater than end port %d", pm.portRange.Start, pm.portRange.End)
	}
	return nil
}