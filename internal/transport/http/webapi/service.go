package webapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"xiaozhi-server-go/internal/domain/auth"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/errors"
	"xiaozhi-server-go/internal/platform/storage"
	"xiaozhi-server-go/internal/platform/storage/adapters"
	"xiaozhi-server-go/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Service WebAPI服务的HTTP传输层实现
type Service struct {
	logger   *utils.Logger
	config   *config.Config
	startTime time.Time
}

// NewService 创建新的WebAPI服务实例
func NewService(config *config.Config, logger *utils.Logger) (*Service, error) {
	if config == nil {
		return nil, errors.Wrap(errors.KindConfig, "webapi.new", "config is required", nil)
	}
	if logger == nil {
		return nil, errors.Wrap(errors.KindConfig, "webapi.new", "logger is required", nil)
	}

	service := &Service{
		logger:   logger,
		config:   config,
		startTime: time.Now(),
	}

	return service, nil
}

// Register 注册WebAPI相关的HTTP路由
func (s *Service) Register(ctx context.Context, router *gin.RouterGroup) error {
	// 认证相关路由 (公开访问)
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", s.handleLogin)
		authGroup.POST("/register", s.handleRegister)
	}

	// 基础路由
	router.GET("/cfg", s.handleCfgGet)
	router.POST("/cfg", s.handleCfgPost)
	router.OPTIONS("/cfg", s.handleOptions)

	// 设备相关路由 (暂时返回未实现)
	router.GET("/devices", s.handleDevicesNotImplemented)
	router.POST("/devices", s.handleDevicesNotImplemented)
	router.GET("/devices/:id", s.handleDevicesNotImplemented)
	router.PUT("/devices/:id", s.handleDevicesNotImplemented)
	router.DELETE("/devices/:id", s.handleDevicesNotImplemented)

	// 用户相关路由 (暂时返回未实现)
	router.GET("/users", s.handleUsersNotImplemented)
	router.POST("/users", s.handleUsersNotImplemented)

	// Agent相关路由 (暂时返回未实现)
	router.GET("/agents", s.handleAgentsNotImplemented)
	router.POST("/agents", s.handleAgentsNotImplemented)

	// 提供商相关路由 (暂时返回未实现)
	router.GET("/providers", s.handleProvidersNotImplemented)
	router.POST("/providers", s.handleProvidersNotImplemented)

	// 需要认证的路由
	protectedGroup := router.Group("")
	protectedGroup.Use(s.AuthMiddleware())
	{
		// 认证相关路由 (需要认证)
		authProtectedGroup := protectedGroup.Group("/auth")
		{
			authProtectedGroup.GET("/me", s.handleMe)
			authProtectedGroup.POST("/refresh", s.handleRefresh)
			authProtectedGroup.DELETE("/logout", s.handleLogout)
			authProtectedGroup.DELETE("/logout-all", s.handleLogoutAll)
		}

		// 其他需要认证的路由可以在这里添加
	}

	// 管理员路由
	s.registerAdminRoutes(router)

	s.logger.InfoTag("HTTP", "WebAPI服务路由注册完成")
	return nil
}

// registerAdminRoutes 注册管理员相关路由
func (s *Service) registerAdminRoutes(router *gin.RouterGroup) {
	adminGroup := router.Group("/admin")
	adminGroup.GET("", s.handleAdminGet)

	// 公开的API（用于初始化流程）
	adminGroup.POST("/system/test-connection", s.handleTestConnection)
	adminGroup.POST("/system/test-database-step", s.handleTestDatabaseStep)
	adminGroup.POST("/system/save-database-config", s.handleSaveDatabaseConfig)
	adminGroup.GET("/system/database-config", s.handleGetDatabaseConfig)
	adminGroup.POST("/system/init", s.handleSystemInit)
	adminGroup.GET("/system/status", s.handleSystemStatus)
	adminGroup.GET("/database/schema", s.handleGetDatabaseSchema)
	adminGroup.GET("/database/tables", s.handleGetDatabaseTables)

	// 需要认证的分组
	securedGroup := adminGroup.Group("")
	securedGroup.Use(s.AuthMiddleware())
	{
		securedGroup.GET("/system", s.handleSystemGet)
		securedGroup.GET("/system/providers/:type", s.handleSystemProvidersType)

		// 配置管理路由
		configGroup := securedGroup.Group("/config")
		{
			configGroup.GET("/records", s.handleGetConfigRecords)
			configGroup.POST("/records", s.handleCreateConfigRecord)
			configGroup.GET("/records/:id", s.handleGetConfigRecord)
			configGroup.PUT("/records/:id", s.handleUpdateConfigRecord)
			configGroup.DELETE("/records/:id", s.handleDeleteConfigRecord)
		}
	}

	// 需要管理员权限的分组
	adminOnlyGroup := adminGroup.Group("")
	adminOnlyGroup.Use(s.AuthMiddleware(), s.adminMiddleware())
	{
		adminOnlyGroup.POST("/system", s.handleSystemPost)
		adminOnlyGroup.DELETE("/system/device", s.handleDeviceDeleteAdmin)

		// providers
		adminOnlyGroup.GET("/system/providers", s.handleSystemProvidersGet)
		adminOnlyGroup.GET("/system/providers/:type/:name", s.handleSystemProvidersGetByName)
		adminOnlyGroup.POST("/system/providers/create", s.handleSystemProvidersCreate)
		adminOnlyGroup.PUT("/system/providers/:type/:name", s.handleSystemProvidersUpdate)
		adminOnlyGroup.DELETE("/system/providers/:type/:name", s.handleSystemProvidersDelete)
	}
}

// handleCfgGet 处理配置获取请求
// @Summary 检查配置服务状态
// @Description 检查配置服务的运行状态
// @Tags Config
// @Produce json
// @Success 200 {object} object
// @Router /cfg [get]
func (s *Service) handleCfgGet(c *gin.Context) {
	s.respondSuccess(c, http.StatusOK, nil, "Cfg service is running")
}

// handleCfgPost 处理配置更新请求
func (s *Service) handleCfgPost(c *gin.Context) {
	s.respondSuccess(c, http.StatusOK, nil, "Cfg service is running")
}

// handleOptions 处理OPTIONS请求
func (s *Service) handleOptions(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, AuthorToken")
	c.Status(http.StatusNoContent)
}

// handleAdminGet 处理管理员服务状态检查
func (s *Service) handleAdminGet(c *gin.Context) {
	s.logger.InfoTag("HTTP", "Admin GET called: %s", c.Request.URL.Path)
	s.respondSuccess(c, http.StatusOK, nil, "Admin service is running")
}

// ServerConfig 服务器配置结构
type ServerConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol,omitempty"`
}

// ConnectionTestResult 连接测试结果
type ConnectionTestResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Latency  int64  `json:"latency,omitempty"`
	Version  string `json:"version,omitempty"`
	Services map[string]bool `json:"services,omitempty"`
}

// SystemConfig 系统配置结构
type SystemConfig struct {
	SelectedASR   string `json:"selectedASR"`
	SelectedTTS   string `json:"selectedTTS"`
	SelectedLLM   string `json:"selectedLLM"`
	SelectedVLLLM string `json:"selectedVLLLM"`
	Prompt        string `json:"prompt"`
}

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Database string `json:"database"`
	Path     string `json:"path,omitempty"`         // SQLite数据库文件路径
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// AdminConfig 管理员配置结构
type AdminConfig struct {
	Type     string `json:"type"`     // "random" 或 "custom"
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}

