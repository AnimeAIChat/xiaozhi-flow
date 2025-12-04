package v1

import "time"

// Workflow 工作流结构
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Tags        []string               `json:"tags"`
	Config      WorkflowConfig         `json:"config"`
	Nodes       []WorkflowNode         `json:"nodes"`
	Edges       []WorkflowEdge         `json:"edges"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	Timeout       int64                  `json:"timeout"`        // 超时时间（毫秒）
	MaxRetries    int                    `json:"max_retries"`    // 最大重试次数
	ParallelLimit int                    `json:"parallel_limit"` // 并行限制
	EnableLog     bool                   `json:"enable_log"`     // 是否启用日志
	Environment   map[string]interface{} `json:"environment"`    // 环境变量
	Variables     map[string]interface{} `json:"variables"`      // 全局变量
	OnFailure     string                 `json:"on_failure"`     // 失败处理策略
}

// WorkflowNode 工作流节点
type WorkflowNode struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`      // pending, running, completed, failed
	Position    Position               `json:"position"`
	StartTime   *time.Time             `json:"start_time,omitempty"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    int64                  `json:"duration,omitempty"` // 持续时间（纳秒）
	Critical    bool                   `json:"critical"`          // 是否为关键节点
	Optional    bool                   `json:"optional"`          // 是否为可选节点
	DependsOn   []string               `json:"depends_on"`       // 依赖的节点ID
	Config      map[string]interface{} `json:"config,omitempty"`  // 节点配置
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // 节点元数据
}

// WorkflowEdge 工作流边
type WorkflowEdge struct {
	ID     string `json:"id"`
	From   string `json:"from"`   // 源节点ID
	To     string `json:"to"`     // 目标节点ID
	Label  string `json:"label"`  // 边标签
	Type   string `json:"type"`   // 边类型（dependency, data, control）
}

// Position 位置信息
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}


// WorkflowListQuery 工作流列表查询参数
type WorkflowListQuery struct {
	Tags  []string `form:"tags"`
	Page  int      `form:"page,default=1"`
	Limit int      `form:"limit,default=20"`
}

