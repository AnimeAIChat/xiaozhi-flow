package workflow

import (
	"context"
	"time"
)

// NodeType 节点类型
type NodeType string

const (
	NodeTypeStart     NodeType = "start"     // 开始节点
	NodeTypeEnd       NodeType = "end"       // 结束节点
	NodeTypeTask      NodeType = "task"      // 任务节点
	NodeTypeCondition NodeType = "condition" // 条件节点
	NodeTypeParallel  NodeType = "parallel"  // 并行节点
	NodeTypeMerge     NodeType = "merge"     // 合并节点
)

// NodeStatus 节点状态
type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"   // 待执行
	NodeStatusRunning   NodeStatus = "running"   // 执行中
	NodeStatusCompleted NodeStatus = "completed" // 已完成
	NodeStatusFailed    NodeStatus = "failed"    // 执行失败
	NodeStatusSkipped   NodeStatus = "skipped"   // 已跳过
)

// Edge 连接边
type Edge struct {
	ID     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Label  string `json:"label,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

// InputSchema 输入Schema定义
type InputSchema struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`        // string, number, boolean, object, array
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
	Validation  *Validation `json:"validation,omitempty"`
}

// OutputSchema 输出Schema定义
type OutputSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// Validation 验证规则
type Validation struct {
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Enum      []string `json:"enum,omitempty"`
}

// Node 节点定义
type Node struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        NodeType       `json:"type"`
	Description string         `json:"description"`
	Plugin      string         `json:"plugin"`      // 关联的插件ID
	Method      string         `json:"method"`      // 调用的方法
	Config      map[string]interface{} `json:"config"` // 节点配置
	Inputs      []InputSchema  `json:"inputs"`      // 输入Schema
	Outputs     []OutputSchema `json:"outputs"`     // 输出Schema
	Position    Position       `json:"position"`    // 画布位置
	Status      NodeStatus     `json:"status"`      // 节点状态
	Error       string         `json:"error,omitempty"` // 错误信息
}

// Position 节点位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Workflow 工作流定义
type Workflow struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Version     string  `json:"version"`
	Nodes       []Node  `json:"nodes"`
	Edges       []Edge  `json:"edges"`
	Config      WorkflowConfig `json:"config"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	Timeout       time.Duration `json:"timeout"`        // 执行超时时间
	MaxRetries    int           `json:"max_retries"`    // 最大重试次数
	ParallelLimit int           `json:"parallel_limit"` // 并行执行限制
	EnableLog     bool          `json:"enable_log"`     // 启用日志
	Variables     map[string]interface{} `json:"variables"` // 全局变量
}

// Execution 执行实例
type Execution struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      ExecutionStatus        `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Context     map[string]interface{} `json:"context"`     // 执行上下文
	NodeResults map[string]*NodeResult `json:"node_results"` // 节点执行结果
	Inputs      map[string]interface{} `json:"inputs"`       // 输入参数
	Outputs     map[string]interface{} `json:"outputs"`      // 输出结果
	Error       string                 `json:"error,omitempty"` // 执行错误
	Logs        []ExecutionLog         `json:"logs"`         // 执行日志
}

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"   // 待执行
	ExecutionStatusRunning   ExecutionStatus = "running"   // 执行中
	ExecutionStatusCompleted ExecutionStatus = "completed" // 已完成
	ExecutionStatusFailed    ExecutionStatus = "failed"    // 执行失败
	ExecutionStatusCancelled ExecutionStatus = "cancelled" // 已取消
)

