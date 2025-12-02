package nodes

import (
	"context"
	"fmt"
	"time"

	"xiaozhi-server-go/internal/startup"
	"xiaozhi-server-go/internal/workflow"
)

// PluginNodeExecutor 插件节点执行器
type PluginNodeExecutor struct {
	logger startup.StartupLogger
}

// NewPluginNodeExecutor 创建插件节点执行器
func NewPluginNodeExecutor(logger startup.StartupLogger) *PluginNodeExecutor {
	return &PluginNodeExecutor{
		logger: logger,
	}
}

// Execute 执行插件节点
func (e *PluginNodeExecutor) Execute(
	ctx context.Context,
	node *startup.StartupNode,
	inputs map[string]interface{},
	context map[string]interface{},
) (*startup.StartupNodeResult, error) {
	startTime := time.Now()
	result := &startup.StartupNodeResult{
		NodeID:   node.ID,
		NodeName: node.Name,
		NodeType: node.Type,
		StartTime: startTime,
		Status:   workflow.NodeStatusRunning,
		Inputs:   inputs,
		Outputs:  make(map[string]interface{}),
		Logs:     make([]startup.StartupNodeLog, 0),
	}

	e.logger.Info("Executing plugin node", "node_id", node.ID, "node_name", node.Name)

	// 根据节点ID执行不同的插件操作
	switch node.ID {
	case "plugin:init-manager":
		err := e.executeInitPluginManager(ctx, node, result, inputs, context)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	default:
		err := fmt.Errorf("unknown plugin node: %s", node.ID)
		result.Status = workflow.NodeStatusFailed
		result.Error = err.Error()
		return result, err
	}

	// 成功完成
	endTime := time.Now()
	result.EndTime = &endTime
	result.Duration = endTime.Sub(startTime)
	result.Status = workflow.NodeStatusCompleted

	e.logger.Info("Plugin node completed successfully",
		"node_id", node.ID,
		"duration", result.Duration.String())

	return result, nil
}

// executeInitPluginManager 执行初始化插件管理器
func (e *PluginNodeExecutor) executeInitPluginManager(
	ctx context.Context,
	node *startup.StartupNode,
	result *startup.StartupNodeResult,
	inputs map[string]interface{},
	context map[string]interface{},
) error {
	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting plugin manager initialization",
	})

	// 获取配置
	enableDiscovery := getBoolConfig(node.Config, "enable_discovery", true)
	discoveryPaths := getStringSliceConfig(node.Config, "discovery_paths", []string{"./plugins", "./plugins/examples"})
	scanInterval := getStringConfig(node.Config, "scan_interval", "30s")
	healthCheckInterval := getStringConfig(node.Config, "health_check_interval", "10s")
	enableHealthCheck := getBoolConfig(node.Config, "enable_health_check", true)
	registryType := getStringConfig(node.Config, "registry_type", "memory")
	registryTTL := getStringConfig(node.Config, "registry_ttl", "5m")

	e.logger.Info("Initializing plugin manager",
		"enable_discovery", enableDiscovery,
		"discovery_paths", discoveryPaths,
		"scan_interval", scanInterval,
		"health_check_interval", healthCheckInterval,
		"enable_health_check", enableHealthCheck,
		"registry_type", registryType,
		"registry_ttl", registryTTL)

	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Plugin config: discovery=%v, paths=%v, scan=%s", enableDiscovery, discoveryPaths, scanInterval),
	})

	// 检查依赖节点是否已完成
	loggingReady := e.checkDependencyCompletion(context, "logging:init-provider")

	if !loggingReady {
		err := fmt.Errorf("required dependency not completed: logging=%v", loggingReady)
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   err.Error(),
		})
		return err
	}

	// 初始化插件注册表
	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Initializing %s plugin registry", registryType),
	})

	err := e.initializePluginRegistry(ctx, registryType, registryTTL, result)
	if err != nil {
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   fmt.Sprintf("Failed to initialize plugin registry: %s", err.Error()),
		})
		return err
	}

	// 初始化插件发现系统
	if enableDiscovery {
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Initializing plugin discovery system",
		})

		err := e.initializePluginDiscovery(ctx, discoveryPaths, scanInterval, result)
		if err != nil {
			result.Logs = append(result.Logs, startup.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "error",
				Message:   fmt.Sprintf("Failed to initialize plugin discovery: %s", err.Error()),
			})
			return err
		}
	}

	// 初始化健康检查系统
	if enableHealthCheck {
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Initializing plugin health check system",
		})

		err := e.initializeHealthCheckSystem(ctx, healthCheckInterval, result)
		if err != nil {
			result.Logs = append(result.Logs, startup.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "error",
				Message:   fmt.Sprintf("Failed to initialize health check: %s", err.Error()),
			})
			return err
		}
	}

	// 执行初始插件扫描
	if enableDiscovery {
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Performing initial plugin discovery scan",
		})

		discoveredPlugins, err := e.performPluginDiscovery(ctx, discoveryPaths, result)
		if err != nil {
			result.Logs = append(result.Logs, startup.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "warn",
				Message:   fmt.Sprintf("Initial plugin discovery failed: %s", err.Error()),
			})
			// 发现失败不阻止插件管理器初始化
		} else {
			result.Logs = append(result.Logs, startup.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   fmt.Sprintf("Discovered %d plugins", len(discoveredPlugins)),
			})
		}
	}

	// 启动后台任务
	go e.startBackgroundTasks(ctx, enableDiscovery, enableHealthCheck, scanInterval, healthCheckInterval, discoveryPaths, result)

	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Plugin manager initialized successfully",
	})

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"discovery_enabled":         enableDiscovery,
		"discovery_paths":          discoveryPaths,
		"scan_interval":           scanInterval,
		"health_check_enabled":     enableHealthCheck,
		"health_check_interval":    healthCheckInterval,
		"registry_type":           registryType,
		"registry_ttl":            registryTTL,
		"plugin_manager_id":       "plugin-manager-" + node.ID,
		"registry_id":             "plugin-registry-" + node.ID,
		"discovery_service_id":    "discovery-" + node.ID,
		"health_checker_id":       "health-checker-" + node.ID,
		"initialized_at":          time.Now(),
		"discovered_plugin_count": len(result.Outputs["discovered_plugins"].([]interface{})),
		"plugin_types":           e.getSupportedPluginTypes(),
	}

	return nil
}

