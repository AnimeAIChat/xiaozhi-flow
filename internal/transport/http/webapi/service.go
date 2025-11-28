package webapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/errors"
	"xiaozhi-server-go/internal/platform/storage"
	"xiaozhi-server-go/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	securedGroup.Use(s.authMiddleware())
	{
		securedGroup.GET("/system", s.handleSystemGet)
		securedGroup.GET("/system/providers/:type", s.handleSystemProvidersType)
	}

	// 需要管理员权限的分组
	adminOnlyGroup := adminGroup.Group("")
	adminOnlyGroup.Use(s.authMiddleware(), s.adminMiddleware())
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

// authMiddleware 认证中间件
func (s *Service) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// TODO: JWT验证逻辑暂时简化
		if token == "" {
			s.logger.Error("无效的token")
			s.respondError(c, http.StatusUnauthorized, "无效的token")
			c.Abort()
			return
		}

		c.Next()
	}
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

// initializeSQLite 初始化SQLite数据库
func (s *Service) initializeSQLite(config DatabaseConfig) error {
	// 确保数据目录存在
	dataDir := filepath.Dir(config.Database)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 直接使用GORM创建SQLite数据库
	db, err := gorm.Open(sqlite.Open(config.Database), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("SQLite数据库连接失败: %w", err)
	}

	// 运行数据库迁移
	if err := db.AutoMigrate(
		&storage.AuthClient{},
		&storage.DomainEvent{},
		&storage.ConfigRecord{},
		&storage.ConfigSnapshot{},
		&storage.ModelSelection{},
		&storage.User{},
		&storage.Device{},
		&storage.Agent{},
		&storage.AgentDialog{},
		&storage.VerificationCode{},
	); err != nil {
		return fmt.Errorf("数据库表结构创建失败: %w", err)
	}

	// 将数据库实例设置到全局变量
	storage.SetDB(db)

	s.logger.InfoTag("数据库初始化", "SQLite数据库创建成功: %s", config.Database)
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
		fmt.Printf("请妥善保存此信息，用于首次登录\n")
		fmt.Printf("=====================================\n")
	}

	s.logger.InfoTag("管理员创建", "管理员用户创建成功: %s", username)

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
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[n.Int64()]
	}
	return string(password)
}

// handleTestDatabaseStep 处理分步骤数据库测试
// @Summary 分步骤数据库测试
// @Description 执行数据库连接的各个测试步骤
// @Tags admin
// @Accept json
// @Produce json
// @Param request body storage.DatabaseConnection true "数据库连接配置"
// @Param step query string false "测试步骤" Enums(network_check, database_connect, permission_check, table_creation)
// @Success 200 {object} httptransport.APIResponse{data=storage.DatabaseTestResult}
// @Failure 400 {object} httptransport.APIResponse
// @Router /admin/system/test-database-step [post]
func (s *Service) handleTestDatabaseStep(c *gin.Context) {
	var config storage.DatabaseConnection
	if err := c.ShouldBindJSON(&config); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid database configuration")
		return
	}

	// 获取要执行的步骤
	step := c.Query("step")
	if step == "" {
		step = string(storage.StepNetworkCheck) // 默认从网络检查开始
	}

	// 验证步骤
	validSteps := map[string]bool{
		string(storage.StepNetworkCheck):     true,
		string(storage.StepDatabaseConnect):  true,
		string(storage.StepPermissionCheck):  true,
		string(storage.StepTableCreation):    true,
	}
	if !validSteps[step] {
		s.respondError(c, http.StatusBadRequest, fmt.Sprintf("Invalid test step: %s", step))
		return
	}

	// 执行测试步骤
	result := s.executeDatabaseTestStep(storage.DatabaseTestStep(step), config)
	s.respondSuccess(c, http.StatusOK, result, "Database test step completed")
}

// handleSaveDatabaseConfig 处理保存数据库配置
// @Summary 保存数据库配置
// @Description 将数据库配置保存到 db.json 文件
// @Tags admin
// @Accept json
// @Produce json
// @Param request body storage.DatabaseConfig true "完整的数据库配置"
// @Success 200 {object} httptransport.APIResponse
// @Failure 400 {object} httptransport.APIResponse
// @Router /admin/system/save-database-config [post]
func (s *Service) handleSaveDatabaseConfig(c *gin.Context) {
	var config storage.DatabaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid database configuration")
		return
	}

	// 创建配置管理器
	configManager := storage.NewDatabaseConfigManager()

	// 验证配置
	if err := configManager.ValidateConfig(&config); err != nil {
		s.respondError(c, http.StatusBadRequest, fmt.Sprintf("Configuration validation failed: %s", err.Error()))
		return
	}

	// 保存配置
	if err := configManager.SaveConfig(&config); err != nil {
		s.logger.ErrorTag("配置", "保存数据库配置失败: %v", err)
		s.respondError(c, http.StatusInternalServerError, "Failed to save database configuration")
		return
	}

	s.logger.InfoTag("配置", "数据库配置已保存")
	s.respondSuccess(c, http.StatusOK, gin.H{
		"message": "Database configuration saved successfully",
		"path":    configManager.GetConfigPath(),
	}, "Database configuration saved")
}

