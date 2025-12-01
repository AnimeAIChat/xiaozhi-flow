package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"

	pluginv1 "xiaozhi-server-go/api/v1"
	"xiaozhi-server-go/internal/plugin/config"
)

// Registry 插件注册表接口
type Registry interface {
	// 注册插件
	Register(info *pluginv1.PluginInfo) error
	// 取消注册插件
	Unregister(pluginID string) error
	// 获取插件信息
	Get(pluginID string) (*pluginv1.PluginInfo, error)
	// 列出所有插件
	List() ([]*pluginv1.PluginInfo, error)
	// 按类型列出插件
	ListByType(pluginType pluginv1.PluginType) ([]*pluginv1.PluginInfo, error)
	// 搜索插件
	Search(query string) ([]*pluginv1.PluginInfo, error)
	// 清理过期插件
	Cleanup() error
	// 停止注册表
	Shutdown(ctx context.Context) error
}

// MemoryRegistry 内存注册表实现
type MemoryRegistry struct {
	logger   hclog.Logger
	config   *config.RegistryConfig
	plugins  map[string]*RegistryEntry
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	cleanupTiker *time.Ticker
}

// RegistryEntry 注册表条目
type RegistryEntry struct {
	Info       *pluginv1.PluginInfo
	RegisteredAt time.Time
	ExpiresAt  time.Time
}

// NewRegistry 创建注册表（工厂函数）
func NewRegistry(config *config.RegistryConfig, logger hclog.Logger) (Registry, error) {
	if config == nil {
		config = &config.RegistryConfig{
			Type: "memory",
			TTL:  5 * time.Minute,
		}
	}

	switch config.Type {
	case "memory":
		return NewMemoryRegistry(config, logger)
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", config.Type)
	}
}

// NewMemoryRegistry 创建内存注册表
func NewMemoryRegistry(config *config.RegistryConfig, logger hclog.Logger) (*MemoryRegistry, error) {
	ctx, cancel := context.WithCancel(context.Background())

	registry := &MemoryRegistry{
		logger:  logger.Named("memory-registry"),
		config:  config,
		plugins: make(map[string]*RegistryEntry),
		ctx:     ctx,
		cancel:  cancel,
	}

	// 启动清理任务
	if config.TTL > 0 {
		registry.startCleanupTask()
	}

	return registry, nil
}

// Register 注册插件
func (r *MemoryRegistry) Register(info *pluginv1.PluginInfo) error {
	if info == nil {
		return fmt.Errorf("plugin info is nil")
	}

	if info.Id == "" {
		return fmt.Errorf("plugin ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	expiresAt := time.Time{}
	if r.config.TTL > 0 {
		expiresAt = time.Now().Add(r.config.TTL)
	}

	entry := &RegistryEntry{
		Info:         info,
		RegisteredAt: time.Now(),
		ExpiresAt:    expiresAt,
	}

	r.plugins[info.Id] = entry

	r.logger.Info("Plugin registered",
		"plugin_id", info.Id,
		"name", info.Name,
		"version", info.Version,
		"type", info.Type.String(),
		"expires_at", expiresAt.String())

	return nil
}

// Unregister 取消注册插件
func (r *MemoryRegistry) Unregister(pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[pluginID]; !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	delete(r.plugins, pluginID)

	r.logger.Info("Plugin unregistered", "plugin_id", pluginID)
	return nil
}

// Get 获取插件信息
func (r *MemoryRegistry) Get(pluginID string) (*pluginv1.PluginInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	// 检查是否过期
	if !r.config.TTL.IsZero() && time.Now().After(entry.ExpiresAt) {
		return nil, fmt.Errorf("plugin %s has expired", pluginID)
	}

	return entry.Info, nil
}

// List 列出所有插件
func (r *MemoryRegistry) List() ([]*pluginv1.PluginInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	var plugins []*pluginv1.PluginInfo

	for _, entry := range r.plugins {
		// 过滤过期插件
		if !r.config.TTL.IsZero() && now.After(entry.ExpiresAt) {
			continue
		}
		plugins = append(plugins, entry.Info)
	}

	return plugins, nil
}

// ListByType 按类型列出插件
func (r *MemoryRegistry) ListByType(pluginType pluginv1.PluginType) ([]*pluginv1.PluginInfo, error) {
	plugins, err := r.List()
	if err != nil {
		return nil, err
	}

	var filtered []*pluginv1.PluginInfo
	for _, plugin := range plugins {
		if plugin.Type == pluginType {
			filtered = append(filtered, plugin)
		}
	}

	return filtered, nil
}

// Search 搜索插件
func (r *MemoryRegistry) Search(query string) ([]*pluginv1.PluginInfo, error) {
	plugins, err := r.List()
	if err != nil {
		return nil, err
	}

	if query == "" {
		return plugins, nil
	}

	query = strings.ToLower(query)
	var results []*pluginv1.PluginInfo

	for _, plugin := range plugins {
		// 搜索名称
		if strings.Contains(strings.ToLower(plugin.Name), query) {
			results = append(results, plugin)
			continue
		}

		// 搜索描述
		if strings.Contains(strings.ToLower(plugin.Description), query) {
			results = append(results, plugin)
			continue
		}

		// 搜索标签
		for _, tag := range plugin.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, plugin)
				break
			}
		}

		// 搜索能力
		for _, capability := range plugin.Capabilities {
			if strings.Contains(strings.ToLower(capability), query) {
				results = append(results, plugin)
				break
			}
		}
	}

	return results, nil
}

// Cleanup 清理过期插件
func (r *MemoryRegistry) Cleanup() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	var cleanedCount int

	for pluginID, entry := range r.plugins {
		if !r.config.TTL.IsZero() && now.After(entry.ExpiresAt) {
			delete(r.plugins, pluginID)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		r.logger.Info("Cleaned up expired plugins", "count", cleanedCount)
	}

	return nil
}

// Shutdown 关闭注册表
func (r *MemoryRegistry) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down memory registry")

	// 停止清理任务
	if r.cleanupTiker != nil {
		r.cleanupTiker.Stop()
	}

	// 取消上下文
	r.cancel()

	r.logger.Info("Memory registry shutdown complete")
	return nil
}

// startCleanupTask 启动清理任务
func (r *MemoryRegistry) startCleanupTask() {
	// 清理间隔为TTL的1/4，或者最少1分钟
	interval := r.config.TTL / 4
	if interval < time.Minute {
		interval = time.Minute
	}

	r.cleanupTiker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return
			case <-r.cleanupTiker.C:
				if err := r.Cleanup(); err != nil {
					r.logger.Error("Failed to cleanup expired plugins", "error", err)
				}
			}
		}
	}()

	r.logger.Info("Cleanup task started", "interval", interval.String())
}