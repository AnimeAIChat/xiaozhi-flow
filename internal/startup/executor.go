package startup

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"xiaozhi-server-go/internal/workflow"
)

// StartupWorkflowExecutor 启动工作流执行器
type StartupWorkflowExecutor struct {
	manager         StartupWorkflowManager
	pluginManager   StartupPluginManager
	eventHandler    StartupEventHandler
	workflowEngine  *workflow.WorkflowExecutor
	logger          StartupLogger

	// 执行状态
	executions      map[string]*StartupWorkflowExecution
	executionsMutex sync.RWMutex

	// 配置
	config          ExecutorConfig
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	DefaultTimeout       time.Duration `json:"default_timeout"`
	MaxConcurrentExecutions int        `json:"max_concurrent_executions"`
	EnableEventLogging    bool        `json:"enable_event_logging"`
	EnableMetrics         bool        `json:"enable_metrics"`
	MetricsInterval       time.Duration `json:"metrics_interval"`
}

// StartupWorkflowExecution 启动工作流执行实例
type StartupWorkflowExecution struct {
	ID             string                        `json:"id"`
	Workflow       *StartupWorkflow              `json:"workflow"`
	Execution      *workflow.Execution           `json:"execution"`
	Context        map[string]interface{}        `json:"context"`
	StartTime      time.Time                     `json:"start_time"`
	EndTime        *time.Time                    `json:"end_time,omitempty"`
	Status         WorkflowExecutionStatus       `json:"status"`
	Error          string                        `json:"error,omitempty"`
	Progress       float64                       `json:"progress"`
	CompletedNodes int                           `json:"completed_nodes"`
	TotalNodes     int                           `json:"total_nodes"`
	NodeExecutions map[string]*NodeExecution     `json:"node_executions"`
	Mutex          sync.RWMutex                  `json:"-"`
}

// NodeExecution 节点执行状态
type NodeExecution struct {
	NodeID      string                    `json:"node_id"`
	NodeName    string                    `json:"node_name"`
	NodeType    StartupNodeType           `json:"node_type"`
	Status      workflow.NodeStatus       `json:"status"`
	StartTime   time.Time                 `json:"start_time"`
	EndTime     *time.Time                `json:"end_time,omitempty"`
	Duration    time.Duration             `json:"duration"`
	Inputs      map[string]interface{}    `json:"inputs"`
	Outputs     map[string]interface{}    `json:"outputs"`
	Error       string                    `json:"error,omitempty"`
	RetryCount  int                       `json:"retry_count"`
	Progress    float64                   `json:"progress"`
	Logs        []NodeLog                 `json:"logs"`
	Mutex       sync.RWMutex              `json:"-"`
}