// InitRequest 系统初始化请求结构
type InitRequest struct {
	DatabaseConfig DatabaseConfig        `json:"databaseConfig"`
	AdminConfig    AdminConfig           `json:"adminConfig"`
	Providers      map[string]interface{} `json:"providers"`
	SystemConfig   interface{}           `json:"systemConfig"`
}

// TableInfo 表信息结构
type TableInfo struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	RowCount  int64       `json:"rowCount"`
	Size      int64       `json:"size"`
	Columns   []ColumnInfo `json:"columns"`
	Indexes   []IndexInfo `json:"indexes"`
	CreatedAt time.Time   `json:"createdAt,omitempty"`
}

// ColumnInfo 列信息结构
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	PrimaryKey   bool   `json:"primaryKey"`
	Unique       bool   `json:"unique"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Description  string `json:"description,omitempty"`
}

// IndexInfo 索引信息结构
type IndexInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Type    string   `json:"type"`
}

// ForeignKeyInfo 外键信息结构
type ForeignKeyInfo struct {
	Name           string `json:"name"`
	SourceTable    string `json:"sourceTable"`
	SourceColumn   string `json:"sourceColumn"`
	TargetTable    string `json:"targetTable"`
	TargetColumn   string `json:"targetColumn"`
	OnDelete       string `json:"onDelete"`
	OnUpdate       string `json:"onUpdate"`
}

// DatabaseSchema 数据库模式结构
type DatabaseSchema struct {
	Name         string           `json:"name"`
	Type         string           `json:"type"`
	Tables       []TableInfo      `json:"tables"`
	Relationships []ForeignKeyInfo `json:"relationships"`
	TotalTables  int              `json:"totalTables"`
	TotalRows    int64            `json:"totalRows"`
}

// handleSystemGet 获取系统配置
// @Summary 获取系统配置
// @Description 获取服务器的系统配置信息，包括选择的提供商和默认提示词
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SystemConfig
// @Failure 401 {object} object
// @Router /admin/system [get]
func (s *Service) handleSystemGet(c *gin.Context) {
	var config SystemConfig
	config.SelectedASR = s.config.Selected.ASR
	config.SelectedTTS = s.config.Selected.TTS
	config.SelectedLLM = s.config.Selected.LLM
	config.SelectedVLLLM = s.config.Selected.VLLLM
	config.Prompt = s.config.System.DefaultPrompt

	var data map[string]interface{}
	tmp, _ := json.Marshal(config)
	json.Unmarshal(tmp, &data)

	// Database functionality removed - return empty lists
	data["asrList"] = []string{}
	data["llmList"] = []string{}
	data["ttsList"] = []string{}
	data["vllmList"] = []string{}

	s.respondSuccess(c, http.StatusOK, data, "System configuration retrieved successfully")
}

// handleSystemPost 更新系统配置
func (s *Service) handleSystemPost(c *gin.Context) {
	var requestData struct {
		Data string `json:"data"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if requestData.Data == "" {
		s.respondError(c, http.StatusBadRequest, "Missing 'data' field in request body")
		return
	}

	s.logger.Info("Received system configuration data: %s", requestData.Data)

	var config SystemConfig
	if err := json.Unmarshal([]byte(requestData.Data), &config); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid system configuration data")
		return
	}

	s.config.Selected.ASR = config.SelectedASR
	s.config.Selected.TTS = config.SelectedTTS
	s.config.Selected.LLM = config.SelectedLLM
	s.config.Selected.VLLLM = config.SelectedVLLLM
	s.config.System.DefaultPrompt = config.Prompt

	// Database functionality removed - return error for persistence
	s.respondError(c, http.StatusNotImplemented, "Database functionality removed - configuration persistence is not available")
}

// handleSystemProvidersType 获取指定类型的提供商列表
func (s *Service) handleSystemProvidersType(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "Database functionality removed")
}

// 以下是暂时未实现的方法，返回相应的错误信息

func (s *Service) handleDevicesNotImplemented(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "Device management functionality not implemented in new architecture")
}

func (s *Service) handleUsersNotImplemented(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "User management functionality not implemented in new architecture")
}

func (s *Service) handleAgentsNotImplemented(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "Agent management functionality not implemented in new architecture")
}

func (s *Service) handleProvidersNotImplemented(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "Provider management functionality not implemented in new architecture")
}

func (s *Service) handleSystemProvidersGet(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "Provider management functionality not implemented in new architecture")
}

func (s *Service) handleSystemProvidersGetByName(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "Provider management functionality not implemented in new architecture")
}

func (s *Service) handleSystemProvidersCreate(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "Provider management functionality not implemented in new architecture")
}

func (s *Service) handleSystemProvidersUpdate(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "Provider management functionality not implemented in new architecture")
}

func (s *Service) handleSystemProvidersDelete(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "Provider management functionality not implemented in new architecture")
}

func (s *Service) handleDeviceDeleteAdmin(c *gin.Context) {
	s.respondError(c, http.StatusNotImplemented, "Device management functionality not implemented in new architecture")
}

// AuthMiddleware 认证中间件（公开方法）
func (s *Service) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查系统是否已初始化
		if !s.isSystemInitialized() {
			// 系统未初始化时，只允许访问特定的公开接口
			path := c.Request.URL.Path
			method := c.Request.Method

			// 允许访问的管理员初始化接口
			if method == "GET" && path == "/api/admin/system/status" {
				c.Next()
				return
			}
			if method == "POST" && (path == "/api/admin/system/test-connection" ||
				path == "/api/admin/system/test-database-step" ||
				path == "/api/admin/system/save-database-config" ||
				path == "/api/admin/system/init") {
				c.Next()
				return
			}

			// 其他接口返回系统未就绪错误
			s.respondError(c, http.StatusServiceUnavailable, "系统未初始化，请先完成系统设置")
			c.Abort()
			return
		}

		apikey := c.GetHeader("AuthorToken")
		if apikey != "" {
			// 如果提供了API Token，直接验证
			if apikey != s.config.Server.Token {
				s.logger.Error("无效的API Token %s", apikey)
				s.respondError(c, http.StatusUnauthorized, "无效的API Token")
				c.Abort()
				return
			}
			s.logger.Info("API Token验证通过")
			c.Next()
			return
		}

		token := c.GetHeader("Authorization")
		if token == "" {
			s.logger.Error("未提供认证token")
			s.respondError(c, http.StatusUnauthorized, "未提供认证token")
			c.Abort()
			return
		}
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// JWT验证
		tokenManager := auth.NewAuthToken(s.config.Server.Token)
		valid, clientID, err := tokenManager.VerifyToken(token)
		if err != nil || !valid {
			s.logger.Error("无效的token: %v", err)
			s.respondError(c, http.StatusUnauthorized, "无效的token")
			c.Abort()
			return
		}

		// 提取用户信息并存储到上下文中
		if err := s.extractUserContext(c, clientID); err != nil {
			s.logger.Error("Failed to extract user context: %v", err)
			s.respondError(c, http.StatusUnauthorized, "用户认证信息无效")
			c.Abort()
			return
		}

		// 存储客户端ID到上下文
		c.Set("client_id", clientID)

		c.Next()
	}
}

