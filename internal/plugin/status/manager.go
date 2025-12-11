package status

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"xiaozhi-server-go/internal/plugin/capability"
	"xiaozhi-server-go/internal/plugin/ports"
	"xiaozhi-server-go/internal/platform/logging"
)

// PluginStatusManager 插件状态管理器
type PluginStatusManager struct {
	plugins       map[string]*PluginStatus
	portManager   *ports.PortManager
	registry      *capability.Registry
	healthChecker *HealthChecker
	mutex         sync.RWMutex
	logger        *logging.Logger
}

// NewPluginStatusManager 创建插件状态管理器
func NewPluginStatusManager(
	registry *capability.Registry,
	portManager *ports.PortManager,
	logger *logging.Logger,
) *PluginStatusManager {
	if logger == nil {
		logger = logging.DefaultLogger
	}

	psm := &PluginStatusManager{
		plugins:       make(map[string]*PluginStatus),
		portManager:   portManager,
		registry:      registry,
		healthChecker: NewHealthChecker(logger),
		logger:        logger,
	}

	// 自动发现插件
	psm.autoDiscoverPlugins()

	return psm
}

// autoDiscoverPlugins 自动发现并注册插件
func (psm *PluginStatusManager) autoDiscoverPlugins() {
	if psm.registry == nil {
		return
	}

	providers := psm.registry.GetAllProviders()
	for pluginID, pluginProviders := range providers {
		if len(pluginProviders) > 0 {
			psm.registerPlugin(pluginID, pluginProviders[0])
		}
	}
}

