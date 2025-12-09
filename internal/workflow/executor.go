package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/plugin/capability"
)

// WorkflowExecutorImpl 工作流执行器实现
type WorkflowExecutorImpl struct {
	config        *config.Config
	registry      *capability.Registry
	dagEngine     DAGEngine
	dataFlow      DataFlow
	logger        Logger

	// 运行时状态
	executions    map[string]*Execution
	executionMu   sync.RWMutex
	cancelFuncs   map[string]context.CancelFunc
	cancelFuncsMu sync.RWMutex
}

// NewWorkflowExecutor 创建工作流执行器
func NewWorkflowExecutor(config *config.Config, registry *capability.Registry, dagEngine DAGEngine, dataFlow DataFlow, logger Logger) WorkflowExecutor {
	return &WorkflowExecutorImpl{
		config:        config,
		registry:      registry,
		dagEngine:     dagEngine,
		dataFlow:      dataFlow,
		logger:        logger,
		executions:    make(map[string]*Execution),
		cancelFuncs:   make(map[string]context.CancelFunc),
	}
}

// Execute 执行工作流
func (e *WorkflowExecutorImpl) Execute(ctx context.Context, workflow *Workflow, inputs map[string]interface{}) (*Execution, error) {
	// 验证工作流
	if err := e.dagEngine.ValidateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// 创建执行实例
	execution := &Execution{
		ID:          e.generateExecutionID(),
		WorkflowID:  workflow.ID,
		Status:      ExecutionStatusPending,
		StartTime:   time.Now(),
		Context:     make(map[string]interface{}),
		NodeResults: make(map[string]*NodeResult),
		Inputs:      inputs,
		Outputs:     make(map[string]interface{}),
		Logs:        make([]ExecutionLog, 0),
	}

	// 保存执行实例
	e.executionMu.Lock()
	e.executions[execution.ID] = execution
	e.executionMu.Unlock()

	// 创建可取消的上下文
	execCtx, cancel := context.WithCancel(ctx)
	e.cancelFuncsMu.Lock()
	e.cancelFuncs[execution.ID] = cancel
	e.cancelFuncsMu.Unlock()

	// 启动执行
	go e.executeWorkflow(execCtx, workflow, execution)

	e.logger.Info("Workflow execution started", "execution_id", execution.ID, "workflow_id", workflow.ID)

	return execution, nil
}