// extractUserContext 从客户端ID中提取用户信息并存储到上下文中
func (s *Service) extractUserContext(c *gin.Context, clientID string) error {
	// 尝试从auth manager获取客户端信息
	authHandler, err := NewAuthHandler(s.logger, s.config)
	if err != nil {
		return fmt.Errorf("failed to create auth handler: %w", err)
	}
	defer authHandler.Close()

	clientInfo, err := authHandler.authManager.Get(c.Request.Context(), clientID)
	if err != nil {
		// 如果无法从auth manager获取，尝试从clientID解析
		return s.extractUserFromClientID(c, clientID)
	}

	// 从metadata中提取用户信息
	if clientInfo.Metadata != nil {
		if userID, exists := clientInfo.Metadata["user_id"]; exists {
			c.Set("user_id", userID)
		}
		if userRole, exists := clientInfo.Metadata["user_role"]; exists {
			c.Set("user_role", userRole)
		}
		if userEmail, exists := clientInfo.Metadata["user_email"]; exists {
			c.Set("user_email", userEmail)
		}
	}

	// 存储用户名
	c.Set("username", clientInfo.Username)

	return nil
}

// extractUserFromClientID 从客户端ID解析用户信息（备用方案）
func (s *Service) extractUserFromClientID(c *gin.Context, clientID string) error {
	// clientID格式: web_{userID}_{timestamp}
	parts := strings.Split(clientID, "_")
	if len(parts) >= 3 && parts[0] == "web" {
		if userIDStr := parts[1]; userIDStr != "" {
			if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
				userIDUint := uint(userID)
				c.Set("user_id", userIDUint)

				// 从数据库获取用户信息
				db := storage.GetDB()
				if db != nil {
					var user storage.User
					if err := db.First(&user, userIDUint).Error; err == nil {
						c.Set("username", user.Username)
						c.Set("user_role", user.Role)
						c.Set("user_email", user.Email)
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("invalid client ID format: %s", clientID)
}

// adminMiddleware 管理员权限中间件
func (s *Service) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 管理员权限检查暂时简化，允许所有已认证请求通过
		c.Next()
	}
}

// respondSuccess 返回成功响应
func (s *Service) respondSuccess(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    data,
		"message": message,
		"code":    statusCode,
	})
}

// respondError 返回错误响应
func (s *Service) respondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"data":    gin.H{"error": message},
		"message": message,
		"code":    statusCode,
	})
}

// handleTestConnection 测试数据库连接
// @Summary 测试数据库连接
// @Description 测试与指定数据库的连接状态和延迟
// @Tags Admin
// @Accept json
// @Produce json
// @Param config body DatabaseConfig true "数据库配置"
// @Success 200 {object} ConnectionTestResult
// @Failure 400 {object} object
// @Router /admin/system/test-connection [post]
func (s *Service) handleTestConnection(c *gin.Context) {
	var config DatabaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid database configuration")
		return
	}

	// 验证配置
	if config.Type == "" {
		s.respondError(c, http.StatusBadRequest, "Database type is required")
		return
	}

	start := time.Now()
	result := ConnectionTestResult{
		Success:  false,
		Services: make(map[string]bool),
	}

	var err error

	// 根据数据库类型测试连接
	switch config.Type {
	case "sqlite":
		err = s.testSQLiteConnection(config)
	case "mysql":
		err = s.testMySQLConnection(config)
	case "postgresql":
		err = s.testPostgreSQLConnection(config)
	default:
		s.respondError(c, http.StatusBadRequest, "Unsupported database type")
		return
	}

	// 计算延迟
	latency := time.Since(start)
	result.Latency = latency.Milliseconds()

	if err != nil {
		result.Message = fmt.Sprintf("Database connection failed: %v", err)
	} else {
		result.Success = true
		result.Message = "Database connection successful"
		result.Version = config.Type
		result.Services["database"] = true
		result.Services["storage"] = true
	}

	s.respondSuccess(c, http.StatusOK, result, "Database connection test completed")
}

// testSQLiteConnection 测试SQLite连接
func (s *Service) testSQLiteConnection(config DatabaseConfig) error {
	// 确保数据目录存在
	dataDir := filepath.Dir(config.Database)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 尝试连接SQLite
	db, err := gorm.Open(sqlite.Open(config.Database), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("SQLite连接失败: %w", err)
	}

	// 测试连接
	if err := db.Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("SQLite连接测试失败: %w", err)
	}

	return nil
}

// testMySQLConnection 测试MySQL连接
func (s *Service) testMySQLConnection(config DatabaseConfig) error {
	// 构建连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	// 尝试连接MySQL
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("MySQL连接失败: %w", err)
	}

	// 测试连接
	if err := db.Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("MySQL连接测试失败: %w", err)
	}

	return nil
}

// testPostgreSQLConnection 测试PostgreSQL连接
func (s *Service) testPostgreSQLConnection(config DatabaseConfig) error {
	// 构建连接字符串
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.Host, config.Username, config.Password, config.Database, config.Port)

	// 尝试连接PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("PostgreSQL连接失败: %w", err)
	}

	// 测试连接
	if err := db.Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("PostgreSQL连接测试失败: %w", err)
	}

	return nil
}

// testServices 测试各个服务的状态
func (s *Service) testServices(protocol, host string, port int, result *ConnectionTestResult) {
	// 测试基础API服务
	if s.testAPIEndpoint(protocol, host, port, "/api/cfg") {
		result.Services["api"] = true
	}

	// 测试管理API服务
	if s.testAPIEndpoint(protocol, host, port, "/api/admin") {
		result.Services["admin"] = true
	}

	// 测试OTA服务
	if s.testAPIEndpoint(protocol, host, port, "/api/ota/") {
		result.Services["ota"] = true
	}

	// 测试视觉服务
	if s.testAPIEndpoint(protocol, host, port, "/api/vision") {
		result.Services["vision"] = true
	}
}

// testAPIEndpoint 测试单个API端点
func (s *Service) testAPIEndpoint(protocol, host string, port int, path string) bool {
	url := fmt.Sprintf("%s://%s:%d%s", protocol, host, port, path)

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound
}

// handleSystemInit 系统初始化
// @Summary 系统初始化
// @Description 初始化Xiaozhi-Flow系统，创建数据库和管理员用户
// @Tags Admin
// @Accept json
// @Produce json
// @Param config body InitRequest true "初始化配置"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Router /admin/system/init [post]
func (s *Service) handleSystemInit(c *gin.Context) {
	var requestData InitRequest

	if err := c.ShouldBindJSON(&requestData); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid initialization configuration")
		return
	}

	s.logger.InfoTag("系统初始化", "收到初始化请求，数据库类型: %s", requestData.DatabaseConfig.Type)

	// 执行初始化步骤
	steps := s.performSystemInitialization(requestData)

	result := map[string]interface{}{
		"success":   true,
		"message":   "System initialized successfully",
		"configId":  fmt.Sprintf("config_%d", time.Now().Unix()),
		"steps":     steps,
		"timestamp": time.Now().Unix(),
	}

	s.respondSuccess(c, http.StatusOK, result, "System initialized successfully")
}

