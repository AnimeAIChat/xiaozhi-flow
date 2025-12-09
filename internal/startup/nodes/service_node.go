package nodes

import (
	"context"
	"fmt"
	"time"

	"xiaozhi-server-go/internal/startup/model"
	"xiaozhi-server-go/internal/workflow"
)

// ServiceNodeExecutor 服务节点执行器
type ServiceNodeExecutor struct {
	logger model.StartupLogger
}

// NewServiceNodeExecutor 创建服务节点执行器
func NewServiceNodeExecutor(logger model.StartupLogger) *ServiceNodeExecutor {
	return &ServiceNodeExecutor{
		logger: logger,
	}
}

// Execute 执行服务节点
func (e *ServiceNodeExecutor) Execute(
	ctx context.Context,
	node *model.StartupNode,
	inputs map[string]interface{},
	context map[string]interface{},
) (*model.StartupNodeResult, error) {
	startTime := time.Now()
	result := &model.StartupNodeResult{
		NodeID:   node.ID,
		NodeName: node.Name,
		NodeType: node.Type,
		StartTime: startTime,
		Status:   workflow.NodeStatusRunning,
		Inputs:   inputs,
		Outputs:  make(map[string]interface{}),
		Logs:     make([]model.StartupNodeLog, 0),
	}

	e.logger.Info("Executing service node", "node_id", node.ID, "node_name", node.Name)

	// 根据节点ID执行不同的服务操作
	switch node.ID {
	case "logging:init-provider":
		err := e.executeInitLogging(ctx, node, result)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	case "components:init-container":
		err := e.executeInitComponents(ctx, node, result)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	case "mcp:init-manager":
		err := e.executeInitMCPManager(ctx, node, result)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	case "observability:setup-hooks":
		err := e.executeSetupObservability(ctx, node, result)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	case "start-services":
		err := e.executeStartServices(ctx, node, result)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	default:
		err := fmt.Errorf("unknown service node: %s", node.ID)
		result.Status = workflow.NodeStatusFailed
		result.Error = err.Error()
		return result, err
	}

	// 成功完成
	endTime := time.Now()
	result.EndTime = &endTime
	result.Duration = endTime.Sub(startTime)
	result.Status = workflow.NodeStatusCompleted

	e.logger.Info("Service node completed successfully",
		"node_id", node.ID,
		"duration", result.Duration.String())

	return result, nil
}

// executeInitLogging 执行初始化日志系统
func (e *ServiceNodeExecutor) executeInitLogging(
	ctx context.Context,
	node *model.StartupNode,
	result *model.StartupNodeResult,
) error {
	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting logging provider initialization",
	})

	// 获取配置
	logLevel := getStringConfig(node.Config, "log_level", "info")
	logDir := getStringConfig(node.Config, "log_dir", "logs/")
	logFile := getStringConfig(node.Config, "log_file", "xiaozhi-server.log")
	maxSize := getStringConfig(node.Config, "max_size", "100MB")
	maxBackups := getIntConfig(node.Config, "max_backups", 10)
	compress := getBoolConfig(node.Config, "compress", true)

	e.logger.Info("Initializing logging provider",
		"log_level", logLevel,
		"log_dir", logDir,
		"log_file", logFile,
		"max_size", maxSize,
		"max_backups", maxBackups,
		"compress", compress)

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Logging configuration: level=%s, dir=%s, file=%s", logLevel, logDir, logFile),
	})

	// 模拟日志系统初始化
	// 在实际实现中，这里会：
	// 1. 创建日志目录
	// 2. 初始化日志提供者
	// 3. 设置日志级别
	// 4. 配置日志轮转
	// 5. 设置事件总线处理器

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Logging provider initialized successfully",
	})

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"log_level":         logLevel,
		"log_dir":          logDir,
		"log_file":         logFile,
		"max_size":         maxSize,
		"max_backups":      maxBackups,
		"compress_enabled": compress,
		"logger_id":        "logging-provider-" + node.ID,
		"initialized_at":   time.Now(),
	}

	return nil
}

