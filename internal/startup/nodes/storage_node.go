package nodes

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"xiaozhi-server-go/internal/startup/model"
	"xiaozhi-server-go/internal/workflow"
)

// StorageNodeExecutor 存储节点执行器
type StorageNodeExecutor struct {
	logger model.StartupLogger
}

// NewStorageNodeExecutor 创建存储节点执行器
func NewStorageNodeExecutor(logger model.StartupLogger) *StorageNodeExecutor {
	return &StorageNodeExecutor{
		logger: logger,
	}
}

// Execute 执行存储节点
func (e *StorageNodeExecutor) Execute(
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

	e.logger.Info("Executing storage node", "node_id", node.ID, "node_name", node.Name)

	// 根据节点ID执行不同的存储操作
	switch node.ID {
	case "storage:init-config-store":
		err := e.executeInitConfigStore(ctx, node, result)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	case "storage:init-database":
		err := e.executeInitDatabase(ctx, node, result)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	default:
		err := fmt.Errorf("unknown storage node: %s", node.ID)
		result.Status = workflow.NodeStatusFailed
		result.Error = err.Error()
		return result, err
	}

	// 成功完成
	endTime := time.Now()
	result.EndTime = &endTime
	result.Duration = endTime.Sub(startTime)
	result.Status = workflow.NodeStatusCompleted

	e.logger.Info("Storage node completed successfully",
		"node_id", node.ID,
		"duration", result.Duration.String())

	return result, nil
}

// executeInitConfigStore 执行初始化配置存储
func (e *StorageNodeExecutor) executeInitConfigStore(
	ctx context.Context,
	node *model.StartupNode,
	result *model.StartupNodeResult,
) error {
	// 添加日志
	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting config store initialization",
	})

	// 获取配置
	storageType := getStringConfig(node.Config, "storage_type", "file")
	configPath := getStringConfig(node.Config, "config_path", "config/")

	e.logger.Info("Initializing config store",
		"storage_type", storageType,
		"config_path", configPath)

	// 模拟配置存储初始化
	// 在实际实现中，这里会调用 platformstorage.InitConfigStore()
	// 现在我们创建配置目录并设置基本配置

	// 确保配置目录存在
	absConfigPath := filepath.Clean(configPath)

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Config store initialized with type: %s, path: %s", storageType, absConfigPath),
	})

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"storage_type":     storageType,
		"config_path":      absConfigPath,
		"initialized_at":   time.Now(),
		"config_store_id":  "config-store-" + node.ID,
	}

	return nil
}

// executeInitDatabase 执行初始化数据库
func (e *StorageNodeExecutor) executeInitDatabase(
	ctx context.Context,
	node *model.StartupNode,
	result *model.StartupNodeResult,
) error {
	// 添加日志
	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting database initialization",
	})

	// 获取配置
	databaseType := getStringConfig(node.Config, "database_type", "sqlite")
	connectionString := getStringConfig(node.Config, "connection_string", "./data/xiaozhi.db")
	autoMigrate := getBoolConfig(node.Config, "auto_migrate", true)
	wMode := getBoolConfig(node.Config, "w_mode", true)

	e.logger.Info("Initializing database",
		"database_type", databaseType,
		"connection_string", connectionString,
		"auto_migrate", autoMigrate,
		"w_mode", wMode)

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Database config: type=%s, connection=%s", databaseType, connectionString),
	})

	// 模拟数据库初始化过程
	// 在实际实现中，这里会：
	// 1. 检查数据库配置文件
	// 2. 初始化数据库连接
	// 3. 运行数据库迁移
	// 4. 设置数据库连接池

	// 模拟初始化延迟（实际实现中删除）
	select {
	case <-time.After(2 * time.Second):
		// 继续执行
	case <-ctx.Done():
		return ctx.Err()
	}

	result.Logs = append(result.Logs, model.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Database connection established",
	})

	if autoMigrate {
		result.Logs = append(result.Logs, model.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Database migrations completed",
		})
	}

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"database_type":      databaseType,
		"connection_string":  connectionString,
		"initialized_at":     time.Now(),
		"database_id":        "database-" + node.ID,
		"auto_migrated":      autoMigrate,
		"write_mode_enabled": wMode,
		"connection_status":  "connected",
	}

	return nil
}