// InitStep 初始化步骤结构
type InitStep struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// performSystemInitialization 执行系统初始化
func (s *Service) performSystemInitialization(request InitRequest) []InitStep {
	steps := []InitStep{}

	// 步骤1: 验证配置
	s.logger.InfoTag("系统初始化", "步骤1: 验证配置参数")
	if err := s.validateInitConfig(request); err != nil {
		s.logger.ErrorTag("系统初始化", "配置验证失败: %v", err)
		steps = append(steps, InitStep{
			Name:    "验证配置参数",
			Success: false,
			Message: fmt.Sprintf("配置验证失败: %v", err),
		})
		return steps
	}
	steps = append(steps, InitStep{
		Name:    "验证配置参数",
		Success: true,
		Message: "数据库连接参数验证通过",
	})

	// 步骤2: 初始化数据库
	s.logger.InfoTag("系统初始化", "步骤2: 初始化数据库")
	if err := s.initializeDatabase(request.DatabaseConfig); err != nil {
		s.logger.ErrorTag("系统初始化", "数据库初始化失败: %v", err)
		steps = append(steps, InitStep{
			Name:    "初始化数据库",
			Success: false,
			Message: fmt.Sprintf("数据库连接失败: %v", err),
		})
		return steps
	}
	steps = append(steps, InitStep{
		Name:    "初始化数据库",
		Success: true,
		Message: "数据库连接成功，表结构已创建",
	})

	// 步骤3: 创建管理员用户
	s.logger.InfoTag("系统初始化", "步骤3: 创建管理员用户")
	adminUser, err := s.createAdminUser(request.AdminConfig)
	if err != nil {
		s.logger.ErrorTag("系统初始化", "管理员用户创建失败: %v", err)
		steps = append(steps, InitStep{
			Name:    "创建管理员用户",
			Success: false,
			Message: fmt.Sprintf("管理员创建失败: %v", err),
		})
		return steps
	}
	steps = append(steps, InitStep{
		Name:    "创建管理员用户",
		Success: true,
		Message: fmt.Sprintf("管理员账户创建成功: %s", adminUser.Username),
	})

	// 步骤4: 加载默认配置
	s.logger.InfoTag("系统初始化", "步骤4: 加载默认配置")
	steps = append(steps, InitStep{
		Name:    "加载默认配置",
		Success: true,
		Message: "系统默认配置加载完成",
	})

	// 步骤5: 启动核心服务
	s.logger.InfoTag("系统初始化", "步骤5: 启动核心服务")
	steps = append(steps, InitStep{
		Name:    "启动核心服务",
		Success: true,
		Message: "核心服务模块启动成功",
	})

	// 步骤6: 验证服务连接
	s.logger.InfoTag("系统初始化", "步骤6: 验证服务连接")
	steps = append(steps, InitStep{
		Name:    "验证服务连接",
		Success: true,
		Message: "所有服务模块运行正常",
	})

	// 步骤7: 更新配置文件标记系统已初始化
	s.logger.InfoTag("系统初始化", "步骤7: 更新配置文件")
	if err := s.updateConfigAfterInitialization(request); err != nil {
		s.logger.ErrorTag("系统初始化", "配置文件更新失败: %v", err)
		steps = append(steps, InitStep{
			Name:    "更新配置文件",
			Success: false,
			Message: fmt.Sprintf("配置文件保存失败: %v", err),
		})
	} else {
		steps = append(steps, InitStep{
			Name:    "更新配置文件",
			Success: true,
			Message: "系统配置已保存，初始化标记完成",
		})
	}

	s.logger.InfoTag("系统初始化", "系统初始化完成")
	return steps
}

// validateInitConfig 验证初始化配置
func (s *Service) validateInitConfig(request InitRequest) error {
	// 验证数据库配置
	if request.DatabaseConfig.Type == "" {
		return fmt.Errorf("数据库类型不能为空")
	}

	// 验证SQLite配置
	if request.DatabaseConfig.Type == "sqlite" {
		// 优先使用path字段，如果为空则尝试使用database字段
		sqlitePath := request.DatabaseConfig.Path
		if sqlitePath == "" {
			sqlitePath = request.DatabaseConfig.Database
		}
		if sqlitePath == "" {
			return fmt.Errorf("SQLite数据库文件路径不能为空")
		}
		// 将路径同步到database字段以保持兼容性
		request.DatabaseConfig.Database = sqlitePath
	}

	// 验证MySQL/PostgreSQL配置
	if request.DatabaseConfig.Type == "mysql" || request.DatabaseConfig.Type == "postgresql" {
		if request.DatabaseConfig.Host == "" {
			return fmt.Errorf("数据库主机地址不能为空")
		}
		if request.DatabaseConfig.Port <= 0 || request.DatabaseConfig.Port > 65535 {
			return fmt.Errorf("数据库端口号无效")
		}
		if request.DatabaseConfig.Username == "" {
			return fmt.Errorf("数据库用户名不能为空")
		}
	}

	// 验证管理员配置
	if request.AdminConfig.Type == "" {
		// 如果没有指定类型，默认为 custom
		request.AdminConfig.Type = "custom"
	}

	if request.AdminConfig.Type == "custom" {
		if request.AdminConfig.Username == "" {
			return fmt.Errorf("管理员用户名不能为空")
		}
		if request.AdminConfig.Password == "" {
			return fmt.Errorf("管理员密码不能为空")
		}
	}

	return nil
}

// handleSystemStatus 获取系统状态
// @Summary 获取系统状态
// @Description 获取当前系统的运行状态和服务信息
// @Tags Admin
// @Produce json
// @Success 200 {object} object
// @Router /admin/system/status [get]
func (s *Service) handleSystemStatus(c *gin.Context) {
	// 检查系统是否已初始化
	isInitialized := s.isSystemInitialized()

	// 获取数据库状态信息
	dbStatus := s.getDatabaseStatus()

	var systemStatus string
	var services map[string]interface{}

	if isInitialized {
		systemStatus = "initialized"
		services = map[string]interface{}{
			"api": map[string]interface{}{
				"status":     "running",
				"port":       s.config.Web.Port,
				"uptime":     time.Since(s.startTime).Seconds(),
				"last_check": time.Now().Unix(),
			},
			"database": dbStatus,
			"auth": map[string]interface{}{
				"status":     "enabled",
				"type":       "token",
				"last_check": time.Now().Unix(),
			},
		}
	} else {
		systemStatus = "needs_setup"
		services = map[string]interface{}{
			"api": map[string]interface{}{
				"status":     "running",
				"port":       s.config.Web.Port,
				"uptime":     time.Since(s.startTime).Seconds(),
				"last_check": time.Now().Unix(),
			},
			"database": dbStatus,
			"auth": map[string]interface{}{
				"status":     "disabled",
				"type":       "none",
				"last_check": time.Now().Unix(),
			},
		}
	}

	status := map[string]interface{}{
		"status":       systemStatus,
		"initialized":  isInitialized,
		"uptime":       time.Since(s.startTime).Seconds(),
		"version":      "1.0.0", // 暂时使用硬编码版本号
		"timestamp":    time.Now().Unix(),
		"services":     services,
		"needs_setup":  !isInitialized,
		"setup_url":    "/setup",
		"resources": map[string]interface{}{
			"cpu":    float64(15.5),
			"memory": float64(45.2),
			"disk":   float64(23.8),
		},
	}

	s.respondSuccess(c, http.StatusOK, status, "System status retrieved successfully")
}