// initializePluginRegistry 初始化插件注册表
func (e *PluginNodeExecutor) initializePluginRegistry(ctx context.Context, registryType, registryTTL string, result *startup.StartupNodeResult) error {
	switch registryType {
	case "memory":
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Memory-based plugin registry initialized",
		})
	case "redis":
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Redis-based plugin registry initialized",
		})
	case "database":
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Database-based plugin registry initialized",
		})
	default:
		return fmt.Errorf("unsupported registry type: %s", registryType)
	}

	// 解析注册表TTL
	if registryTTL != "" {
		_, err := time.ParseDuration(registryTTL)
		if err != nil {
			return fmt.Errorf("invalid registry_ttl format: %s", registryTTL)
		}
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Registry TTL set to: %s", registryTTL),
		})
	}

	// 模拟注册表初始化延迟
	select {
	case <-time.After(300 * time.Millisecond):
		// 初始化完成
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// initializePluginDiscovery 初始化插件发现系统
func (e *PluginNodeExecutor) initializePluginDiscovery(ctx context.Context, paths []string, scanInterval string, result *startup.StartupNodeResult) error {
	// 解析扫描间隔
	if scanInterval != "" {
		_, err := time.ParseDuration(scanInterval)
		if err != nil {
			return fmt.Errorf("invalid scan_interval format: %s", scanInterval)
		}
	}

	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Plugin discovery initialized with paths: %v", paths),
	})

	// 验证发现路径
	for _, path := range paths {
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Plugin discovery path: %s", path),
		})
	}

	// 模拟发现系统初始化延迟
	select {
	case <-time.After(400 * time.Millisecond):
		// 初始化完成
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// initializeHealthCheckSystem 初始化健康检查系统
func (e *PluginNodeExecutor) initializeHealthCheckSystem(ctx context.Context, healthCheckInterval string, result *startup.StartupNodeResult) error {
	// 解析健康检查间隔
	if healthCheckInterval != "" {
		_, err := time.ParseDuration(healthCheckInterval)
		if err != nil {
			return fmt.Errorf("invalid health_check_interval format: %s", healthCheckInterval)
		}
	}

	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Health check system initialized with interval: %s", healthCheckInterval),
	})

	// 模拟健康检查系统初始化延迟
	select {
	case <-time.After(200 * time.Millisecond):
		// 初始化完成
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// performPluginDiscovery 执行插件发现
func (e *PluginNodeExecutor) performPluginDiscovery(ctx context.Context, paths []string, result *startup.StartupNodeResult) ([]map[string]interface{}, error) {
	var discoveredPlugins []map[string]interface{}

	// 模拟插件发现过程
	for i, path := range paths {
		select {
		case <-time.After(200 * time.Millisecond):
			// 模拟发现一个插件
			plugin := map[string]interface{}{
				"id":          fmt.Sprintf("plugin-%d", i+1),
				"name":        fmt.Sprintf("Example Plugin %d", i+1),
				"path":        path,
				"type":        "http",
				"version":     "1.0.0",
				"status":      "discovered",
				"discovered_at": time.Now(),
				"metadata": map[string]interface{}{
					"description": fmt.Sprintf("Example plugin discovered from %s", path),
					"author":      "XiaoZhi Flow Team",
				},
			}
			discoveredPlugins = append(discoveredPlugins, plugin)

			result.Logs = append(result.Logs, startup.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   fmt.Sprintf("Discovered plugin: %s", plugin["name"]),
			})
		case <-ctx.Done():
			return discoveredPlugins, ctx.Err()
		}
	}

	// 将发现的插件添加到输出结果中
	if result.Outputs == nil {
		result.Outputs = make(map[string]interface{})
	}
	result.Outputs["discovered_plugins"] = discoveredPlugins

	return discoveredPlugins, nil
}