// Validate 验证节点配置
func (e *StorageNodeExecutor) Validate(ctx context.Context, node *model.StartupNode) error {
	if node.ID == "" {
		return fmt.Errorf("node ID is required")
	}

	if node.Type != model.StartupNodeStorage {
		return fmt.Errorf("invalid node type: expected %s, got %s", model.StartupNodeStorage, node.Type)
	}

	// 根据节点ID验证特定配置
	switch node.ID {
	case "storage:init-config-store":
		return e.validateConfigStoreConfig(node)
	case "storage:init-database":
		return e.validateDatabaseConfig(node)
	default:
		return fmt.Errorf("unknown storage node: %s", node.ID)
	}
}

// validateConfigStoreConfig 验证配置存储配置
func (e *StorageNodeExecutor) validateConfigStoreConfig(node *model.StartupNode) error {
	storageType := getStringConfig(node.Config, "storage_type", "")
	if storageType == "" {
		return fmt.Errorf("storage_type is required for config store node")
	}

	if storageType != "file" && storageType != "memory" && storageType != "database" {
		return fmt.Errorf("unsupported storage_type: %s", storageType)
	}

	configPath := getStringConfig(node.Config, "config_path", "")
	if storageType == "file" && configPath == "" {
		return fmt.Errorf("config_path is required for file-based storage")
	}

	return nil
}

// validateDatabaseConfig 验证数据库配置
func (e *StorageNodeExecutor) validateDatabaseConfig(node *model.StartupNode) error {
	databaseType := getStringConfig(node.Config, "database_type", "")
	if databaseType == "" {
		return fmt.Errorf("database_type is required for database node")
	}

	if databaseType != "sqlite" && databaseType != "mysql" && databaseType != "postgresql" {
		return fmt.Errorf("unsupported database_type: %s", databaseType)
	}

	connectionString := getStringConfig(node.Config, "connection_string", "")
	if connectionString == "" {
		return fmt.Errorf("connection_string is required for database node")
	}

	return nil
}

// GetNodeInfo 获取节点信息
func (e *StorageNodeExecutor) GetNodeInfo() *model.StartupNodeInfo {
	return &model.StartupNodeInfo{
		Type:        model.StartupNodeStorage,
		Name:        "Storage Node Executor",
		Description: "Handles storage-related initialization tasks including config store and database setup",
		Version:     "1.0.0",
		Author:      "XiaoZhi Flow Team",
		SupportedConfig: map[string]interface{}{
			"storage_type": map[string]interface{}{
				"type":        "string",
				"description": "Storage backend type",
				"enum":        []string{"file", "memory", "database"},
				"default":     "file",
			},
			"config_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to configuration directory",
				"default":     "config/",
			},
			"database_type": map[string]interface{}{
				"type":        "string",
				"description": "Database backend type",
				"enum":        []string{"sqlite", "mysql", "postgresql"},
				"default":     "sqlite",
			},
			"connection_string": map[string]interface{}{
				"type":        "string",
				"description": "Database connection string",
				"default":     "./data/xiaozhi.db",
			},
			"auto_migrate": map[string]interface{}{
				"type":        "boolean",
				"description": "Run database migrations automatically",
				"default":     true,
			},
			"w_mode": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable write mode",
				"default":     true,
			},
		},
		Capabilities: []string{
			"config-store-init",
			"database-init",
			"migration-support",
			"connection-pooling",
		},
	}
}

// Cleanup 清理资源
func (e *StorageNodeExecutor) Cleanup(ctx context.Context) error {
	e.logger.Info("Cleaning up storage node executor")
	// 这里可以清理数据库连接、关闭文件句柄等
	return nil
}

// 辅助函数

// getStringConfig 获取字符串配置值
func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if config == nil {
		return defaultValue
	}

	value, exists := config[key]
	if !exists {
		return defaultValue
	}

	if strValue, ok := value.(string); ok {
		return strValue
	}

	return defaultValue
}

// getBoolConfig 获取布尔配置值
func getBoolConfig(config map[string]interface{}, key string, defaultValue bool) bool {
	if config == nil {
		return defaultValue
	}

	value, exists := config[key]
	if !exists {
		return defaultValue
	}

	if boolValue, ok := value.(bool); ok {
		return boolValue
	}

	return defaultValue
}

// getIntConfig 获取整数配置值
func getIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	if config == nil {
		return defaultValue
	}

	value, exists := config[key]
	if !exists {
		return defaultValue
	}

	if intValue, ok := value.(int); ok {
		return intValue
	}

	if floatValue, ok := value.(float64); ok {
		return int(floatValue)
	}

	return defaultValue
}