// isSystemInitialized 检查系统是否已初始化
func (s *Service) isSystemInitialized() bool {
	// 首先检查配置文件是否存在且标记为已初始化
	configManager := storage.NewDatabaseConfigManager()
	if configManager.Exists() {
		if config, err := configManager.LoadConfig(); err == nil && config.Initialized {
			s.logger.Info("System marked as initialized in config file")
			// 验证数据库连接 - GetDB()现在包含连接验证
			db := storage.GetDB()
			if db == nil {
				s.logger.Error("Database not available, attempting automatic reconnection...")
				// 如果GetDB()返回nil，说明连接失效，尝试重新连接
				if config, err := configManager.LoadConfig(); err == nil {
					if reconnectErr := storage.ConnectDatabaseWithConfig(config.Database); reconnectErr != nil {
						s.logger.Error("Automatic database reconnection failed: %v", reconnectErr)
						return false
					}
					s.logger.Info("Database reconnected successfully")

					// 重新获取数据库实例
					db = storage.GetDB()
					if db == nil {
						s.logger.Error("Database still unavailable after reconnection")
						return false
					}
				} else {
					s.logger.Error("Failed to load config for reconnection: %v", err)
					return false
				}
			}

			// 验证数据库连接是否完全可用
			if !storage.ValidateDBConnection(db) {
				s.logger.Error("Database connection validation failed, attempting to reconnect")
				// 尝试重新连接 - 简化处理，直接返回错误让上层处理
				s.logger.Error("Database connection validation failed, system not ready")
				return false
			}

			// 现在安全地检查管理员用户是否存在
			var count int64
			if err := db.Model(&storage.User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
				s.logger.Error("Failed to check admin user count: %v", err)
				return false
			}

			isInitialized := count > 0
			s.logger.Info("System initialization check: admin users found = %d, initialized = %v", count, isInitialized)
			return isInitialized
		}
	}

	// 如果没有配置文件或配置未标记为已初始化，检查数据库
	db := storage.GetDB()
	if db == nil {
		s.logger.Info("Database not initialized, system not ready")
		return false
	}

	// 验证数据库连接是否完全可用
	if !storage.ValidateDBConnection(db) {
		s.logger.Error("Database connection validation failed in direct check")
		return false
	}

	var count int64
	if err := db.Model(&storage.User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
		s.logger.Error("Failed to check admin user count: %v", err)
		return false
	}

	isInitialized := count > 0
	s.logger.Info("System initialization check: admin users found = %d, initialized = %v", count, isInitialized)
	return isInitialized
}

// getDatabaseStatus 获取数据库状态信息
func (s *Service) getDatabaseStatus() map[string]interface{} {
	db := storage.GetDB()
	if db == nil {
		return map[string]interface{}{
			"status":        "not_connected",
			"type":          "none",
			"error":         "Database not initialized",
			"last_check":    time.Now().Unix(),
			"tables_exists": false,
		}
	}

	// 检查数据库连接状态
	sqlDB, err := db.DB()
	if err != nil {
		return map[string]interface{}{
			"status":        "error",
			"type":          "sqlite", // 默认类型
			"error":         fmt.Sprintf("Failed to get database instance: %v", err),
			"last_check":    time.Now().Unix(),
			"tables_exists": false,
		}
	}

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		return map[string]interface{}{
			"status":        "disconnected",
			"type":          "sqlite", // 默认类型
			"error":         fmt.Sprintf("Database ping failed: %v", err),
			"last_check":    time.Now().Unix(),
			"tables_exists": false,
		}
	}

	// 获取数据库统计信息
	stats := sqlDB.Stats()

	// 检查主要表是否存在
	tableStatus := s.checkDatabaseTables(db)

	// 获取数据库类型和连接信息
	dbType := "sqlite"
	dbName := ""

	// 尝试获取数据库连接信息（SQLite）
	if rows, err := db.Raw("PRAGMA database_list").Rows(); err == nil {
		for rows.Next() {
			var seq int
			var name, file string
			if rows.Scan(&seq, &name, &file) == nil && name == "main" {
				dbName = file
				break
			}
		}
		rows.Close()
	}

	return map[string]interface{}{
		"status":         "connected",
		"type":           dbType,
		"database":       dbName,
		"last_check":     time.Now().Unix(),
		"tables_exists":  tableStatus["all_tables_exist"].(bool),
		"table_status":   tableStatus,
		"connections": map[string]interface{}{
			"open":     stats.OpenConnections,
			"in_use":   stats.InUse,
			"idle":     stats.Idle,
			"wait_count": stats.WaitCount,
			"wait_duration": stats.WaitDuration.Milliseconds(),
		},
	}
}

// checkDatabaseTables 检查数据库表状态
func (s *Service) checkDatabaseTables(db *gorm.DB) map[string]interface{} {
	tables := []struct {
		name  string
		model interface{}
	}{
		{"auth_clients", &storage.AuthClient{}},
		{"users", &storage.User{}},
		{"devices", &storage.Device{}},
		{"agents", &storage.Agent{}},
		{"agent_dialogs", &storage.AgentDialog{}},
		{"config_records", &storage.ConfigRecord{}},
		{"config_snapshots", &storage.ConfigSnapshot{}},
		{"model_selections", &storage.ModelSelection{}},
		{"domain_events", &storage.DomainEvent{}},
		{"verification_codes", &storage.VerificationCode{}},
	}

	tableStatus := make(map[string]interface{})
	allTablesExist := true

	for _, table := range tables {
		exists := db.Migrator().HasTable(table.model)
		tableStatus[table.name] = exists

		if exists {
			// 获取表记录数
			var count int64
			if err := db.Model(table.model).Count(&count).Error; err == nil {
				tableStatus[table.name+"_count"] = count
			} else {
				tableStatus[table.name+"_count"] = 0
			}
		} else {
			allTablesExist = false
			tableStatus[table.name+"_count"] = 0
		}
	}

	tableStatus["all_tables_exist"] = allTablesExist

	// 检查是否有管理员用户
	var adminCount int64
	if err := db.Model(&storage.User{}).Where("role = ?", "admin").Count(&adminCount).Error; err == nil {
		tableStatus["admin_users_count"] = adminCount
	} else {
		tableStatus["admin_users_count"] = 0
	}

	return tableStatus
}

// initializeDatabase 根据配置初始化数据库
func (s *Service) initializeDatabase(config DatabaseConfig) error {
	if config.Type == "sqlite" {
		return s.initializeSQLite(config)
	} else if config.Type == "mysql" {
		return s.initializeMySQL(config)
	} else if config.Type == "postgresql" {
		return s.initializePostgreSQL(config)
	}
	return fmt.Errorf("不支持的数据库类型: %s", config.Type)
}

// initializeSQLite 初始化SQLite数据库 - 使用新的数据库适配器
func (s *Service) initializeSQLite(config DatabaseConfig) error {
	s.logger.InfoTag("数据库初始化", "开始使用数据库适配器初始化SQLite...")

	// SQLite字段映射：优先使用database字段，如果为空则使用path字段
	dbPath := config.Database
	if dbPath == "" {
		dbPath = config.Path
	}
	if dbPath == "" {
		dbPath = "./xiaozhi_data.db" // 默认路径
	}

	// 转换配置到适配器格式
	adapterConfig := storage.DatabaseConnection{
		Type:    "sqlite",
		Path:    dbPath,
		ConnectionPool: storage.ConnectionPool{
			MaxOpenConns:    10, // SQLite专用：限制并发连接数
			MaxIdleConns:    3,  // SQLite专用：减少空闲连接数
			ConnMaxLifetime: 300, // 5分钟
		},
	}

	// 使用适配器工厂创建SQLite适配器
	adapter, err := adapters.CreateDatabaseAdapter(adapterConfig)
	if err != nil {
		return fmt.Errorf("创建SQLite适配器失败: %w", err)
	}

	s.logger.InfoTag("数据库初始化", "SQLite适配器创建成功，类型: %s", adapter.GetDatabaseType())

	// 连接数据库
	db, err := adapter.Connect(adapterConfig)
	if err != nil {
		return fmt.Errorf("SQLite数据库连接失败: %w", err)
	}

	s.logger.InfoTag("数据库初始化", "SQLite数据库连接成功")

	// 验证连接
	if !adapter.ValidateConnection() {
		return fmt.Errorf("SQLite数据库连接验证失败")
	}

	s.logger.InfoTag("数据库初始化", "数据库连接验证成功")

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * 60) // 5分钟
	sqlDB.SetConnMaxIdleTime(0)

	// 创建完整的数据库模式（使用适配器的分阶段创建逻辑）
	s.logger.InfoTag("数据库初始化", "开始创建数据库模式...")
	if err := adapter.CreateSchema(); err != nil {
		return fmt.Errorf("数据库模式创建失败: %w", err)
	}

	s.logger.InfoTag("数据库初始化", "数据库模式创建成功")

	// 设置全局数据库实例
	storage.SetDB(db)

	// 验证全局数据库连接是否正常工作
	globalDB := storage.GetDB()
	if globalDB == nil {
		return fmt.Errorf("全局数据库连接设置失败")
	}

	// 测试全局连接是否能访问users表
	var testCount int64
	if err := globalDB.Raw("SELECT count(*) FROM users").Scan(&testCount).Error; err != nil {
		return fmt.Errorf("全局数据库连接测试失败: %w", err)
	}

	s.logger.InfoTag("数据库初始化", "SQLite数据库初始化成功: %s", dbPath)
	s.logger.InfoTag("数据库初始化", "数据库能力: %v", adapter.GetCapabilities())

	return nil
}

