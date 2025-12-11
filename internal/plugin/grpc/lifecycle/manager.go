package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/plugin/capability"
	"xiaozhi-server-go/internal/plugin/grpc/discovery"
)

// PluginStatus 插件状态
type PluginStatus string

const (
	StatusUnknown    PluginStatus = "unknown"
	StatusInstalled  PluginStatus = "installed"
	StatusEnabled    PluginStatus = "enabled"
	StatusDisabled   PluginStatus = "disabled"
	StatusError      PluginStatus = "error"
	StatusRunning    PluginStatus = "running"
	StatusStopped    PluginStatus = "stopped"
)

// PluginMetadata 插件元数据
type PluginMetadata struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Status      PluginStatus           `json:"status"`
	Config      map[string]interface{} `json:"config,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// LifecycleManager 插件生命周期管理器
type LifecycleManager struct {
	registry      *capability.Registry
	discovery     *discovery.DiscoveryService
	plugins       map[string]*PluginMetadata
	pluginPorts   map[string]int
	mu            sync.RWMutex
	logger        *logging.Logger
}

// NewLifecycleManager 创建插件生命周期管理器
func NewLifecycleManager(
	registry *capability.Registry,
	discovery *discovery.DiscoveryService,
	logger *logging.Logger,
) *LifecycleManager {
	return &LifecycleManager{
		registry:    registry,
		discovery:   discovery,
		plugins:     make(map[string]*PluginMetadata),
		pluginPorts: getDefaultPluginPorts(),
		logger:      logger,
	}
}

// getDefaultPluginPorts 获取默认插件端口分配
func getDefaultPluginPorts() map[string]int {
	return map[string]int{
		"openai":   15501,
		"ollama":   15502,
		"coze":     15503,
		"doubao":   15504,
		"chatglm":  15505,
		"deepgram": 15506,
		"gosherpa": 15507,
		"stepfun":  15508,
		"edge":     15509,
	}
}

// InstallPlugin 安装插件
func (lm *LifecycleManager) InstallPlugin(ctx context.Context, pluginID string, config map[string]interface{}) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.logger != nil {
		lm.logger.InfoTag("lifecycle", "安装插件",
			"plugin_id", pluginID)
	}

	// 检查插件是否已安装
	if _, exists := lm.plugins[pluginID]; exists {
		return fmt.Errorf("plugin %s is already installed", pluginID)
	}

	// 创建插件元数据
	metadata := &PluginMetadata{
		ID:        pluginID,
		Status:    StatusInstalled,
		Config:    config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 从能力注册表获取插件信息
	if provider, exists := lm.registry.GetProvider(pluginID); exists {
		if pluginInfo := lm.getPluginInfoFromProvider(pluginID, provider); pluginInfo != nil {
			metadata.Name = pluginInfo.Name
			metadata.Type = pluginInfo.Type
			metadata.Description = pluginInfo.Description
			metadata.Version = pluginInfo.Version
		}
	}

	lm.plugins[pluginID] = metadata

	if lm.logger != nil {
		lm.logger.InfoTag("lifecycle", "插件安装成功",
			"plugin_id", pluginID,
			"name", metadata.Name)
	}

	return nil
}

// UninstallPlugin 卸载插件
func (lm *LifecycleManager) UninstallPlugin(ctx context.Context, pluginID string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.logger != nil {
		lm.logger.InfoTag("lifecycle", "卸载插件",
			"plugin_id", pluginID)
	}

	// 检查插件是否已安装
	metadata, exists := lm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s is not installed", pluginID)
	}

	// 如果插件正在运行，先停止
	if metadata.Status == StatusRunning {
		if err := lm.stopPluginUnsafe(ctx, pluginID); err != nil {
			return fmt.Errorf("failed to stop plugin before uninstall: %w", err)
		}
	}

	// 从发现服务注销
	lm.discovery.UnregisterPlugin(pluginID)

	// 删除插件元数据
	delete(lm.plugins, pluginID)

	if lm.logger != nil {
		lm.logger.InfoTag("lifecycle", "插件卸载成功",
			"plugin_id", pluginID)
	}

	return nil
}

// EnablePlugin 启用插件
func (lm *LifecycleManager) EnablePlugin(ctx context.Context, pluginID string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.logger != nil {
		lm.logger.InfoTag("lifecycle", "启用插件",
			"plugin_id", pluginID)
	}

	metadata, exists := lm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s is not installed", pluginID)
	}

	if metadata.Status == StatusEnabled || metadata.Status == StatusRunning {
		return nil // 插件已启用
	}

	// 启动插件
	if err := lm.startPluginUnsafe(ctx, pluginID); err != nil {
		metadata.Status = StatusError
		return err
	}

	metadata.Status = StatusRunning
	metadata.UpdatedAt = time.Now()

	if lm.logger != nil {
		lm.logger.InfoTag("lifecycle", "插件启用成功",
			"plugin_id", pluginID)
	}

	return nil
}

// DisablePlugin 禁用插件
func (lm *LifecycleManager) DisablePlugin(ctx context.Context, pluginID string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.logger != nil {
		lm.logger.InfoTag("lifecycle", "禁用插件",
			"plugin_id", pluginID)
	}

	metadata, exists := lm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s is not installed", pluginID)
	}

	if metadata.Status == StatusDisabled {
		return nil // 插件已禁用
	}

	// 如果插件正在运行，停止它
	if metadata.Status == StatusRunning {
		if err := lm.stopPluginUnsafe(ctx, pluginID); err != nil {
			return err
		}
	}

	metadata.Status = StatusDisabled
	metadata.UpdatedAt = time.Now()

	if lm.logger != nil {
		lm.logger.InfoTag("lifecycle", "插件禁用成功",
			"plugin_id", pluginID)
	}

	return nil
}

// startPluginUnsafe 启动插件（非线程安全，调用者需要持有锁）
func (lm *LifecycleManager) startPluginUnsafe(ctx context.Context, pluginID string) error {
	// 获取端口
	port, exists := lm.pluginPorts[pluginID]
	if !exists {
		return fmt.Errorf("no port allocated for plugin %s", pluginID)
	}

	address := fmt.Sprintf("0.0.0.0:%d", port)

	// 注册到发现服务
	if err := lm.discovery.RegisterPlugin(ctx, pluginID, address); err != nil {
		return fmt.Errorf("failed to register plugin %s: %w", pluginID, err)
	}

	return nil
}

// stopPluginUnsafe 停止插件（非线程安全，调用者需要持有锁）
func (lm *LifecycleManager) stopPluginUnsafe(ctx context.Context, pluginID string) error {
	// 从发现服务注销
	if err := lm.discovery.UnregisterPlugin(pluginID); err != nil {
		return fmt.Errorf("failed to unregister plugin %s: %w", pluginID, err)
	}

	return nil
}

// GetPluginStatus 获取插件状态
func (lm *LifecycleManager) GetPluginStatus(pluginID string) (*PluginMetadata, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	metadata, exists := lm.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s is not installed", pluginID)
	}

	return metadata, nil
}

// GetAllPlugins 获取所有插件状态
func (lm *LifecycleManager) GetAllPlugins() []*PluginMetadata {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	plugins := make([]*PluginMetadata, 0, len(lm.plugins))
	for _, metadata := range lm.plugins {
		plugins = append(plugins, metadata)
	}

	return plugins
}

// GetRunningPlugins 获取正在运行的插件
func (lm *LifecycleManager) GetRunningPlugins() []*PluginMetadata {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	var running []*PluginMetadata
	for _, metadata := range lm.plugins {
		if metadata.Status == StatusRunning {
			running = append(running, metadata)
		}
	}

	return running
}

// AutoDiscoverPlugins 自动发现已安装的插件
func (lm *LifecycleManager) AutoDiscoverPlugins(ctx context.Context) error {
	if lm.logger != nil {
		lm.logger.InfoTag("lifecycle", "自动发现已安装插件")
	}

	// 从能力注册表获取所有提供者
	providers := lm.registry.GetAllProviders()

	for pluginID, pluginProviders := range providers {
		if len(pluginProviders) == 0 {
			continue
		}

		// 检查是否已安装
		if _, exists := lm.plugins[pluginID]; !exists {
			// 自动安装插件
			metadata := &PluginMetadata{
				ID:        pluginID,
				Status:    StatusInstalled,
				Config:    make(map[string]interface{}),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// 获取插件信息
			provider := pluginProviders[0]
			if pluginInfo := lm.getPluginInfoFromProvider(pluginID, provider); pluginInfo != nil {
				metadata.Name = pluginInfo.Name
				metadata.Type = pluginInfo.Type
				metadata.Description = pluginInfo.Description
				metadata.Version = pluginInfo.Version
			}

			lm.plugins[pluginID] = metadata

			if lm.logger != nil {
				lm.logger.InfoTag("lifecycle", "自动发现插件",
					"plugin_id", pluginID,
					"name", metadata.Name)
			}
		}
	}

	return nil
}

// getPluginInfoFromProvider 从提供者获取插件信息
func (lm *LifecycleManager) getPluginInfoFromProvider(pluginID string, provider capability.Provider) *PluginMetadata {
	capabilities := provider.GetCapabilities()
	if len(capabilities) == 0 {
		return nil
	}

	metadata := &PluginMetadata{}
	firstCap := capabilities[0]
	metadata.Type = string(firstCap.Type)

	// 根据插件ID推断名称
	switch pluginID {
	case "openai":
		metadata.Name = "OpenAI"
		metadata.Description = "OpenAI GPT API Service"
	case "ollama":
		metadata.Name = "Ollama"
		metadata.Description = "Ollama Local LLM Service"
	case "coze":
		metadata.Name = "Coze"
		metadata.Description = "Coze AI Platform Service"
	case "doubao":
		metadata.Name = "Doubao"
		metadata.Description = "Doubao AI Service Platform"
	case "chatglm":
		metadata.Name = "ChatGLM"
		metadata.Description = "ChatGLM Language Model Service"
	case "deepgram":
		metadata.Name = "Deepgram"
		metadata.Description = "Deepgram Speech Recognition Service"
	case "gosherpa":
		metadata.Name = "GoSherpa"
		metadata.Description = "GoSherpa Speech Recognition Service"
	case "stepfun":
		metadata.Name = "StepFun"
		metadata.Description = "StepFun AI Service"
	case "edge":
		metadata.Name = "Microsoft Edge TTS"
		metadata.Description = "Microsoft Edge Text-to-Speech Service"
	default:
		metadata.Name = pluginID
		metadata.Description = fmt.Sprintf("%s Plugin Service", pluginID)
	}

	metadata.Version = "1.0.0"
	return metadata
}

// Close 关闭生命周期管理器
func (lm *LifecycleManager) Close() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.logger != nil {
		lm.logger.InfoTag("lifecycle", "关闭插件生命周期管理器")
	}

	// 停止所有运行中的插件
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for pluginID, metadata := range lm.plugins {
		if metadata.Status == StatusRunning {
			if err := lm.stopPluginUnsafe(ctx, pluginID); err != nil && lm.logger != nil {
				lm.logger.ErrorTag("lifecycle", "停止插件失败",
					"plugin_id", pluginID,
					"error", err.Error())
			}
		}
	}

	return nil
}