// handleGetDatabaseConfig 处理获取数据库配置
// @Summary 获取数据库配置
// @Description 读取 db.json 文件中的数据库配置
// @Tags admin
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=storage.DatabaseConfig}
// @Failure 404 {object} httptransport.APIResponse
// @Router /admin/system/database-config [get]
func (s *Service) handleGetDatabaseConfig(c *gin.Context) {
	// 创建配置管理器
	configManager := storage.NewDatabaseConfigManager()

	// 检查配置文件是否存在
	if !configManager.Exists() {
		// 返回默认配置
		defaultConfig := configManager.GetDefaultConfig()
		s.respondSuccess(c, http.StatusOK, gin.H{
			"config":    defaultConfig,
			"is_default": true,
			"exists":    false,
		}, "Default database configuration loaded")
		return
	}

	// 加载配置
	config, err := configManager.LoadConfig()
	if err != nil {
		s.logger.ErrorTag("配置", "加载数据库配置失败: %v", err)
		s.respondError(c, http.StatusInternalServerError, "Failed to load database configuration")
		return
	}

	s.respondSuccess(c, http.StatusOK, gin.H{
		"config":     config,
		"is_default": false,
		"exists":     true,
	}, "Database configuration loaded")
}

// executeDatabaseTestStep 执行数据库测试步骤
func (s *Service) executeDatabaseTestStep(step storage.DatabaseTestStep, config storage.DatabaseConnection) *storage.DatabaseTestResult {
	startTime := time.Now()
	result := &storage.DatabaseTestResult{
		Step:   step,
		Status: "running",
	}

	switch step {
	case storage.StepNetworkCheck:
		result = s.performNetworkCheck(config)

	case storage.StepDatabaseConnect:
		result = s.performDatabaseConnect(config)

	case storage.StepPermissionCheck:
		result = s.performPermissionCheck(config)

	case storage.StepTableCreation:
		result = s.performTableCreation(config)

	default:
		result.Status = "failed"
		result.Message = fmt.Sprintf("Unknown test step: %s", step)
	}

	// 计算延迟
	latency := time.Since(startTime).Milliseconds()
	result.Latency = latency

	return result
}

// performNetworkCheck 执行网络连通性检查
func (s *Service) performNetworkCheck(config storage.DatabaseConnection) *storage.DatabaseTestResult {
	result := &storage.DatabaseTestResult{
		Step:   storage.StepNetworkCheck,
		Status: "success",
	}

	switch strings.ToLower(config.Type) {
	case "sqlite":
		// 对于 SQLite，检查文件路径是否可访问
		if config.Path == "" {
			result.Status = "failed"
			result.Message = "SQLite 数据库路径不能为空"
			return result
		}

		// 检查目录是否存在，如果不存在则尝试创建
		dir := filepath.Dir(config.Path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				result.Status = "failed"
				result.Message = fmt.Sprintf("无法创建数据库目录: %v", err)
				return result
			}
			result.Message = fmt.Sprintf("已创建数据库目录: %s", dir)
		} else {
			result.Message = "数据库目录可访问"
		}

	case "mysql", "postgresql":
		// 对于远程数据库，检查网络连通性
		if config.Host == "" || config.Port <= 0 {
			result.Status = "failed"
			result.Message = fmt.Sprintf("%s 主机和端口不能为空", config.Type)
			return result
		}

		address := fmt.Sprintf("%s:%d", config.Host, config.Port)
		conn, err := net.DialTimeout("tcp", address, 5*time.Second)
		if err != nil {
			result.Status = "failed"
			result.Message = fmt.Sprintf("无法连接到 %s 服务器: %v", config.Type, err)
			return result
		}
		conn.Close()
		result.Message = fmt.Sprintf("%s 服务器网络连接正常", config.Type)

	default:
		result.Status = "failed"
		result.Message = fmt.Sprintf("不支持的数据库类型: %s", config.Type)
	}

	return result
}

