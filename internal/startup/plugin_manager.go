package startup

import (
	"context"
	"fmt"
	"sync"

	"xiaozhi-server-go/internal/startup/nodes"
)

// DefaultPluginManager 默认启动插件管理器
type DefaultPluginManager struct {
	executors map[StartupNodeType]StartupNodeExecutor
	logger    StartupLogger
	mutex     sync.RWMutex
}

// NewDefaultPluginManager 创建默认插件管理器
func NewDefaultPluginManager(logger StartupLogger) *DefaultPluginManager {
	return &DefaultPluginManager{
		executors: make(map[StartupNodeType]StartupNodeExecutor),
		logger:    logger,
	}
}

// RegisterExecutor 注册节点执行器
func (p *DefaultPluginManager) RegisterExecutor(nodeType StartupNodeType, executor StartupNodeExecutor) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.executors[nodeType]; exists {
		return fmt.Errorf("executor for node type %s already registered", nodeType)
	}

	p.executors[nodeType] = executor
	p.logger.Info("Registered node executor", "node_type", nodeType, "executor", executor.GetNodeInfo().Name)

	return nil
}

// GetExecutor 获取节点执行器
func (p *DefaultPluginManager) GetExecutor(nodeType StartupNodeType) (StartupNodeExecutor, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	executor, exists := p.executors[nodeType]
	return executor, exists
}

// ListExecutors 列出所有执行器
func (p *DefaultPluginManager) ListExecutors() map[StartupNodeType]StartupNodeInfo {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	result := make(map[StartupNodeType]StartupNodeInfo)
	for nodeType, executor := range p.executors {
		result[nodeType] = *executor.GetNodeInfo()
	}

	return result
}

// UnregisterExecutor 注销节点执行器
func (p *DefaultPluginManager) UnregisterExecutor(nodeType StartupNodeType) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.executors[nodeType]; !exists {
		return fmt.Errorf("executor for node type %s not found", nodeType)
	}

	// 清理执行器资源
	executor := p.executors[nodeType]
	if err := executor.Cleanup(context.Background()); err != nil {
		p.logger.Error("Failed to cleanup executor", "node_type", nodeType, "error", err)
	}

	delete(p.executors, nodeType)
	p.logger.Info("Unregistered node executor", "node_type", nodeType)

	return nil
}

// InitializeDefaults 初始化默认执行器
func (p *DefaultPluginManager) InitializeDefaults() error {
	p.logger.Info("Initializing default node executors")

	// 注册存储节点执行器
	storageExecutor := nodes.NewStorageNodeExecutor(p.logger)
	if err := p.RegisterExecutor(StartupNodeStorage, storageExecutor); err != nil {
		return fmt.Errorf("failed to register storage executor: %w", err)
	}

	// 注册服务节点执行器
	serviceExecutor := nodes.NewServiceNodeExecutor(p.logger)
	if err := p.RegisterExecutor(StartupNodeService, serviceExecutor); err != nil {
		return fmt.Errorf("failed to register service executor: %w", err)
	}

	// 注册配置节点执行器
	configExecutor := nodes.NewConfigNodeExecutor(p.logger)
	if err := p.RegisterExecutor(StartupNodeConfig, configExecutor); err != nil {
		return fmt.Errorf("failed to register config executor: %w", err)
	}

	// 注册认证节点执行器
	authExecutor := nodes.NewAuthNodeExecutor(p.logger)
	if err := p.RegisterExecutor(StartupNodeAuth, authExecutor); err != nil {
		return fmt.Errorf("failed to register auth executor: %w", err)
	}

	// 注册插件节点执行器
	pluginExecutor := nodes.NewPluginNodeExecutor(p.logger)
	if err := p.RegisterExecutor(StartupNodePlugin, pluginExecutor); err != nil {
		return fmt.Errorf("failed to register plugin executor: %w", err)
	}

	p.logger.Info("Default node executors initialized successfully",
		"total_executors", len(p.executors))

	return nil
}

// Cleanup 清理所有执行器
func (p *DefaultPluginManager) Cleanup(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.logger.Info("Cleaning up all node executors")

	var errors []error
	for nodeType, executor := range p.executors {
		if err := executor.Cleanup(ctx); err != nil {
			p.logger.Error("Failed to cleanup executor", "node_type", nodeType, "error", err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup completed with %d errors", len(errors))
	}

	p.logger.Info("All node executors cleaned up successfully")
	return nil
}

// GetExecutorStats 获取执行器统计信息
func (p *DefaultPluginManager) GetExecutorStats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_executors": len(p.executors),
		"executors":       make([]interface{}, 0),
	}

	for nodeType, executor := range p.executors {
		info := executor.GetNodeInfo()
		executorStats := map[string]interface{}{
			"type":         nodeType,
			"name":         info.Name,
			"description":  info.Description,
			"version":      info.Version,
			"author":       info.Author,
			"capabilities": info.Capabilities,
		}
		stats["executors"] = append(stats["executors"].([]interface{}), executorStats)
	}

	return stats
}

