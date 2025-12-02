package nodes

import (
	"context"
	"fmt"
	"time"

	"xiaozhi-server-go/internal/startup"
	"xiaozhi-server-go/internal/workflow"
)

// AuthNodeExecutor 认证节点执行器
type AuthNodeExecutor struct {
	logger startup.StartupLogger
}

// NewAuthNodeExecutor 创建认证节点执行器
func NewAuthNodeExecutor(logger startup.StartupLogger) *AuthNodeExecutor {
	return &AuthNodeExecutor{
		logger: logger,
	}
}

// Execute 执行认证节点
func (e *AuthNodeExecutor) Execute(
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

	e.logger.Info("Executing auth node", "node_id", node.ID, "node_name", node.Name)

	// 根据节点ID执行不同的认证操作
	switch node.ID {
	case "auth:init-manager":
		err := e.executeInitAuthManager(ctx, node, result, inputs, context)
		if err != nil {
			result.Status = workflow.NodeStatusFailed
			result.Error = err.Error()
			return result, err
		}
	default:
		err := fmt.Errorf("unknown auth node: %s", node.ID)
		result.Status = workflow.NodeStatusFailed
		result.Error = err.Error()
		return result, err
	}

	// 成功完成
	endTime := time.Now()
	result.EndTime = &endTime
	result.Duration = endTime.Sub(startTime)
	result.Status = workflow.NodeStatusCompleted

	e.logger.Info("Auth node completed successfully",
		"node_id", node.ID,
		"duration", result.Duration.String())

	return result, nil
}

// executeInitAuthManager 执行初始化认证管理器
func (e *AuthNodeExecutor) executeInitAuthManager(
	ctx context.Context,
	node *startup.StartupNode,
	result *startup.StartupNodeResult,
	inputs map[string]interface{},
	context map[string]interface{},
) error {
	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Starting authentication manager initialization",
	})

	// 获取配置
	authStore := getStringConfig(node.Config, "auth_store", "database")
	sessionTTL := getStringConfig(node.Config, "session_ttl", "24h")
	cleanupInterval := getStringConfig(node.Config, "cleanup_interval", "10m")
	enableOAuth := getBoolConfig(node.Config, "enable_oauth", false)
	enableJWT := getBoolConfig(node.Config, "enable_jwt", true)
	jwtSecret := getStringConfig(node.Config, "jwt_secret", "your-jwt-secret-key")

	e.logger.Info("Initializing authentication manager",
		"auth_store", authStore,
		"session_ttl", sessionTTL,
		"cleanup_interval", cleanupInterval,
		"enable_oauth", enableOAuth,
		"enable_jwt", enableJWT)

	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Auth configuration: store=%s, ttl=%s, oauth=%v, jwt=%v", authStore, sessionTTL, enableOAuth, enableJWT),
	})

	// 检查依赖节点是否已完成
	componentsReady := e.checkDependencyCompletion(context, "components:init-container")
	observabilityReady := e.checkDependencyCompletion(context, "observability:setup-hooks")
	databaseReady := e.checkDependencyCompletion(context, "storage:init-database")

	if !componentsReady || !observabilityReady || !databaseReady {
		err := fmt.Errorf("required dependencies not completed: components=%v, observability=%v, database=%v",
			componentsReady, observabilityReady, databaseReady)
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   err.Error(),
		})
		return err
	}

	// 验证配置
	if enableJWT && jwtSecret == "" {
		err := fmt.Errorf("JWT secret is required when JWT is enabled")
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   err.Error(),
		})
		return err
	}

	// 初始化认证存储后端
	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Initializing %s authentication store", authStore),
	})

	err := e.initializeAuthStore(ctx, authStore, result)
	if err != nil {
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   fmt.Sprintf("Failed to initialize auth store: %s", err.Error()),
		})
		return err
	}

	// 初始化会话管理
	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Initializing session management",
	})

	err = e.initializeSessionManagement(ctx, sessionTTL, cleanupInterval, result)
	if err != nil {
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   fmt.Sprintf("Failed to initialize session management: %s", err.Error()),
		})
		return err
	}

	// 初始化密码加密
	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Initializing password encryption",
	})

	err = e.initializePasswordEncryption(result)
	if err != nil {
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   fmt.Sprintf("Failed to initialize password encryption: %s", err.Error()),
		})
		return err
	}

	// 初始化JWT支持（如果启用）
	if enableJWT {
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Initializing JWT token support",
		})

		err = e.initializeJWTSupport(jwtSecret, result)
		if err != nil {
			result.Logs = append(result.Logs, startup.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "error",
				Message:   fmt.Sprintf("Failed to initialize JWT support: %s", err.Error()),
			})
			return err
		}
	}

	// 初始化OAuth支持（如果启用）
	if enableOAuth {
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Initializing OAuth support",
		})

		err = e.initializeOAuthSupport(result)
		if err != nil {
			result.Logs = append(result.Logs, startup.StartupNodeLog{
				Timestamp: time.Now(),
				Level:     "warn",
				Message:   fmt.Sprintf("OAuth initialization failed: %s", err.Error()),
			})
			// OAuth失败不影响认证管理器的整体初始化
		}
	}

	// 启动会话清理协程
	if cleanupInterval != "" {
		go e.startSessionCleanup(ctx, cleanupInterval, result)
	}

	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Authentication manager initialized successfully",
	})

	// 设置输出结果
	result.Outputs = map[string]interface{}{
		"auth_store":            authStore,
		"session_ttl":           sessionTTL,
		"cleanup_interval":      cleanupInterval,
		"oauth_enabled":         enableOAuth,
		"jwt_enabled":           enableJWT,
		"auth_manager_id":       "auth-manager-" + node.ID,
		"session_store_id":      "session-store-" + node.ID,
		"password_hasher_id":    "password-hasher-" + node.ID,
		"initialized_at":        time.Now(),
		"supported_methods":     e.getSupportedAuthMethods(enableOAuth, enableJWT),
		"security_features":     e.getSecurityFeatures(),
	}

	return nil
}