// registerPlugin 注册插件到状态管理器
func (psm *PluginStatusManager) registerPlugin(pluginID string, provider capability.Provider) {
	psm.mutex.Lock()
	defer psm.mutex.Unlock()

	capabilities := provider.GetCapabilities()
	capabilityDefs := make([]CapabilityDef, len(capabilities))
	for i, cap := range capabilities {
		capabilityDefs[i] = ConvertFromCapability(cap)
	}

	plugin := &PluginStatus{
		ID:           pluginID,
		Name:         psm.getPluginName(pluginID),
		Type:         psm.getPluginType(pluginID),
		Description:  psm.getPluginDescription(pluginID),
		Version:      "1.0.0",
		Status:       StatusInstalled,
		Capabilities: capabilityDefs,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	psm.plugins[pluginID] = plugin

	if psm.logger != nil {
		psm.logger.InfoTag("plugin_manager", "注册插件",
			"plugin_id", pluginID,
			"name", plugin.Name,
			"type", plugin.Type,
			"capabilities", len(capabilityDefs))
	}
}

// StartPlugin 启动插件
func (psm *PluginStatusManager) StartPlugin(pluginID string) error {
	return psm.StartPluginWithConfig(pluginID, nil)
}

// StartPluginWithConfig 使用配置启动插件
func (psm *PluginStatusManager) StartPluginWithConfig(pluginID string, config map[string]interface{}) error {
	psm.mutex.Lock()
	defer psm.mutex.Unlock()

	plugin, exists := psm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	if plugin.Status == StatusRunning {
		return fmt.Errorf("plugin %s is already running", pluginID)
	}

	if psm.logger != nil {
		psm.logger.InfoTag("plugin_manager", "开始启动插件",
			"plugin_id", pluginID,
			"current_status", plugin.Status)
	}

	// 分配端口
	port, err := psm.portManager.AllocatePortWithRetry(pluginID, 3, 1*time.Second)
	if err != nil {
		plugin.Status = StatusError
		plugin.Error = fmt.Sprintf("端口分配失败: %v", err)
		plugin.UpdatedAt = time.Now()
		return fmt.Errorf("failed to allocate port for plugin %s: %w", pluginID, err)
	}

	// 更新插件状态
	plugin.Status = StatusEnabled
	plugin.Port = port
	plugin.Address = fmt.Sprintf("0.0.0.0:%d", port)
	plugin.UpdatedAt = time.Now()
	plugin.Error = ""

	if psm.logger != nil {
		psm.logger.InfoTag("plugin_manager", "插件端口分配成功",
			"plugin_id", pluginID,
			"port", port,
			"address", plugin.Address)
	}

	return nil
}

// StopPlugin 停止插件
func (psm *PluginStatusManager) StopPlugin(pluginID string) error {
	psm.mutex.Lock()
	defer psm.mutex.Unlock()

	plugin, exists := psm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	if plugin.Status != StatusRunning && plugin.Status != StatusEnabled {
		return fmt.Errorf("plugin %s is not running", pluginID)
	}

	if psm.logger != nil {
		psm.logger.InfoTag("plugin_manager", "停止插件",
			"plugin_id", pluginID,
			"current_status", plugin.Status)
	}

	// 释放端口
	if plugin.Port > 0 {
		psm.portManager.ReleasePort(plugin.Port)
	}

	// 更新状态
	plugin.Status = StatusStopped
	plugin.Port = 0
	plugin.Address = ""
	plugin.UpdatedAt = time.Now()
	plugin.Error = ""

	if psm.logger != nil {
		psm.logger.InfoTag("plugin_manager", "插件已停止",
			"plugin_id", pluginID)
	}

	return nil
}

// RestartPlugin 重启插件
func (psm *PluginStatusManager) RestartPlugin(pluginID string) error {
	if err := psm.StopPlugin(pluginID); err != nil {
		return fmt.Errorf("failed to stop plugin %s: %w", pluginID, err)
	}

	time.Sleep(1 * time.Second) // 等待停止完成

	if err := psm.StartPlugin(pluginID); err != nil {
		return fmt.Errorf("failed to start plugin %s: %w", pluginID, err)
	}

	return nil
}

// ReallocatePort 重新分配端口
func (psm *PluginStatusManager) ReallocatePort(pluginID string) error {
	psm.mutex.Lock()
	defer psm.mutex.Unlock()

	plugin, exists := psm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	if plugin.Status != StatusRunning && plugin.Status != StatusEnabled {
		return fmt.Errorf("plugin %s is not running", pluginID)
	}

	oldPort := plugin.Port

	// 释放旧端口
	if oldPort > 0 {
		psm.portManager.ReleasePort(oldPort)
	}

	// 分配新端口
	newPort, err := psm.portManager.FindAvailablePort(pluginID)
	if err != nil {
		plugin.Status = StatusError
		plugin.Error = fmt.Sprintf("端口重新分配失败: %v", err)
		plugin.UpdatedAt = time.Now()
		return fmt.Errorf("failed to reallocate port for plugin %s: %w", pluginID, err)
	}

	// 更新状态
	plugin.Port = newPort
	plugin.Address = fmt.Sprintf("0.0.0.0:%d", newPort)
	plugin.UpdatedAt = time.Now()
	plugin.Error = ""

	if psm.logger != nil {
		psm.logger.InfoTag("plugin_manager", "插件端口重新分配成功",
			"plugin_id", pluginID,
			"old_port", oldPort,
			"new_port", newPort)
	}

	return nil
}

// UpdatePluginHealth 更新插件健康状态
func (psm *PluginStatusManager) UpdatePluginHealth(pluginID string, status HealthStatus, details string) {
	psm.mutex.Lock()
	defer psm.mutex.Unlock()

	plugin, exists := psm.plugins[pluginID]
	if !exists {
		return
	}

	plugin.HealthStatus = status
	plugin.LastHealthCheck = time.Now()
	plugin.UpdatedAt = time.Now()

	if details != "" {
		plugin.Error = details
	} else if status == HealthStatusHealthy {
		plugin.Error = ""
	}

	if psm.logger != nil {
		psm.logger.DebugTag("plugin_manager", "更新插件健康状态",
			"plugin_id", pluginID,
			"health_status", status,
			"details", details)
	}
}

// GetPluginStatus 获取插件状态
func (psm *PluginStatusManager) GetPluginStatus(pluginID string) (*PluginStatus, error) {
	psm.mutex.RLock()
	defer psm.mutex.RUnlock()

	plugin, exists := psm.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	// 返回副本以避免并发修改
	pluginCopy := *plugin
	return &pluginCopy, nil
}

// ListPlugins 列出所有插件
func (psm *PluginStatusManager) ListPlugins(filter PluginFilter) (*PluginListResponse, error) {
	psm.mutex.RLock()
	defer psm.mutex.RUnlock()

	// 筛选插件
	filteredPlugins := make([]PluginStatus, 0)
	for _, plugin := range psm.plugins {
		if psm.matchesFilter(plugin, filter) {
			filteredPlugins = append(filteredPlugins, *plugin)
		}
	}

	// 排序
	psm.sortPlugins(filteredPlugins, filter)

	// 分页
	total := len(filteredPlugins)
	page := filter.Page
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	var paginatedPlugins []PluginStatus
	if start < total {
		paginatedPlugins = filteredPlugins[start:end]
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &PluginListResponse{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Plugins:    paginatedPlugins,
	}, nil
}

// GetStats 获取插件统计信息
func (psm *PluginStatusManager) GetStats() PluginStats {
	psm.mutex.RLock()
	defer psm.mutex.RUnlock()

	stats := PluginStats{
		TotalPlugins:     len(psm.plugins),
		RunningPlugins:   0,
		StoppedPlugins:   0,
		ErrorPlugins:     0,
		HealthyPlugins:   0,
		UnhealthyPlugins: 0,
		ByType:           make(map[string]int),
		ByStatus:         make(map[PluginStatusType]int),
	}

	for _, plugin := range psm.plugins {
		// 按状态统计
		stats.ByStatus[plugin.Status]++
		switch plugin.Status {
		case StatusRunning:
			stats.RunningPlugins++
		case StatusStopped, StatusDisabled:
			stats.StoppedPlugins++
		case StatusError:
			stats.ErrorPlugins++
		}

		// 按类型统计
		if plugin.Type != "" {
			stats.ByType[plugin.Type]++
		}

		// 按健康状态统计
		switch plugin.HealthStatus {
		case HealthStatusHealthy:
			stats.HealthyPlugins++
		case HealthStatusUnhealthy:
			stats.UnhealthyPlugins++
		}
	}

	// 计算平均端口使用率
	portStats := psm.portManager.GetStats()
	stats.AveragePortUsage = portStats.UsagePercent

	return stats
}

// StartHealthCheck 启动健康检查
func (psm *PluginStatusManager) StartHealthCheck(ctx context.Context, interval time.Duration) {
	psm.healthChecker.Start(ctx, psm, interval)
}

// matchesFilter 检查插件是否匹配筛选条件
func (psm *PluginStatusManager) matchesFilter(plugin *PluginStatus, filter PluginFilter) bool {
	// 类型筛选
	if filter.Type != "" && plugin.Type != filter.Type {
		return false
	}

	// 状态筛选
	if filter.Status != "" && plugin.Status != filter.Status {
		return false
	}

	// 健康状态筛选
	if filter.HealthStatus != "" && plugin.HealthStatus != filter.HealthStatus {
		return false
	}

	// 搜索筛选
	if filter.Search != "" {
		searchLower := FilterToLower(filter.Search)
		if !Contains(PluginToLower(plugin.Name), searchLower) &&
			!Contains(PluginToLower(plugin.Description), searchLower) &&
			!Contains(plugin.ID, searchLower) {
			return false
		}
	}

	return true
}

// sortPlugins 排序插件列表
func (psm *PluginStatusManager) sortPlugins(plugins []PluginStatus, filter PluginFilter) {
	// 实现排序逻辑
	// 这里简化处理，实际可以根据需要实现复杂的排序
}

// getPluginName 获取插件名称
func (psm *PluginStatusManager) getPluginName(pluginID string) string {
	switch pluginID {
	case "openai":
		return "OpenAI"
	case "ollama":
		return "Ollama"
	case "coze":
		return "Coze"
	case "doubao":
		return "Doubao"
	case "chatglm":
		return "ChatGLM"
	case "deepgram":
		return "Deepgram"
	case "gosherpa":
		return "GoSherpa"
	case "stepfun":
		return "StepFun"
	case "edge":
		return "Microsoft Edge TTS"
	default:
		return pluginID
	}
}

// getPluginType 获取插件类型
func (psm *PluginStatusManager) getPluginType(pluginID string) string {
	// 根据插件能力推断类型
	plugin, exists := psm.plugins[pluginID]
	if exists && len(plugin.Capabilities) > 0 {
		return plugin.Capabilities[0].Type
	}
	return "Unknown"
}

// getPluginDescription 获取插件描述
func (psm *PluginStatusManager) getPluginDescription(pluginID string) string {
	switch pluginID {
	case "openai":
		return "OpenAI GPT模型提供者"
	case "ollama":
		return "Ollama本地大语言模型服务"
	case "coze":
		return "Coze AI平台服务"
	case "doubao":
		return "Doubao AI服务平台"
	case "chatglm":
		return "ChatGLM大语言模型服务"
	case "deepgram":
		return "Deepgram语音识别和语音合成服务"
	case "gosherpa":
		return "GoSherpa语音识别和语音合成服务"
	case "stepfun":
		return "StepFun实时语音识别服务"
	case "edge":
		return "Microsoft Edge文本转语音服务"
	default:
		return fmt.Sprintf("%s插件服务", pluginID)
	}
}

// 辅助函数
func FilterToLower(s string) string {
	// 简化的字符串转小写
	return strings.ToLower(s)
}

func PluginToLower(s string) string {
	// 简化的字符串转小写
	return strings.ToLower(s)
}

func Contains(s, substr string) bool {
	// 简化的包含检查
	return len(s) > 0 && len(substr) > 0 && strings.Contains(s, substr)
}