// ValidateExecutor 验证执行器
func (p *DefaultPluginManager) ValidateExecutor(nodeType StartupNodeType, node *StartupNode) error {
	executor, exists := p.GetExecutor(nodeType)
	if !exists {
		return fmt.Errorf("no executor found for node type: %s", nodeType)
	}

	return executor.Validate(context.Background(), node)
}

// ExecutorInfo 执行器信息
type ExecutorInfo struct {
	Type         StartupNodeType        `json:"type"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	Author       string                 `json:"author"`
	Capabilities []string               `json:"capabilities"`
	Config       map[string]interface{} `json:"supported_config"`
	Registered   bool                   `json:"registered"`
}

// ListExecutorInfos 列出执行器详细信息
func (p *DefaultPluginManager) ListExecutorInfos() []ExecutorInfo {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var infos []ExecutorInfo
	for nodeType, executor := range p.executors {
		info := executor.GetNodeInfo()
		infos = append(infos, ExecutorInfo{
			Type:         nodeType,
			Name:         info.Name,
			Description:  info.Description,
			Version:      info.Version,
			Author:       info.Author,
			Capabilities: info.Capabilities,
			Config:       info.SupportedConfig,
			Registered:   true,
		})
	}

	return infos
}

// HealthCheck 健康检查
func (p *DefaultPluginManager) HealthCheck(ctx context.Context) map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	health := map[string]interface{}{
		"status":   "healthy",
		"executors": make(map[string]interface{}),
	}

	unhealthyCount := 0
	for nodeType, executor := range p.executors {
		executorHealth := map[string]interface{}{
			"status": "healthy",
			"error":  nil,
		}

		// 尝试验证执行器（简单的健康检查）
		testNode := &StartupNode{
			ID:   "health-check",
			Name: "Health Check Node",
			Type: nodeType,
			Config: map[string]interface{}{
				"test": true,
			},
		}

		if err := executor.Validate(ctx, testNode); err != nil {
			executorHealth["status"] = "unhealthy"
			executorHealth["error"] = err.Error()
			unhealthyCount++
		}

		health["executors"].(map[string]interface{})[string(nodeType)] = executorHealth
	}

	if unhealthyCount > 0 {
		health["status"] = "degraded"
	}

	health["unhealthy_count"] = unhealthyCount
	health["healthy_count"] = len(p.executors) - unhealthyCount

	return health
}

// GetExecutorByCapability 根据能力获取执行器
func (p *DefaultPluginManager) GetExecutorByCapability(capability string) []StartupNodeExecutor {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var matchingExecutors []StartupNodeExecutor
	for _, executor := range p.executors {
		info := executor.GetNodeInfo()
		for _, cap := range info.Capabilities {
			if cap == capability {
				matchingExecutors = append(matchingExecutors, executor)
				break
			}
		}
	}

	return matchingExecutors
}

// ReloadExecutor 重新加载执行器
func (p *DefaultPluginManager) ReloadExecutor(nodeType StartupNodeType) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	executor, exists := p.executors[nodeType]
	if !exists {
		return fmt.Errorf("executor for node type %s not found", nodeType)
	}

	// 清理旧的执行器
	if err := executor.Cleanup(context.Background()); err != nil {
		p.logger.Error("Failed to cleanup old executor", "node_type", nodeType, "error", err)
	}

	// 根据类型创建新的执行器
	var newExecutor StartupNodeExecutor
	var err error

	switch nodeType {
	case StartupNodeStorage:
		newExecutor = nodes.NewStorageNodeExecutor(p.logger)
	case StartupNodeService:
		newExecutor = nodes.NewServiceNodeExecutor(p.logger)
	case StartupNodeConfig:
		newExecutor = nodes.NewConfigNodeExecutor(p.logger)
	case StartupNodeAuth:
		newExecutor = nodes.NewAuthNodeExecutor(p.logger)
	case StartupNodePlugin:
		newExecutor = nodes.NewPluginNodeExecutor(p.logger)
	default:
		return fmt.Errorf("unknown node type for reload: %s", nodeType)
	}

	if err != nil {
		return fmt.Errorf("failed to create new executor: %w", err)
	}

	p.executors[nodeType] = newExecutor
	p.logger.Info("Reloaded node executor", "node_type", nodeType)

	return nil
}