// executeWorkflow 执行工作流的具体逻辑
func (e *WorkflowExecutorImpl) executeWorkflow(ctx context.Context, workflow *Workflow, execution *Execution) {
	defer func() {
		if r := recover(); r != nil {
			e.logger.Error("Workflow execution panic", "execution_id", execution.ID, "panic", r)
			e.markExecutionFailed(execution, fmt.Sprintf("Execution panic: %v", r))
		}
	}()

	// 设置执行状态
	execution.Status = ExecutionStatusRunning
	e.addLog(execution, "info", "", "Workflow execution started")

	// 执行超时控制
	timeoutCtx := ctx
	if workflow.Config.Timeout > 0 {
		var cancel context.CancelFunc
		timeoutCtx, cancel = context.WithTimeout(ctx, workflow.Config.Timeout)
		defer cancel()
	}

	// 拓扑排序获取执行顺序
	_, err := e.dagEngine.TopologicalSort(workflow.Nodes, workflow.Edges)
	if err != nil {
		e.markExecutionFailed(execution, fmt.Sprintf("Topological sort failed: %w", err))
		return
	}

	// 执行节点
	for {
		select {
		case <-timeoutCtx.Done():
			e.markExecutionFailed(execution, "Execution timeout")
			return
		default:
			// 获取可执行节点
			executableNodes, err := e.dagEngine.GetExecutableNodes(execution, workflow)
			if err != nil {
				e.markExecutionFailed(execution, fmt.Sprintf("Failed to get executable nodes: %w", err))
				return
			}

			if len(executableNodes) == 0 {
				// 没有更多可执行节点，检查是否完成
				if e.isExecutionCompleted(workflow, execution) {
					e.markExecutionCompleted(execution)
					return
				}
				// 等待一段时间后重试
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 并行执行节点（考虑并行限制）
			e.executeNodes(ctx, workflow, execution, executableNodes)
		}
	}
}

// executeNodes 执行节点
func (e *WorkflowExecutorImpl) executeNodes(ctx context.Context, workflow *Workflow, execution *Execution, nodeIDs []string) {
	// 限制并行执行数量
	parallelLimit := workflow.Config.ParallelLimit
	if parallelLimit <= 0 {
		parallelLimit = 5 // 默认限制
	}

	if len(nodeIDs) > parallelLimit {
		// 分批执行
		for i := 0; i < len(nodeIDs); i += parallelLimit {
			end := i + parallelLimit
			if end > len(nodeIDs) {
				end = len(nodeIDs)
			}

			batch := nodeIDs[i:end]
			e.executeNodesBatch(ctx, workflow, execution, batch)
		}
	} else {
		e.executeNodesBatch(ctx, workflow, execution, nodeIDs)
	}
}

// executeNodesBatch 批量执行节点
func (e *WorkflowExecutorImpl) executeNodesBatch(ctx context.Context, workflow *Workflow, execution *Execution, nodeIDs []string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, len(nodeIDs))

	for _, nodeID := range nodeIDs {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(id string) {
			defer func() {
				<-semaphore // 释放信号量
				wg.Done()
			}()

			e.executeSingleNode(ctx, workflow, execution, id)
		}(nodeID)
	}

	wg.Wait()
}

// executeSingleNode 执行单个节点
func (e *WorkflowExecutorImpl) executeSingleNode(ctx context.Context, workflow *Workflow, execution *Execution, nodeID string) {
	// 获取节点定义
	var node *Node
	for i := range workflow.Nodes {
		if workflow.Nodes[i].ID == nodeID {
			node = &workflow.Nodes[i]
			break
		}
	}

	if node == nil {
		e.markNodeFailed(execution, nodeID, "Node not found")
		return
	}

	e.addLog(execution, "info", nodeID, fmt.Sprintf("Starting node execution: %s", node.Name))

	// 创建节点结果
	result := &NodeResult{
		NodeID:    nodeID,
		Status:    NodeStatusRunning,
		StartTime: time.Now(),
		Inputs:    make(map[string]interface{}),
		Outputs:   make(map[string]interface{}),
	}

	execution.NodeResults[nodeID] = result

	// 根据节点类型执行
	switch node.Type {
	case NodeTypeStart:
		e.executeStartNode(ctx, workflow, execution, node, result)
	case NodeTypeEnd:
		e.executeEndNode(ctx, workflow, execution, node, result)
	case NodeTypeTask:
		e.executeTaskNode(ctx, workflow, execution, node, result)
	case NodeTypeCondition:
		e.executeConditionNode(ctx, workflow, execution, node, result)
	case NodeTypeParallel:
		e.executeParallelNode(ctx, workflow, execution, node, result)
	case NodeTypeMerge:
		e.executeMergeNode(ctx, workflow, execution, node, result)
	default:
		e.markNodeFailed(execution, nodeID, fmt.Sprintf("Unknown node type: %s", node.Type))
	}
}

// executeStartNode 执行开始节点
func (e *WorkflowExecutorImpl) executeStartNode(ctx context.Context, workflow *Workflow, execution *Execution, node *Node, result *NodeResult) {
	// 开始节点通常只是传递输入数据
	result.Inputs = execution.Inputs
	result.Outputs = make(map[string]interface{})

	// 传递所有输入数据到输出
	for key, value := range execution.Inputs {
		result.Outputs[key] = value
	}

	// 合并全局变量
	for key, value := range workflow.Config.Variables {
		result.Outputs[fmt.Sprintf("global.%s", key)] = value
	}

	e.markNodeCompleted(execution, result)
}