// NodeLog 节点执行日志
type NodeLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                `json:"level"`
	Message   string                `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// WorkflowExecutionStatus 工作流执行状态
type WorkflowExecutionStatus string

const (
	WorkflowStatusPending    WorkflowExecutionStatus = "pending"
	WorkflowStatusRunning    WorkflowExecutionStatus = "running"
	WorkflowStatusPaused     WorkflowExecutionStatus = "paused"
	WorkflowStatusCompleted  WorkflowExecutionStatus = "completed"
	WorkflowStatusFailed     WorkflowExecutionStatus = "failed"
	WorkflowStatusCancelled  WorkflowExecutionStatus = "cancelled"
)

// StartupLogger 启动日志接口
type StartupLogger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// NewStartupWorkflowExecutor 创建启动工作流执行器
func NewStartupWorkflowExecutor(
	manager StartupWorkflowManager,
	pluginManager StartupPluginManager,
	eventHandler StartupEventHandler,
	logger StartupLogger,
) *StartupWorkflowExecutor {
	config := ExecutorConfig{
		DefaultTimeout:            DefaultTimeout,
		MaxConcurrentExecutions:   5,
		EnableEventLogging:        true,
		EnableMetrics:            true,
		MetricsInterval:          30 * time.Second,
	}

	return &StartupWorkflowExecutor{
		manager:        manager,
		pluginManager:  pluginManager,
		eventHandler:   eventHandler,
		logger:         logger,
		executions:     make(map[string]*StartupWorkflowExecution),
		config:         config,
	}
}

// ExecuteWorkflow 执行启动工作流
func (s *StartupWorkflowExecutor) ExecuteWorkflow(
	ctx context.Context,
	workflowID string,
	inputs map[string]interface{},
) (*StartupWorkflowExecution, error) {
	// 获取工作流定义
	workflowDef, err := s.manager.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// 检查并发执行限制
	s.executionsMutex.RLock()
	if len(s.executions) >= s.config.MaxConcurrentExecutions {
		s.executionsMutex.RUnlock()
		return nil, NewStartupError("MAX_CONCURRENT_EXECUTIONS",
			fmt.Sprintf("maximum concurrent executions (%d) reached", s.config.MaxConcurrentExecutions))
	}
	s.executionsMutex.RUnlock()

	// 创建执行实例
	execution := &StartupWorkflowExecution{
		ID:             generateExecutionID(),
		Workflow:       workflowDef,
		Context:        make(map[string]interface{}),
		StartTime:      time.Now(),
		Status:         WorkflowStatusPending,
		Progress:       0.0,
		CompletedNodes: 0,
		TotalNodes:     len(workflowDef.Nodes),
		NodeExecutions: make(map[string]*NodeExecution),
	}

	// 初始化上下文
	if inputs != nil {
		for k, v := range inputs {
			execution.Context[k] = v
		}
	}

	// 添加到执行列表
	s.executionsMutex.Lock()
	s.executions[execution.ID] = execution
	s.executionsMutex.Unlock()

	// 发送执行开始事件
	if s.eventHandler != nil {
		s.eventHandler.OnExecutionStart(ctx, execution)
	}

	// 启动执行
	go s.executeWorkflowAsync(ctx, execution)

	return execution, nil
}

// executeWorkflowAsync 异步执行工作流
func (s *StartupWorkflowExecutor) executeWorkflowAsync(
	ctx context.Context,
	execution *StartupWorkflowExecution,
) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Workflow execution panic", "execution_id", execution.ID, "panic", r)
			s.markExecutionFailed(execution, fmt.Sprintf("panic: %v", r))
		}
	}()

	// 更新状态为运行中
	s.updateExecutionStatus(execution, WorkflowStatusRunning, "")

	// 获取执行超时配置
	timeout := execution.Workflow.Config.Timeout
	if timeout == 0 {
		timeout = s.config.DefaultTimeout
	}

	// 创建带超时的上下文
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 执行工作流
	err := s.executeNodes(execCtx, execution)

	// 记录结束时间
	now := time.Now()
	execution.EndTime = &now

	if err != nil {
		s.logger.Error("Workflow execution failed",
			"execution_id", execution.ID,
			"error", err)
		s.markExecutionFailed(execution, err.Error())
	} else {
		s.markExecutionCompleted(execution)
	}

	// 发送执行结束事件
	if s.eventHandler != nil {
		s.eventHandler.OnExecutionEnd(ctx, execution)
	}

	// 清理执行实例（延迟一段时间以支持状态查询）
	go func() {
		time.Sleep(5 * time.Minute)
		s.executionsMutex.Lock()
		delete(s.executions, execution.ID)
		s.executionsMutex.Unlock()
	}()
}

// executeNodes 执行所有节点
func (s *StartupWorkflowExecutor) executeNodes(
	ctx context.Context,
	execution *StartupWorkflowExecution,
) error {
	// 拓扑排序获取执行顺序
	workflowDef := execution.Workflow
	executionOrder, err := s.topologicalSort(workflowDef)
	if err != nil {
		return fmt.Errorf("failed to sort workflow nodes: %w", err)
	}

	// 按层次执行节点
	for level, nodes := range executionOrder {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 并行执行同一层级的节点
		err = s.executeNodesAtLevel(ctx, execution, nodes, level)
		if err != nil {
			return err
		}
	}

	return nil
}

// topologicalSort 拓扑排序获取执行层次
func (s *StartupWorkflowExecutor) topologicalSort(
	workflow *StartupWorkflow,
) ([][]string, error) {
	// 构建节点映射和依赖图
	nodeMap := make(map[string]*StartupNode)
	for _, node := range workflow.Nodes {
		nodeMap[node.ID] = &node
	}

	// 计算每个节点的入度
	inDegree := make(map[string]int)
	for _, node := range workflow.Nodes {
		inDegree[node.ID] = len(node.DependsOn)
	}

	// 使用Kahn算法进行拓扑排序
	var queue []string
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	var levels [][]string
	for len(queue) > 0 {
		var currentLevel []string
		var nextQueue []string

		// 处理当前层级的所有节点
		for _, nodeID := range queue {
			currentLevel = append(currentLevel, nodeID)

			// 更新依赖此节点的其他节点的入度
			node := nodeMap[nodeID]
			for _, edge := range workflow.Edges {
				if edge.From == nodeID {
					inDegree[edge.To]--
					if inDegree[edge.To] == 0 {
						nextQueue = append(nextQueue, edge.To)
					}
				}
			}
		}

		levels = append(levels, currentLevel)
		queue = nextQueue
	}

	// 检查是否还有未处理的节点（循环依赖）
	unprocessed := 0
	for _, degree := range inDegree {
		if degree > 0 {
			unprocessed++
		}
	}

	if unprocessed > 0 {
		return nil, ErrCircularDependency
	}

	return levels, nil
}

// executeNodesAtLevel 执行指定层级的节点
func (s *StartupWorkflowExecutor) executeNodesAtLevel(
	ctx context.Context,
	execution *StartupWorkflowExecution,
	nodeIDs []string,
	level int,
) error {
	// 创建等待组和错误通道
	var wg sync.WaitGroup
	errChan := make(chan error, len(nodeIDs))

	// 并行执行节点
	for _, nodeID := range nodeIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			err := s.executeNode(ctx, execution, id)
			if err != nil {
				errChan <- fmt.Errorf("node %s failed: %w", id, err)
			}
		}(nodeID)
	}

	// 等待所有节点完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		// 从错误消息中提取节点ID
		nodeID := extractNodeIDFromError(err.Error())
		// 如果是关键节点失败，返回错误
		node := execution.Workflow.GetNodeByID(nodeID)
		if node != nil && node.Critical {
			return err
		}
		// 非关键节点失败，记录日志但继续执行
		s.logger.Warn("Non-critical node failed", "node_id", nodeID, "error", err)
	}

	return nil
}

// executeNode 执行单个节点
func (s *StartupWorkflowExecutor) executeNode(
	ctx context.Context,
	execution *StartupWorkflowExecution,
	nodeID string,
) error {
	// 获取节点定义
	node := execution.Workflow.GetNodeByID(nodeID)
	if node == nil {
		return NewStartupError("NODE_NOT_FOUND", fmt.Sprintf("node %s not found", nodeID))
	}

	// 创建节点执行记录
	nodeExec := &NodeExecution{
		NodeID:   nodeID,
		NodeName: node.Name,
		NodeType: node.Type,
		Status:   workflow.NodeStatusPending,
		StartTime: time.Now(),
		Inputs:   make(map[string]interface{}),
		Outputs:  make(map[string]interface{}),
		Logs:     make([]NodeLog, 0),
	}

	execution.NodeExecutions[nodeID] = nodeExec

	// 发送节点开始事件
	if s.eventHandler != nil {
		s.eventHandler.OnNodeStart(ctx, execution, node)
	}

	// 更新节点状态
	s.updateNodeStatus(nodeExec, workflow.NodeStatusRunning, "")

	// 获取节点执行器
	executor, exists := s.pluginManager.GetExecutor(node.Type)
	if !exists {
		return NewStartupError("EXECUTOR_NOT_FOUND",
			fmt.Sprintf("no executor found for node type %s", node.Type))
	}

	// 准备执行输入
	inputs := s.prepareNodeInputs(ctx, execution, node)

	// 执行节点
	result, err := executor.Execute(ctx, node, inputs, execution.Context)

	// 记录结束时间
	endTime := time.Now()
	nodeExec.EndTime = &endTime
	nodeExec.Duration = endTime.Sub(nodeExec.StartTime)

	if err != nil {
		s.updateNodeStatus(nodeExec, workflow.NodeStatusFailed, err.Error())

		// 发送节点错误事件
		if s.eventHandler != nil {
			s.eventHandler.OnNodeError(ctx, execution, node, err)
		}

		// 检查是否需要重试
		if node.Retry != nil && nodeExec.RetryCount < node.Retry.MaxAttempts {
			return s.retryNode(ctx, execution, node, nodeExec)
		}

		return err
	}

	// 更新节点执行结果
	nodeExec.Outputs = result.Outputs
	nodeExec.Status = workflow.NodeStatusCompleted
	nodeExec.Progress = 100.0

	// 更新执行上下文
	if result.Outputs != nil {
		for k, v := range result.Outputs {
			execution.Context[k] = v
		}
	}

	// 更新工作流进度
	s.updateExecutionProgress(execution)

	// 发送节点完成事件
	if s.eventHandler != nil {
		s.eventHandler.OnNodeComplete(ctx, execution, node, result)
	}

	return nil
}

// prepareNodeInputs 准备节点执行输入
func (s *StartupWorkflowExecutor) prepareNodeInputs(
	ctx context.Context,
	execution *StartupWorkflowExecution,
	node *StartupNode,
) map[string]interface{} {
	inputs := make(map[string]interface{})

	// 添加全局配置
	for k, v := range execution.Workflow.Config.Variables {
		inputs[k] = v
	}

	// 添加执行上下文
	for k, v := range execution.Context {
		inputs[k] = v
	}

	// 添加节点配置
	if node.Config != nil {
		for k, v := range node.Config {
			inputs[k] = v
		}
	}

	return inputs
}

// retryNode 重试节点执行
func (s *StartupWorkflowExecutor) retryNode(
	ctx context.Context,
	execution *StartupWorkflowExecution,
	node *StartupNode,
	nodeExec *NodeExecution,
) error {
	if node.Retry == nil {
		return fmt.Errorf("retry configuration not found")
	}

	nodeExec.RetryCount++

	// 计算重试延迟
	delay := node.Retry.Delay
	if node.Retry.Backoff {
		delay = time.Duration(nodeExec.RetryCount) * node.Retry.Delay
		if delay > node.Retry.MaxDelay {
			delay = node.Retry.MaxDelay
		}
	}

	// 等待重试延迟
	select {
	case <-time.After(delay):
		// 继续重试
	case <-ctx.Done():
		return ctx.Err()
	}

	// 发送重试事件
	if s.eventHandler != nil {
		s.eventHandler.OnNodeRetry(ctx, execution, node)
	}

	// 重新执行节点
	return s.executeNode(ctx, execution, node.ID)
}

// updateExecutionStatus 更新执行状态
func (s *StartupWorkflowExecutor) updateExecutionStatus(
	execution *StartupWorkflowExecution,
	status WorkflowExecutionStatus,
	errorMsg string,
) {
	execution.Mutex.Lock()
	defer execution.Mutex.Unlock()

	execution.Status = status
	if errorMsg != "" {
		execution.Error = errorMsg
	}
}

// updateNodeStatus 更新节点状态
func (s *StartupWorkflowExecutor) updateNodeStatus(
	nodeExec *NodeExecution,
	status workflow.NodeStatus,
	errorMsg string,
) {
	nodeExec.Mutex.Lock()
	defer nodeExec.Mutex.Unlock()

	nodeExec.Status = status
	if errorMsg != "" {
		nodeExec.Error = errorMsg
	}

	// 添加日志
	nodeExec.Logs = append(nodeExec.Logs, NodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Node status changed to %s", status),
	})
}

// updateExecutionProgress 更新执行进度
func (s *StartupWorkflowExecutor) updateExecutionProgress(execution *StartupWorkflowExecution) {
	execution.Mutex.Lock()
	defer execution.Mutex.Unlock()

	completed := 0
	for _, nodeExec := range execution.NodeExecutions {
		if nodeExec.Status == workflow.NodeStatusCompleted {
			completed++
		}
	}

	execution.CompletedNodes = completed
	if execution.TotalNodes > 0 {
		execution.Progress = float64(completed) / float64(execution.TotalNodes) * 100
	}
}

// markExecutionFailed 标记执行失败
func (s *StartupWorkflowExecutor) markExecutionFailed(execution *StartupWorkflowExecution, errorMsg string) {
	s.updateExecutionStatus(execution, WorkflowStatusFailed, errorMsg)
	s.logger.Error("Workflow execution failed", "execution_id", execution.ID, "error", errorMsg)
}

// markExecutionCompleted 标记执行完成
func (s *StartupWorkflowExecutor) markExecutionCompleted(execution *StartupWorkflowExecution) {
	s.updateExecutionStatus(execution, WorkflowStatusCompleted, "")
	execution.Progress = 100.0
	s.logger.Info("Workflow execution completed", "execution_id", execution.ID)
}

// GetExecution 获取执行实例
func (s *StartupWorkflowExecutor) GetExecution(executionID string) (*StartupWorkflowExecution, bool) {
	s.executionsMutex.RLock()
	defer s.executionsMutex.RUnlock()

	execution, exists := s.executions[executionID]
	return execution, exists
}

// CancelExecution 取消执行
func (s *StartupWorkflowExecutor) CancelExecution(executionID string) error {
	s.executionsMutex.Lock()
	defer s.executionsMutex.Unlock()

	execution, exists := s.executions[executionID]
	if !exists {
		return ErrExecutionNotFound
	}

	if execution.Status == WorkflowStatusCompleted || execution.Status == WorkflowStatusFailed {
		return NewStartupError("EXECUTION_ALREADY_FINISHED", "execution already finished")
	}

	s.updateExecutionStatus(execution, WorkflowStatusCancelled, "cancelled by user")
	return nil
}

// PauseExecution 暂停执行
func (s *StartupWorkflowExecutor) PauseExecution(executionID string) error {
	s.executionsMutex.Lock()
	defer s.executionsMutex.Unlock()

	execution, exists := s.executions[executionID]
	if !exists {
		return ErrExecutionNotFound
	}

	if execution.Status != WorkflowStatusRunning {
		return NewStartupError("EXECUTION_NOT_RUNNING", "execution is not running")
	}

	s.updateExecutionStatus(execution, WorkflowStatusPaused, "paused by user")
	return nil
}

// ResumeExecution 恢复执行
func (s *StartupWorkflowExecutor) ResumeExecution(executionID string) error {
	s.executionsMutex.Lock()
	defer s.executionsMutex.Unlock()

	execution, exists := s.executions[executionID]
	if !exists {
		return ErrExecutionNotFound
	}

	if execution.Status != WorkflowStatusPaused {
		return NewStartupError("EXECUTION_NOT_PAUSED", "execution is not paused")
	}

	s.updateExecutionStatus(execution, WorkflowStatusRunning, "")

	// 恢复执行
	go s.executeWorkflowAsync(context.Background(), execution)

	return nil
}

// generateExecutionID 生成执行ID
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

// extractNodeIDFromError 从错误消息中提取节点ID
func extractNodeIDFromError(errMsg string) string {
	// 错误消息格式通常是: "node node_id failed: ..."
	// 使用简单的字符串解析提取节点ID
	const prefix = "node "
	const suffix = " failed"

	startIndex := strings.Index(errMsg, prefix)
	if startIndex == -1 {
		return ""
	}

	startIndex += len(prefix)
	endIndex := strings.Index(errMsg[startIndex:], suffix)
	if endIndex == -1 {
		return ""
	}

	return errMsg[startIndex : startIndex+endIndex]
}