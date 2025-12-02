package startup

import (
	"context"
	"fmt"
	"time"

	"xiaozhi-server-go/internal/workflow"
)

// StartupWorkflow 启动工作流定义
type StartupWorkflow struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Version     string              `json:"version"`
	Nodes       []StartupNode       `json:"nodes"`
	Edges       []workflow.Edge     `json:"edges"`
	Config      StartupWorkflowConfig `json:"config"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Tags        []string            `json:"tags,omitempty"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
}

// GetNodeByID 根据ID获取节点
func (w *StartupWorkflow) GetNodeByID(nodeID string) *StartupNode {
	for _, node := range w.Nodes {
		if node.ID == nodeID {
			return &node
		}
	}
	return nil
}

// StartupWorkflowConfig 启动工作流配置
type StartupWorkflowConfig struct {
	Timeout       time.Duration         `json:"timeout"`        // 执行超时时间
	MaxRetries    int                   `json:"max_retries"`     // 最大重试次数
	ParallelLimit int                   `json:"parallel_limit"`  // 并行执行限制
	EnableLog     bool                  `json:"enable_log"`      // 启用日志
	Environment   map[string]interface{} `json:"environment"`    // 环境变量
	Variables     map[string]interface{} `json:"variables"`      // 全局变量
	OnFailure     FailureAction         `json:"on_failure"`     // 失败处理策略
}

// FailureAction 失败处理策略
type FailureAction string

const (
	FailureActionStop    FailureAction = "stop"     // 停止执行
	FailureActionRetry   FailureAction = "retry"    // 重试失败节点
	FailureActionSkip    FailureAction = "skip"     // 跳过失败节点
	FailureActionIgnore  FailureAction = "ignore"   // 忽略失败继续执行
)

// StartupNode 启动工作流节点
type StartupNode struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          StartupNodeType        `json:"type"`
	Description   string                 `json:"description"`
	DependsOn     []string               `json:"depends_on"`
	Config        map[string]interface{} `json:"config"`
	Status        workflow.NodeStatus    `json:"status"`
	Timeout       time.Duration          `json:"timeout"`
	Retry         *RetryConfig           `json:"retry,omitempty"`
	Critical      bool                   `json:"critical"`       // 是否为关键节点
	Optional      bool                   `json:"optional"`       // 是否为可选节点
	Position      workflow.Position      `json:"position"`      // 画布位置
	Metadata      map[string]string      `json:"metadata,omitempty"`
}

// StartupNodeType 启动节点类型
type StartupNodeType string

const (
	StartupNodeStorage      StartupNodeType = "storage"       // 存储相关
	StartupNodeConfig       StartupNodeType = "config"        // 配置相关
	StartupNodeService      StartupNodeType = "service"       // 服务相关
	StartupNodeAuth         StartupNodeType = "auth"          // 认证相关
	StartupNodePlugin       StartupNodeType = "plugin"        // 插件相关
	StartupNodeCustom       StartupNodeType = "custom"        // 自定义节点
	StartupNodeParallel     StartupNodeType = "parallel"      // 并行节点
	StartupNodeMerge        StartupNodeType = "merge"         // 合并节点
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts"`
	Delay       time.Duration `json:"delay"`
	Backoff     bool          `json:"backoff"`
	MaxDelay    time.Duration `json:"max_delay"`
}

// StartupExecution 启动工作流执行实例
type StartupExecution struct {
	ID            string                        `json:"id"`
	WorkflowID    string                        `json:"workflow_id"`
	WorkflowName  string                        `json:"workflow_name"`
	Status        workflow.ExecutionStatus       `json:"status"`
	StartTime     time.Time                     `json:"start_time"`
	EndTime       *time.Time                    `json:"end_time,omitempty"`
	Duration      time.Duration                 `json:"duration"`
	NodeResults   map[string]*StartupNodeResult  `json:"node_results"`
	Context       map[string]interface{}        `json:"context"`
	Inputs        map[string]interface{}        `json:"inputs"`
	Outputs       map[string]interface{}        `json:"outputs"`
	Error         string                        `json:"error,omitempty"`
	Progress      float64                       `json:"progress"`        // 0-100
	TotalNodes    int                           `json:"total_nodes"`
	CompletedNodes int                           `json:"completed_nodes"`
	FailedNodes   int                           `json:"failed_nodes"`
	CurrentNodes  []string                      `json:"current_nodes"`  // 当前正在执行的节点
	Environment   map[string]interface{}        `json:"environment"`   // 执行环境
	TriggeredBy   string                        `json:"triggered_by"`    // 触发方式
}