// executeEndNode 执行结束节点
func (e *WorkflowExecutorImpl) executeEndNode(ctx context.Context, workflow *Workflow, execution *Execution, node *Node, result *NodeResult) {
	// 结束节点收集所有前置节点的输出
	dependencies := e.dagEngine.GetNodeDependencies(node.ID, workflow.Edges)

	outputs := make(map[string]interface{})

	for _, depID := range dependencies {
		if depResult, exists := execution.NodeResults[depID]; exists && depResult.Status == NodeStatusCompleted {
			for key, value := range depResult.Outputs {
				outputs[fmt.Sprintf("%s.%s", depID, key)] = value
			}
		}
	}

	// 设置工作流最终输出
	execution.Outputs = outputs
	result.Outputs = outputs

	e.markNodeCompleted(execution, result)
}

// executeTaskNode 执行任务节点
func (e *WorkflowExecutorImpl) executeTaskNode(ctx context.Context, workflow *Workflow, execution *Execution, node *Node, result *NodeResult) {
	// 获取节点输入数据
	inputs, err := e.dataFlow.GetNodeInputs(execution, node, workflow)
	if err != nil {
		e.markNodeFailed(execution, node.ID, fmt.Sprintf("Failed to get inputs: %w", err))
		return
	}

	result.Inputs = inputs

	// 调用插件
	// 假设 node.Plugin 存储的是 capabilityID (例如 "openai_chat")
	// 如果 node.Plugin 为空，尝试使用 node.Type 或其他元数据
	capabilityID := node.Plugin
	if capabilityID == "" {
		e.markNodeFailed(execution, node.ID, "No plugin/capability specified")
		return
	}

	executor, err := e.registry.GetExecutor(capabilityID)
	if err != nil {
		e.markNodeFailed(execution, node.ID, fmt.Sprintf("Failed to get executor for capability %s: %v", capabilityID, err))
		return
	}
	// 准备配置
	// 这里的 node.Config 是 map[string]interface{}，直接传递给 Executor
	config := node.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	// 合并全局配置
	config = e.mergeGlobalConfig(capabilityID, config)

	pluginOutputs, err := executor.Execute(ctx, config, inputs)
	if err != nil {
		e.markNodeFailed(execution, node.ID, fmt.Sprintf("Plugin execution failed: %w", err))
		return
	}

	// 处理插件输出
	result.Outputs = pluginOutputs

	// 验证输出Schema
	if err := e.validateNodeOutputs(node, result.Outputs); err != nil {
		e.markNodeFailed(execution, node.ID, fmt.Sprintf("Output validation failed: %w", err))
		return
	}

	e.markNodeCompleted(execution, result)
}

// executeConditionNode 执行条件节点
func (e *WorkflowExecutorImpl) executeConditionNode(ctx context.Context, workflow *Workflow, execution *Execution, node *Node, result *NodeResult) {
	// 获取条件输入
	inputs, err := e.dataFlow.GetNodeInputs(execution, node, workflow)
	if err != nil {
		e.markNodeFailed(execution, node.ID, fmt.Sprintf("Failed to get inputs: %w", err))
		return
	}

	result.Inputs = inputs

	// 简单的条件判断逻辑
	condition, ok := inputs["condition"].(string)
	if !ok {
		e.markNodeFailed(execution, node.ID, "Condition not found or invalid")
		return
	}

	// 评估条件
	result.Outputs = map[string]interface{}{
		"condition": condition,
		"result":    e.evaluateCondition(condition, inputs),
	}

	e.markNodeCompleted(execution, result)
}

// executeParallelNode 执行并行节点
func (e *WorkflowExecutorImpl) executeParallelNode(ctx context.Context, workflow *Workflow, execution *Execution, node *Node, result *NodeResult) {
	// 并行节点主要用于标记并行执行的开始
	// 具体的并行执行逻辑在拓扑排序中处理
	inputs, err := e.dataFlow.GetNodeInputs(execution, node, workflow)
	if err != nil {
		e.markNodeFailed(execution, node.ID, fmt.Sprintf("Failed to get inputs: %w", err))
		return
	}

	result.Inputs = inputs
	result.Outputs = map[string]interface{}{
		"parallel": true,
		"inputs":   inputs,
	}

	e.markNodeCompleted(execution, result)
}