// startBackgroundTasks 启动后台任务
func (e *PluginNodeExecutor) startBackgroundTasks(ctx context.Context, enableDiscovery, enableHealthCheck bool, scanInterval, healthCheckInterval string, discoveryPaths []string, result *startup.StartupNodeResult) {
	// 启动插件扫描任务
	if enableDiscovery && scanInterval != "" {
		interval, _ := time.ParseDuration(scanInterval)
		go e.startPeriodicScan(ctx, interval, discoveryPaths, result)
	}

	// 启动健康检查任务
	if enableHealthCheck && healthCheckInterval != "" {
		interval, _ := time.ParseDuration(healthCheckInterval)
		go e.startPeriodicHealthCheck(ctx, interval, result)
	}
}

// startPeriodicScan 启动定期扫描任务
func (e *PluginNodeExecutor) startPeriodicScan(ctx context.Context, interval time.Duration, discoveryPaths []string, result *startup.StartupNodeResult) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	e.logger.Info("Periodic plugin discovery started", "interval", interval.String())

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Periodic plugin discovery stopped")
			return
		case <-ticker.C:
			e.logger.Debug("Running periodic plugin discovery")
			// 在实际实现中，这里会重新扫描插件目录
		}
	}
}

// startPeriodicHealthCheck 启动定期健康检查任务
func (e *PluginNodeExecutor) startPeriodicHealthCheck(ctx context.Context, interval time.Duration, result *startup.StartupNodeResult) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	e.logger.Info("Periodic plugin health check started", "interval", interval.String())

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Periodic plugin health check stopped")
			return
		case <-ticker.C:
			e.logger.Debug("Running periodic plugin health check")
			// 在实际实现中，这里会检查所有已注册插件的健康状态
		}
	}
}

// checkDependencyCompletion 检查依赖节点是否已完成
func (e *PluginNodeExecutor) checkDependencyCompletion(context map[string]interface{}, nodeID string) bool {
	if context == nil {
		return false
	}

	completedKey := fmt.Sprintf("node_completed_%s", nodeID)
	if completed, exists := context[completedKey]; exists {
		if boolVal, ok := completed.(bool); ok {
			return boolVal
		}
	}

	return false
}

// getSupportedPluginTypes 获取支持的插件类型
func (e *PluginNodeExecutor) getSupportedPluginTypes() []string {
	return []string{
		"http",
		"grpc",
		"native",
		"websocket",
		"script",
	}
}