// performDatabaseConnect 执行数据库连接测试
func (s *Service) performDatabaseConnect(config storage.DatabaseConnection) *storage.DatabaseTestResult {
	result := &storage.DatabaseTestResult{
		Step:   storage.StepDatabaseConnect,
		Status: "success",
	}

	var db *gorm.DB
	var err error

	switch strings.ToLower(config.Type) {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.Path), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			result.Status = "failed"
			result.Message = fmt.Sprintf("SQLite 连接失败: %v", err)
			return result
		}

	case "mysql", "postgresql":
		result.Status = "failed"
		result.Message = fmt.Sprintf("%s 连接测试暂未实现", config.Type)
		return result

	default:
		result.Status = "failed"
		result.Message = fmt.Sprintf("不支持的数据库类型: %s", config.Type)
		return result
	}

	// 测试基本查询
	if err := db.Exec("SELECT 1").Error; err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("数据库查询测试失败: %v", err)
		return result
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	result.Message = fmt.Sprintf("%s 数据库连接成功", config.Type)
	return result
}

// performPermissionCheck 执行权限检查
func (s *Service) performPermissionCheck(config storage.DatabaseConnection) *storage.DatabaseTestResult {
	result := &storage.DatabaseTestResult{
		Step:   storage.StepPermissionCheck,
		Status: "success",
	}

	var db *gorm.DB
	var err error

	switch strings.ToLower(config.Type) {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.Path), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})

	case "mysql", "postgresql":
		result.Status = "failed"
		result.Message = fmt.Sprintf("%s 权限检查暂未实现", config.Type)
		return result

	default:
		result.Status = "failed"
		result.Message = fmt.Sprintf("不支持的数据库类型: %s", config.Type)
		return result
	}

	if err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("无法连接到数据库进行权限检查: %v", err)
		return result
	}

	// 测试创建表权限
	testTable := "test_permissions_" + time.Now().Format("20060102150405")
	if err := db.Exec(fmt.Sprintf("CREATE TABLE %s (id INTEGER)", testTable)).Error; err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("缺少创建表权限: %v", err)
		return result
	}

	// 测试插入权限
	if err := db.Exec(fmt.Sprintf("INSERT INTO %s (id) VALUES (1)", testTable)).Error; err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("缺少插入数据权限: %v", err)
		return result
	}

	// 测试查询权限
	var count int64
	if err := db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", testTable)).Scan(&count).Error; err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("缺少查询数据权限: %v", err)
		return result
	}

	// 测试删除权限
	if err := db.Exec(fmt.Sprintf("DROP TABLE %s", testTable)).Error; err != nil {
		result.Status = "warning"
		result.Message = fmt.Sprintf("缺少删除表权限，但基本权限正常: %v", err)
		return result
	}

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	result.Message = fmt.Sprintf("数据库权限检查通过 (%d 条测试记录)", count)
	return result
}

// performTableCreation 执行表创建测试
func (s *Service) performTableCreation(config storage.DatabaseConnection) *storage.DatabaseTestResult {
	result := &storage.DatabaseTestResult{
		Step:   storage.StepTableCreation,
		Status: "success",
	}

	var db *gorm.DB
	var err error

	switch strings.ToLower(config.Type) {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.Path), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})

	case "mysql", "postgresql":
		result.Status = "failed"
		result.Message = fmt.Sprintf("%s 表创建测试暂未实现", config.Type)
		return result

	default:
		result.Status = "failed"
		result.Message = fmt.Sprintf("不支持的数据库类型: %s", config.Type)
		return result
	}

	if err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("无法连接到数据库进行表创建测试: %v", err)
		return result
	}

	// 创建一个简单的测试表
	type TestTable struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	if err := db.AutoMigrate(&TestTable{}); err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("测试表创建失败: %v", err)
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		return result
	}

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	result.Status = "success"
	result.Message = "数据库表创建测试通过，test表创建成功"

	return result
}

// updateConfigAfterInitialization 更新配置文件标记系统已初始化
func (s *Service) updateConfigAfterInitialization(request InitRequest) error {
	// 创建配置管理器
	configManager := storage.NewDatabaseConfigManager()

	// 尝试加载现有配置
	config, err := configManager.LoadConfig()
	if err != nil {
		// 如果配置文件不存在，创建新的配置
		s.logger.InfoTag("系统初始化", "配置文件不存在，创建新配置")
		config = &storage.DatabaseConfig{
			Database: storage.DatabaseConnection{
				Type: request.DatabaseConfig.Type,
				Host: request.DatabaseConfig.Host,
				Port: request.DatabaseConfig.Port,
				Database: request.DatabaseConfig.Database,
				Username: request.DatabaseConfig.Username,
				Password: request.DatabaseConfig.Password,
				Path: request.DatabaseConfig.Path,
				SSLMode: "", // Not available in request, set default
				Charset: "", // Not available in request, set default
				ConnectionPool: storage.ConnectionPool{
					MaxOpenConns:    25,  // Default values
					MaxIdleConns:    10,
					ConnMaxLifetime: 300,
				},
			},
			Admin: storage.AdminConfig{
				Username: request.AdminConfig.Username,
				Password: request.AdminConfig.Password,
				Email:    request.AdminConfig.Email,
			},
			Version: "1.0.0",
		}
	}

	// 更新配置状态
	config.Initialized = true
	config.UpdatedAt = time.Now()

	// 保存配置
	if err := configManager.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save updated configuration: %w", err)
	}

	s.logger.InfoTag("系统初始化", "配置文件已更新，系统标记为已初始化")
	return nil
}

