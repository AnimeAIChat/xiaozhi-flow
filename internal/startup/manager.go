package startup

import (
	"xiaozhi-server-go/internal/platform/logging"
	"context"
	"fmt"
	"sync"
	"time"

	"xiaozhi-server-go/internal/platform/storage"
	"xiaozhi-server-go/internal/workflow"
	"xiaozhi-server-go/internal/startup/model"
)

// StartupWorkflowManagerImpl 启动工作流管理器实现
type StartupWorkflowManagerImpl struct {
	// 存储
	workflowStorage model.WorkflowStorage
	executionStorage model.ExecutionStorage
	templateStorage model.TemplateStorage

	// 状态管理
	workflows     map[string]*model.StartupWorkflow
	executions    map[string]*model.StartupExecution
	templates     map[string]*model.StartupWorkflowTemplate
	pluginManager model.StartupPluginManager
	eventHandlers []model.StartupEventHandler

	// 同步锁
	workflowMu sync.RWMutex
	executionMu sync.RWMutex
	templateMu  sync.RWMutex

	// 配置
	config *ManagerConfig

	// 日志和监控
	logger *logging.Logger
	metrics *model.StartupMetrics
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	MaxExecutions    int           `json:"max_executions"`
	ExecutionTimeout time.Duration `json:"execution_timeout"`
	CleanupInterval  time.Duration `json:"cleanup_interval"`
	EnablePersistence bool          `json:"enable_persistence"`
	StorageConfig    *StorageConfig `json:"storage_config"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type     string `json:"type"`     // memory, file, database
	FilePath string `json:"file_path"`
	Database *storage.DatabaseConfig `json:"database"`
}







// NewStartupWorkflowManager 创建启动工作流管理器
func NewStartupWorkflowManager(config *ManagerConfig, logger *logging.Logger, pluginManager model.StartupPluginManager) *StartupWorkflowManagerImpl {
	mgr := &StartupWorkflowManagerImpl{
		workflows:     make(map[string]*model.StartupWorkflow),
		executions:    make(map[string]*model.StartupExecution),
		templates:     make(map[string]*model.StartupWorkflowTemplate),
		pluginManager: pluginManager,
		eventHandlers: make([]model.StartupEventHandler, 0),
		config:        config,
		logger:        logger,
		metrics:       &model.StartupMetrics{},
	}

	// 设置默认配置
	if mgr.config == nil {
		mgr.config = &ManagerConfig{
			MaxExecutions:    100,
			ExecutionTimeout: 30 * time.Minute,
			CleanupInterval:  1 * time.Hour,
			EnablePersistence: false,
		}
	}

	// 初始化存储
	mgr.initializeStorage()

	// 启动后台任务
	go mgr.startBackgroundTasks()

	return mgr
}

// initializeStorage 初始化存储
func (m *StartupWorkflowManagerImpl) initializeStorage() {
	if m.config.EnablePersistence {
		switch m.config.StorageConfig.Type {
		case "file":
			m.workflowStorage = NewFileWorkflowStorage(m.config.StorageConfig.FilePath)
			m.executionStorage = NewFileExecutionStorage(m.config.StorageConfig.FilePath)
			m.templateStorage = NewFileTemplateStorage(m.config.StorageConfig.FilePath)
		case "database":
			m.workflowStorage = NewDatabaseWorkflowStorage(m.config.StorageConfig.Database)
			m.executionStorage = NewDatabaseExecutionStorage(m.config.StorageConfig.Database)
			m.templateStorage = NewDatabaseTemplateStorage(m.config.StorageConfig.Database)
		default:
			m.logger.Warn("Unknown storage type, using memory storage", "type", m.config.StorageConfig.Type)
			m.workflowStorage = NewMemoryWorkflowStorage()
			m.executionStorage = NewMemoryExecutionStorage()
			m.templateStorage = NewMemoryTemplateStorage()
		}
	} else {
		m.workflowStorage = NewMemoryWorkflowStorage()
		m.executionStorage = NewMemoryExecutionStorage()
		m.templateStorage = NewMemoryTemplateStorage()
	}

	// 加载持久化数据
	m.loadPersistedData()
}

// loadPersistedData 加载持久化数据
func (m *StartupWorkflowManagerImpl) loadPersistedData() {
	ctx := context.Background()

	// 加载工作流
	workflows, err := m.workflowStorage.List(ctx)
	if err != nil {
		m.logger.Error("Failed to load workflows from storage", "error", err)
		return
	}

	for _, workflow := range workflows {
		m.workflows[workflow.ID] = workflow
	}

	// 加载模板
	templates, err := m.templateStorage.List(ctx)
	if err != nil {
		m.logger.Error("Failed to load templates from storage", "error", err)
		return
	}

	for _, template := range templates {
		m.templates[template.ID] = template
	}

	m.logger.Info("Loaded persisted data", "workflows", len(m.workflows), "templates", len(m.templates))
}

// startBackgroundTasks 启动后台任务
func (m *StartupWorkflowManagerImpl) startBackgroundTasks() {
	// 清理任务
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupOldExecutions()
		}
	}
}

// cleanupOldExecutions 清理旧的执行记录
func (m *StartupWorkflowManagerImpl) cleanupOldExecutions() {
	ctx := context.Background()
	cutoff := time.Now().Add(-24 * time.Hour) // 清理24小时前的记录

	err := m.executionStorage.Cleanup(ctx, cutoff)
	if err != nil {
		m.logger.Error("Failed to cleanup old executions", "error", err)
	} else {
		m.logger.Debug("Cleaned up old executions", "cutoff", cutoff)
	}
}

// 创建工作流相关方法

func (m *StartupWorkflowManagerImpl) CreateWorkflow(ctx context.Context, workflow *model.StartupWorkflow) (*model.StartupWorkflow, error) {
	// 验证工作流
	if err := m.ValidateWorkflow(ctx, workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// 生成ID和时间戳
	if workflow.ID == "" {
		workflow.ID = fmt.Sprintf("workflow_%d", time.Now().UnixNano())
	}
	if workflow.CreatedAt.IsZero() {
		workflow.CreatedAt = time.Now()
	}
	workflow.UpdatedAt = time.Now()

	// 保存到内存
	m.workflowMu.Lock()
	m.workflows[workflow.ID] = workflow
	m.workflowMu.Unlock()

	// 持久化存储
	if err := m.workflowStorage.Save(ctx, workflow); err != nil {
		m.logger.Error("Failed to save workflow to storage", "error", err)
		return nil, fmt.Errorf("failed to save workflow: %w", err)
	}

	m.logger.Info("Workflow created successfully", "id", workflow.ID, "name", workflow.Name)
	return workflow, nil
}

func (m *StartupWorkflowManagerImpl) GetWorkflow(ctx context.Context, id string) (*model.StartupWorkflow, error) {
	m.workflowMu.RLock()
	defer m.workflowMu.RUnlock()

	workflow, exists := m.workflows[id]
	if !exists {
		return nil, model.ErrWorkflowNotFound
	}

	return workflow, nil
}

func (m *StartupWorkflowManagerImpl) UpdateWorkflow(ctx context.Context, workflow *model.StartupWorkflow) (*model.StartupWorkflow, error) {
	// 验证工作流
	if err := m.ValidateWorkflow(ctx, workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// 检查是否存在
	existingWorkflow, err := m.GetWorkflow(ctx, workflow.ID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// 更新时间戳
	workflow.UpdatedAt = time.Now()
	workflow.CreatedAt = existingWorkflow.CreatedAt

	// 保存到内存
	m.workflowMu.Lock()
	m.workflows[workflow.ID] = workflow
	m.workflowMu.Unlock()

	// 持久化存储
	if err := m.workflowStorage.Update(ctx, workflow); err != nil {
		m.logger.Error("Failed to update workflow in storage", "error", err)
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	m.logger.Info("Workflow updated successfully", "id", workflow.ID)
	return workflow, nil
}

func (m *StartupWorkflowManagerImpl) DeleteWorkflow(ctx context.Context, id string) error {
	// 检查是否正在使用
	if m.isWorkflowInUse(ctx, id) {
		return fmt.Errorf("workflow is currently in use and cannot be deleted")
	}

	// 从内存删除
	m.workflowMu.Lock()
	delete(m.workflows, id)
	m.workflowMu.Unlock()

	// 从存储删除
	if err := m.workflowStorage.Delete(ctx, id); err != nil {
		m.logger.Error("Failed to delete workflow from storage", "error", err)
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	m.logger.Info("Workflow deleted successfully", "id", id)
	return nil
}

func (m *StartupWorkflowManagerImpl) ListWorkflows(ctx context.Context) ([]*model.StartupWorkflow, error) {
	m.workflowMu.RLock()
	defer m.workflowMu.RUnlock()

	workflows := make([]*model.StartupWorkflow, 0, len(m.workflows))
	for _, workflow := range m.workflows {
		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

func (m *StartupWorkflowManagerImpl) ValidateWorkflow(ctx context.Context, workflow *model.StartupWorkflow) error {
	// 基本验证
	if workflow.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if len(workflow.Nodes) == 0 {
		return fmt.Errorf("workflow must have at least one node")
	}

	// 节点ID唯一性
	nodeIDs := make(map[string]bool)
	for _, node := range workflow.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node ID is required")
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true
	}

	// 边的有效性
	for _, edge := range workflow.Edges {
		if !nodeIDs[edge.From] {
			return fmt.Errorf("edge references non-existent node: %s", edge.From)
		}
		if !nodeIDs[edge.To] {
			return fmt.Errorf("edge references non-existent node: %s", edge.To)
		}
		if edge.From == edge.To {
			return fmt.Errorf("self-loop detected for node: %s", edge.From)
		}
	}

	// 循环依赖检测
	if m.hasCircularDependency(workflow.Nodes, workflow.Edges) {
		return model.ErrCircularDependency
	}

	// 节点验证
	for _, node := range workflow.Nodes {
		if err := m.validateNode(ctx, &node); err != nil {
			return fmt.Errorf("node validation failed for %s: %w", node.ID, err)
		}
	}

	return nil
}

func (m *StartupWorkflowManagerImpl) hasCircularDependency(nodes []model.StartupNode, edges []workflow.Edge) bool {
	// 构建邻接表
	adjacency := make(map[string][]string)
	nodeSet := make(map[string]bool)

	for _, node := range nodes {
		nodeSet[node.ID] = true
		adjacency[node.ID] = make([]string, 0)
	}

	for _, edge := range edges {
		if nodeSet[edge.From] && nodeSet[edge.To] {
			adjacency[edge.From] = append(adjacency[edge.From], edge.To)
		}
	}

	// DFS检测循环
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for nodeID := range nodeSet {
		if !visited[nodeID] {
			if m.hasCircularDependencyDFS(nodeID, adjacency, visited, recursionStack) {
				return true
			}
		}
	}

	return false
}

func (m *StartupWorkflowManagerImpl) hasCircularDependencyDFS(nodeID string, adjacency map[string][]string, visited, recursionStack map[string]bool) bool {
	visited[nodeID] = true
	recursionStack[nodeID] = true

	for _, neighbor := range adjacency[nodeID] {
		if !visited[neighbor] {
			if m.hasCircularDependencyDFS(neighbor, adjacency, visited, recursionStack) {
				return true
			}
		} else if recursionStack[neighbor] {
			return true
		}
	}

	recursionStack[nodeID] = false
	return false
}

func (m *StartupWorkflowManagerImpl) validateNode(ctx context.Context, node *model.StartupNode) error {
	// 检查节点类型是否支持
	if _, err := m.pluginManager.GetExecutor(node.Type); err != nil {
		return fmt.Errorf("unsupported node type: %s", node.Type)
	}

	// 验证超时配置
	if node.Timeout <= 0 {
		return fmt.Errorf("invalid timeout value: %v", node.Timeout)
	}

	// 验证重试配置
	if node.Retry != nil {
		if node.Retry.MaxAttempts <= 0 {
			return fmt.Errorf("invalid max retry attempts: %d", node.Retry.MaxAttempts)
		}
		if node.Retry.Delay <= 0 {
			return fmt.Errorf("invalid retry delay: %v", node.Retry.Delay)
		}
	}

	return nil
}

// isWorkflowInUse 检查工作流是否正在使用
func (m *StartupWorkflowManagerImpl) isWorkflowInUse(ctx context.Context, workflowID string) bool {
	m.executionMu.RLock()
	defer m.executionMu.RUnlock()

	for _, execution := range m.executions {
		if execution.WorkflowID == workflowID &&
		   (execution.Status == workflow.ExecutionStatusRunning ||
		    execution.Status == workflow.ExecutionStatusPending) {
			return true
		}
	}

	return false
}

// 添加事件处理器
func (m *StartupWorkflowManagerImpl) AddEventHandler(handler model.StartupEventHandler) {
	m.eventHandlers = append(m.eventHandlers, handler)
}

// 触发事件
func (m *StartupWorkflowManagerImpl) triggerEvent(ctx context.Context, event *model.StartupEvent) error {
	for _, handler := range m.eventHandlers {
		var err error
		switch event.EventType {
		case model.EventTypeExecutionStart:
			err = handler.OnExecutionStart(ctx, event.Data["execution"].(*model.StartupExecution))
		case model.EventTypeNodeStart:
			err = handler.OnNodeStart(ctx,
				event.Data["execution"].(*model.StartupExecution),
				event.Data["node"].(*model.StartupNode))
		case model.EventTypeNodeProgress:
			err = handler.OnNodeProgress(ctx,
				event.Data["execution"].(*model.StartupExecution),
				event.Data["node"].(*model.StartupNode),
				event.Data["progress"].(float64))
		case model.EventTypeNodeComplete:
			err = handler.OnNodeComplete(ctx,
				event.Data["execution"].(*model.StartupExecution),
				event.Data["node"].(*model.StartupNode),
				event.Data["result"].(*model.StartupNodeResult))
		case model.EventTypeNodeError:
			err = handler.OnNodeError(ctx,
				event.Data["execution"].(*model.StartupExecution),
				event.Data["node"].(*model.StartupNode),
				fmt.Errorf(event.Data["error"].(string)))
		case model.EventTypeExecutionEnd:
			err = handler.OnExecutionEnd(ctx, event.Data["execution"].(*model.StartupExecution))
		}

		if err != nil {
			m.logger.Error("Event handler error", "event_type", event.EventType, "error", err)
		}
	}

	return nil
}

// GetSystemStatus 获取系统状态
func (m *StartupWorkflowManagerImpl) GetSystemStatus(ctx context.Context) (*model.StartupSystemStatus, error) {
	m.executionMu.RLock()
	defer m.executionMu.RUnlock()

	status := &model.StartupSystemStatus{
		IsRunning:     true,
		StartTime:     time.Now(), // TODO: 获取实际启动时间
		Version:       "1.0.0",      // TODO: 获取版本信息
		Components:    make(map[string]model.ComponentStatus),
		TotalExecutions: m.metrics.TotalExecutions,
		SuccessfulRuns: m.metrics.SuccessfulExecutions,
		FailedRuns:     m.metrics.FailedExecutions,
	}

	// 查找当前执行
	for _, execution := range m.executions {
		if execution.Status == workflow.ExecutionStatusRunning {
			status.CurrentExecution = execution
			break
		}
	}

	// 查找最近的执行
	if len(m.executions) > 0 {
		var latest *model.StartupExecution
		for _, execution := range m.executions {
			if latest == nil || execution.StartTime.After(latest.StartTime) {
				latest = execution
			}
		}
		status.LastExecution = latest
	}

	// 组件状态
	// TODO: 实现组件健康检查
	status.Components["workflow_manager"] = model.ComponentStatus{
		Status:    "healthy",
		LastCheck: time.Now(),
		Message:   "Running normally",
	}

	return status, nil
}

// GetMetrics 获取指标
func (m *StartupWorkflowManagerImpl) GetMetrics(ctx context.Context) (*model.StartupMetrics, error) {
	// 更新指标
	m.updateMetrics()

	return m.metrics, nil
}

// updateMetrics 更新指标
func (m *StartupWorkflowManagerImpl) updateMetrics() {
	m.executionMu.RLock()
	defer m.executionMu.RUnlock()

	// 统计执行次数
	total := len(m.executions)
	successful := 0
	failed := 0
	var totalDuration time.Duration
	var minDuration, maxDuration time.Duration

	first := true
	for _, execution := range m.executions {
		switch execution.Status {
		case workflow.ExecutionStatusCompleted:
			successful++
		case workflow.ExecutionStatusFailed:
			failed++
		}

		if !execution.EndTime.IsZero() {
			duration := execution.EndTime.Sub(execution.StartTime)
			totalDuration += duration

			if first {
				minDuration = duration
				maxDuration = duration
				first = false
			} else {
				if duration < minDuration {
					minDuration = duration
				}
				if duration > maxDuration {
					maxDuration = duration
				}
			}
		}
	}

	// 计算平均时间
	var avgDuration time.Duration
	if successful+failed > 0 {
		avgDuration = totalDuration / time.Duration(successful+failed)
	}

	// 更新指标
	m.metrics.TotalExecutions = int64(total)
	m.metrics.SuccessfulExecutions = int64(successful)
	m.metrics.FailedExecutions = int64(failed)
	m.metrics.AverageExecutionTime = avgDuration
	m.metrics.MinExecutionTime = minDuration
	m.metrics.MaxExecutionTime = maxDuration
	m.metrics.CalculatedAt = time.Now()

	// 统计最近执行
	recent := make([]*model.StartupExecution, 0, 10)
	for _, execution := range m.executions {
		if len(recent) >= 10 {
			break
		}
		recent = append(recent, execution)
	}
	m.metrics.RecentExecutions = recent

	// TODO: 统计节点执行次数和错误次数
	m.metrics.NodeExecutionCounts = make(map[string]int64)
	m.metrics.ErrorCounts = make(map[string]int64)
}

// 执行管理相关方法

func (m *StartupWorkflowManagerImpl) ExecuteWorkflow(ctx context.Context, workflowID string, inputs map[string]interface{}) (*model.StartupExecution, error) {
	return m.ExecuteWorkflowWithConfig(ctx, workflowID, inputs, nil)
}

func (m *StartupWorkflowManagerImpl) ExecuteWorkflowWithConfig(ctx context.Context, workflowID string, inputs map[string]interface{}, config *model.StartupWorkflowConfig) (*model.StartupExecution, error) {
	// 获取工作流
	wf, err := m.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// 检查并发执行限制
	if m.getRunningExecutionCount() >= m.config.MaxExecutions {
		return nil, fmt.Errorf("maximum concurrent executions reached: %d", m.config.MaxExecutions)
	}

	// 创建执行实例
	execution := &model.StartupExecution{
		ID:             fmt.Sprintf("exec_%d", time.Now().UnixNano()),
		WorkflowID:     workflowID,
		WorkflowName:   wf.Name,
		Status:         workflow.ExecutionStatusPending,
		StartTime:      time.Now(),
		NodeResults:    make(map[string]*model.StartupNodeResult),
		Context:        make(map[string]interface{}),
		Inputs:         inputs,
		Outputs:        make(map[string]interface{}),
		TotalNodes:     len(wf.Nodes),
		Environment:    make(map[string]interface{}),
		TriggeredBy:    "user",
	}

	// 应用配置
	if config != nil {
		execution.Context["config"] = config
	} else {
		execution.Context["config"] = wf.Config
	}

	// 保存执行实例
	m.executionMu.Lock()
	m.executions[execution.ID] = execution
	m.executionMu.Unlock()

	// 触发执行开始事件
	event := &model.StartupEvent{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		EventType: model.EventTypeExecutionStart,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"execution":   execution,
			"execution_id": execution.ID,
			"workflow_id":  workflowID,
			"source":       "workflow_manager",
		},
	}
	if err := m.triggerEvent(ctx, event); err != nil {
		m.logger.Warn("Failed to trigger execution start event", "error", err)
	}

	// 异步执行工作流
	go m.executeWorkflowAsync(ctx, wf, execution)

	m.logger.Info("Workflow execution started", "execution_id", execution.ID, "workflow_id", workflowID)
	return execution, nil
}

func (m *StartupWorkflowManagerImpl) executeWorkflowAsync(ctx context.Context, wf *model.StartupWorkflow, execution *model.StartupExecution) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("Workflow execution panic recovered", "execution_id", execution.ID, "panic", r)
			execution.Status = workflow.ExecutionStatusFailed
			execution.Error = &model.StartupError{
				Code:    "PANIC",
				Message: fmt.Sprintf("Execution panic: %v", r),
			}
			m.markExecutionCompleted(ctx, execution)
		}
	}()

	// 设置执行状态
	execution.Status = workflow.ExecutionStatusRunning

	// 转换为工作流格式并执行
	workflowNodes, workflowEdges := m.convertToWorkflowFormat(wf.Nodes, wf.Edges)

	// 创建工作流执行器
	workflowExecutor := workflow.NewWorkflowExecutor(
		nil, // TODO: 传递配置
		nil, // TODO: 传递插件管理器
		nil, // TODO: 传递DAG引擎
		nil, // TODO: 传递数据流引擎
		nil, // TODO: 传递日志器
	)

	// 执行工作流
	workflowExecution, err := workflowExecutor.Execute(ctx, &workflow.Workflow{
		ID:    wf.ID,
		Name:  wf.Name,
		Nodes: workflowNodes,
		Edges: workflowEdges,
		// Config:   wf.Config, // TODO: Convert config
	}, execution.Inputs)

	if err != nil {
		m.logger.Error("Workflow execution failed", "execution_id", execution.ID, "error", err)
		execution.Status = workflow.ExecutionStatusFailed
		execution.Error = &model.StartupError{Message: err.Error()}
	} else {
		// 映射执行结果
		m.mapExecutionResults(execution, workflowExecution)
		execution.Status = workflow.ExecutionStatusCompleted
	}

	// 计算执行时间
	now := time.Now()
	execution.EndTime = &now
	execution.Duration = execution.EndTime.Sub(execution.StartTime)

	// 触发执行结束事件
	event := &model.StartupEvent{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		EventType: model.EventTypeExecutionEnd,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"execution":    execution,
			"execution_id": execution.ID,
			"workflow_id":  wf.ID,
			"source":       "workflow_manager",
		},
	}
	if err := m.triggerEvent(ctx, event); err != nil {
		m.logger.Warn("Failed to trigger execution end event", "error", err)
	}

	m.markExecutionCompleted(ctx, execution)
}

func (m *StartupWorkflowManagerImpl) convertToWorkflowFormat(nodes []model.StartupNode, edges []workflow.Edge) ([]workflow.Node, []workflow.Edge) {
	workflowNodes := make([]workflow.Node, len(nodes))
	for i, node := range nodes {
		// 转换输入Schema
		inputs := make([]workflow.InputSchema, 0)
		for key, value := range node.Config {
			inputs = append(inputs, workflow.InputSchema{
				Name:    key,
				Type:    inferType(value),
				Required: false,
			})
		}

		workflowNodes[i] = workflow.Node{
			ID:          node.ID,
			Name:        node.Name,
			Type:        m.mapStartupNodeType(node.Type),
			Description: node.Description,
			Inputs:      inputs,
			Outputs:     []workflow.OutputSchema{{Name: "result", Type: "object"}},
			Position:    node.Position,
			Status:      node.Status,
		}
	}

	return workflowNodes, edges
}

func (m *StartupWorkflowManagerImpl) mapStartupNodeType(nodeType model.StartupNodeType) workflow.NodeType {
	switch nodeType {
	case model.StartupNodeStorage:
		return workflow.NodeTypeTask
	case model.StartupNodeConfig:
		return workflow.NodeTypeTask
	case model.StartupNodeService:
		return workflow.NodeTypeTask
	case model.StartupNodeAuth:
		return workflow.NodeTypeTask
	case model.StartupNodePlugin:
		return workflow.NodeTypeTask
	case model.StartupNodeParallel:
		return workflow.NodeTypeParallel
	case model.StartupNodeMerge:
		return workflow.NodeTypeMerge
	default:
		return workflow.NodeTypeTask
	}
}

func inferType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32, int64, float32, float64:
		return "number"
	case bool:
		return "boolean"
	case map[string]interface{}, []interface{}:
		return "object"
	default:
		return "string"
	}
}

func (m *StartupWorkflowManagerImpl) mapExecutionResults(execution *model.StartupExecution, workflowExecution *workflow.Execution) {
	// 映射节点结果
	for nodeID, nodeResult := range workflowExecution.NodeResults {
		execution.NodeResults[nodeID] = &model.StartupNodeResult{
			NodeID:    nodeID,
			Status:    nodeResult.Status,
			StartTime: nodeResult.StartTime,
			EndTime:   nodeResult.EndTime,
			Inputs:    nodeResult.Inputs,
			Outputs:   nodeResult.Outputs,
			Error:     nodeResult.Error,
			Duration: nodeResult.ElapsedTime,
		}
	}

	// 计算进度
	completedNodes := 0
	var failedNodes []string
	for _, result := range execution.NodeResults {
		if result.Status == workflow.NodeStatusCompleted {
			completedNodes++
		}
		if result.Status == workflow.NodeStatusFailed {
			failedNodes = append(failedNodes, result.NodeID)
		}
	}
	if execution.TotalNodes > 0 {
		execution.Progress = float64(completedNodes) / float64(execution.TotalNodes) * 100
	}
	execution.CompletedNodes = completedNodes
	execution.FailedNodes = failedNodes
}

func (m *StartupWorkflowManagerImpl) markExecutionCompleted(ctx context.Context, execution *model.StartupExecution) {
	m.executionMu.Lock()
	defer m.executionMu.Unlock()

	// 持久化存储
	if err := m.executionStorage.Save(ctx, execution); err != nil {
		m.logger.Error("Failed to save execution to storage", "execution_id", execution.ID, "error", err)
	}

	// 更新指标
	m.updateMetrics()
}

func (m *StartupWorkflowManagerImpl) GetExecution(ctx context.Context, executionID string) (*model.StartupExecution, error) {
	m.executionMu.RLock()
	defer m.executionMu.RUnlock()

	execution, exists := m.executions[executionID]
	if !exists {
		return nil, model.ErrExecutionNotFound
	}

	return execution, nil
}

func (m *StartupWorkflowManagerImpl) ListExecutions(ctx context.Context, workflowID string) ([]*model.StartupExecution, error) {
	m.executionMu.RLock()
	defer m.executionMu.RUnlock()

	executions := make([]*model.StartupExecution, 0)
	for _, execution := range m.executions {
		if workflowID == "" || execution.WorkflowID == workflowID {
			executions = append(executions, execution)
		}
	}

	return executions, nil
}

func (m *StartupWorkflowManagerImpl) CancelExecution(ctx context.Context, executionID string) error {
	m.executionMu.Lock()
	defer m.executionMu.Unlock()

	execution, exists := m.executions[executionID]
	if !exists {
		return model.ErrExecutionNotFound
	}

	if execution.Status != workflow.ExecutionStatusRunning {
		return fmt.Errorf("execution is not running")
	}

	// TODO: 实现执行取消逻辑
	execution.Status = workflow.ExecutionStatusCancelled
	now := time.Now()
	execution.EndTime = &now
	execution.Error = &model.StartupError{Message: "Execution cancelled by user"}

	// 持久化更改
	if err := m.executionStorage.Save(ctx, execution); err != nil {
		m.logger.Error("Failed to save cancelled execution", "error", err)
	}

	m.logger.Info("Execution cancelled", "execution_id", executionID)
	return nil
}

func (m *StartupWorkflowManagerImpl) PauseExecution(ctx context.Context, executionID string) error {
	m.executionMu.Lock()
	defer m.executionMu.Unlock()

	execution, exists := m.executions[executionID]
	if !exists {
		return model.ErrExecutionNotFound
	}

	if execution.Status != workflow.ExecutionStatusRunning {
		return fmt.Errorf("execution is not running")
	}

	// TODO: 实现执行暂停逻辑
	execution.Status = workflow.ExecutionStatusPaused

	m.logger.Info("Execution paused", "execution_id", executionID)
	return nil
}

func (m *StartupWorkflowManagerImpl) ResumeExecution(ctx context.Context, executionID string) error {
	m.executionMu.Lock()
	defer m.executionMu.Unlock()

	execution, exists := m.executions[executionID]
	if !exists {
		return model.ErrExecutionNotFound
	}

	if execution.Status != workflow.ExecutionStatusPaused {
		return fmt.Errorf("execution is not paused")
	}

	// TODO: 实现执行恢复逻辑
	execution.Status = workflow.ExecutionStatusRunning

	m.logger.Info("Execution resumed", "execution_id", executionID)
	return nil
}

func (m *StartupWorkflowManagerImpl) getRunningExecutionCount() int {
	m.executionMu.RLock()
	defer m.executionMu.RUnlock()

	count := 0
	for _, execution := range m.executions {
		if execution.Status == workflow.ExecutionStatusRunning {
			count++
		}
	}

	return count
}

// 模板管理相关方法

func (m *StartupWorkflowManagerImpl) CreateTemplate(ctx context.Context, template *model.StartupWorkflowTemplate) (*model.StartupWorkflowTemplate, error) {
	// 验证模板
	if template.ID == "" {
		template.ID = fmt.Sprintf("template_%d", time.Now().UnixNano())
	}
	if template.Name == "" {
		return nil, fmt.Errorf("template name is required")
	}
	if template.Workflow == nil {
		return nil, fmt.Errorf("template workflow is required")
	}

	// 验证工作流
	if err := m.ValidateWorkflow(ctx, template.Workflow); err != nil {
		return nil, fmt.Errorf("template workflow validation failed: %w", err)
	}

	// 设置时间戳
	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}
	template.UpdatedAt = time.Now()

	// 保存到内存
	m.templateMu.Lock()
	m.templates[template.ID] = template
	m.templateMu.Unlock()

	// 持久化存储
	if err := m.templateStorage.Save(ctx, template); err != nil {
		m.logger.Error("Failed to save template to storage", "error", err)
		return nil, fmt.Errorf("failed to save template: %w", err)
	}

	m.logger.Info("Template created successfully", "id", template.ID, "name", template.Name)
	return template, nil
}

func (m *StartupWorkflowManagerImpl) GetTemplate(ctx context.Context, id string) (*model.StartupWorkflowTemplate, error) {
	m.templateMu.RLock()
	defer m.templateMu.RUnlock()

	template, exists := m.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found")
	}

	return template, nil
}

func (m *StartupWorkflowManagerImpl) UpdateTemplate(ctx context.Context, template *model.StartupWorkflowTemplate) (*model.StartupWorkflowTemplate, error) {
	// 验证模板
	if template.Workflow != nil {
		if err := m.ValidateWorkflow(ctx, template.Workflow); err != nil {
			return nil, fmt.Errorf("template workflow validation failed: %w", err)
		}
	}

	// 检查是否存在
	_, err := m.GetTemplate(ctx, template.ID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// 更新时间戳
	template.UpdatedAt = time.Now()

	// 保存到内存
	m.templateMu.Lock()
	m.templates[template.ID] = template
	m.templateMu.Unlock()

	// 持久化存储
	if err := m.templateStorage.Update(ctx, template); err != nil {
		m.logger.Error("Failed to update template in storage", "error", err)
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	m.logger.Info("Template updated successfully", "id", template.ID)
	return template, nil
}

func (m *StartupWorkflowManagerImpl) DeleteTemplate(ctx context.Context, id string) error {
	// 从内存删除
	m.templateMu.Lock()
	delete(m.templates, id)
	m.templateMu.Unlock()

	// 从存储删除
	if err := m.templateStorage.Delete(ctx, id); err != nil {
		m.logger.Error("Failed to delete template from storage", "error", err)
		return fmt.Errorf("failed to delete template: %w", err)
	}

	m.logger.Info("Template deleted successfully", "id", id)
	return nil
}

func (m *StartupWorkflowManagerImpl) ListTemplates(ctx context.Context) ([]*model.StartupWorkflowTemplate, error) {
	m.templateMu.RLock()
	defer m.templateMu.RUnlock()

	templates := make([]*model.StartupWorkflowTemplate, 0, len(m.templates))
	for _, template := range m.templates {
		templates = append(templates, template)
	}

	return templates, nil
}

func (m *StartupWorkflowManagerImpl) DeployFromTemplate(ctx context.Context, templateID string, name string) (*model.StartupWorkflow, error) {
	// 获取模板
	template, err := m.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// 创建工作流副本
	workflow := &model.StartupWorkflow{
		ID:          fmt.Sprintf("workflow_%d", time.Now().UnixNano()),
		Name:        name,
		Description: fmt.Sprintf("Deployed from template: %s", template.Name),
		Version:     template.Version,
		Nodes:       template.Workflow.Nodes,
		Edges:       template.Workflow.Edges,
		Config:      template.Workflow.Config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata: map[string]string{
			"source_template": templateID,
			"deployed_at":    time.Now().Format(time.RFC3339),
		},
	}

	// 保存工作流
	return m.CreateWorkflow(ctx, workflow)
}