// executeMergeNode 执行合并节点
func (e *WorkflowExecutorImpl) executeMergeNode(ctx context.Context, workflow *Workflow, execution *Execution, node *Node, result *NodeResult) {
	// 获取所有前置并行节点的输出
	dependencies := e.dagEngine.GetNodeDependencies(node.ID, workflow.Edges)

	mergedData, err := e.dataFlow.MergeParallelData(execution, dependencies)
	if err != nil {
		e.markNodeFailed(execution, node.ID, fmt.Sprintf("Failed to merge parallel data: %w", err))
		return
	}

	result.Inputs = map[string]interface{}{
		"dependencies": dependencies,
	}
	result.Outputs = mergedData

	e.markNodeCompleted(execution, result)
}

// evaluateCondition 评估条件
func (e *WorkflowExecutorImpl) evaluateCondition(condition string, inputs map[string]interface{}) bool {
	// 简单的条件评估实现
	// 在实际应用中应该使用更复杂的表达式求值器

	switch condition {
	case "true", "True", "TRUE":
		return true
	case "false", "False", "FALSE":
		return false
	default:
		// 尝试从输入中查找条件值
		if value, exists := inputs["value"]; exists {
			if boolValue, ok := value.(bool); ok {
				return boolValue
			}
		}
		return false
	}
}

// validateNodeOutputs 验证节点输出
func (e *WorkflowExecutorImpl) validateNodeOutputs(node *Node, outputs map[string]interface{}) error {
	// 验证输出Schema
	for _, outputSchema := range node.Outputs {
		if _, exists := outputs[outputSchema.Name]; !exists {
			return fmt.Errorf("required output %s is missing", outputSchema.Name)
		}
	}

	return nil
}

// markNodeCompleted 标记节点完成
func (e *WorkflowExecutorImpl) markNodeCompleted(execution *Execution, result *NodeResult) {
	result.Status = NodeStatusCompleted
	endTime := time.Now()
	result.EndTime = &endTime
	result.ElapsedTime = endTime.Sub(result.StartTime)

	e.addLog(execution, "info", result.NodeID, fmt.Sprintf("Node completed in %v", result.ElapsedTime))
}

// markNodeFailed 标记节点失败
func (e *WorkflowExecutorImpl) markNodeFailed(execution *Execution, nodeID, errorMsg string) {
	if result, exists := execution.NodeResults[nodeID]; exists {
		result.Status = NodeStatusFailed
		result.Error = errorMsg
		endTime := time.Now()
		result.EndTime = &endTime
		if !result.StartTime.IsZero() {
			result.ElapsedTime = endTime.Sub(result.StartTime)
		}
	}

	e.addLog(execution, "error", nodeID, errorMsg)
}

// markExecutionCompleted 标记执行完成
func (e *WorkflowExecutorImpl) markExecutionCompleted(execution *Execution) {
	execution.Status = ExecutionStatusCompleted
	endTime := time.Now()
	execution.EndTime = &endTime

	e.addLog(execution, "info", "", "Workflow execution completed")
	e.logger.Info("Workflow execution completed", "execution_id", execution.ID, "duration", endTime.Sub(execution.StartTime))
}

// markExecutionFailed 标记执行失败
func (e *WorkflowExecutorImpl) markExecutionFailed(execution *Execution, errorMsg string) {
	execution.Status = ExecutionStatusFailed
	execution.Error = errorMsg
	endTime := time.Now()
	execution.EndTime = &endTime

	e.addLog(execution, "error", "", errorMsg)
	e.logger.Error("Workflow execution failed", "execution_id", execution.ID, "error", errorMsg)
}

