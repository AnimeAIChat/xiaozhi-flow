package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	v1 "xiaozhi-server-go/api/v1"
)

// SimplePlugin 简化插件接口
type SimplePlugin interface {
	// 基础生命周期
	Initialize(ctx context.Context, config *InitializeConfig) error
	Shutdown(ctx context.Context) error

	// 健康检查
	HealthCheck(ctx context.Context) *v1.HealthStatus

	// 指标收集
	GetMetrics(ctx context.Context) *v1.Metrics

	// 插件信息
	GetInfo() *v1.PluginInfo

	// 日志记录
	Logger() hclog.Logger

	// 工具调用
	CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse
	ListTools(ctx context.Context) *v1.ListToolsResponse
	GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse
}

// SimplePluginImpl 简化插件实现
type SimplePluginImpl struct {
	info     *v1.PluginInfo
	logger   hclog.Logger
	config   *InitializeConfig
	mu       sync.RWMutex
	started  bool
	metrics  *v1.Metrics
}

// NewSimplePlugin 创建简化插件
func NewSimplePlugin(info *v1.PluginInfo, logger hclog.Logger) *SimplePluginImpl {
	return &SimplePluginImpl{
		info: info,
		logger: logger,
		metrics: &v1.Metrics{
			Counters:   make(map[string]float64),
			Gauges:     make(map[string]float64),
			Histograms: make(map[string]*v1.Histogram),
		},
	}
}

// Initialize 初始化插件
func (p *SimplePluginImpl) Initialize(ctx context.Context, config *InitializeConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("plugin already initialized")
	}

	p.config = config
	p.started = true

	p.logger.Info("Plugin initialized",
		"name", p.info.Name,
		"version", p.info.Version,
		"type", p.info.Type.String())

	// 记录初始化指标
	p.IncrementCounter("plugin.initialize.total")
	p.SetGauge("plugin.uptime", 0)

	return nil
}

// Shutdown 关闭插件
func (p *SimplePluginImpl) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return fmt.Errorf("plugin not initialized")
	}

	p.logger.Info("Plugin shutting down", "name", p.info.Name)
	p.started = false
	p.IncrementCounter("plugin.shutdown.total")

	return nil
}

// HealthCheck 健康检查
func (p *SimplePluginImpl) HealthCheck(ctx context.Context) *v1.HealthStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	healthy := p.started
	status := "stopped"
	if healthy {
		status = "running"
	}

	return &v1.HealthStatus{
		Healthy: healthy,
		Status:  status,
		Checks:  []string{"initialized", "memory"},
		Details: map[string]string{
			"version": p.info.Version,
			"type":    p.info.Type.String(),
		},
		Timestamp: time.Now(),
	}
}

// GetMetrics 获取指标
func (p *SimplePluginImpl) GetMetrics(ctx context.Context) *v1.Metrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 更新运行时间指标
	if p.started {
		p.SetGauge("plugin.uptime", time.Since(time.Now().Add(-time.Duration(p.metrics.Gauges["plugin.uptime"]) * time.Second)).Seconds())
	}

	p.metrics.Timestamp = time.Now()
	return p.metrics
}

// GetInfo 获取插件信息
func (p *SimplePluginImpl) GetInfo() *v1.PluginInfo {
	return p.info
}

// Logger 获取日志记录器
func (p *SimplePluginImpl) Logger() hclog.Logger {
	return p.logger
}

// 指标操作方法

// IncrementCounter 增加计数器
func (p *SimplePluginImpl) IncrementCounter(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.metrics.Counters == nil {
		p.metrics.Counters = make(map[string]float64)
	}
	p.metrics.Counters[name]++
}

// SetGauge 设置仪表盘值
func (p *SimplePluginImpl) SetGauge(name string, value float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.metrics.Gauges == nil {
		p.metrics.Gauges = make(map[string]float64)
	}
	p.metrics.Gauges[name] = value
}

// RecordHistogram 记录直方图数据
func (p *SimplePluginImpl) RecordHistogram(name string, value float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.metrics.Histograms == nil {
		p.metrics.Histograms = make(map[string]*v1.Histogram)
	}

	hist := p.metrics.Histograms[name]
	if hist == nil {
		hist = &v1.Histogram{
			Buckets:      []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
			BucketCounts: make([]uint64, 6),
		}
		p.metrics.Histograms[name] = hist
	}

	hist.Count++
	hist.Sum += value

	// 更新桶计数
	for i, bucket := range hist.Buckets {
		if value <= bucket {
			hist.BucketCounts[i]++
		}
	}
}

