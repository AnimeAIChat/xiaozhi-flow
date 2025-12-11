package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginpb "xiaozhi-server-go/gen/go/api/proto"
	"xiaozhi-server-go/internal/platform/logging"
)

// PluginInfo 插件信息
type PluginInfo struct {
	ID           string
	Name         string
	Type         string
	Description  string
	Version      string
	Status       string
	Address      string
	Capabilities []string
	LastSeen     time.Time
}

// DiscoveryService gRPC插件发现服务
type DiscoveryService struct {
	plugins map[string]*PluginInfo
	clients map[string]*grpc.ClientConn
	mu      sync.RWMutex
	logger  *logging.Logger
}

// NewDiscoveryService 创建插件发现服务
func NewDiscoveryService(logger *logging.Logger) *DiscoveryService {
	return &DiscoveryService{
		plugins: make(map[string]*PluginInfo),
		clients: make(map[string]*grpc.ClientConn),
		logger:  logger,
	}
}

// RegisterPlugin 注册插件
func (ds *DiscoveryService) RegisterPlugin(ctx context.Context, pluginID, address string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.logger != nil {
		ds.logger.InfoTag("discovery", "注册插件",
			"plugin_id", pluginID,
			"address", address)
	}

	// 创建gRPC连接
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create connection to plugin %s: %w", pluginID, err)
	}

	// 获取插件信息
	client := pluginpb.NewPluginServiceClient(conn)
	infoResp, err := client.GetPluginInfo(ctx, &pluginpb.GetPluginInfoRequest{
		PluginId: pluginID,
	})
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to get plugin info for %s: %w", pluginID, err)
	}

	// 健康检查
	healthResp, err := client.HealthCheck(ctx, &pluginpb.HealthCheckRequest{})
	if err != nil {
		conn.Close()
		return fmt.Errorf("plugin %s health check failed: %w", pluginID, err)
	}

	// 转换能力信息
	capabilities := make([]string, len(infoResp.Capabilities))
	for i, cap := range infoResp.Capabilities {
		capabilities[i] = cap.Id
	}

	// 保存插件信息
	pluginInfo := &PluginInfo{
		ID:           pluginID,
		Name:         infoResp.PluginInfo.Name,
		Type:         infoResp.PluginInfo.Type,
		Description:  infoResp.PluginInfo.Description,
		Version:      infoResp.PluginInfo.Version,
		Status:       healthResp.Status,
		Address:      address,
		Capabilities: capabilities,
		LastSeen:     time.Now(),
	}

	// 如果插件已存在，关闭旧连接
	if oldConn, exists := ds.clients[pluginID]; exists {
		oldConn.Close()
	}

	ds.plugins[pluginID] = pluginInfo
	ds.clients[pluginID] = conn

	if ds.logger != nil {
		ds.logger.InfoTag("discovery", "插件注册成功",
			"plugin_id", pluginID,
			"name", pluginInfo.Name,
			"capabilities", len(capabilities))
	}

	return nil
}

// UnregisterPlugin 注销插件
func (ds *DiscoveryService) UnregisterPlugin(pluginID string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.logger != nil {
		ds.logger.InfoTag("discovery", "注销插件",
			"plugin_id", pluginID)
	}

	// 关闭连接
	if conn, exists := ds.clients[pluginID]; exists {
		conn.Close()
		delete(ds.clients, pluginID)
	}

	// 删除插件信息
	delete(ds.plugins, pluginID)

	return nil
}

// GetPlugin 获取插件信息
func (ds *DiscoveryService) GetPlugin(pluginID string) (*PluginInfo, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	plugin, exists := ds.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	return plugin, nil
}

// GetAllPlugins 获取所有插件信息
func (ds *DiscoveryService) GetAllPlugins() []*PluginInfo {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	plugins := make([]*PluginInfo, 0, len(ds.plugins))
	for _, plugin := range ds.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// GetPluginsByCapability 根据能力获取插件列表
func (ds *DiscoveryService) GetPluginsByCapability(capabilityID string) []*PluginInfo {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var plugins []*PluginInfo
	for _, plugin := range ds.plugins {
		for _, cap := range plugin.Capabilities {
			if cap == capabilityID {
				plugins = append(plugins, plugin)
				break
			}
		}
	}

	return plugins
}

// GetClient 获取插件客户端
func (ds *DiscoveryService) GetClient(pluginID string) (pluginpb.PluginServiceClient, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	conn, exists := ds.clients[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	return pluginpb.NewPluginServiceClient(conn), nil
}

// HealthCheck 对所有插件进行健康检查
func (ds *DiscoveryService) HealthCheck(ctx context.Context) map[string]error {
	ds.mu.RLock()
	plugins := make(map[string]*PluginInfo)
	for id, plugin := range ds.plugins {
		plugins[id] = plugin
	}
	ds.mu.RUnlock()

	results := make(map[string]error)

	for pluginID := range plugins {
		client, err := ds.GetClient(pluginID)
		if err != nil {
			results[pluginID] = err
			continue
		}

		healthResp, err := client.HealthCheck(ctx, &pluginpb.HealthCheckRequest{})
		if err != nil {
			results[pluginID] = err
			continue
		}

		// 更新插件状态
		ds.mu.Lock()
		if pluginInfo, exists := ds.plugins[pluginID]; exists {
			pluginInfo.Status = healthResp.Status
			pluginInfo.LastSeen = time.Now()
		}
		ds.mu.Unlock()

		results[pluginID] = nil
	}

	return results
}

// StartHealthCheckLoop 启动健康检查循环
func (ds *DiscoveryService) StartHealthCheckLoop(ctx context.Context, interval time.Duration) {
	if ds.logger != nil {
		ds.logger.InfoTag("discovery", "启动插件健康检查循环",
			"interval", interval.String())
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if ds.logger != nil {
				ds.logger.InfoTag("discovery", "插件健康检查循环停止")
			}
			return
		case <-ticker.C:
			ds.HealthCheck(ctx)
		}
	}
}

// Close 关闭发现服务
func (ds *DiscoveryService) Close() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.logger != nil {
		ds.logger.InfoTag("discovery", "关闭插件发现服务")
	}

	// 关闭所有连接
	for pluginID, conn := range ds.clients {
		if err := conn.Close(); err != nil && ds.logger != nil {
			ds.logger.ErrorTag("discovery", "关闭插件连接失败",
				"plugin_id", pluginID,
				"error", err.Error())
		}
	}

	// 清空数据
	ds.clients = make(map[string]*grpc.ClientConn)
	ds.plugins = make(map[string]*PluginInfo)

	return nil
}