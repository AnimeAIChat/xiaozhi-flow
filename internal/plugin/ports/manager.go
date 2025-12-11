package ports

import (
	"context"
	"time"

	"xiaozhi-server-go/internal/platform/logging"
)

// PortManager 端口管理器，提供高级端口管理功能
type PortManager struct {
	allocator *PortAllocator
	logger    *logging.Logger
}

// NewPortManager 创建端口管理器
func NewPortManager(portRange PortRange, logger *logging.Logger) *PortManager {
	allocator := NewPortAllocator(portRange, logger)

	return &PortManager{
		allocator: allocator,
		logger:    logger,
	}
}

// NewDefaultPortManager 创建使用默认配置的端口管理器
func NewDefaultPortManager(logger *logging.Logger) *PortManager {
	return NewPortManager(DefaultPortRange(), logger)
}

// AllocatePortWithRetry 带重试的端口分配
func (pm *PortManager) AllocatePortWithRetry(pluginID string, maxRetries int, retryDelay time.Duration) (int, error) {
	if pm.logger != nil {
		pm.logger.InfoTag("port_manager", "开始为插件分配端口",
			"plugin_id", pluginID,
			"max_retries", maxRetries,
			"retry_delay", retryDelay.String())
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		port, err := pm.allocator.FindAvailablePort(pluginID)
		if err == nil {
			if pm.logger != nil {
				pm.logger.InfoTag("port_manager", "端口分配成功",
					"plugin_id", pluginID,
					"port", port,
					"attempt", attempt)
			}
			return port, nil
		}

		lastErr = err
		if attempt < maxRetries {
			if pm.logger != nil {
				pm.logger.WarnTag("port_manager", "端口分配失败，准备重试",
					"plugin_id", pluginID,
					"attempt", attempt,
					"max_retries", maxRetries,
					"error", err.Error())
			}

			// 指数退避
			delay := time.Duration(attempt) * retryDelay
			time.Sleep(delay)
		}
	}

	if pm.logger != nil {
		pm.logger.ErrorTag("port_manager", "端口分配最终失败",
			"plugin_id", pluginID,
			"max_retries", maxRetries,
			"error", lastErr.Error())
	}

	return 0, lastErr
}

// ValidatePort 验证端口是否在有效范围内
func (pm *PortManager) ValidatePort(port int) error {
	stats := pm.allocator.GetStats()
	if port < stats.TotalPorts {
		return nil
	}
	return ErrPortOutOfRange
}

// GetRecommendedPortRange 获取推荐的端口范围
func (pm *PortManager) GetRecommendedPortRange() PortRange {
	return pm.allocator.portRange
}

// StartCleanupTask 启动清理任务
func (pm *PortManager) StartCleanupTask(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if pm.logger != nil {
		pm.logger.InfoTag("port_manager", "启动端口记录清理任务",
			"interval", interval.String())
	}

	for {
		select {
		case <-ctx.Done():
			if pm.logger != nil {
				pm.logger.InfoTag("port_manager", "端口记录清理任务已停止")
			}
			return
		case <-ticker.C:
			pm.allocator.CleanupExpiredRecords(24 * time.Hour) // 清理24小时前的记录
		}
	}
}

// 错误定义
var (
	ErrPortOutOfRange = &PortError{
		Code:    "PORT_OUT_OF_RANGE",
		Message: "port is out of valid range",
	}
	ErrPortNotAvailable = &PortError{
		Code:    "PORT_NOT_AVAILABLE",
		Message: "port is not available for allocation",
	}
	ErrPluginAlreadyAllocated = &PortError{
		Code:    "PLUGIN_ALREADY_ALLOCATED",
		Message: "plugin already has allocated ports",
	}
)

// ReleasePort 释放端口
func (pm *PortManager) ReleasePort(port int) {
	pm.allocator.ReleasePort(port)
}

// FindAvailablePort 查找可用端口
func (pm *PortManager) FindAvailablePort(pluginID string) (int, error) {
	return pm.allocator.FindAvailablePort(pluginID)
}

// GetStats 获取端口统计信息
func (pm *PortManager) GetStats() PortStats {
	return pm.allocator.GetStats()
}

// PortError 端口相关错误
type PortError struct {
	Code    string
	Message string
}

func (e *PortError) Error() string {
	return e.Message
}

func (e *PortError) ErrorCode() string {
	return e.Code
}