// StartupNodeResult 启动节点执行结果
type StartupNodeResult struct {
	NodeID       string                     `json:"node_id"`
	NodeName     string                     `json:"node_name"`
	NodeType     StartupNodeType            `json:"node_type"`
	Status       workflow.NodeStatus        `json:"status"`
	StartTime    time.Time                  `json:"start_time"`
	EndTime      *time.Time                 `json:"end_time,omitempty"`
	Duration     time.Duration              `json:"duration"`
	Inputs       map[string]interface{}     `json:"inputs"`
	Outputs      map[string]interface{}     `json:"outputs"`
	Error        string                     `json:"error,omitempty"`
	ErrorMessage string                     `json:"error_message,omitempty"`
	RetryCount   int                        `json:"retry_count"`
	Metrics      map[string]interface{}     `json:"metrics,omitempty"`
	Logs         []StartupNodeLog            `json:"logs,omitempty"`
}

// StartupNodeLog 节点执行日志
type StartupNodeLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                `json:"level"`     // debug, info, warn, error
	Message   string                `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// StartupWorkflowTemplate 启动工作流模板
type StartupWorkflowTemplate struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Version     string              `json:"version"`
	Category    string              `json:"category"`
	Workflow    *StartupWorkflow    `json:"workflow"`
	Preview     string              `json:"preview"`     // 预览图URL
	Author      string              `json:"author"`
	Official    bool                `json:"official"`    // 是否为官方模板
	Downloads   int                 `json:"downloads"`
	Rating      float64             `json:"rating"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Tags        []string            `json:"tags"`
}

// StartupEvent 启动工作流事件
type StartupEvent struct {
	ID          string                 `json:"id"`
	EventType   StartupEventType       `json:"event_type"`
	ExecutionID string                 `json:"execution_id"`
	NodeID      string                 `json:"node_id,omitempty"`
	WorkflowID  string                 `json:"workflow_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	Source      string                 `json:"source"`      // event source
	Level       string                 `json:"level"`       // debug, info, warn, error
}

// StartupEventType 启动事件类型
type StartupEventType string

const (
	EventTypeExecutionStart    StartupEventType = "execution_start"
	EventTypeExecutionEnd      StartupEventType = "execution_end"
	EventTypeNodeStart        StartupEventType = "node_start"
	EventTypeNodeProgress     StartupEventType = "node_progress"
	EventTypeNodeComplete     StartupEventType = "node_complete"
	EventTypeNodeError        StartupEventType = "node_error"
	EventTypeNodeRetry        StartupEventType = "node_retry"
	EventTypeExecutionPause   StartupEventType = "execution_pause"
	EventTypeExecutionResume  StartupEventType = "execution_resume"
	EventTypeExecutionCancel  StartupEventType = "execution_cancel"
)

// StartupWorkflowManager 启动工作流管理器接口
type StartupWorkflowManager interface {
	// 工作流管理
	CreateWorkflow(ctx context.Context, workflow *StartupWorkflow) (*StartupWorkflow, error)
	GetWorkflow(ctx context.Context, id string) (*StartupWorkflow, error)
	UpdateWorkflow(ctx context.Context, workflow *StartupWorkflow) (*StartupWorkflow, error)
	DeleteWorkflow(ctx context.Context, id string) error
	ListWorkflows(ctx context.Context) ([]*StartupWorkflow, error)
	ValidateWorkflow(ctx context.Context, workflow *StartupWorkflow) error

	// 执行管理
	ExecuteWorkflow(ctx context.Context, workflowID string, inputs map[string]interface{}) (*StartupExecution, error)
	ExecuteWorkflowWithConfig(ctx context.Context, workflowID string, inputs map[string]interface{}, config *StartupWorkflowConfig) (*StartupExecution, error)
	GetExecution(ctx context.Context, executionID string) (*StartupExecution, error)
	ListExecutions(ctx context.Context, workflowID string) ([]*StartupExecution, error)
	CancelExecution(ctx context.Context, executionID string) error
	PauseExecution(ctx context.Context, executionID string) error
	ResumeExecution(ctx context.Context, executionID string) error

	// 模板管理
	CreateTemplate(ctx context.Context, template *StartupWorkflowTemplate) (*StartupWorkflowTemplate, error)
	GetTemplate(ctx context.Context, id string) (*StartupWorkflowTemplate, error)
	UpdateTemplate(ctx context.Context, template *StartupWorkflowTemplate) (*StartupWorkflowTemplate, error)
	DeleteTemplate(ctx context.Context, id string) error
	ListTemplates(ctx context.Context) ([]*StartupWorkflowTemplate, error)
	DeployFromTemplate(ctx context.Context, templateID string, name string) (*StartupWorkflow, error)

	// 系统管理
	GetSystemStatus(ctx context.Context) (*StartupSystemStatus, error)
	GetMetrics(ctx context.Context) (*StartupMetrics, error)
}

// StartupSystemStatus 启动系统状态
type StartupSystemStatus struct {
	IsRunning        bool                  `json:"is_running"`
	CurrentExecution *StartupExecution     `json:"current_execution,omitempty"`
	TotalExecutions   int64                 `json:"total_executions"`
	SuccessfulRuns   int64                 `json:"successful_runs"`
	FailedRuns       int64                 `json:"failed_runs"`
	LastExecution    *StartupExecution     `json:"last_execution,omitempty"`
	StartTime        time.Time             `json:"start_time"`
	Version          string                `json:"version"`
	Components       map[string]ComponentStatus `json:"components"`
}

// ComponentStatus 组件状态
type ComponentStatus struct {
	Status      string    `json:"status"`      // healthy, unhealthy, unknown
	LastCheck   time.Time `json:"last_check"`
	Message     string    `json:"message"`
	Details     map[string]string `json:"details"`
}

// StartupMetrics 启动指标
type StartupMetrics struct {
	TotalExecutions      int64         `json:"total_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions    int64         `json:"failed_executions"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	MinExecutionTime     time.Duration `json:"min_execution_time"`
	MaxExecutionTime     time.Duration `json:"max_execution_time"`
	NodeExecutionCounts   map[string]int64 `json:"node_execution_counts"`
	ErrorCounts          map[string]int64 `json:"error_counts"`
	RecentExecutions     []*StartupExecution `json:"recent_executions"`
	CalculatedAt         time.Time       `json:"calculated_at"`
}

