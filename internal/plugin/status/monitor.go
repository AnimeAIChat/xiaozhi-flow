package status

import (
	"context"
	"fmt"
	"net"
	"time"

	"xiaozhi-server-go/internal/platform/logging"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	logger *logging.Logger
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(logger *logging.Logger) *HealthChecker {
	if logger == nil {
		logger = logging.DefaultLogger
	}

	return &HealthChecker{
		logger: logger,
	}
}

// Start 启动健康检查
func (hc *HealthChecker) Start(ctx context.Context, manager *PluginStatusManager, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if hc.logger != nil {
		hc.logger.InfoTag("health_checker", "启动插件健康检查",
			"interval", interval.String())
	}

	for {
		select {
		case <-ctx.Done():
			if hc.logger != nil {
				hc.logger.InfoTag("health_checker", "插件健康检查已停止")
			}
			return
		case <-ticker.C:
			hc.performHealthCheck(manager)
		}
	}
}

// performHealthCheck 执行健康检查
func (hc *HealthChecker) performHealthCheck(manager *PluginStatusManager) {
	// 获取所有插件状态
	response, err := manager.ListPlugins(DefaultPluginFilter())
	if err != nil {
		if hc.logger != nil {
			hc.logger.ErrorTag("health_checker", "获取插件列表失败", "error", err.Error())
		}
		return
	}

	// 对每个插件进行健康检查
	for _, plugin := range response.Plugins {
		if plugin.Status == StatusRunning || plugin.Status == StatusEnabled {
			hc.checkPluginHealth(manager, plugin)
		}
	}
}

// checkPluginHealth 检查单个插件健康状态
func (hc *HealthChecker) checkPluginHealth(manager *PluginStatusManager, plugin PluginStatus) {
	// 简单的健康检查逻辑
	// 检查端口是否可访问
	if plugin.Port > 0 && plugin.Address != "" {
		isHealthy := hc.checkTCPPort(plugin.Port)
		var status HealthStatus
		var details string

		if isHealthy {
			status = HealthStatusHealthy
			details = "端口连接正常"
		} else {
			status = HealthStatusUnhealthy
			details = "端口连接失败"
		}

		manager.UpdatePluginHealth(plugin.ID, status, details)
	} else {
		manager.UpdatePluginHealth(plugin.ID, HealthStatusUnknown, "插件未启动")
	}
}

// checkTCPPort 检查TCP端口是否可访问
func (hc *HealthChecker) checkTCPPort(port int) bool {
	timeout := 3 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}