package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"xiaozhi-server-go/internal/platform/logging"
	pluginpb "xiaozhi-server-go/gen/go/api/proto"
)

// ClientConn gRPC客户端连接封装
type ClientConn struct {
	client pluginpb.PluginServiceClient
	conn   *grpc.ClientConn
	info   *PluginInfo
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID        string
	Name      string
	Type      string
	Address   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ClientPool gRPC客户端连接池
type ClientPool struct {
	connections map[string]*ClientConn
	mu          sync.RWMutex
	logger      *logging.Logger
}

// NewClientPool 创建新的客户端连接池
func NewClientPool(logger *logging.Logger) *ClientPool {
	if logger == nil {
		logger = logging.DefaultLogger
	}

	return &ClientPool{
		connections: make(map[string]*ClientConn),
		logger:      logger,
	}
}

// AddConnection 添加新的插件连接
func (p *ClientPool) AddConnection(pluginID string, address string, info *PluginInfo) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查是否已存在连接
	if _, exists := p.connections[pluginID]; exists {
		return fmt.Errorf("plugin %s already has a connection", pluginID)
	}

	// 创建gRPC连接
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to plugin %s at %s: %w", pluginID, address, err)
	}

	// 创建客户端
	client := pluginpb.NewPluginServiceClient(conn)

	// 设置默认插件信息
	if info == nil {
		info = &PluginInfo{
			ID:      pluginID,
			Address: address,
			Status:  "active",
		}
	}

	// 保存连接
	p.connections[pluginID] = &ClientConn{
		client: client,
		conn:   conn,
		info:   info,
	}

	p.logger.InfoTag("gRPC", "插件连接已添加",
		"plugin_id", pluginID,
		"address", address)

	return nil
}

// RemoveConnection 移除插件连接
func (p *ClientPool) RemoveConnection(pluginID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, exists := p.connections[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s connection not found", pluginID)
	}

	// 关闭连接
	if conn.conn != nil {
		conn.conn.Close()
	}

	// 从池中移除
	delete(p.connections, pluginID)

	p.logger.InfoTag("gRPC", "插件连接已移除",
		"plugin_id", pluginID)

	return nil
}

// GetClient 获取插件客户端
func (p *ClientPool) GetClient(pluginID string) (pluginpb.PluginServiceClient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	conn, exists := p.connections[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s connection not found", pluginID)
	}

	return conn.client, nil
}

// GetConnection 获取插件连接信息
func (p *ClientPool) GetConnection(pluginID string) (*ClientConn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	conn, exists := p.connections[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s connection not found", pluginID)
	}

	return conn, nil
}

// ListConnections 列出所有连接
func (p *ClientPool) ListConnections() map[string]*PluginInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]*PluginInfo)
	for pluginID, conn := range p.connections {
		result[pluginID] = conn.info
	}

	return result
}

// HealthCheck 健康检查所有连接
func (p *ClientPool) HealthCheck(ctx context.Context) map[string]error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	results := make(map[string]error)

	for pluginID, conn := range p.connections {
		if conn.client == nil {
			results[pluginID] = fmt.Errorf("client is nil")
			continue
		}

		// 调用健康检查
		resp, err := conn.client.HealthCheck(ctx, &pluginpb.HealthCheckRequest{
			PluginId: pluginID,
		})

		if err != nil {
			results[pluginID] = fmt.Errorf("health check failed: %w", err)
			p.logger.WarnTag("gRPC", "插件健康检查失败",
				"plugin_id", pluginID,
				"error", err.Error())
		} else if resp.Status != "healthy" {
			results[pluginID] = fmt.Errorf("plugin is not healthy: %s", resp.Status)
			p.logger.WarnTag("gRPC", "插件状态不健康",
				"plugin_id", pluginID,
				"status", resp.Status)
		} else {
			results[pluginID] = nil
			p.logger.DebugTag("gRPC", "插件健康检查通过",
				"plugin_id", pluginID)
		}
	}

	return results
}

// Close 关闭所有连接
func (p *ClientPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for pluginID, conn := range p.connections {
		if conn.conn != nil {
			conn.conn.Close()
		}
		p.logger.InfoTag("gRPC", "插件连接已关闭",
			"plugin_id", pluginID)
	}

	p.connections = make(map[string]*ClientConn)
}

// UpdatePluginStatus 更新插件状态
func (p *ClientPool) UpdatePluginStatus(pluginID string, status string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, exists := p.connections[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s connection not found", pluginID)
	}

	conn.info.Status = status
	conn.info.UpdatedAt = time.Now()

	return nil
}

// IsPluginActive 检查插件是否活跃
func (p *ClientPool) IsPluginActive(pluginID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	conn, exists := p.connections[pluginID]
	if !exists {
		return false
	}

	return conn.info.Status == "active"
}

// GetActivePlugins 获取所有活跃的插件
func (p *ClientPool) GetActivePlugins() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var activePlugins []string
	for pluginID, conn := range p.connections {
		if conn.info.Status == "active" {
			activePlugins = append(activePlugins, pluginID)
		}
	}

	return activePlugins
}

// ReconnectConnection 重新连接插件
func (p *ClientPool) ReconnectConnection(pluginID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, exists := p.connections[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s connection not found", pluginID)
	}

	// 关闭旧连接
	if conn.conn != nil {
		conn.conn.Close()
	}

	// 创建新连接
	newConn, err := grpc.Dial(conn.info.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		conn.info.Status = "error"
		return fmt.Errorf("failed to reconnect to plugin %s: %w", pluginID, err)
	}

	// 更新连接
	conn.conn = newConn
	conn.client = pluginpb.NewPluginServiceClient(newConn)
	conn.info.Status = "active"
	conn.info.UpdatedAt = time.Now()

	p.logger.InfoTag("gRPC", "插件重新连接成功",
		"plugin_id", pluginID,
		"address", conn.info.Address)

	return nil
}