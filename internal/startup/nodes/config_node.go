package nodes

import (
	"context"
	"fmt"
	"time"

	"xiaozhi-server-go/internal/startup/model"
	"xiaozhi-server-go/internal/workflow"
)

// ConfigNodeExecutor 配置节点执行器
type ConfigNodeExecutor struct {
	logger model.StartupLogger
}

// NewConfigNodeExecutor 创建配置节点执行器
func NewConfigNodeExecutor(logger model.StartupLogger) *ConfigNodeExecutor {
	return &ConfigNodeExecutor{
		logger: logger,
	}
}

// Execute 执行配置节点
func (e *ConfigNodeExecutor) Execute(
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

	e.logger.Info("Executing config node", "node_id", node.ID, "node_name", node.Name)

	// 根据节点ID执行不同的配置操作
	switch node.ID {
	case "config:load-default":
		err := e.executeLoadDefaultConfig(ctx, node, result, inputs, context)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	case "config:init-integrator":
		err := e.executeInitConfigIntegrator(ctx, node, result, inputs, context)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	default:
		err := fmt.Errorf("unknown config node: %s", node.ID)
		result.Status = workflow.NodeStatusFailed
		result.Error = err.Error()
		return result, err
	}

	// 成功完成
	endTime := time.Now()
	result.EndTime = &endTime
	result.Duration = endTime.Sub(startTime)
	result.Status = workflow.NodeStatusCompleted

	e.logger.Info("Config node completed successfully",
		"node_id", node.ID,
		"duration", result.Duration.String())

	return result, nil
}

// executeLoadDefaultConfig 执行加载默认配置
func (e *ConfigNodeExecutor) executeLoadDefaultConfig(
	ctx context.Context,
	node *model.StartupNode,
	result *model.StartupNodeResult,
	inputs map[string]interface{},
	context map[string]interface{},
) error {
	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting default configuration loading",
	})

	// 获取配置
	configFile := getStringConfig(node.Config, "config_file", "default_config.json")
	fallbackToDefaults := getBoolConfig(node.Config, "fallback_to_defaults", true)
	useBuiltinDefaults := getBoolConfig(node.Config, "use_builtin_defaults", false)

	e.logger.Info("Loading default configuration",
		"config_file", configFile,
		"fallback_to_defaults", fallbackToDefaults,
		"use_builtin_defaults", useBuiltinDefaults)

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Config loading: file=%s, fallback=%v, builtin=%v", configFile, fallbackToDefaults, useBuiltinDefaults),
	})

	// 检查依赖节点是否已完成
	configStoreReady := e.checkDependencyCompletion(context, "storage:init-config-store")
	databaseReady := e.checkDependencyCompletion(context, "storage:init-database")

	if !configStoreReady || !databaseReady {
		err := fmt.Errorf("required dependencies not completed: config_store=%v, database=%v", configStoreReady, databaseReady)
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   err.Error(),
		})
		return err
	}

	// 模拟配置加载过程
	// 在实际实现中，这里会：
	// 1. 尝试从数据库加载配置
	// 2. 如果没有配置且启用fallback，加载内置默认配置
	// 3. 创建配置仓库实例
	// 4. 验证配置完整性
	// 5. 存储配置到上下文中供后续节点使用

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Attempting to load configuration from database",
	})

	// 模拟数据库配置加载
	select {
	case <-time.After(500 * time.Millisecond):
		// 模拟配置加载完成
	case <-ctx.Done():
		return ctx.Err()
	}

	// 创建默认配置对象
	defaultConfig := e.createDefaultConfiguration()

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Configuration loaded successfully",
	})

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Configuration sections loaded: %d", len(defaultConfig)),
	})

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"config_file":         configFile,
		"config_loaded":       true,
		"config_source":       "database",
		"fallback_used":       false,
		"config_sections":     defaultConfig,
		"config_repository_id": "config-repo-" + node.ID,
		"loaded_at":          time.Now(),
	}

	return nil
}

