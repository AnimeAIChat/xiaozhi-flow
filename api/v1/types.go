package v1

import (
	"time"
)

// 简化的插件类型，避免依赖 protobuf

// PluginType 插件类型枚举
type PluginType int

const (
	PluginTypeUnspecified PluginType = iota
	PluginTypeAudio
	PluginTypeLLM
	PluginTypeDevice
	PluginTypeUtility
	PluginTypeCustom
)

func (t PluginType) String() string {
	switch t {
	case PluginTypeAudio:
		return "audio"
	case PluginTypeLLM:
		return "llm"
	case PluginTypeDevice:
		return "device"
	case PluginTypeUtility:
		return "utility"
	case PluginTypeCustom:
		return "custom"
	default:
		return "unspecified"
	}
}

// PluginStatus 插件状态枚举
type PluginStatus int

const (
	PluginStatusUnspecified PluginStatus = iota
	PluginStatusStopped
	PluginStatusStarting
	PluginStatusRunning
	PluginStatusStopping
	PluginStatusError
)

func (s PluginStatus) String() string {
	switch s {
	case PluginStatusStopped:
		return "stopped"
	case PluginStatusStarting:
		return "starting"
	case PluginStatusRunning:
		return "running"
	case PluginStatusStopping:
		return "stopping"
	case PluginStatusError:
		return "error"
	default:
		return "unspecified"
	}
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Type        PluginType             `json:"type"`
	Tags        []string               `json:"tags"`
	Capabilities []string              `json:"capabilities"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Healthy   bool              `json:"healthy"`
	Status    string            `json:"status"`
	Checks    []string          `json:"checks"`
	Details   map[string]string `json:"details"`
	Timestamp time.Time         `json:"timestamp"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details string            `json:"details"`
	Context map[string]string `json:"context"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Error     *ErrorInfo             `json:"error"`
	Timestamp time.Time              `json:"timestamp"`
}

// Metrics 指标信息
type Metrics struct {
	Counters   map[string]float64   `json:"counters"`
	Gauges     map[string]float64   `json:"gauges"`
	Histograms map[string]*Histogram `json:"histograms"`
	Timestamp  time.Time            `json:"timestamp"`
}

// Histogram 直方图指标
type Histogram struct {
	Count        uint64    `json:"count"`
	Sum          float64   `json:"sum"`
	Buckets      []float64 `json:"buckets"`
	BucketCounts []uint64  `json:"bucket_counts"`
}

// Usage 使用情况统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
	Metadata    map[string]string      `json:"metadata"`
}

// CallToolRequest 工具调用请求
type CallToolRequest struct {
	ToolName string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
	Options   map[string]string      `json:"options"`
}

// CallToolResponse 工具调用响应
type CallToolResponse struct {
	Success bool                   `json:"success"`
	Result  map[string]interface{} `json:"result"`
	Output  string                 `json:"output"`
	Error   *ErrorInfo             `json:"error"`
}

// ListToolsRequest 列出工具请求
type ListToolsRequest struct{}

// ListToolsResponse 列出工具响应
type ListToolsResponse struct {
	Success bool        `json:"success"`
	Tools   []*ToolInfo `json:"tools"`
	Error   *ErrorInfo  `json:"error"`
}

// GetToolSchemaRequest 获取工具模式请求
type GetToolSchemaRequest struct {
	ToolName string `json:"tool_name"`
}

// GetToolSchemaResponse 获取工具模式响应
type GetToolSchemaResponse struct {
	Success bool                   `json:"success"`
	Schema  map[string]interface{} `json:"schema"`
	Error   *ErrorInfo             `json:"error"`
}

// Audio plugin types

// ProcessAudioRequest 音频处理请求
type ProcessAudioRequest struct {
	AudioData []byte            `json:"audio_data"`
	Format    string            `json:"format"`
	Options   map[string]string `json:"options"`
}

// ProcessAudioResponse 音频处理响应
type ProcessAudioResponse struct {
	Success   bool                   `json:"success"`
	AudioData []byte                 `json:"audio_data"`
	Format    string                 `json:"format"`
	Message   string                 `json:"message"`
	Error     *ErrorInfo             `json:"error"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// LLM plugin types

// GenerateTextRequest 文本生成请求
type GenerateTextRequest struct {
	Prompt      string            `json:"prompt"`
	MaxTokens   int               `json:"max_tokens"`
	Temperature float64           `json:"temperature"`
	Options     map[string]string `json:"options"`
}

// GenerateTextResponse 文本生成响应
type GenerateTextResponse struct {
	Success     bool                   `json:"success"`
	Text        string                 `json:"text"`
	Usage       *Usage                 `json:"usage"`
	Message     string                 `json:"message"`
	Error       *ErrorInfo             `json:"error"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// StreamGenerateTextRequest 流式文本生成请求
type StreamGenerateTextRequest struct {
	Prompt      string            `json:"prompt"`
	MaxTokens   int               `json:"max_tokens"`
	Temperature float64           `json:"temperature"`
	Options     map[string]string `json:"options"`
}

// StreamGenerateTextResponse 流式文本生成响应
type StreamGenerateTextResponse struct {
	Success bool                   `json:"success"`
	Text    string                 `json:"text"`
	Done    bool                   `json:"done"`
	Error   *ErrorInfo             `json:"error"`
}

// Device plugin types

// ControlDeviceRequest 设备控制请求
type ControlDeviceRequest struct {
	DeviceID   string                 `json:"device_id"`
	Command    string                 `json:"command"`
	Parameters map[string]interface{} `json:"parameters"`
	Options    map[string]string      `json:"options"`
}

// ControlDeviceResponse 设备控制响应
type ControlDeviceResponse struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Result    map[string]interface{} `json:"result"`
	Error     *ErrorInfo             `json:"error"`
	Timestamp time.Time              `json:"timestamp"`
}

// GetDeviceStatusRequest 获取设备状态请求
type GetDeviceStatusRequest struct {
	DeviceID string            `json:"device_id"`
	Options  map[string]string `json:"options"`
}

// GetDeviceStatusResponse 获取设备状态响应
type GetDeviceStatusResponse struct {
	Success   bool                   `json:"success"`
	DeviceID  string                 `json:"device_id"`
	Status    map[string]interface{} `json:"status"`
	Message   string                 `json:"message"`
	Error     *ErrorInfo             `json:"error"`
	Timestamp time.Time              `json:"timestamp"`
}

// ListDevicesRequest 列出设备请求
type ListDevicesRequest struct {
	Type    string            `json:"type"`
	Options map[string]string `json:"options"`
}

// ListDevicesResponse 列出设备响应
type ListDevicesResponse struct {
	Success bool        `json:"success"`
	Devices []*DeviceInfo `json:"devices"`
	Message string      `json:"message"`
	Error   *ErrorInfo  `json:"error"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata"`
}

// gRPC service request/response types

// GetInfoRequest 获取信息请求
type GetInfoRequest struct{}

// GetInfoResponse 获取信息响应
type GetInfoResponse struct {
	Info *PluginInfo `json:"info"`
}

// InitializeRequest 初始化请求
type InitializeRequest struct {
	Config      map[string]interface{} `json:"config"`
	Environment map[string]string      `json:"environment"`
}

// InitializeResponse 初始化响应
type InitializeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ExecuteRequest 执行请求
type ExecuteRequest struct {
	Command  string                 `json:"command"`
	Args     map[string]interface{} `json:"args"`
	Options  map[string]string      `json:"options"`
}

// ExecuteResponse 执行响应
type ExecuteResponse struct {
	Result *ExecutionResult `json:"result"`
}

// HealthCheckRequest 健康检查请求
type HealthCheckRequest struct{}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status *HealthStatus `json:"status"`
}

// GetMetricsRequest 获取指标请求
type GetMetricsRequest struct{}

// GetMetricsResponse 获取指标响应
type GetMetricsResponse struct {
	Metrics *Metrics `json:"metrics"`
}

// ShutdownRequest 关闭请求
type ShutdownRequest struct {
	Graceful bool `json:"graceful"`
}

// ShutdownResponse 关闭响应
type ShutdownResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}