package ports

import (
	"time"
)

// PortRange 端口范围配置
type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// PortAllocation 端口分配记录
type PortAllocation struct {
	Port      int       `json:"port"`
	PluginID  string    `json:"plugin_id"`
	Address   string    `json:"address"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // "allocated", "released", "reserved"
}

// PortAllocationStatus 端口分配状态
type PortAllocationStatus string

const (
	StatusAllocated PortAllocationStatus = "allocated"
	StatusReleased  PortAllocationStatus = "released"
	StatusReserved  PortAllocationStatus = "reserved"
	StatusError     PortAllocationStatus = "error"
)

// PortStats 端口统计信息
type PortStats struct {
	TotalPorts     int `json:"total_ports"`
	AllocatedPorts int `json:"allocated_ports"`
	AvailablePorts int `json:"available_ports"`
	ReservedPorts  int `json:"reserved_ports"`
	UsagePercent   float64 `json:"usage_percent"`
}

// DefaultPortRange 默认端口范围
func DefaultPortRange() PortRange {
	return PortRange{
		Start: 20000,
		End:   29999,
	}
}