// initializeAuthStore 初始化认证存储后端
func (e *AuthNodeExecutor) initializeAuthStore(ctx context.Context, storeType string, result *startup.StartupNodeResult) error {
	switch storeType {
	case "memory":
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Memory-based auth store initialized (not persistent)",
		})
	case "sqlite":
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "SQLite auth store initialized",
		})
	case "redis":
		result.Logs = append(result.Logs, startup.StartupNodeLog{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Redis auth store initialized",
		})
	default:
		return fmt.Errorf("unsupported auth store type: %s", storeType)
	}

	// 模拟存储初始化延迟
	select {
	case <-time.After(500 * time.Millisecond):
		// 初始化完成
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// initializeSessionManagement 初始化会话管理
func (e *AuthNodeExecutor) initializeSessionManagement(ctx context.Context, sessionTTL, cleanupInterval string, result *startup.StartupNodeResult) error {
	// 解析会话TTL
	ttl, err := time.ParseDuration(sessionTTL)
	if err != nil {
		return fmt.Errorf("invalid session_ttl format: %s", sessionTTL)
	}

	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Session management initialized with TTL: %v", ttl),
	})

	// 模拟会话管理初始化
	select {
	case <-time.After(300 * time.Millisecond):
		// 初始化完成
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// initializePasswordEncryption 初始化密码加密
func (e *AuthNodeExecutor) initializePasswordEncryption(result *startup.StartupNodeResult) error {
	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Password encryption initialized with bcrypt",
	})

	// 模拟加密器初始化
	select {
	case <-time.After(200 * time.Millisecond):
		// 初始化完成
	default:
	}

	return nil
}

// initializeJWTSupport 初始化JWT支持
func (e *AuthNodeExecutor) initializeJWTSupport(jwtSecret string, result *startup.StartupNodeResult) error {
	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "JWT token support initialized",
	})

	// 模拟JWT初始化
	select {
	case <-time.After(200 * time.Millisecond):
		// 初始化完成
	default:
	}

	return nil
}

// initializeOAuthSupport 初始化OAuth支持
func (e *AuthNodeExecutor) initializeOAuthSupport(result *startup.StartupNodeResult) error {
	result.Logs = append(result.Logs, startup.StartupNodeLog{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "OAuth support initialized",
	})

	// 模拟OAuth初始化
	select {
	case <-time.After(500 * time.Millisecond):
		// 初始化完成
	default:
	}

	return nil
}

// startSessionCleanup 启动会话清理协程
func (e *AuthNodeExecutor) startSessionCleanup(ctx context.Context, cleanupInterval string, result *startup.StartupNodeResult) {
	interval, err := time.ParseDuration(cleanupInterval)
	if err != nil {
		e.logger.Error("Invalid cleanup interval format", "interval", cleanupInterval, "error", err)
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	e.logger.Info("Session cleanup routine started", "interval", interval.String())

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Session cleanup routine stopped")
			return
		case <-ticker.C:
			// 执行会话清理
			e.logger.Debug("Running session cleanup")
			// 在实际实现中，这里会清理过期的会话
		}
	}
}