// NodeResult 节点执行结果
type NodeResult struct {
	NodeID      string                 `json:"node_id"`
	Status      NodeStatus             `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
	Error       string                 `json:"error,omitempty"`
	ElapsedTime time.Duration          `json:"elapsed_time"`
}

// ExecutionLog 执行日志
type ExecutionLog struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // info, warn, error
	NodeID    string    `json:"node_id,omitempty"`
	Message   string    `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// PluginManager 插件管理器接口
type PluginManager interface {
	// 启动插件进程
	StartPlugin(ctx context.Context, pluginID string) (*PluginProcess, error)
	// 停止插件进程
	StopPlugin(ctx context.Context, pluginID string) error
	// 获取插件进程
	GetPlugin(pluginID string) (*PluginProcess, bool)
	// 健康检查
	HealthCheck(ctx context.Context, pluginID string) (*PluginHealth, error)
	// 调用插件方法
	CallPlugin(ctx context.Context, pluginID, method string, payload map[string]interface{}) (map[string]interface{}, error)
	// 获取所有插件状态
	ListPlugins() map[string]*PluginProcess
}

// PluginProcess 插件进程信息
type PluginProcess struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Version   string              `json:"version"`
	Status    PluginStatus        `json:"status"`
	Health    *PluginHealth       `json:"health"`
	Config    PluginConfig        `json:"config"`
	StartTime time.Time           `json:"start_time"`
	EndTime   *time.Time          `json:"end_time,omitempty"`
	Metadata  map[string]string   `json:"metadata"`
	Stats     *PluginStats        `json:"stats"`
}

// PluginStatus 插件状态
type PluginStatus string

const (
	PluginStatusStopped   PluginStatus = "stopped"   // 已停止
	PluginStatusStarting  PluginStatus = "starting"  // 启动中
	PluginStatusRunning   PluginStatus = "running"   // 运行中
	PluginStatusStopping  PluginStatus = "stopping"  // 停止中
	PluginStatusError     PluginStatus = "error"     // 错误状态
	PluginStatusUnknown   PluginStatus = "unknown"   // 未知状态
)

// PluginHealth 插件健康状态
type PluginHealth struct {
	Status      string            `json:"status"`
	LastCheck   time.Time         `json:"last_check"`
	ResponseTime time.Duration    `json:"response_time"`
	Message     string            `json:"message"`
	Details     map[string]string `json:"details"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	HTTP   *HTTPPluginConfig   `json:"http,omitempty"`
	GRPC   *GRPCPluginConfig   `json:"grpc,omitempty"`
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// HTTPPluginConfig HTTP插件配置
type HTTPPluginConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Timeout time.Duration     `json:"timeout"`
}

// GRPCPluginConfig gRPC插件配置
type GRPCPluginConfig struct {
	Address string `json:"address"`
	Service string `json:"service"`
	Method  string `json:"method"`
}

// PluginStats 插件统计信息
type PluginStats struct {
	CallCount    int64         `json:"call_count"`
	SuccessCount int64         `json:"success_count"`
	ErrorCount   int64         `json:"error_count"`
	AvgLatency   time.Duration `json:"avg_latency"`
	LastCalled   time.Time     `json:"last_called"`
}

// WorkflowExecutor 工作流执行器接口
type WorkflowExecutor interface {
	// 执行工作流
	Execute(ctx context.Context, workflow *Workflow, inputs map[string]interface{}) (*Execution, error)
	// 取消执行
	Cancel(executionID string) error
	// 获取执行状态
	GetExecution(executionID string) (*Execution, bool)
	// 获取执行日志
	GetExecutionLogs(executionID string) ([]ExecutionLog, error)
}

// DAGEngine DAG引擎接口
type DAGEngine interface {
	// 拓扑排序
	TopologicalSort(nodes []Node, edges []Edge) ([]string, error)
	// 检查循环依赖
	HasCycle(nodes []Node, edges []Edge) bool
	// 获取可执行节点
	GetExecutableNodes(execution *Execution, workflow *Workflow) ([]string, error)
	// 获取节点依赖
	GetNodeDependencies(nodeID string, edges []Edge) []string
}

// DataFlow 数据流传递接口
type DataFlow interface {
	// 传递数据到节点
	PassDataToNode(execution *Execution, nodeID string, data map[string]interface{}) error
	// 获取节点输入数据
	GetNodeInputs(execution *Execution, node *Node, workflow *Workflow) (map[string]interface{}, error)
	// 合并并行节点数据
	MergeParallelData(execution *Execution, nodeIDs []string) (map[string]interface{}, error)
}