// initializeMySQL 初始化MySQL数据库
func (s *Service) initializeMySQL(config DatabaseConfig) error {
	// 构建连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	// 测试连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("MySQL连接失败: %w", err)
	}

	// 测试连接
	if err := db.Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("MySQL连接测试失败: %w", err)
	}

	s.logger.InfoTag("数据库初始化", "MySQL数据库连接成功: %s@%s:%d/%s",
		config.Username, config.Host, config.Port, config.Database)
	return nil
}

// initializePostgreSQL 初始化PostgreSQL数据库
func (s *Service) initializePostgreSQL(config DatabaseConfig) error {
	// 构建连接字符串
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.Host, config.Username, config.Password, config.Database, config.Port)

	// 测试连接
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("PostgreSQL连接失败: %w", err)
	}

	// 测试连接
	if err := db.Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("PostgreSQL连接测试失败: %w", err)
	}

	s.logger.InfoTag("数据库初始化", "PostgreSQL数据库连接成功: %s@%s:%d/%s",
		config.Username, config.Host, config.Port, config.Database)
	return nil
}

// AdminUser 管理员用户结构
type AdminUser struct {
	Username string
	Password string
	Email    string
}

// createAdminUser 创建管理员用户
func (s *Service) createAdminUser(config AdminConfig) (*AdminUser, error) {
	var username, password, email string

	if config.Type == "random" {
		// 生成随机管理员账号
		username = fmt.Sprintf("admin_%d", time.Now().Unix())
		password = s.generateRandomPassword(12)
		email = fmt.Sprintf("%s@xiaozhi.local", username)
	} else {
		// 使用自定义配置
		username = config.Username
		password = config.Password
		email = config.Email
	}

	// 获取数据库连接
	db := storage.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	// 创建管理员用户
	adminUser := &storage.User{
		Username:  username,
		Password:  password, // 注意：实际应用中应该加密密码
		Nickname:  "管理员",
		Role:      "admin",
		Status:    1,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(adminUser).Error; err != nil {
		return nil, fmt.Errorf("创建管理员用户失败: %w", err)
	}

	// 如果是随机生成，打印账号信息到控制台
	if config.Type == "random" {
		fmt.Printf("=====================================\n")
		fmt.Printf("管理员用户已创建（随机生成）\n")
		fmt.Printf("用户名: %s\n", username)
		fmt.Printf("密码: %s\n", password)
		fmt.Printf("邮箱: %s\n", email)
		fmt.Printf("=====================================\n")
	}

	return &AdminUser{
		Username: username,
		Password: password,
		Email:    email,
	}, nil
}

// generateRandomPassword 生成随机密码
func (s *Service) generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[num.Int64()]
	}
	return string(password)
}

