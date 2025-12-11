package ports

import (
	"fmt"
	"net"
	"sync"
	"time"

	"xiaozhi-server-go/internal/platform/logging"
)

// PortAllocator 动态端口分配器
type PortAllocator struct {
	portRange PortRange
	allocated map[int]bool              // 已分配端口
	reserved  map[string]int            // 预留端口 plugin_id -> port
	records   map[int]*PortAllocation   // 端口分配记录
	mutex     sync.RWMutex               // 读写锁
	logger    *logging.Logger            // 日志记录器
}

// NewPortAllocator 创建新的端口分配器
func NewPortAllocator(portRange PortRange, logger *logging.Logger) *PortAllocator {
	if logger == nil {
		logger = logging.DefaultLogger
	}

	return &PortAllocator{
		portRange: portRange,
		allocated: make(map[int]bool),
		reserved:  make(map[string]int),
		records:   make(map[int]*PortAllocation),
		logger:    logger,
	}
}

// NewDefaultPortAllocator 创建使用默认端口范围的分配器
func NewDefaultPortAllocator(logger *logging.Logger) *PortAllocator {
	return NewPortAllocator(DefaultPortRange(), logger)
}

// FindAvailablePort 为指定插件查找可用端口
func (pa *PortAllocator) FindAvailablePort(pluginID string) (int, error) {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()

	if pa.logger != nil {
		pa.logger.DebugTag("port_allocator", "开始为插件分配端口",
			"plugin_id", pluginID,
			"port_range", fmt.Sprintf("%d-%d", pa.portRange.Start, pa.portRange.End))
	}

	// 首先检查是否已有预分配端口
	if port, exists := pa.reserved[pluginID]; exists {
		if pa.IsPortAvailableUnlocked(port) {
			pa.allocated[port] = true
			pa.updateRecord(port, pluginID, StatusAllocated)

			if pa.logger != nil {
				pa.logger.InfoTag("port_allocator", "使用预分配端口",
					"plugin_id", pluginID,
					"port", port)
			}
			return port, nil
		}
		// 预分配端口不可用，清除记录
		delete(pa.reserved, pluginID)
		delete(pa.allocated, port)
		pa.updateRecord(port, pluginID, StatusError)
	}

	// 在指定范围内寻找可用端口
	for port := pa.portRange.Start; port <= pa.portRange.End; port++ {
		if !pa.allocated[port] && pa.IsPortAvailableUnlocked(port) {
			pa.allocated[port] = true
			pa.reserved[pluginID] = port
			pa.updateRecord(port, pluginID, StatusAllocated)

			if pa.logger != nil {
				pa.logger.InfoTag("port_allocator", "成功分配端口",
					"plugin_id", pluginID,
					"port", port)
			}
			return port, nil
		}
	}

	if pa.logger != nil {
		pa.logger.ErrorTag("port_allocator", "无可用端口",
			"plugin_id", pluginID,
			"port_range", fmt.Sprintf("%d-%d", pa.portRange.Start, pa.portRange.End))
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", pa.portRange.Start, pa.portRange.End)
}

// IsPortAvailable 检查指定端口是否可用
func (pa *PortAllocator) IsPortAvailable(port int) bool {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()

	// 检查是否在范围内
	if port < pa.portRange.Start || port > pa.portRange.End {
		return false
	}

	// 检查是否已分配
	if pa.allocated[port] {
		return false
	}

	return pa.IsPortAvailableUnlocked(port)
}

// IsPortAvailableUnlocked 不加锁检查端口可用性（内部使用）
func (pa *PortAllocator) IsPortAvailableUnlocked(port int) bool {
	// 尝试监听端口来检查可用性
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ReservePort 预留指定端口给插件
func (pa *PortAllocator) ReservePort(pluginID string, port int) error {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()

	// 检查端口是否在范围内
	if port < pa.portRange.Start || port > pa.portRange.End {
		return fmt.Errorf("port %d is out of range %d-%d", port, pa.portRange.Start, pa.portRange.End)
	}

	// 检查端口是否已分配
	if pa.allocated[port] {
		return fmt.Errorf("port %d is already allocated", port)
	}

	// 检查端口是否可用
	if !pa.IsPortAvailableUnlocked(port) {
		return fmt.Errorf("port %d is not available", port)
	}

	// 预留端口
	pa.allocated[port] = true
	pa.reserved[pluginID] = port
	pa.updateRecord(port, pluginID, StatusReserved)

	if pa.logger != nil {
		pa.logger.InfoTag("port_allocator", "预留端口成功",
			"plugin_id", pluginID,
			"port", port)
	}

	return nil
}

// ReleasePort 释放端口
func (pa *PortAllocator) ReleasePort(port int) {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()

	if !pa.allocated[port] {
		return
	}

	// 查找对应的插件ID
	pluginID := ""
	for id, p := range pa.reserved {
		if p == port {
			pluginID = id
			break
		}
	}

	if pluginID != "" {
		delete(pa.reserved, pluginID)
	}

	delete(pa.allocated, port)
	pa.updateRecord(port, pluginID, StatusReleased)

	if pa.logger != nil {
		pa.logger.InfoTag("port_allocator", "释放端口",
			"plugin_id", pluginID,
			"port", port)
	}
}

// ReleasePluginPorts 释放插件的所有端口
func (pa *PortAllocator) ReleasePluginPorts(pluginID string) {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()

	if port, exists := pa.reserved[pluginID]; exists {
		delete(pa.reserved, pluginID)
		delete(pa.allocated, port)
		pa.updateRecord(port, pluginID, StatusReleased)

		if pa.logger != nil {
			pa.logger.InfoTag("port_allocator", "释放插件端口",
				"plugin_id", pluginID,
				"port", port)
		}
	}
}

// GetPluginPort 获取插件分配的端口
func (pa *PortAllocator) GetPluginPort(pluginID string) (int, bool) {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()

	port, exists := pa.reserved[pluginID]
	return port, exists
}

// GetStats 获取端口统计信息
func (pa *PortAllocator) GetStats() PortStats {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()

	totalPorts := pa.portRange.End - pa.portRange.Start + 1
	allocatedPorts := len(pa.allocated)
	reservedPorts := len(pa.reserved)
	availablePorts := totalPorts - allocatedPorts

	usagePercent := 0.0
	if totalPorts > 0 {
		usagePercent = float64(allocatedPorts) / float64(totalPorts) * 100
	}

	return PortStats{
		TotalPorts:     totalPorts,
		AllocatedPorts: allocatedPorts,
		AvailablePorts: availablePorts,
		ReservedPorts:  reservedPorts,
		UsagePercent:   usagePercent,
	}
}

// GetAllocations 获取所有端口分配记录
func (pa *PortAllocator) GetAllocations() []PortAllocation {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()

	allocations := make([]PortAllocation, 0, len(pa.records))
	for _, record := range pa.records {
		allocations = append(allocations, *record)
	}

	return allocations
}

// updateRecord 更新端口分配记录
func (pa *PortAllocator) updateRecord(port int, pluginID string, status PortAllocationStatus) {
	if _, exists := pa.records[port]; !exists {
		pa.records[port] = &PortAllocation{
			Port:      port,
			PluginID:  pluginID,
			Address:   fmt.Sprintf("0.0.0.0:%d", port),
			Timestamp: time.Now(),
			Status:    string(status),
		}
	} else {
		pa.records[port].PluginID = pluginID
		pa.records[port].Status = string(status)
		pa.records[port].Timestamp = time.Now()
	}
}

// CleanupExpiredRecords 清理过期的记录
func (pa *PortAllocator) CleanupExpiredRecords(maxAge time.Duration) {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for port, record := range pa.records {
		if record.Timestamp.Before(cutoff) && record.Status == string(StatusReleased) {
			delete(pa.records, port)
		}
	}

	if pa.logger != nil {
		pa.logger.DebugTag("port_allocator", "清理过期记录",
			"max_age", maxAge.String())
	}
}