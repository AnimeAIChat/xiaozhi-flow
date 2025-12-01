package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/protobuf/types/known/timestamppb"

	pluginv1 "github.com/kalicyh/xiaozhi-flow/api/v1"
	"github.com/kalicyh/xiaozhi-flow/internal/plugin/config"
	"github.com/kalicyh/xiaozhi-flow/internal/plugin/discovery"
	"github.com/kalicyh/xiaozhi-flow/internal/plugin/registry"
	"github.com/kalicyh/xiaozhi-flow/internal/plugin/runtime"
	"github.com/kalicyh/xiaozhi-flow/internal/plugin/sdk"
)

// PluginManager 插件管理器接口
type PluginManager interface {
	// 插件生命周期管理
	LoadPlugin(ctx context.Context, config *config.PluginConfig) (*LoadedPlugin, error)
	UnloadPlugin(ctx context.Context, pluginID string) error
	RestartPlugin(ctx context.Context, pluginID string) error

	// 插件发现和注册
	DiscoverPlugins(ctx context.Context) ([]*pluginv1.PluginInfo, error)
	RegisterPlugin(plugin *LoadedPlugin) error
	UnregisterPlugin(pluginID string) error

	// 插件查询
	GetPlugin(pluginID string) (*LoadedPlugin, error)
	ListPlugins() ([]*LoadedPlugin, error)
	GetPluginsByType(pluginType pluginv1.PluginType) ([]*LoadedPlugin, error)

	// 健康检查和监控
	HealthCheckAll(ctx context.Context) map[string]*pluginv1.HealthStatus
	GetMetrics(ctx context.Context, pluginID string) (*pluginv1.Metrics, error)

	// 生命周期管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// PluginConfig 插件配置
type PluginConfig struct {
	Enabled    bool                   `yaml:"enabled"`
	Discovery  *DiscoveryConfig       `yaml:"discovery"`
	Registry   *RegistryConfig        `yaml:"registry"`
	HealthCheck *HealthCheckConfig    `yaml:"health_check"`
}

// DiscoveryConfig 发现配置
type DiscoveryConfig struct {
	Enabled     bool          `yaml:"enabled"`
	ScanInterval time.Duration `yaml:"scan_interval"`
	Paths       []string      `yaml:"paths"`
}

// RegistryConfig 注册表配置
type RegistryConfig struct {
	Type string        `yaml:"type"` // memory, redis, etcd
	TTL  time.Duration `yaml:"ttl"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Interval         time.Duration `yaml:"interval"`
	Timeout          time.Duration `yaml:"timeout"`
	FailureThreshold int           `yaml:"failure_threshold"`
}

// LoadedPlugin 已加载的插件
type LoadedPlugin struct {
	ID        string                    `json:"id"`
	Config    *config.PluginConfig      `json:"config"`
	Runtime   runtime.Runtime           `json:"-"`
	Client    plugin.PluginClient       `json:"-"`
	Plugin    interface{}               `json:"-"`
	Info      *pluginv1.PluginInfo      `json:"info"`
	Status    pluginv1.PluginStatus     `json:"status"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`

	// 健康检查相关
	healthCheckCount int           `json:"-"`
	lastHealthCheck  time.Time     `json:"-"`

	// 互斥锁
	mu sync.RWMutex
}

// pluginManagerImpl 插件管理器实现
type pluginManagerImpl struct {
	config       *PluginConfig
	logger       hclog.Logger
	registry     registry.Registry
	discovery    discovery.Discovery
	runtimeMgr   runtime.Manager

	// 已加载的插件
	plugins map[string]*LoadedPlugin
	mu      sync.RWMutex

	// 上下文和取消
	ctx    context.Context
	cancel context.CancelFunc

	// 健康检查
	healthCheckTicker *time.Ticker
}

// NewPluginManager 创建插件管理器
func NewPluginManager(config *PluginConfig, logger hclog.Logger) (PluginManager, error) {
	if config == nil {
		config = &PluginConfig{
			Enabled: true,
			Discovery: &DiscoveryConfig{
				Enabled:     true,
				ScanInterval: 30 * time.Second,
				Paths:       []string{"./plugins", "/opt/xiaozhi-flow/plugins"},
			},
			Registry: &RegistryConfig{
				Type: "memory",
				TTL:  5 * time.Minute,
			},
			HealthCheck: &HealthCheckConfig{
				Interval:         10 * time.Second,
				Timeout:          5 * time.Second,
				FailureThreshold: 3,
			},
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 创建注册表
	reg, err := registry.NewRegistry(config.Registry, logger.Named("registry"))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}

	// 创建发现服务
	disc, err := discovery.NewDiscovery(config.Discovery, logger.Named("discovery"))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create discovery: %w", err)
	}

	// 创建运行时管理器
	runtimeMgr, err := runtime.NewManager(logger.Named("runtime"))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create runtime manager: %w", err)
	}

	return &pluginManagerImpl{
		config:     config,
		logger:     logger,
		registry:   reg,
		discovery:  disc,
		runtimeMgr: runtimeMgr,
		plugins:    make(map[string]*LoadedPlugin),
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start 启动插件管理器
func (pm *pluginManagerImpl) Start(ctx context.Context) error {
	pm.logger.Info("Starting plugin manager")

	// 启动发现服务
	if pm.config.Discovery.Enabled {
		if err := pm.discovery.Start(ctx); err != nil {
			return fmt.Errorf("failed to start discovery: %w", err)
		}
		pm.logger.Info("Plugin discovery started")
	}

	// 启动健康检查
	pm.startHealthCheck()

	// 初始发现插件
	if pm.config.Discovery.Enabled {
		go pm.discoveryLoop()
	}

	pm.logger.Info("Plugin manager started successfully")
	return nil
}

// Stop 停止插件管理器
func (pm *pluginManagerImpl) Stop(ctx context.Context) error {
	pm.logger.Info("Stopping plugin manager")

	// 停止健康检查
	if pm.healthCheckTicker != nil {
		pm.healthCheckTicker.Stop()
	}

	// 卸载所有插件
	pm.mu.Lock()
	pluginIDs := make([]string, 0, len(pm.plugins))
	for id := range pm.plugins {
		pluginIDs = append(pluginIDs, id)
	}
	pm.mu.Unlock()

	for _, id := range pluginIDs {
		if err := pm.UnloadPlugin(ctx, id); err != nil {
			pm.logger.Error("Failed to unload plugin", "plugin_id", id, "error", err)
		}
	}

	// 停止发现服务
	if pm.discovery != nil {
		pm.discovery.Stop(ctx)
	}

	// 取消上下文
	pm.cancel()

	pm.logger.Info("Plugin manager stopped")
	return nil
}

// LoadPlugin 加载插件
func (pm *pluginManagerImpl) LoadPlugin(ctx context.Context, config *config.PluginConfig) (*LoadedPlugin, error) {
	pm.logger.Info("Loading plugin", "plugin_id", config.ID)

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 检查插件是否已加载
	if _, exists := pm.plugins[config.ID]; exists {
		return nil, fmt.Errorf("plugin %s already loaded", config.ID)
	}

	// 选择运行时
	rt, err := pm.runtimeMgr.GetRuntime(config.Deployment.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime: %w", err)
	}

	// 启动插件
	client, err := rt.Start(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	// 获取插件实例
	pluginImpl, err := client.Dispense("plugin")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to dispense plugin: %w", err)
	}

	// 获取插件信息
	var info *pluginv1.PluginInfo
	if basePlugin, ok := pluginImpl.(sdk.BasePlugin); ok {
		info = basePlugin.GetInfo()
	} else {
		// 如果无法获取信息，创建基本信息
		info = &pluginv1.PluginInfo{
			Id:          config.ID,
			Name:        config.Name,
			Version:     config.Version,
			Description: config.Description,
			Type:        pluginv1.PluginType_PLUGIN_TYPE_CUSTOM,
		}
	}

	// 初始化插件
	if basePlugin, ok := pluginImpl.(sdk.BasePlugin); ok {
		initConfig := &sdk.InitializeConfig{
			Config:      config.Config,
			Environment: config.Environment,
		}
		if err := basePlugin.Initialize(ctx, initConfig); err != nil {
			client.Kill()
			return nil, fmt.Errorf("failed to initialize plugin: %w", err)
		}
	}

	// 创建已加载插件实例
	loadedPlugin := &LoadedPlugin{
		ID:        config.ID,
		Config:    config,
		Runtime:   rt,
		Client:    client,
		Plugin:    pluginImpl,
		Info:      info,
		Status:    pluginv1.PluginStatus_PLUGIN_STATUS_RUNNING,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 注册插件
	pm.plugins[config.ID] = loadedPlugin
	if err := pm.registry.Register(info); err != nil {
		pm.logger.Warn("Failed to register plugin in registry", "error", err)
	}

	pm.logger.Info("Plugin loaded successfully",
		"plugin_id", config.ID,
		"name", info.Name,
		"version", info.Version)

	return loadedPlugin, nil
}

// UnloadPlugin 卸载插件
func (pm *pluginManagerImpl) UnloadPlugin(ctx context.Context, pluginID string) error {
	pm.logger.Info("Unloading plugin", "plugin_id", pluginID)

	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 关闭插件
	if basePlugin, ok := plugin.Plugin.(sdk.BasePlugin); ok {
		if err := basePlugin.Shutdown(ctx); err != nil {
			pm.logger.Warn("Failed to shutdown plugin gracefully", "error", err)
		}
	}

	// 终止客户端
	if err := plugin.Client.Kill(); err != nil {
		pm.logger.Warn("Failed to kill plugin client", "error", err)
	}

	// 从运行时清理
	if err := plugin.Runtime.Stop(ctx, pluginID); err != nil {
		pm.logger.Warn("Failed to stop runtime", "error", err)
	}

	// 从注册表移除
	if err := pm.registry.Unregister(pluginID); err != nil {
		pm.logger.Warn("Failed to unregister plugin", "error", err)
	}

	// 从内存中移除
	delete(pm.plugins, pluginID)

	pm.logger.Info("Plugin unloaded successfully", "plugin_id", pluginID)
	return nil
}

// RestartPlugin 重启插件
func (pm *pluginManagerImpl) RestartPlugin(ctx context.Context, pluginID string) error {
	pm.logger.Info("Restarting plugin", "plugin_id", pluginID)

	plugin, exists := pm.GetPlugin(pluginID)
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	config := plugin.Config

	// 先卸载
	if err := pm.UnloadPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to unload plugin for restart: %w", err)
	}

	// 等待一下确保清理完成
	time.Sleep(1 * time.Second)

	// 重新加载
	_, err := pm.LoadPlugin(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to reload plugin: %w", err)
	}

	pm.logger.Info("Plugin restarted successfully", "plugin_id", pluginID)
	return nil
}

// DiscoverPlugins 发现插件
func (pm *pluginManagerImpl) DiscoverPlugins(ctx context.Context) ([]*pluginv1.PluginInfo, error) {
	return pm.discovery.Discover(ctx)
}

// RegisterPlugin 注册插件
func (pm *pluginManagerImpl) RegisterPlugin(plugin *LoadedPlugin) error {
	return pm.registry.Register(plugin.Info)
}

// UnregisterPlugin 取消注册插件
func (pm *pluginManagerImpl) UnregisterPlugin(pluginID string) error {
	return pm.registry.Unregister(pluginID)
}

// GetPlugin 获取插件
func (pm *pluginManagerImpl) GetPlugin(pluginID string) (*LoadedPlugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (pm *pluginManagerImpl) ListPlugins() ([]*LoadedPlugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]*LoadedPlugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// GetPluginsByType 按类型获取插件
func (pm *pluginManagerImpl) GetPluginsByType(pluginType pluginv1.PluginType) ([]*LoadedPlugin, error) {
	plugins, err := pm.ListPlugins()
	if err != nil {
		return nil, err
	}

	var filtered []*LoadedPlugin
	for _, plugin := range plugins {
		if plugin.Info.Type == pluginType {
			filtered = append(filtered, plugin)
		}
	}

	return filtered, nil
}

// HealthCheckAll 健康检查所有插件
func (pm *pluginManagerImpl) HealthCheckAll(ctx context.Context) map[string]*pluginv1.HealthStatus {
	pm.mu.RLock()
	plugins := make(map[string]*LoadedPlugin)
	for id, plugin := range pm.plugins {
		plugins[id] = plugin
	}
	pm.mu.RUnlock()

	results := make(map[string]*pluginv1.HealthStatus)

	for id, plugin := range plugins {
		ctx, cancel := context.WithTimeout(ctx, pm.config.HealthCheck.Timeout)

		if basePlugin, ok := plugin.Plugin.(sdk.BasePlugin); ok {
			status, err := basePlugin.HealthCheck(ctx)
			if err != nil {
				status = &pluginv1.HealthStatus{
					Healthy: false,
					Status:  "error",
					Details: map[string]string{
						"error": err.Error(),
					},
					Timestamp: timestamppb.Now(),
				}
			}
			results[id] = status
		} else {
			results[id] = &pluginv1.HealthStatus{
				Healthy:   false,
				Status:    "no_health_check",
				Timestamp: timestamppb.Now(),
			}
		}

		cancel()
	}

	return results
}

// GetMetrics 获取插件指标
func (pm *pluginManagerImpl) GetMetrics(ctx context.Context, pluginID string) (*pluginv1.Metrics, error) {
	plugin, err := pm.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	if basePlugin, ok := plugin.Plugin.(sdk.BasePlugin); ok {
		return basePlugin.GetMetrics(ctx)
	}

	return &pluginv1.Metrics{}, nil
}

// startHealthCheck 启动健康检查
func (pm *pluginManagerImpl) startHealthCheck() {
	if pm.config.HealthCheck == nil {
		return
	}

	pm.healthCheckTicker = time.NewTicker(pm.config.HealthCheck.Interval)
	go func() {
		for {
			select {
			case <-pm.ctx.Done():
				return
			case <-pm.healthCheckTicker.C:
				pm.performHealthCheck()
			}
		}
	}()

	pm.logger.Info("Health check started", "interval", pm.config.HealthCheck.Interval)
}

// performHealthCheck 执行健康检查
func (pm *pluginManagerImpl) performHealthCheck() {
	ctx, cancel := context.WithTimeout(pm.ctx, pm.config.HealthCheck.Timeout)
	defer cancel()

	healthStatuses := pm.HealthCheckAll(ctx)

	for pluginID, status := range healthStatuses {
		pm.mu.Lock()
		if plugin, exists := pm.plugins[pluginID]; exists {
			if status.Healthy {
				plugin.healthCheckCount = 0
				plugin.lastHealthCheck = time.Now()
				if plugin.Status != pluginv1.PluginStatus_PLUGIN_STATUS_RUNNING {
					plugin.Status = pluginv1.PluginStatus_PLUGIN_STATUS_RUNNING
					plugin.UpdatedAt = time.Now()
				}
			} else {
				plugin.healthCheckCount++
				plugin.lastHealthCheck = time.Now()

				// 如果连续失败次数达到阈值，标记为错误状态
				if plugin.healthCheckCount >= pm.config.HealthCheck.FailureThreshold {
					if plugin.Status != pluginv1.PluginStatus_PLUGIN_STATUS_ERROR {
						plugin.Status = pluginv1.PluginStatus_PLUGIN_STATUS_ERROR
						plugin.UpdatedAt = time.Now()
						pm.logger.Error("Plugin marked as unhealthy",
							"plugin_id", pluginID,
							"failure_count", plugin.healthCheckCount,
							"last_error", status.Details["error"])
					}
				}
			}
		}
		pm.mu.Unlock()
	}
}

// discoveryLoop 发现循环
func (pm *pluginManagerImpl) discoveryLoop() {
	ticker := time.NewTicker(pm.config.Discovery.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.discoverAndLoadPlugins()
		}
	}
}

// discoverAndLoadPlugins 发现并加载插件
func (pm *pluginManagerImpl) discoverAndLoadPlugins() {
	ctx, cancel := context.WithTimeout(pm.ctx, 30*time.Second)
	defer cancel()

	pluginInfos, err := pm.DiscoverPlugins(ctx)
	if err != nil {
		pm.logger.Error("Failed to discover plugins", "error", err)
		return
	}

	for _, info := range pluginInfos {
		// 检查插件是否已加载
		if _, err := pm.GetPlugin(info.Id); err == nil {
			continue // 已加载，跳过
		}

		// 这里可以根据插件信息创建配置并自动加载
		pm.logger.Info("Discovered new plugin",
			"plugin_id", info.Id,
			"name", info.Name,
			"type", info.Type.String())

		// 自动加载逻辑可以根据需要实现
		// pm.LoadPlugin(ctx, createConfigFromInfo(info))
	}
}

// getDefaultPluginPaths 获取默认插件路径
func getDefaultPluginPaths() []string {
	paths := []string{"./plugins"}

	// 添加系统插件路径
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".xiaozhi-flow", "plugins"))
	}

	paths = append(paths, "/opt/xiaozhi-flow/plugins")

	// 过滤存在的路径
	var existingPaths []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existingPaths = append(existingPaths, path)
		}
	}

	return existingPaths
}