// isExecutionCompleted 检查执行是否完成
func (e *WorkflowExecutorImpl) isExecutionCompleted(workflow *Workflow, execution *Execution) bool {
	// 检查所有节点是否已完成
	for _, node := range workflow.Nodes {
		if result, exists := execution.NodeResults[node.ID]; exists {
			if result.Status != NodeStatusCompleted {
				return false
			}
		} else {
			return false
		}
	}

	// 至少有一个结束节点已完成
	for _, node := range workflow.Nodes {
		if node.Type == NodeTypeEnd {
			if result, exists := execution.NodeResults[node.ID]; exists && result.Status == NodeStatusCompleted {
				return true
			}
		}
	}

	return false
}

// addLog 添加执行日志
func (e *WorkflowExecutorImpl) addLog(execution *Execution, level, nodeID, message string) {
	log := ExecutionLog{
		Timestamp: time.Now(),
		Level:     level,
		NodeID:    nodeID,
		Message:   message,
	}

	execution.Logs = append(execution.Logs, log)
}

// generateExecutionID 生成执行ID
func (e *WorkflowExecutorImpl) generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

// Cancel 取消执行
func (e *WorkflowExecutorImpl) Cancel(executionID string) error {
	e.cancelFuncsMu.RLock()
	cancel, exists := e.cancelFuncs[executionID]
	e.cancelFuncsMu.RUnlock()

	if !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}

	// 调用取消函数
	cancel()

	// 更新执行状态
	e.executionMu.Lock()
	if execution, exists := e.executions[executionID]; exists {
		execution.Status = ExecutionStatusCancelled
		endTime := time.Now()
		execution.EndTime = &endTime
		execution.Error = "Execution cancelled by user"
	}
	e.executionMu.Unlock()

	// 清理取消函数
	e.cancelFuncsMu.Lock()
	delete(e.cancelFuncs, executionID)
	e.cancelFuncsMu.Unlock()

	e.logger.Info("Execution cancelled", "execution_id", executionID)

	return nil
}

// GetExecution 获取执行状态
func (e *WorkflowExecutorImpl) GetExecution(executionID string) (*Execution, bool) {
	e.executionMu.RLock()
	defer e.executionMu.RUnlock()

	execution, exists := e.executions[executionID]
	if !exists {
		return nil, false
	}

	// 返回副本以避免并发问题
	executionCopy := *execution
	return &executionCopy, true
}

// GetExecutionLogs 获取执行日志
func (e *WorkflowExecutorImpl) GetExecutionLogs(executionID string) ([]ExecutionLog, error) {
	e.executionMu.RLock()
	defer e.executionMu.RUnlock()

	execution, exists := e.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	// 返回日志副本
	logs := make([]ExecutionLog, len(execution.Logs))
	copy(logs, execution.Logs)

	return logs, nil
}

// mergeGlobalConfig 合并全局配置到节点配置
func (e *WorkflowExecutorImpl) mergeGlobalConfig(capabilityID string, nodeConfig map[string]interface{}) map[string]interface{} {
	if e.config == nil {
		return nodeConfig
	}

	// 复制节点配置
	newConfig := make(map[string]interface{})
	for k, v := range nodeConfig {
		newConfig[k] = v
	}

	// 根据 capabilityID 注入配置
	// 目前主要处理 LLM 配置
	if capabilityID == "openai_chat" {
		if llmConfig, ok := e.config.LLM["openai"]; ok {
			if _, exists := newConfig["api_key"]; !exists || newConfig["api_key"] == "" {
				newConfig["api_key"] = llmConfig.APIKey
			}
			if _, exists := newConfig["base_url"]; !exists || newConfig["base_url"] == "" {
				newConfig["base_url"] = llmConfig.BaseURL
			}
			if _, exists := newConfig["model"]; !exists || newConfig["model"] == "" {
				newConfig["model"] = llmConfig.ModelName
			}
		}
	}

	// TODO: 处理其他类型的配置注入

	return newConfig
}