// executeInitComponents 执行初始化组件容器
func (e *ServiceNodeExecutor) executeInitComponents(
	ctx context.Context,
	node *model.StartupNode,
	result *model.StartupNodeResult,
) error {
	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting components container initialization",
	})

	// 获取配置
	enableDIContainer := getBoolConfig(node.Config, "enable_di_container", false)
	enableRuntimeContainer := getBoolConfig(node.Config, "enable_runtime_container", false)

	e.logger.Info("Initializing components container",
		"enable_di_container", enableDIContainer,
		"enable_runtime_container", enableRuntimeContainer)

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Components configuration: DI=%v, Runtime=%v", enableDIContainer, enableRuntimeContainer),
	})

	// 模拟组件容器初始化
	// 在实际实现中，这里会：
	// 1. 创建BootstrapManager实例
	// 2. 初始化依赖注入容器
	// 3. 设置运行时容器
	// 4. 初始化组件适配器
	// 5. 设置组件生命周期管理

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Bootstrap manager created",
	})

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Component container initialized successfully",
	})

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"di_container_enabled":     enableDIContainer,
		"runtime_container_enabled": enableRuntimeContainer,
		"container_id":            "components-container-" + node.ID,
		"bootstrap_manager_id":    "bootstrap-manager-" + node.ID,
		"initialized_at":          time.Now(),
	}

	return nil
}

// executeInitMCPManager 执行初始化MCP管理器
func (e *ServiceNodeExecutor) executeInitMCPManager(
	ctx context.Context,
	node *model.StartupNode,
	result *model.StartupNodeResult,
) error {
	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting MCP manager initialization",
	})

	// 获取配置
	enableGlobalManager := getBoolConfig(node.Config, "enable_global_manager", true)
	enableLocalClients := getBoolConfig(node.Config, "enable_local_clients", true)
	toolTimeout := getStringConfig(node.Config, "tool_timeout", "30s")
	maxRetries := getIntConfig(node.Config, "max_retries", 3)

	e.logger.Info("Initializing MCP manager",
		"enable_global_manager", enableGlobalManager,
		"enable_local_clients", enableLocalClients,
		"tool_timeout", toolTimeout,
		"max_retries", maxRetries)

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("MCP configuration: global=%v, local=%v, timeout=%s", enableGlobalManager, enableLocalClients, toolTimeout),
	})

	// 模拟MCP管理器初始化
	// 在实际实现中，这里会：
	// 1. 初始化全局MCP工具管理器
	// 2. 创建域特定的MCP管理器
	// 3. 设置MCP客户端连接
	// 4. 配置工具超时和重试策略

	if enableGlobalManager {
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Global MCP manager initialized",
		})
	}

	if enableLocalClients {
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Local MCP clients initialized",
		})
	}

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"global_manager_enabled": enableGlobalManager,
		"local_clients_enabled":  enableLocalClients,
		"tool_timeout":          toolTimeout,
		"max_retries":           maxRetries,
		"mcp_manager_id":        "mcp-manager-" + node.ID,
		"initialized_at":        time.Now(),
	}

	return nil
}

// executeSetupObservability 执行设置可观测性钩子
func (e *ServiceNodeExecutor) executeSetupObservability(
	ctx context.Context,
	node *model.StartupNode,
	result *model.StartupNodeResult,
) error {
	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting observability setup",
	})

	// 获取配置
	enableMetrics := getBoolConfig(node.Config, "enable_metrics", true)
	enableTracing := getBoolConfig(node.Config, "enable_tracing", false)
	metricsEndpoint := getStringConfig(node.Config, "metrics_endpoint", "/metrics")
	tracingEndpoint := getStringConfig(node.Config, "tracing_endpoint", "")

	e.logger.Info("Setting up observability hooks",
		"enable_metrics", enableMetrics,
		"enable_tracing", enableTracing,
		"metrics_endpoint", metricsEndpoint,
		"tracing_endpoint", tracingEndpoint)

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Observability configuration: metrics=%v, tracing=%v", enableMetrics, enableTracing),
	})

	// 模拟可观测性设置
	// 在实际实现中，这里会：
	// 1. 设置指标收集
	// 2. 配置分布式追踪
	// 3. 设置性能监控钩子
	// 4. 初始化监控仪表板

	if enableMetrics {
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Metrics collection enabled at %s", metricsEndpoint),
		})
	}

	if enableTracing && tracingEndpoint != "" {
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Distributed tracing enabled with endpoint: %s", tracingEndpoint),
		})
	}

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"metrics_enabled":       enableMetrics,
		"tracing_enabled":       enableTracing,
		"metrics_endpoint":      metricsEndpoint,
		"tracing_endpoint":      tracingEndpoint,
		"observability_id":      "observability-" + node.ID,
		"initialized_at":        time.Now(),
	}

	return nil
}

