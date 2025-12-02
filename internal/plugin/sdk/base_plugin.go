package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	pluginv1 "xiaozhi-server-go/api/v1"
)

// 插件握手配置
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "XIAOZHI_PLUGIN",
	MagicCookieValue: "xiaozhi-flow-plugin-system",
}

// 插件映射配置
var PluginMap = map[string]plugin.Plugin{
	"plugin":     &PluginPlugin{},
	"audio":      &AudioPluginPlugin{},
	"llm":        &LLMPluginPlugin{},
	"device":     &DevicePluginPlugin{},
	"utility":    &UtilityPluginPlugin{},
}

// BasePlugin 插件基础接口
type BasePlugin interface {
	// 基础生命周期
	Initialize(ctx context.Context, config *InitializeConfig) error
	Shutdown(ctx context.Context) error

	// 健康检查
	HealthCheck(ctx context.Context) (*pluginv1.HealthStatus, error)

	// 指标收集
	GetMetrics(ctx context.Context) (*pluginv1.Metrics, error)

	// 插件信息
	GetInfo() *pluginv1.PluginInfo

	// 日志记录
	Logger() hclog.Logger
}

// InitializeConfig 插件初始化配置
type InitializeConfig struct {
	Config      map[string]interface{} `json:"config"`
	Environment map[string]string      `json:"environment"`
}

// BasePluginImpl 插件基础实现
type BasePluginImpl struct {
	info     *pluginv1.PluginInfo
	logger   hclog.Logger
	config   *InitializeConfig
	mu       sync.RWMutex
	started  bool
	metrics  *pluginv1.Metrics
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(info *pluginv1.PluginInfo, logger hclog.Logger) *BasePluginImpl {
	return &BasePluginImpl{
		info:    info,
		logger:  logger,
		metrics: &pluginv1.Metrics{
			Counters:   make(map[string]float64),
			Gauges:     make(map[string]float64),
			Histograms: make(map[string]*pluginv1.Histogram),
		},
	}
}

// Initialize 初始化插件
func (p *BasePluginImpl) Initialize(ctx context.Context, config *InitializeConfig) error {
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
func (p *BasePluginImpl) Shutdown(ctx context.Context) error {
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
func (p *BasePluginImpl) HealthCheck(ctx context.Context) (*pluginv1.HealthStatus, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	healthy := p.started
	status := "stopped"
	if healthy {
		status = "running"
	}

	return &pluginv1.HealthStatus{
		Healthy:   healthy,
		Status:    status,
		Checks:    []string{"initialized", "memory"},
		Details: map[string]string{
			"version": p.info.Version,
			"type":    p.info.Type.String(),
		},
		Timestamp: time.Now(),
	}, nil
}

// GetMetrics 获取指标
func (p *BasePluginImpl) GetMetrics(ctx context.Context) (*pluginv1.Metrics, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 更新运行时间指标
	if p.started {
		uptime := time.Since(time.Now().Add(-time.Duration(p.metrics.Gauges["plugin.uptime"]) * time.Second))
		p.SetGauge("plugin.uptime", uptime.Seconds())
	}

	return p.metrics, nil
}

// GetInfo 获取插件信息
func (p *BasePluginImpl) GetInfo() *pluginv1.PluginInfo {
	return p.info
}

// Logger 获取日志记录器
func (p *BasePluginImpl) Logger() hclog.Logger {
	return p.logger
}

// 指标操作方法

// IncrementCounter 增加计数器
func (p *BasePluginImpl) IncrementCounter(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.metrics.Counters == nil {
		p.metrics.Counters = make(map[string]float64)
	}
	p.metrics.Counters[name]++
}

// SetGauge 设置仪表盘值
func (p *BasePluginImpl) SetGauge(name string, value float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.metrics.Gauges == nil {
		p.metrics.Gauges = make(map[string]float64)
	}
	p.metrics.Gauges[name] = value
}

// RecordHistogram 记录直方图数据
func (p *BasePluginImpl) RecordHistogram(name string, value float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.metrics.Histograms == nil {
		p.metrics.Histograms = make(map[string]*pluginv1.Histogram)
	}

	hist := p.metrics.Histograms[name]
	if hist == nil {
		hist = &pluginv1.Histogram{
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
func (p *BasePluginImpl) GetConfig() *InitializeConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.config
}

// IsStarted 检查插件是否已启动
func (p *BasePluginImpl) IsStarted() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.started
}

// 具体插件类型接口

// AudioPlugin 音频处理插件接口
type AudioPlugin interface {
	BasePlugin
	ProcessAudio(ctx context.Context, req *pluginv1.ProcessAudioRequest) (*pluginv1.ProcessAudioResponse, error)
}

// LLMPlugin 大模型插件接口
type LLMPlugin interface {
	BasePlugin
	GenerateText(ctx context.Context, req *pluginv1.GenerateTextRequest) (*pluginv1.GenerateTextResponse, error)
	StreamGenerateText(ctx context.Context, req *pluginv1.StreamGenerateTextRequest) (<-chan *pluginv1.StreamGenerateTextResponse, error)
}

// DevicePlugin 设备控制插件接口
type DevicePlugin interface {
	BasePlugin
	ControlDevice(ctx context.Context, req *pluginv1.ControlDeviceRequest) (*pluginv1.ControlDeviceResponse, error)
	GetDeviceStatus(ctx context.Context, req *pluginv1.GetDeviceStatusRequest) (*pluginv1.GetDeviceStatusResponse, error)
	ListDevices(ctx context.Context, req *pluginv1.ListDevicesRequest) (*pluginv1.ListDevicesResponse, error)
}

// UtilityPlugin 通用功能插件接口
type UtilityPlugin interface {
	BasePlugin
	CallTool(ctx context.Context, req *pluginv1.CallToolRequest) (*pluginv1.CallToolResponse, error)
	ListTools(ctx context.Context, req *pluginv1.ListToolsRequest) (*pluginv1.ListToolsResponse, error)
	GetToolSchema(ctx context.Context, req *pluginv1.GetToolSchemaRequest) (*pluginv1.GetToolSchemaResponse, error)
}

// Note: gRPC implementation removed to avoid protobuf dependencies
// Can be added later when protobuf definitions are available

// 插件实现

type PluginPlugin struct {
	plugin.Plugin
	Impl BasePlugin
}

// Note: GRPC methods removed to avoid protobuf dependencies
// Can be added later when protobuf definitions are available

// 其他插件类型的类似实现（AudioPlugin, LLMPlugin, DevicePlugin, UtilityPlugin）
// 由于篇幅限制，这里只展示基础结构

type AudioPluginPlugin struct {
	plugin.Plugin
	Impl AudioPlugin
}

type LLMPluginPlugin struct {
	plugin.Plugin
	Impl LLMPlugin
}

type DevicePluginPlugin struct {
	plugin.Plugin
	Impl DevicePlugin
}

type UtilityPluginPlugin struct {
	plugin.Plugin
	Impl UtilityPlugin
}