// StartupEventHandler 启动事件处理器
type StartupEventHandler interface {
	OnExecutionStart(ctx context.Context, execution *StartupExecution) error
	OnExecutionEnd(ctx context.Context, execution *StartupExecution) error
	OnNodeStart(ctx context.Context, execution *StartupExecution, node *StartupNode) error
	OnNodeProgress(ctx context.Context, execution *StartupExecution, node *StartupNode, progress float64) error
	OnNodeComplete(ctx context.Context, execution *StartupExecution, node *StartupNode, result *StartupNodeResult) error
	OnNodeError(ctx context.Context, execution *StartupExecution, node *StartupNode, err error) error
}

// StartupNodeExecutor 启动节点执行器接口
type StartupNodeExecutor interface {
	Execute(ctx context.Context, node *StartupNode, inputs map[string]interface{}, context map[string]interface{}) (*StartupNodeResult, error)
	Validate(ctx context.Context, node *StartupNode) error
	GetNodeInfo() *StartupNodeInfo
	Cleanup(ctx context.Context) error
}

// StartupNodeInfo 节点信息
type StartupNodeInfo struct {
	Type        StartupNodeType     `json:"type"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Version     string              `json:"version"`
	Author      string              `json:"author"`
	SupportedConfig map[string]interface{} `json:"supported_config"`
	Capabilities []string            `json:"capabilities"`
}

// StartupPluginManager 启动插件管理器
type StartupPluginManager interface {
	RegisterExecutor(nodeType StartupNodeType, executor StartupNodeExecutor) error
	GetExecutor(nodeType StartupNodeType) (StartupNodeExecutor, bool)
	ListExecutors() map[StartupNodeType]StartupNodeInfo
	UnregisterExecutor(nodeType StartupNodeType) error
}

// 默认配置常量
const (
	DefaultTimeout            = 5 * time.Minute
	DefaultMaxRetries         = 3
	DefaultParallelLimit      = 5
	DefaultRetryDelay         = 1 * time.Second
	DefaultMaxRetryDelay      = 30 * time.Second
	DefaultWebSocketHeartbeat = 30 * time.Second
	DefaultEventBufferSize     = 1000
)

// 错误定义
var (
	ErrWorkflowNotFound     = NewStartupError("WORKFLOW_NOT_FOUND", "workflow not found")
	ErrExecutionNotFound    = NewStartupError("EXECUTION_NOT_FOUND", "execution not found")
	ErrNodeExecutionFailed  = NewStartupError("NODE_EXECUTION_FAILED", "node execution failed")
	ErrWorkflowValidation   = NewStartupError("WORKFLOW_VALIDATION", "workflow validation failed")
	ErrExecutionTimeout     = NewStartupError("EXECUTION_TIMEOUT", "execution timeout")
	ErrExecutionCancelled   = NewStartupError("EXECUTION_CANCELLED", "execution cancelled")
	ErrInvalidConfig         = NewStartupError("INVALID_CONFIG", "invalid configuration")
	ErrNodeNotFound          = NewStartupError("NODE_NOT_FOUND", "node not found")
	ErrCircularDependency    = NewStartupError("CIRCULAR_DEPENDENCY", "circular dependency detected")
)

// StartupError 启动工作流错误
type StartupError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *StartupError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewStartupError 创建启动工作流错误
func NewStartupError(code, message string) *StartupError {
	return &StartupError{
		Code:    code,
		Message: message,
	}
}

// WithDetails 添加错误详情
func (e *StartupError) WithDetails(details string) *StartupError {
	e.Details = details
	return e
}

// IsCode 检查错误代码
func (e *StartupError) IsCode(code string) bool {
	return e.Code == code
}