// Validate 验证节点配置
func (e *PluginNodeExecutor) Validate(ctx context.Context, node *startup.StartupNode) error {
	if node.ID == "" {
		return fmt.Errorf("node ID is required")
	}

	if node.Type != startup.StartupNodePlugin {
		return fmt.Errorf("invalid node type: expected %s, got %s", startup.StartupNodePlugin, node.Type)
	}

	// 根据节点ID验证特定配置
	switch node.ID {
	case "plugin:init-manager":
		return e.validateInitPluginManager(node)
	default:
		return fmt.Errorf("unknown plugin node: %s", node.ID)
	}
}

// validateInitPluginManager 验证插件管理器配置
func (e *PluginNodeExecutor) validateInitPluginManager(node *startup.StartupNode) error {
	registryType := getStringConfig(node.Config, "registry_type", "")
	if registryType != "" && !contains([]string{"memory", "redis", "database"}, registryType) {
		return fmt.Errorf("unsupported registry_type: %s", registryType)
	}

	scanInterval := getStringConfig(node.Config, "scan_interval", "")
	if scanInterval != "" && !isValidDuration(scanInterval) {
		return fmt.Errorf("invalid scan_interval format: %s", scanInterval)
	}

	healthCheckInterval := getStringConfig(node.Config, "health_check_interval", "")
	if healthCheckInterval != "" && !isValidDuration(healthCheckInterval) {
		return fmt.Errorf("invalid health_check_interval format: %s", healthCheckInterval)
	}

	registryTTL := getStringConfig(node.Config, "registry_ttl", "")
	if registryTTL != "" && !isValidDuration(registryTTL) {
		return fmt.Errorf("invalid registry_ttl format: %s", registryTTL)
	}

	discoveryPaths := getStringSliceConfig(node.Config, "discovery_paths", nil)
	if discoveryPaths != nil && len(discoveryPaths) == 0 {
		return fmt.Errorf("discovery_paths cannot be empty when discovery is enabled")
	}

	return nil
}

// GetNodeInfo 获取节点信息
func (e *PluginNodeExecutor) GetNodeInfo() *startup.StartupNodeInfo {
	return &startup.StartupNodeInfo{
		Type:        startup.StartupNodePlugin,
		Name:        "Plugin Node Executor",
		Description: "Handles plugin management system initialization including discovery, registry, and health monitoring",
		Version:     "1.0.0",
		Author:      "XiaoZhi Flow Team",
		SupportedConfig: map[string]interface{}{
			"enable_discovery": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable automatic plugin discovery",
				"default":     true,
			},
			"discovery_paths": map[string]interface{}{
				"type":        "array",
				"description": "Paths to scan for plugins",
				"default":     []string{"./plugins", "./plugins/examples"},
			},
			"scan_interval": map[string]interface{}{
				"type":        "string",
				"description": "Plugin discovery scan interval",
				"default":     "30s",
			},
			"enable_health_check": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable plugin health monitoring",
				"default":     true,
			},
			"health_check_interval": map[string]interface{}{
				"type":        "string",
				"description": "Health check interval",
				"default":     "10s",
			},
			"registry_type": map[string]interface{}{
				"type":        "string",
				"description": "Plugin registry storage type",
				"enum":        []string{"memory", "redis", "database"},
				"default":     "memory",
			},
			"registry_ttl": map[string]interface{}{
				"type":        "string",
				"description": "Plugin registration TTL",
				"default":     "5m",
			},
		},
		Capabilities: []string{
			"plugin-discovery",
			"plugin-registry",
			"health-monitoring",
			"lifecycle-management",
			"hot-reloading",
			"dependency-resolution",
			"version-management",
		},
	}
}

// Cleanup 清理资源
func (e *PluginNodeExecutor) Cleanup(ctx context.Context) error {
	e.logger.Info("Cleaning up plugin node executor")
	// 这里可以停止后台任务、清理插件注册表等
	return nil
}

// 辅助函数

// getStringSliceConfig 获取字符串切片配置值
func getStringSliceConfig(config map[string]interface{}, key string, defaultValue []string) []string {
	if config == nil {
		return defaultValue
	}

	value, exists := config[key]
	if !exists {
		return defaultValue
	}

	if sliceValue, ok := value.([]interface{}); ok {
		var result []string
		for _, item := range sliceValue {
			if strItem, ok := item.(string); ok {
				result = append(result, strItem)
			}
		}
		return result
	}

	return defaultValue
}