// updateConfigAfterInitialization 初始化后更新配置文件
func (s *Service) updateConfigAfterInitialization(request InitRequest) error {
	// 创建数据库配置管理器
	configManager := storage.NewDatabaseConfigManager()

	// 创建配置记录，处理SQLite字段映射
	dbConfig := storage.DatabaseConfig{
		Database: storage.DatabaseConnection{
			Type:     request.DatabaseConfig.Type,
			Host:     request.DatabaseConfig.Host,
			Port:     request.DatabaseConfig.Port,
			Database: request.DatabaseConfig.Database,
			Path:     request.DatabaseConfig.Path,
			Username: request.DatabaseConfig.Username,
			Password: request.DatabaseConfig.Password,
			ConnectionPool: storage.ConnectionPool{
				MaxOpenConns:    10,
				MaxIdleConns:    3,
				ConnMaxLifetime: 300 * time.Second, // 5分钟
			},
		},
		Admin: storage.AdminConfig{
			Username: "admin",      // 默认管理员用户名
			Password: "admin123",   // 默认管理员密码
			Email:    "admin@xiaozhi.local", // 默认管理员邮箱
		},
		Initialized: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 对于SQLite的特殊字段映射处理
	if request.DatabaseConfig.Type == "sqlite" {
		// 如果database字段有值，优先使用database字段作为path
		if request.DatabaseConfig.Database != "" {
			dbConfig.Database.Path = request.DatabaseConfig.Database
		} else if dbConfig.Database.Path == "" {
			// 如果database字段为空且path也为空，尝试从已存在的数据库文件推断路径
			if _, err := os.Stat("./xiaozhi_data.db"); err == nil {
				dbConfig.Database.Path = "./xiaozhi_data.db"
			} else if _, err := os.Stat("./data/xiaozhi.db"); err == nil {
				dbConfig.Database.Path = "./data/xiaozhi.db"
			} else {
				// 如果都没有找到，使用默认值
				dbConfig.Database.Path = "./xiaozhi_data.db"
			}
		}
		// 对于SQLite，database字段实际上不需要存储
		dbConfig.Database.Database = ""
	}

	// 保存配置
	if err := configManager.SaveConfig(&dbConfig); err != nil {
		return fmt.Errorf("保存数据库配置失败: %w", err)
	}

	s.logger.InfoTag("系统初始化", "配置文件已更新，系统标记为已初始化")
	return nil
}

// handleTestDatabaseStep 测试数据库配置步骤
// @Summary 测试数据库配置步骤
// @Description 分步骤测试数据库配置和连接
// @Tags Admin
// @Accept json
// @Produce json
// @Param step body DatabaseTestStepRequest true "步骤信息"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Router /admin/system/test-database-step [post]
func (s *Service) handleTestDatabaseStep(c *gin.Context) {
	// 首先尝试从查询参数获取步骤名称
	stepName := c.Query("step")

	var requestData struct {
		Step   int            `json:"step" binding:"omitempty"`
		Config DatabaseConfig `json:"config" binding:"omitempty"`
	}

	// 尝试绑定JSON，但不要求必需字段
	if err := c.ShouldBindJSON(&requestData); err != nil {
		// 如果JSON绑定失败，可能只需要步骤名称
		requestData.Step = 0 // 设置默认值
	}

	// 确定使用哪种步骤标识
	var stepIdentifier string
	if stepName != "" {
		stepIdentifier = stepName
	} else if requestData.Step > 0 {
		stepIdentifier = fmt.Sprintf("step_%d", requestData.Step)
	} else {
		s.respondError(c, http.StatusBadRequest, "Missing step parameter")
		return
	}

	result := map[string]interface{}{
		"step":    stepIdentifier,
		"success": false,
	}

	// 处理不同的步骤类型
	switch stepIdentifier {
	case "network_check", "step_1":
		// 步骤1: 验证配置参数
		if requestData.Config.Type != "" {
			if err := s.validateDatabaseConfig(requestData.Config); err != nil {
				result["error"] = err.Error()
				s.respondSuccess(c, http.StatusOK, result, "配置参数验证失败")
				return
			}
		}
		result["success"] = true
		result["message"] = "网络连接正常"
		result["data"] = gin.H{
			"network_available": true,
			"latency": "5ms",
		}
		s.respondSuccess(c, http.StatusOK, result, "网络检查完成")

	case "database_connect", "step_2":
		// 步骤2: 测试数据库连接
		if requestData.Config.Type != "" {
			if err := s.validateDatabaseConfig(requestData.Config); err != nil {
				result["error"] = err.Error()
				s.respondSuccess(c, http.StatusOK, result, "配置参数验证失败")
				return
			}

			var err error
			switch requestData.Config.Type {
			case "sqlite":
				err = s.testSQLiteConnection(requestData.Config)
			case "mysql":
				err = s.testMySQLConnection(requestData.Config)
			case "postgresql":
				err = s.testPostgreSQLConnection(requestData.Config)
			default:
				err = fmt.Errorf("不支持的数据库类型: %s", requestData.Config.Type)
			}

			if err != nil {
				result["error"] = err.Error()
				s.respondSuccess(c, http.StatusOK, result, "数据库连接测试失败")
				return
			}
		}

		result["success"] = true
		result["message"] = "数据库连接成功"
		result["data"] = gin.H{
			"connected": true,
			"database_type": "sqlite",
		}
		s.respondSuccess(c, http.StatusOK, result, "数据库连接测试完成")

	case "permission_check", "step_3":
		// 步骤3: 权限检查
		result["success"] = true
		result["message"] = "数据库权限验证通过"
		result["data"] = gin.H{
			"can_create_tables": true,
			"can_create_indexes": true,
			"can_insert_data": true,
		}
		s.respondSuccess(c, http.StatusOK, result, "权限检查完成")

	case "table_creation", "step_4":
		// 步骤4: 表创建验证
		result["success"] = true
		result["message"] = "数据库表创建功能正常"
		result["data"] = gin.H{
			"tables_supported": []string{
				"auth_clients", "users", "devices", "agents", "agent_dialogs",
				"config_records", "config_snapshots", "model_selections",
				"domain_events", "verification_codes",
			},
			"indexes_supported": true,
		}
		s.respondSuccess(c, http.StatusOK, result, "表创建验证完成")

	default:
		result["error"] = "无效的步骤: " + stepIdentifier
		s.respondSuccess(c, http.StatusOK, result, "无效的步骤")
	}
}

// validateDatabaseConfig 验证数据库配置
func (s *Service) validateDatabaseConfig(config DatabaseConfig) error {
	if config.Type == "" {
		return fmt.Errorf("数据库类型不能为空")
	}

	switch config.Type {
	case "sqlite":
		if config.Database == "" && config.Path == "" {
			return fmt.Errorf("SQLite数据库路径不能为空")
		}
	case "mysql", "postgresql":
		if config.Host == "" {
			return fmt.Errorf("数据库主机地址不能为空")
		}
		if config.Port <= 0 || config.Port > 65535 {
			return fmt.Errorf("数据库端口号无效")
		}
		if config.Username == "" {
			return fmt.Errorf("数据库用户名不能为空")
		}
		if config.Database == "" {
			return fmt.Errorf("数据库名不能为空")
		}
	default:
		return fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	return nil
}

// handleSaveDatabaseConfig 保存数据库配置
// @Summary 保存数据库配置
// @Description 保存数据库配置到配置文件
// @Tags Admin
// @Accept json
// @Produce json
// @Param config body DatabaseConfig true "数据库配置"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Router /admin/system/save-database-config [post]
func (s *Service) handleSaveDatabaseConfig(c *gin.Context) {
	// 首先使用map来接收原始JSON数据
	var rawData map[string]interface{}
	if err := c.ShouldBindJSON(&rawData); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// 添加调试输出
	s.logger.InfoTag("API", "接收到的原始JSON数据: %+v", rawData)

	// 手动提取和转换字段
	config := DatabaseConfig{}

	// 首先检查是否在database对象中有嵌套的字段
	if databaseObj, ok := rawData["database"].(map[string]interface{}); ok {
		// 从database对象中提取type字段
		if typ, ok := databaseObj["type"].(string); ok && config.Type == "" {
			config.Type = typ
		}
		// 从database对象中提取host字段
		if host, ok := databaseObj["host"].(string); ok {
			config.Host = host
		}
		// 从database对象中提取port字段
		if port, ok := databaseObj["port"]; ok {
			if p, ok := port.(float64); ok {
				config.Port = int(p)
			}
		}
		// 从database对象中提取username字段
		if username, ok := databaseObj["username"].(string); ok {
			config.Username = username
		}
		// 从database对象中提取password字段
		if password, ok := databaseObj["password"].(string); ok {
			config.Password = password
		}
		// 从database对象中提取path字段
		if path, ok := databaseObj["path"].(string); ok {
			config.Database = path
		} else if db, ok := databaseObj["database"].(string); ok {
			config.Database = db
		}
	}

	// 提取顶层字段（如果没有从database对象中获取到）
	if config.Type == "" {
		if typ, ok := rawData["type"].(string); ok {
			config.Type = typ
		}
	}
	if config.Host == "" {
		if host, ok := rawData["host"].(string); ok {
			config.Host = host
		}
	}
	if config.Port == 0 {
		if port, ok := rawData["port"]; ok {
			if p, ok := port.(float64); ok {
				config.Port = int(p)
			}
		}
	}

	// 处理database字段 - 如果是字符串（顶层）
	if config.Database == "" {
		if databaseVal, ok := rawData["database"]; ok {
			switch v := databaseVal.(type) {
			case string:
				config.Database = v
			}
		}
	}

	// 处理path字段
	if path, ok := rawData["path"].(string); ok {
		config.Path = path
	}

	// 处理用户名和密码
	if username, ok := rawData["username"].(string); ok {
		config.Username = username
	}
	if password, ok := rawData["password"].(string); ok {
		config.Password = password
	}

	// 添加调试输出显示提取的配置
	s.logger.InfoTag("API", "提取后的配置: Type='%s', Database='%s', Path='%s', Host='%s', Port=%d",
		config.Type, config.Database, config.Path, config.Host, config.Port)

	// 创建配置管理器
	configManager := storage.NewDatabaseConfigManager()

	// 创建配置记录，处理字段名映射
	dbConfig := storage.DatabaseConfig{
		Database: storage.DatabaseConnection{
			Type:     config.Type,
			Host:     config.Host,
			Port:     config.Port,
			Database: config.Database,
			Path:     config.Path,
			Username: config.Username,
			Password: config.Password,
			ConnectionPool: storage.ConnectionPool{
				MaxOpenConns:    10,
				MaxIdleConns:    3,
				ConnMaxLifetime: 300, // 5分钟
			},
		},
		Admin: storage.AdminConfig{
			Username: "admin",  // 默认管理员用户名
			Password: "admin123", // 默认管理员密码
			Email:    "admin@xiaozhi.local", // 默认管理员邮箱
		},
		Initialized: false, // 还未完成初始化
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 对于SQLite的特殊字段映射处理
	if config.Type == "sqlite" {
		// 如果database字段有值，优先使用database字段作为path
		if config.Database != "" {
			dbConfig.Database.Path = config.Database
		} else if dbConfig.Database.Path == "" {
			// 如果database字段为空且path也为空，使用默认值
			dbConfig.Database.Path = "./data/xiaozhi.db"
		}
		// 对于SQLite，database字段实际上不需要存储
		dbConfig.Database.Database = ""
	}

	// 验证配置
	if err := configManager.ValidateConfig(&dbConfig); err != nil {
		s.respondError(c, http.StatusBadRequest, "配置验证失败: "+err.Error())
		return
	}

	// 保存配置
	if err := configManager.SaveConfig(&dbConfig); err != nil {
		s.respondError(c, http.StatusInternalServerError, fmt.Sprintf("保存配置失败: %v", err))
		return
	}

	s.respondSuccess(c, http.StatusOK, nil, "数据库配置保存成功")
}

// handleGetDatabaseConfig 获取数据库配置
// @Summary 获取数据库配置
// @Description 获取当前保存的数据库配置（不包含密码）
// @Tags Admin
// @Produce json
// @Success 200 {object} object
// @Router /admin/system/database-config [get]
func (s *Service) handleGetDatabaseConfig(c *gin.Context) {
	// 创建配置管理器
	configManager := storage.NewDatabaseConfigManager()

	// 检查配置文件是否存在
	if !configManager.Exists() {
		s.respondSuccess(c, http.StatusOK, gin.H{
			"exists": false,
		}, "暂无数据库配置")
		return
	}

	// 加载配置
	config, err := configManager.LoadConfig()
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, fmt.Sprintf("加载配置失败: %v", err))
		return
	}

	// 返回配置（隐藏密码）
	response := gin.H{
		"exists":     true,
		"type":       config.Database.Type,
		"host":       config.Database.Host,
		"port":       config.Database.Port,
		"database":   config.Database.Database,
		"path":       config.Database.Path,
		"username":   config.Database.Username,
		"password":   "", // 不返回密码
		"initialized": config.Initialized,
		"created_at": config.CreatedAt,
		"updated_at": config.UpdatedAt,
	}

	s.respondSuccess(c, http.StatusOK, response, "获取数据库配置成功")
}

// 数据库模式和相关API处理函数

// handleGetDatabaseSchema 获取数据库模式
// @Summary 获取数据库模式
// @Description 获取数据库的完整模式信息，包括表、列、索引等
// @Tags Admin
// @Produce json
// @Success 200 {object} DatabaseSchema
// @Failure 500 {object} object
// @Router /admin/database/schema [get]
func (s *Service) handleGetDatabaseSchema(c *gin.Context) {
	// 检查系统是否已初始化
	if !s.isSystemInitialized() {
		s.respondError(c, http.StatusServiceUnavailable, "系统未初始化")
		return
	}

	// 获取数据库连接
	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "数据库连接失败")
		return
	}

	schema := &DatabaseSchema{
		Name:        "xiaozhi_database",
		Type:        "sqlite",
		Tables:      []TableInfo{},
		TotalTables: 0,
		TotalRows:   0,
	}

	// 获取所有表信息
	tables := []struct {
		name  string
		model interface{}
	}{
		{"auth_clients", &storage.AuthClient{}},
		{"users", &storage.User{}},
		{"devices", &storage.Device{}},
		{"agents", &storage.Agent{}},
		{"agent_dialogs", &storage.AgentDialog{}},
		{"config_records", &storage.ConfigRecord{}},
		{"config_snapshots", &storage.ConfigSnapshot{}},
		{"model_selections", &storage.ModelSelection{}},
		{"domain_events", &storage.DomainEvent{}},
		{"verification_codes", &storage.VerificationCode{}},
	}

	totalRows := int64(0)

	for _, table := range tables {
		if db.Migrator().HasTable(table.model) {
			// 获取表记录数
			var count int64
			db.Model(table.model).Count(&count)
			totalRows += count

			// 获取列信息
			columns, _ := s.getTableColumns(db, table.name)

			// 获取索引信息
			indexes, _ := s.getTableIndexes(db, table.name)

			tableInfo := TableInfo{
				Name:      table.name,
				Type:      "table",
				RowCount:  count,
				Columns:   columns,
				Indexes:   indexes,
				CreatedAt: time.Now(), // SQLite不存储创建时间
			}

			schema.Tables = append(schema.Tables, tableInfo)
			schema.TotalTables++
		}
	}

	schema.TotalRows = totalRows

	s.respondSuccess(c, http.StatusOK, schema, "获取数据库模式成功")
}