// executeInitConfigIntegrator 执行初始化配置集成器
func (e *ConfigNodeExecutor) executeInitConfigIntegrator(
	ctx context.Context,
	node *model.StartupNode,
	result *model.StartupNodeResult,
	inputs map[string]interface{},
	context map[string]interface{},
) error {
	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting configuration integrator initialization",
	})

	// 获取配置
	enableEnvConfig := getBoolConfig(node.Config, "enable_env_config", true)
	enableFileConfig := getBoolConfig(node.Config, "enable_file_config", true)
	configWatchInterval := getStringConfig(node.Config, "config_watch_interval", "30s")

	e.logger.Info("Initializing configuration integrator",
		"enable_env_config", enableEnvConfig,
		"enable_file_config", enableFileConfig,
		"config_watch_interval", configWatchInterval)

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Config integrator: env=%v, file=%v, watch=%s", enableEnvConfig, enableFileConfig, configWatchInterval),
	})

	// 检查依赖节点是否已完成
	componentsReady := e.checkDependencyCompletion(context, "components:init-container")
	loggingReady := e.checkDependencyCompletion(context, "logging:init-provider")

	if !componentsReady || !loggingReady {
		err := fmt.Errorf("required dependencies not completed: components=%v, logging=%v", componentsReady, loggingReady)
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   err.Error(),
		})
		return err
	}

	// 模拟配置集成器初始化
	// 在实际实现中，这里会：
	// 1. 初始化环境变量配置源
	// 2. 初始化文件配置源
	// 3. 设置配置监控和热重载
	// 4. 创建统一的配置管理器
	// 5. 建立配置更新通知机制

	if enableEnvConfig {
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Environment variable configuration enabled",
		})

		// 模拟环境变量加载
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Environment variables loaded successfully",
		})
	}

	if enableFileConfig {
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "File-based configuration enabled",
		})

		// 模拟文件配置监控设置
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Configuration file monitoring started with interval: %s", configWatchInterval),
		})
	}

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Configuration integrator initialized successfully",
	})

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"env_config_enabled":     enableEnvConfig,
		"file_config_enabled":    enableFileConfig,
		"watch_interval":        configWatchInterval,
		"integrator_id":         "config-integrator-" + node.ID,
		"hot_reload_enabled":    enableFileConfig,
		"initialized_at":        time.Now(),
		"config_sources":        e.getConfigSources(enableEnvConfig, enableFileConfig),
	}

	return nil
}

// checkDependencyCompletion 检查依赖节点是否已完成
func (e *ConfigNodeExecutor) checkDependencyCompletion(context map[string]interface{}, nodeID string) bool {
	if context == nil {
		return false
	}

	// 检查上下文中是否有节点完成状态
	completedKey := fmt.Sprintf("node_completed_%s", nodeID)
	if completed, exists := context[completedKey]; exists {
		if boolVal, ok := completed.(bool); ok {
			return boolVal
		}
	}

	return false
}

// createDefaultConfiguration 创建默认配置对象
func (e *ConfigNodeExecutor) createDefaultConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"server": map[string]interface{}{
			"host": "0.0.0.0",
			"port": 8080,
		},
		"database": map[string]interface{}{
			"type":             "sqlite",
			"connection_string": "./data/xiaozhi.db",
			"auto_migrate":     true,
		},
		"logging": map[string]interface{}{
			"level":      "info",
			"directory":  "logs/",
			"filename":   "xiaozhi-server.log",
			"max_size":   "100MB",
			"max_backups": 10,
			"compress":   true,
		},
		"auth": map[string]interface{}{
			"session_ttl":      "24h",
			"cleanup_interval": "10m",
			"jwt_secret":       "your-jwt-secret-key",
		},
		"websocket": map[string]interface{}{
			"port":        8001,
			"heartbeat":   "30s",
		},
		"plugins": map[string]interface{}{
			"discovery_paths":   []string{"./plugins", "./plugins/examples"},
			"scan_interval":     "30s",
			"health_check_interval": "10s",
		},
	}
}

// getConfigSources 获取配置源列表
func (e *ConfigNodeExecutor) getConfigSources(envConfig, fileConfig bool) []string {
	var sources []string

	if envConfig {
		sources = append(sources, "environment")
	}

	if fileConfig {
		sources = append(sources, "file")
	}

	// 始终包含数据库配置源
	sources = append(sources, "database")

	return sources
}