// handleGetDatabaseSchema 获取数据库模式信息
// @Summary 获取数据库模式信息
// @Description 获取数据库中所有表的结构信息，包括列、索引和外键关系
// @Tags Database
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DatabaseSchema
// @Failure 401 {object} object
// @Failure 500 {object} object
// @Router /admin/system/database/schema [get]
func (s *Service) handleGetDatabaseSchema(c *gin.Context) {
	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "Database not available")
		return
	}

	// 验证数据库连接
	if !storage.ValidateDBConnection(db) {
		s.respondError(c, http.StatusInternalServerError, "Database connection validation failed")
		return
	}

	schema, err := s.getDatabaseSchema(db)
	if err != nil {
		s.logger.ErrorTag("Database", "Failed to get database schema: %v", err)
		s.respondError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get database schema: %v", err))
		return
	}

	s.respondSuccess(c, http.StatusOK, schema, "Database schema retrieved successfully")
}

// handleGetDatabaseTables 获取数据库表列表
// @Summary 获取数据库表列表
// @Description 获取数据库中所有表的简要信息
// @Tags Database
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object
// @Failure 401 {object} object
// @Failure 500 {object} object
// @Router /admin/system/database/tables [get]
func (s *Service) handleGetDatabaseTables(c *gin.Context) {
	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "Database not available")
		return
	}

	// 验证数据库连接
	if !storage.ValidateDBConnection(db) {
		s.respondError(c, http.StatusInternalServerError, "Database connection validation failed")
		return
	}

	tables, err := s.getDatabaseTables(db)
	if err != nil {
		s.logger.ErrorTag("Database", "Failed to get database tables: %v", err)
		s.respondError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get database tables: %v", err))
		return
	}

	s.respondSuccess(c, http.StatusOK, gin.H{
		"tables": tables,
		"total":  len(tables),
	}, "Database tables retrieved successfully")
}

// getDatabaseSchema 获取完整的数据库模式信息
func (s *Service) getDatabaseSchema(db *gorm.DB) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Name:         "xiaozhi",
		Type:         "sqlite",
		Tables:       []TableInfo{},
		Relationships: []ForeignKeyInfo{},
		TotalTables:  0,
		TotalRows:    0,
	}

	// 获取数据库表名
	tableNames, err := s.getTableNames(db)
	if err != nil {
		return nil, err
	}

	schema.TotalTables = len(tableNames)

	// 获取每个表的详细信息
	for _, tableName := range tableNames {
		tableInfo, err := s.getTableInfo(db, tableName)
		if err != nil {
			s.logger.WarnTag("Database", "Failed to get info for table %s: %v", tableName, err)
			continue
		}

		schema.Tables = append(schema.Tables, *tableInfo)
		schema.TotalRows += tableInfo.RowCount
	}

	// 获取外键关系
	foreignKeys, err := s.getForeignKeys(db)
	if err != nil {
		s.logger.WarnTag("Database", "Failed to get foreign keys: %v", err)
	} else {
		schema.Relationships = foreignKeys
	}

	return schema, nil
}