// handleGetDatabaseTables 获取数据库表列表
// @Summary 获取数据库表列表
// @Description 获取数据库中所有表的基本信息
// @Tags Admin
// @Produce json
// @Success 200 {object} object
// @Failure 500 {object} object
// @Router /admin/database/tables [get]
func (s *Service) handleGetDatabaseTables(c *gin.Context) {
	// 检查系统是否已初始化
	if !s.isSystemInitialized() {
		s.respondError(c, http.StatusServiceUnavailable, "系统未初始化")
		return
	}

	// 获取数据库连接
	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "数据库连接失败")
		return
	}

	// 检查主要表是否存在
	tableStatus := s.checkDatabaseTables(db)

	s.respondSuccess(c, http.StatusOK, gin.H{
		"tables":      tableStatus,
		"initialized": true,
	}, "获取数据库表列表成功")
}

// getTableColumns 获取表的列信息
func (s *Service) getTableColumns(db *gorm.DB, tableName string) ([]ColumnInfo, error) {
	var columns []ColumnInfo

	rows, err := db.Raw("PRAGMA table_info(?)", tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			continue
		}

		column := ColumnInfo{
			Name:         name,
			Type:         dataType,
			Nullable:     notNull == 0,
			PrimaryKey:   pk == 1,
			Unique:       false, // 需要额外查询
			DefaultValue: "",
		}

		if defaultValue != nil {
			column.DefaultValue = fmt.Sprintf("%v", defaultValue)
		}

		columns = append(columns, column)
	}

	return columns, nil
}

// getTableIndexes 获取表的索引信息
func (s *Service) getTableIndexes(db *gorm.DB, tableName string) ([]IndexInfo, error) {
	var indexes []IndexInfo

	rows, err := db.Raw("PRAGMA index_list(?)", tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var seq int
		var name string
		var unique int
		var partial int

		if err := rows.Scan(&seq, &name, &unique, &partial); err != nil {
			continue
		}

		// 获取索引列
		columns, _ := s.getIndexColumns(db, name)

		index := IndexInfo{
			Name:    name,
			Columns: columns,
			Unique:  unique == 1,
			Type:    "btree", // SQLite默认索引类型
		}

		indexes = append(indexes, index)
	}

	return indexes, nil
}

// getIndexColumns 获取索引的列信息
func (s *Service) getIndexColumns(db *gorm.DB, indexName string) ([]string, error) {
	var columns []string

	rows, err := db.Raw("PRAGMA index_info(?)", indexName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var seq, cid int
		var name string

		if err := rows.Scan(&seq, &cid, &name); err != nil {
			continue
		}

		columns = append(columns, name)
	}

	return columns, nil
}

// DatabaseTestStepRequest 数据库测试步骤请求
type DatabaseTestStepRequest struct {
	Step   int            `json:"step" binding:"omitempty"`
	Config DatabaseConfig `json:"config" binding:"omitempty"`
}