// executeStartServices 执行启动系统服务
func (e *ServiceNodeExecutor) executeStartServices(
	ctx context.Context,
	node *model.StartupNode,
	result *model.StartupNodeResult,
) error {
	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting system services",
	})

	// 获取配置
	startHTTPServer := getBoolConfig(node.Config, "start_http_server", true)
	startWebSocketServer := getBoolConfig(node.Config, "start_websocket_server", true)
	startOTAService := getBoolConfig(node.Config, "start_ota_service", true)
	httpPort := getIntConfig(node.Config, "http_port", 8080)
	websocketPort := getIntConfig(node.Config, "websocket_port", 8001)
	enableStaticServing := getBoolConfig(node.Config, "enable_static_serving", true)
	enableAPIDocs := getBoolConfig(node.Config, "enable_api_docs", true)

	e.logger.Info("Starting system services",
		"http_server", startHTTPServer,
		"websocket_server", startWebSocketServer,
		"ota_service", startOTAService,
		"http_port", httpPort,
		"websocket_port", websocketPort)

	// 模拟服务启动过程
	// 在实际实现中，这里会：
	// 1. 启动HTTP服务器（API端点、静态文件服务）
	// 2. 启动WebSocket服务器
	// 3. 启动OTA服务
	// 4. 设置优雅关闭处理

	if startHTTPServer {
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Starting HTTP server on port %d", httpPort),
		})

		if enableStaticServing {
			result.Logs = append(result.Logs, model.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "Static file serving enabled",
			})
		}

		if enableAPIDocs {
			result.Logs = append(result.Logs, model.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "API documentation enabled",
			})
		}

		// 模拟HTTP服务器启动时间
		select {
		case <-time.After(1 * time.Second):
			result.Logs = append(result.Logs, model.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   fmt.Sprintf("HTTP server started successfully on port %d", httpPort),
			})
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if startWebSocketServer {
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Starting WebSocket server on port %d", websocketPort),
		})

		// 模拟WebSocket服务器启动时间
		select {
		case <-time.After(500 * time.Millisecond):
			result.Logs = append(result.Logs, model.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   fmt.Sprintf("WebSocket server started successfully on port %d", websocketPort),
			})
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if startOTAService {
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Starting OTA service",
		})

		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "OTA service started successfully",
		})
	}

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"http_server_enabled":       startHTTPServer,
		"websocket_server_enabled": startWebSocketServer,
		"ota_service_enabled":       startOTAService,
		"http_port":                httpPort,
		"websocket_port":           websocketPort,
		"static_serving_enabled":   enableStaticServing,
		"api_docs_enabled":         enableAPIDocs,
		"services_started_at":      time.Now(),
		"service_status":           "running",
	}

	return nil
}

// Validate 验证节点配置
func (e *ServiceNodeExecutor) Validate(ctx context.Context, node *model.StartupNode) error {
	if node.ID == "" {
		return fmt.Errorf("node ID is required")
	}

	if node.Type != model.StartupNodeService {
		return fmt.Errorf("invalid node type: expected %s, got %s", model.StartupNodeService, node.Type)
	}

	// 根据节点ID验证特定配置
	switch node.ID {
	case "logging:init-provider":
		return e.validateLoggingConfig(node)
	case "components:init-container":
		return e.validateComponentsConfig(node)
	case "mcp:init-manager":
		return e.validateMCPConfig(node)
	case "observability:setup-hooks":
		return e.validateObservabilityConfig(node)
	case "start-services":
		return e.validateStartServicesConfig(node)
	default:
		return fmt.Errorf("unknown service node: %s", node.ID)
	}
}