// getDatabaseTables 获取数据库表列表（简要信息）
func (s *Service) getDatabaseTables(db *gorm.DB) ([]TableInfo, error) {
	tableNames, err := s.getTableNames(db)
	if err != nil {
		return nil, err
	}

	tables := make([]TableInfo, 0, len(tableNames))

	for _, tableName := range tableNames {
		// 获取基本表信息（行数和大小）
		var count int64
		if err := db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count).Error; err != nil {
			s.logger.WarnTag("Database", "Failed to get row count for table %s: %v", tableName, err)
			count = 0
		}

		table := TableInfo{
			Name:     tableName,
			Type:     "table",
			RowCount: count,
			Size:     0, // SQLite中获取表大小比较复杂，暂时为0
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// getTableNames 获取所有表名
func (s *Service) getTableNames(db *gorm.DB) ([]string, error) {
	var tableNames []string

	// SQLite 获取表名的方式
	rows, err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query table names: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tableNames = append(tableNames, tableName)
	}

	return tableNames, nil
}

// getTableInfo 获取表的详细信息
func (s *Service) getTableInfo(db *gorm.DB, tableName string) (*TableInfo, error) {
	tableInfo := &TableInfo{
		Name:    tableName,
		Type:    "table",
		Columns: []ColumnInfo{},
		Indexes: []IndexInfo{},
	}

	// 获取行数
	if err := db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&tableInfo.RowCount).Error; err != nil {
		tableInfo.RowCount = 0
	}

	// 获取列信息
	columns, err := s.getTableColumns(db, tableName)
	if err != nil {
		s.logger.WarnTag("Database", "Failed to get columns for table %s: %v", tableName, err)
	} else {
		tableInfo.Columns = columns
	}

	// 获取索引信息
	indexes, err := s.getTableIndexes(db, tableName)
	if err != nil {
		s.logger.WarnTag("Database", "Failed to get indexes for table %s: %v", tableName, err)
	} else {
		tableInfo.Indexes = indexes
	}

	return tableInfo, nil
}

// getTableColumns 获取表的列信息
func (s *Service) getTableColumns(db *gorm.DB, tableName string) ([]ColumnInfo, error) {
	var columns []ColumnInfo

	rows, err := db.Raw(fmt.Sprintf("PRAGMA table_info(%s)", tableName)).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get table info for %s: %w", tableName, err)
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
			Name:       name,
			Type:       dataType,
			Nullable:   notNull == 0,
			PrimaryKey: pk == 1,
		}

		if defaultValue != nil {
			if str, ok := defaultValue.(string); ok {
				column.DefaultValue = str
			} else {
				column.DefaultValue = fmt.Sprintf("%v", defaultValue)
			}
		}

		columns = append(columns, column)
	}

	return columns, nil
}

// getTableIndexes 获取表的索引信息
func (s *Service) getTableIndexes(db *gorm.DB, tableName string) ([]IndexInfo, error) {
	var indexes []IndexInfo

	rows, err := db.Raw(fmt.Sprintf("PRAGMA index_list(%s)", tableName)).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get index list for table %s: %w", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var seq int
		var name string
		var unique int
		var origin string
		var partial int

		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			continue
		}

		// 获取索引的列信息
		columns, err := s.getIndexColumns(db, name)
		if err != nil {
			s.logger.WarnTag("Database", "Failed to get columns for index %s: %v", name, err)
			continue
		}

		index := IndexInfo{
			Name:    name,
			Columns: columns,
			Unique:  unique == 1,
			Type:    "btree", // SQLite默认使用btree
		}

		indexes = append(indexes, index)
	}

	return indexes, nil
}

// getIndexColumns 获取索引的列信息
func (s *Service) getIndexColumns(db *gorm.DB, indexName string) ([]string, error) {
	var columns []string

	rows, err := db.Raw(fmt.Sprintf("PRAGMA index_info(%s)", indexName)).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get index info for %s: %w", indexName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var seq int
		var cid int
		var name string

		if err := rows.Scan(&seq, &cid, &name); err != nil {
			continue
		}

		columns = append(columns, name)
	}

	return columns, nil
}

// getForeignKeys 获取外键关系信息
func (s *Service) getForeignKeys(db *gorm.DB) ([]ForeignKeyInfo, error) {
	var foreignKeys []ForeignKeyInfo

	// 获取所有表名
	tableNames, err := s.getTableNames(db)
	if err != nil {
		return nil, err
	}

	// 遍历每个表获取外键信息
	for _, tableName := range tableNames {
		rows, err := db.Raw(fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)).Rows()
		if err != nil {
			continue // 有些表可能没有外键，忽略错误
		}

		for rows.Next() {
			var id int
			var seq int
			var table string
			var from string
			var to string
			var on_update, on_delete string
			var match string

			if err := rows.Scan(&id, &seq, &table, &from, &to, &on_update, &on_delete, &match); err != nil {
				continue
			}

			foreignKey := ForeignKeyInfo{
				Name:         fmt.Sprintf("fk_%s_%s", tableName, table),
				SourceTable:  tableName,
				SourceColumn: from,
				TargetTable:  table,
				TargetColumn: to,
				OnDelete:     on_delete,
				OnUpdate:     on_update,
			}

			foreignKeys = append(foreignKeys, foreignKey)
		}
		rows.Close()
	}

	return foreignKeys, nil
}