// GetConfig 获取配置
func (p *SimplePluginImpl) GetConfig() *InitializeConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.config
}

// IsStarted 检查插件是否已启动
func (p *SimplePluginImpl) IsStarted() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.started
}

// 工具调用的默认实现
func (p *SimplePluginImpl) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	p.logger.Info("CallTool called", "tool", req.ToolName)

	return &v1.CallToolResponse{
		Success: false,
		Error: &v1.ErrorInfo{
			Code:    "NOT_IMPLEMENTED",
			Message: fmt.Sprintf("Tool %s not implemented", req.ToolName),
		},
	}
}

func (p *SimplePluginImpl) ListTools(ctx context.Context) *v1.ListToolsResponse {
	return &v1.ListToolsResponse{
		Success: false,
		Error: &v1.ErrorInfo{
			Code:    "NOT_IMPLEMENTED",
			Message: "Tool listing not implemented",
		},
	}
}

func (p *SimplePluginImpl) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
	return &v1.GetToolSchemaResponse{
		Success: false,
		Error: &v1.ErrorInfo{
			Code:    "NOT_IMPLEMENTED",
			Message: fmt.Sprintf("Tool schema for %s not implemented", req.ToolName),
		},
	}
}

// SimplePluginRPC 简化的 RPC 插件实现
type SimplePluginRPC struct {
	plugin.Plugin
	Impl SimplePlugin
}

func (p *SimplePluginRPC) GRPCServer(broker *plugin.GRPCBroker, s interface{}) error {
	// 简化实现，暂时不使用 gRPC
	return nil
}

func (p *SimplePluginRPC) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c interface{}) (interface{}, error) {
	// 简化实现，暂时不使用 gRPC
	return &SimplePluginRPCClient{}, nil
}

// SimplePluginRPCClient 简化的 RPC 客户端
type SimplePluginRPCClient struct {
}

func (c *SimplePluginRPCClient) Initialize(ctx context.Context, config *InitializeConfig) error {
	// 简化实现
	return nil
}

func (c *SimplePluginRPCClient) Shutdown(ctx context.Context) error {
	// 简化实现
	return nil
}

func (c *SimplePluginRPCClient) HealthCheck(ctx context.Context) *v1.HealthStatus {
	return &v1.HealthStatus{
		Healthy:   true,
		Status:    "unknown",
		Timestamp: time.Now(),
	}
}

func (c *SimplePluginRPCClient) GetMetrics(ctx context.Context) *v1.Metrics {
	return &v1.Metrics{
		Timestamp: time.Now(),
	}
}

func (c *SimplePluginRPCClient) GetInfo() *v1.PluginInfo {
	return &v1.PluginInfo{
		ID:   "unknown",
		Name: "Unknown Plugin",
	}
}

func (c *SimplePluginRPCClient) Logger() hclog.Logger {
	return hclog.Default()
}

func (c *SimplePluginRPCClient) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	return &v1.CallToolResponse{
		Success: false,
		Error: &v1.ErrorInfo{
			Code:    "NOT_CONNECTED",
			Message: "Plugin not connected",
		},
	}
}

func (c *SimplePluginRPCClient) ListTools(ctx context.Context) *v1.ListToolsResponse {
	return &v1.ListToolsResponse{
		Success: false,
		Error: &v1.ErrorInfo{
			Code:    "NOT_CONNECTED",
			Message: "Plugin not connected",
		},
	}
}

func (c *SimplePluginRPCClient) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
	return &v1.GetToolSchemaResponse{
		Success: false,
		Error: &v1.ErrorInfo{
			Code:    "NOT_CONNECTED",
			Message: "Plugin not connected",
		},
	}
}

// 更新握手配置和插件映射
var SimpleHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "XIAOZHI_PLUGIN",
	MagicCookieValue: "xiaozhi-flow-plugin-system",
}

var SimplePluginMap = map[string]plugin.Plugin{
	"plugin": &SimplePluginRPC{},
}