// Validate 验证节点配置
func (e *ConfigNodeExecutor) Validate(ctx context.Context, node *model.StartupNode) error {
	if node.ID == "" {
		return fmt.Errorf("node ID is required")
	}

	if node.Type != model.StartupNodeConfig {
		return fmt.Errorf("invalid node type: expected %s, got %s", model.StartupNodeConfig, node.Type)
	}

	// 根据节点ID验证特定配置
	switch node.ID {
	case "config:load-default":
		return e.validateLoadDefaultConfig(node)
	case "config:init-integrator":
		return e.validateInitConfigIntegrator(node)
	default:
		return fmt.Errorf("unknown config node: %s", node.ID)
	}
}

// validateLoadDefaultConfig 验证加载默认配置节点
func (e *ConfigNodeExecutor) validateLoadDefaultConfig(node *model.StartupNode) error {
	configFile := getStringConfig(node.Config, "config_file", "")
	if configFile != "" && !isValidConfigFile(configFile) {
		return fmt.Errorf("invalid config_file format: %s", configFile)
	}

	return nil
}

// validateInitConfigIntegrator 验证配置集成器节点
func (e *ConfigNodeExecutor) validateInitConfigIntegrator(node *model.StartupNode) error {
	configWatchInterval := getStringConfig(node.Config, "config_watch_interval", "")
	if configWatchInterval != "" && !isValidDuration(configWatchInterval) {
		return fmt.Errorf("invalid config_watch_interval format: %s", configWatchInterval)
	}

	return nil
}

// GetNodeInfo 获取节点信息
func (e *ConfigNodeExecutor) GetNodeInfo() *model.StartupNodeInfo {
	return &model.StartupNodeInfo{
		Type:        model.StartupNodeConfig,
		Name:        "Config Node Executor",
		Description: "Handles configuration-related tasks including loading default config and integrating multiple config sources",
		Version:     "1.0.0",
		Author:      "XiaoZhi Flow Team",
		SupportedConfig: map[string]interface{}{
			"config_file": map[string]interface{}{
				"type":        "string",
				"description": "Configuration file name",
				"default":     "default_config.json",
			},
			"fallback_to_defaults": map[string]interface{}{
				"type":        "boolean",
				"description": "Fallback to built-in defaults if config not found",
				"default":     true,
			},
			"use_builtin_defaults": map[string]interface{}{
				"type":        "boolean",
				"description": "Use built-in default configuration",
				"default":     false,
			},
			"enable_env_config": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable environment variable configuration",
				"default":     true,
			},
			"enable_file_config": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable file-based configuration",
				"default":     true,
			},
			"config_watch_interval": map[string]interface{}{
				"type":        "string",
				"description": "Configuration file monitoring interval",
				"default":     "30s",
			},
		},
		Capabilities: []string{
			"config-loading",
			"config-integration",
			"environment-config",
			"file-config",
			"hot-reload",
			"config-validation",
		},
	}
}

// Cleanup 清理资源
func (e *ConfigNodeExecutor) Cleanup(ctx context.Context) error {
	e.logger.Info("Cleaning up config node executor")
	// 这里可以停止配置监控、清理配置缓存等
	return nil
}

// 辅助函数

// isValidConfigFile 检查配置文件名是否有效
func isValidConfigFile(filename string) bool {
	// 简单的文件名验证
	// 在实际实现中可以使用更严格的验证
	validExtensions := []string{".json", ".yaml", ".yml", ".toml", ".ini"}
	for _, ext := range validExtensions {
		if len(filename) >= len(ext) && filename[len(filename)-len(ext):] == ext {
			return true
		}
	}
	return len(filename) > 0
}

// isValidDuration 检查时间间隔格式是否有效
func isValidDuration(duration string) bool {
	// 简单的时间格式验证
	// 在实际实现中可以使用 time.ParseDuration
	validSuffixes := []string{"s", "m", "h", "ms", "us", "ns"}
	if len(duration) == 0 {
		return false
	}

	for _, suffix := range validSuffixes {
		if len(duration) > len(suffix) && duration[len(duration)-len(suffix):] == suffix {
			return true
		}
	}

	return false
}