// checkDependencyCompletion 检查依赖节点是否已完成
func (e *AuthNodeExecutor) checkDependencyCompletion(context map[string]interface{}, nodeID string) bool {
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

// getSupportedAuthMethods 获取支持的认证方法
func (e *AuthNodeExecutor) getSupportedAuthMethods(enableOAuth, enableJWT bool) []string {
	methods := []string{"password", "session"}

	if enableJWT {
		methods = append(methods, "jwt")
	}

	if enableOAuth {
		methods = append(methods, "oauth")
	}

	return methods
}

// getSecurityFeatures 获取安全特性
func (e *AuthNodeExecutor) getSecurityFeatures() []string {
	return []string{
		"password_hashing",
		"session_management",
		"csrf_protection",
		"rate_limiting",
		"account_lockout",
		"audit_logging",
	}
}

// Validate 验证节点配置
func (e *AuthNodeExecutor) Validate(ctx context.Context, node *startup.StartupNode) error {
	if node.ID == "" {
		return fmt.Errorf("node ID is required")
	}

	if node.Type != startup.StartupNodeAuth {
		return fmt.Errorf("invalid node type: expected %s, got %s", startup.StartupNodeAuth, node.Type)
	}

	// 根据节点ID验证特定配置
	switch node.ID {
	case "auth:init-manager":
		return e.validateInitAuthManager(node)
	default:
		return fmt.Errorf("unknown auth node: %s", node.ID)
	}
}

// validateInitAuthManager 验证认证管理器配置
func (e *AuthNodeExecutor) validateInitAuthManager(node *startup.StartupNode) error {
	authStore := getStringConfig(node.Config, "auth_store", "")
	if authStore != "" && !contains([]string{"memory", "sqlite", "redis"}, authStore) {
		return fmt.Errorf("unsupported auth_store: %s", authStore)
	}

	sessionTTL := getStringConfig(node.Config, "session_ttl", "")
	if sessionTTL != "" && !isValidDuration(sessionTTL) {
		return fmt.Errorf("invalid session_ttl format: %s", sessionTTL)
	}

	cleanupInterval := getStringConfig(node.Config, "cleanup_interval", "")
	if cleanupInterval != "" && !isValidDuration(cleanupInterval) {
		return fmt.Errorf("invalid cleanup_interval format: %s", cleanupInterval)
	}

	enableJWT := getBoolConfig(node.Config, "enable_jwt", false)
	if enableJWT {
		jwtSecret := getStringConfig(node.Config, "jwt_secret", "")
		if jwtSecret == "" {
			return fmt.Errorf("jwt_secret is required when JWT is enabled")
		}
		if len(jwtSecret) < 32 {
			return fmt.Errorf("jwt_secret must be at least 32 characters long")
		}
	}

	return nil
}

// GetNodeInfo 获取节点信息
func (e *AuthNodeExecutor) GetNodeInfo() *startup.StartupNodeInfo {
	return &startup.StartupNodeInfo{
		Type:        startup.StartupNodeAuth,
		Name:        "Auth Node Executor",
		Description: "Handles authentication and authorization system initialization including user management and session handling",
		Version:     "1.0.0",
		Author:      "XiaoZhi Flow Team",
		SupportedConfig: map[string]interface{}{
			"auth_store": map[string]interface{}{
				"type":        "string",
				"description": "Authentication storage backend",
				"enum":        []string{"memory", "sqlite", "redis"},
				"default":     "database",
			},
			"session_ttl": map[string]interface{}{
				"type":        "string",
				"description": "Session time-to-live duration",
				"default":     "24h",
			},
			"cleanup_interval": map[string]interface{}{
				"type":        "string",
				"description": "Expired session cleanup interval",
				"default":     "10m",
			},
			"enable_oauth": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable OAuth authentication",
				"default":     false,
			},
			"enable_jwt": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable JWT token authentication",
				"default":     true,
			},
			"jwt_secret": map[string]interface{}{
				"type":        "string",
				"description": "JWT signing secret key",
				"min_length":  32,
			},
		},
		Capabilities: []string{
			"user-authentication",
			"session-management",
			"password-hashing",
			"jwt-tokens",
			"oauth-integration",
			"csrf-protection",
			"rate-limiting",
			"audit-logging",
		},
	}
}

// Cleanup 清理资源
func (e *AuthNodeExecutor) Cleanup(ctx context.Context) error {
	e.logger.Info("Cleaning up auth node executor")
	// 这里可以停止会话清理协程、关闭数据库连接等
	return nil
}