// validateLoggingConfig 验证日志配置
func (e *ServiceNodeExecutor) validateLoggingConfig(node *model.StartupNode) error {
	logLevel := getStringConfig(node.Config, "log_level", "")
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if logLevel != "" && !contains(validLogLevels, logLevel) {
		return fmt.Errorf("invalid log_level: %s", logLevel)
	}

	maxBackups := getIntConfig(node.Config, "max_backups", -1)
	if maxBackups < 0 || maxBackups > 100 {
		return fmt.Errorf("max_backups must be between 0 and 100")
	}

	return nil
}

// validateComponentsConfig 验证组件配置
func (e *ServiceNodeExecutor) validateComponentsConfig(node *model.StartupNode) error {
	// 组件配置相对简单，主要是布尔值配置
	return nil
}

// validateMCPConfig 验证MCP配置
func (e *ServiceNodeExecutor) validateMCPConfig(node *model.StartupNode) error {
	maxRetries := getIntConfig(node.Config, "max_retries", -1)
	if maxRetries < 0 || maxRetries > 10 {
		return fmt.Errorf("max_retries must be between 0 and 10")
	}

	return nil
}

// validateObservabilityConfig 验证可观测性配置
func (e *ServiceNodeExecutor) validateObservabilityConfig(node *model.StartupNode) error {
	metricsEndpoint := getStringConfig(node.Config, "metrics_endpoint", "")
	if metricsEndpoint != "" && !isValidEndpoint(metricsEndpoint) {
		return fmt.Errorf("invalid metrics_endpoint format: %s", metricsEndpoint)
	}

	tracingEndpoint := getStringConfig(node.Config, "tracing_endpoint", "")
	if tracingEndpoint != "" && !isValidEndpoint(tracingEndpoint) {
		return fmt.Errorf("invalid tracing_endpoint format: %s", tracingEndpoint)
	}

	return nil
}

// validateStartServicesConfig 验证启动服务配置
func (e *ServiceNodeExecutor) validateStartServicesConfig(node *model.StartupNode) error {
	httpPort := getIntConfig(node.Config, "http_port", -1)
	if httpPort != -1 && (httpPort < 1024 || httpPort > 65535) {
		return fmt.Errorf("http_port must be between 1024 and 65535")
	}

	websocketPort := getIntConfig(node.Config, "websocket_port", -1)
	if websocketPort != -1 && (websocketPort < 1024 || websocketPort > 65535) {
		return fmt.Errorf("websocket_port must be between 1024 and 65535")
	}

	if httpPort != -1 && websocketPort != -1 && httpPort == websocketPort {
		return fmt.Errorf("http_port and websocket_port must be different")
	}

	return nil
}

// GetNodeInfo 获取节点信息
func (e *ServiceNodeExecutor) GetNodeInfo() *model.StartupNodeInfo {
	return &model.StartupNodeInfo{
		Type:        model.StartupNodeService,
		Name:        "Service Node Executor",
		Description: "Handles service-related initialization tasks including logging, components, MCP, observability, and system services",
		Version:     "1.0.0",
		Author:      "XiaoZhi Flow Team",
		SupportedConfig: map[string]interface{}{
			"log_level": map[string]interface{}{
				"type":        "string",
				"description": "Logging level",
				"enum":        []string{"debug", "info", "warn", "error"},
				"default":     "info",
			},
			"http_port": map[string]interface{}{
				"type":        "integer",
				"description": "HTTP server port",
				"min":         1024,
				"max":         65535,
				"default":     8080,
			},
			"websocket_port": map[string]interface{}{
				"type":        "integer",
				"description": "WebSocket server port",
				"min":         1024,
				"max":         65535,
				"default":     8001,
			},
		},
		Capabilities: []string{
			"logging-provider",
			"components-container",
			"mcp-management",
			"observability-hooks",
			"system-services",
		},
	}
}

// Cleanup 清理资源
func (e *ServiceNodeExecutor) Cleanup(ctx context.Context) error {
	e.logger.Info("Cleaning up service node executor")
	// 这里可以停止服务、关闭连接等
	return nil
}

// 辅助函数

// contains 检查字符串切片是否包含指定值
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// isValidEndpoint 检查端点格式是否有效
func isValidEndpoint(endpoint string) bool {
	// 简单的端点格式验证
	// 在实际实现中可以使用更严格的验证
	return len(endpoint) > 0 && (endpoint[0] == '/' || endpoint[0] == 'h' || endpoint[0] == 'w')
}



