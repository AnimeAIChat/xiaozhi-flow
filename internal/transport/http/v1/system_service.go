package v1

import (
	"xiaozhi-server-go/internal/platform/logging"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/transport/http/types/v1"
	httpUtils "xiaozhi-server-go/internal/transport/http/utils"
)

// SystemServiceV1 V1版本系统服务
type SystemServiceV1 struct {
	logger *logging.Logger
	config *config.Config
	// TODO: 添加实际的业务逻辑依赖
}

// NewSystemServiceV1 创建系统服务V1实例
func NewSystemServiceV1(config *config.Config, logger *logging.Logger) (*SystemServiceV1, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &SystemServiceV1{
		logger: logger,
		config: config,
	}, nil
}

// Register 注册系统API路由
func (s *SystemServiceV1) Register(router *gin.RouterGroup) {
	// 新的统一系统API（推荐使用）
	router.GET("/system", s.getUnifiedSystemInfo)           // 获取系统信息（统一接口）
	router.POST("/system", s.executeSystemOperation)         // 执行系统操作（统一接口）

	// === 旧版API（保持向后兼容，将来可以废弃） ===

	// 系统状态和健康检查
	system := router.Group("/system")
	{
		system.GET("/status", s.getSystemStatus)        // 获取系统状态
		system.GET("/health", s.healthCheck)            // 健康检查
		system.POST("/health", s.detailedHealthCheck)   // 详细健康检查
		system.GET("/time", s.getServerTime)            // 获取服务器时间
	}

	// 系统配置管理（需要管理员权限）
	configs := router.Group("/config")
	{
		configs.GET("", s.listConfigs)                 // 获取系统配置
		configs.POST("", s.updateConfig)               // 更新系统配置
	}

	// 供应商管理
	providers := router.Group("/providers")
	{
		providers.GET("", s.listProviders)             // 获取供应商列表
		providers.POST("", s.createProvider)           // 创建供应商配置
		providers.PUT("/:type/:name", s.updateProvider) // 更新供应商配置
		providers.DELETE("/:type/:name", s.deleteProvider) // 删除供应商配置
	}
}

// getSystemStatus 获取系统状态
// @Summary 获取系统状态
// @Description 获取系统的运行状态和统计信息
// @Tags System
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=v1.SystemStatus}
// @Router /v1/system/status [get]
func (s *SystemServiceV1) getSystemStatus(c *gin.Context) {
	s.logger.InfoTag("API", "获取系统状态",
		"request_id", getRequestID(c),
	)

	// 模拟系统状态数据
	status := s.getMockSystemStatus()

	httpUtils.Response.Success(c, status, "获取系统状态成功")
}

// healthCheck 健康检查
// @Summary 健康检查
// @Description 快速健康检查
// @Tags System
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=v1.HealthCheckResponse}
// @Router /v1/system/health [get]
func (s *SystemServiceV1) healthCheck(c *gin.Context) {
	s.logger.InfoTag("API", "健康检查",
		"request_id", getRequestID(c),
	)

	// 模拟健康检查结果
	response := v1.HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Duration:  15, // 15ms
		Checks: []v1.HealthCheckResult{
			{
				Name:     "database",
				Status:   "healthy",
				Duration: 5,
			},
			{
				Name:     "memory",
				Status:   "healthy",
				Duration: 3,
			},
			{
				Name:     "disk",
				Status:   "healthy",
				Duration: 7,
			},
		},
		Overall: "healthy",
	}

	httpUtils.Response.Success(c, response, "健康检查完成")
}

// detailedHealthCheck 详细健康检查
// @Summary 详细健康检查
// @Description 执行详细的健康检查，支持指定检查项目
// @Tags System
// @Accept json
// @Produce json
// @Param request body v1.HealthCheckRequest false "健康检查参数"
// @Success 200 {object} httptransport.APIResponse{data=v1.HealthCheckResponse}
// @Router /v1/system/health [post]
func (s *SystemServiceV1) detailedHealthCheck(c *gin.Context) {
	var request v1.HealthCheckRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "详细健康检查",
		"checks", request.Checks,
		"timeout", request.Timeout,
		"request_id", getRequestID(c),
	)

	// 模拟详细健康检查
	checks := []v1.HealthCheckResult{
		{
			Name:     "database",
			Status:   "healthy",
			Message:  "数据库连接正常",
			Duration: 12,
			Details:  gin.H{"connections": 5, "queries_per_second": 120},
		},
		{
			Name:     "redis",
			Status:   "healthy",
			Message:  "Redis连接正常",
			Duration: 3,
			Details:  gin.H{"memory_usage": "45MB", "connected_clients": 3},
		},
	}

	if len(request.Checks) > 0 {
		// 过滤指定的检查项
		var filteredChecks []v1.HealthCheckResult
		for _, checkName := range request.Checks {
			for _, check := range checks {
				if check.Name == checkName {
					filteredChecks = append(filteredChecks, check)
					break
				}
			}
		}
		checks = filteredChecks
	}

	response := v1.HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Duration:  25,
		Checks:    checks,
		Overall:   "healthy",
	}

	httpUtils.Response.Success(c, response, "详细健康检查完成")
}

// getServerTime 获取服务器时间
// @Summary 获取服务器时间
// @Description 获取服务器当前时间和相关信息
// @Tags System
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=v1.ServerTimeInfo}
// @Router /v1/system/time [get]
func (s *SystemServiceV1) getServerTime(c *gin.Context) {
	s.logger.InfoTag("API", "获取服务器时间",
		"request_id", getRequestID(c),
	)

	uptime := time.Since(time.Now().Add(-24 * time.Hour)).Seconds() // 模拟24小时运行时间
	serverTime := v1.ServerTimeInfo{
		CurrentTime: time.Now(),
		Timezone:    "UTC",
		Uptime:      int64(uptime),
		Load:        0.45, // 模拟CPU负载
	}

	httpUtils.Response.Success(c, serverTime, "获取服务器时间成功")
}

// listConfigs 获取系统配置
// @Summary 获取系统配置
// @Description 获取系统配置项列表
// @Tags System
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=[]v1.SystemConfigResponse}
// @Router /v1/system/config [get]
func (s *SystemServiceV1) listConfigs(c *gin.Context) {
	s.logger.InfoTag("API", "获取系统配置",
		"request_id", getRequestID(c),
	)

	// 模拟系统配置数据
	configs := []v1.SystemConfigResponse{
		{
			Key:       "app_name",
			Value:     "XiaoZhi Server",
			Type:      "string",
			UpdatedBy: "admin",
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			Key:       "max_connections",
			Value:     1000,
			Type:      "number",
			UpdatedBy: "admin",
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			Key:       "enable_metrics",
			Value:     true,
			Type:      "boolean",
			UpdatedBy: "admin",
			UpdatedAt: time.Now().Add(-30 * time.Minute),
		},
	}

	httpUtils.Response.Success(c, configs, "获取系统配置成功")
}



// updateConfig 更新系统配置
// @Summary 更新系统配置
// @Description 更新指定的系统配置项
// @Tags System
// @Accept json
// @Produce json
// @Param request body v1.SystemConfigRequest true "配置更新信息"
// @Success 200 {object} httptransport.APIResponse{data=v1.SystemConfigResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Router /v1/system/config [post]
func (s *SystemServiceV1) updateConfig(c *gin.Context) {
	var request v1.SystemConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "更新系统配置",
		"setting", request.Setting,
		"request_id", getRequestID(c),
	)

	// 验证配置项是否存在
	validSettings := []string{"app_name", "max_connections", "enable_metrics", "debug_mode"}
	valid := false
	for _, setting := range validSettings {
		if request.Setting == setting {
			valid = true
			break
		}
	}

	if !valid {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInvalidInput, "无效的配置项")
		return
	}

	// 模拟配置更新
	response := v1.SystemConfigResponse{
		Key:       request.Setting,
		Value:     request.Value,
		Type:      "unknown", // 根据实际值类型判断
		UpdatedBy: "admin",   // 实际应该从认证信息中获取
		UpdatedAt: time.Now(),
	}

	httpUtils.Response.Success(c, response, "配置更新成功")
}

// listProviders 获取供应商列表
// @Summary 获取供应商列表
// @Description 获取所有配置的AI服务供应商
// @Tags Providers
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=v1.ProviderList}
// @Router /v1/providers [get]
func (s *SystemServiceV1) listProviders(c *gin.Context) {
	s.logger.InfoTag("API", "获取供应商列表",
		"request_id", getRequestID(c),
	)

	// 模拟供应商数据
	providers := []v1.Provider{
		{
			Type:     "llm",
			Name:     "openai",
			Status:   "active",
			Config:   gin.H{"api_key": "sk-***", "model": "gpt-3.5-turbo"},
			Metadata: gin.H{"description": "OpenAI GPT模型"},
			Enabled:  true,
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			Type:     "tts",
			Name:     "azure-tts",
			Status:   "active",
			Config:   gin.H{"api_key": "***", "region": "eastus"},
			Metadata: gin.H{"description": "Azure文字转语音"},
			Enabled:  true,
			CreatedAt: time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			Type:     "asr",
			Name:     "whisper",
			Status:   "inactive",
			Config:   gin.H{},
			Metadata: gin.H{"description": "OpenAI Whisper语音识别"},
			Enabled:  false,
			CreatedAt: time.Now().Add(-3 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	response := v1.ProviderList{
		Providers: providers,
	}

	httpUtils.Response.Success(c, response, "获取供应商列表成功")
}

// createProvider 创建供应商配置
// @Summary 创建供应商配置
// @Description 创建新的AI服务供应商配置
// @Tags Providers
// @Accept json
// @Produce json
// @Param request body v1.ProviderConfigRequest true "供应商配置"
// @Success 201 {object} httptransport.APIResponse{data=v1.Provider}
// @Failure 400 {object} httptransport.APIResponse
// @Router /v1/providers [post]
func (s *SystemServiceV1) createProvider(c *gin.Context) {
	var request v1.ProviderConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "创建供应商配置",
		"type", request.Type,
		"name", request.Name,
		"enabled", request.Enabled,
		"request_id", getRequestID(c),
	)

	// 检查供应商是否已存在
	if s.providerExists(request.Type, request.Name) {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeProviderExists, "供应商已存在")
		return
	}

	// 模拟创建供应商
	provider := v1.Provider{
		Type:     request.Type,
		Name:     request.Name,
		Status:   "active",
		Config:   request.Config,
		Metadata: gin.H{},
		Enabled:  request.Enabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	httpUtils.Response.Created(c, provider, "供应商创建成功")
}

// updateProvider 更新供应商配置
// @Summary 更新供应商配置
// @Description 更新指定供应商的配置信息
// @Tags Providers
// @Accept json
// @Produce json
// @Param type path string true "供应商类型"
// @Param name path string true "供应商名称"
// @Param request body v1.ProviderConfigRequest true "供应商配置"
// @Success 200 {object} httptransport.APIResponse{data=v1.Provider}
// @Failure 400 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Router /v1/providers/{type}/{name} [put]
func (s *SystemServiceV1) updateProvider(c *gin.Context) {
	providerType := c.Param("type")
	providerName := c.Param("name")

	if providerType == "" || providerName == "" {
		httpUtils.Response.BadRequest(c, "供应商类型和名称不能为空")
		return
	}

	var request v1.ProviderConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "更新供应商配置",
		"type", providerType,
		"name", providerName,
		"enabled", request.Enabled,
		"request_id", getRequestID(c),
	)

	// 检查供应商是否存在
	if !s.providerExists(providerType, providerName) {
		httpUtils.Response.NotFound(c, "供应商")
		return
	}

	// 模拟更新供应商
	provider := v1.Provider{
		Type:     providerType,
		Name:     providerName,
		Status:   "active",
		Config:   request.Config,
		Metadata: gin.H{},
		Enabled:  request.Enabled,
		CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	httpUtils.Response.Success(c, provider, "供应商配置更新成功")
}

// deleteProvider 删除供应商配置
// @Summary 删除供应商配置
// @Description 删除指定的供应商配置
// @Tags Providers
// @Produce json
// @Param type path string true "供应商类型"
// @Param name path string true "供应商名称"
// @Success 200 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Router /v1/providers/{type}/{name} [delete]
func (s *SystemServiceV1) deleteProvider(c *gin.Context) {
	providerType := c.Param("type")
	providerName := c.Param("name")

	if providerType == "" || providerName == "" {
		httpUtils.Response.BadRequest(c, "供应商类型和名称不能为空")
		return
	}

	s.logger.InfoTag("API", "删除供应商配置",
		"type", providerType,
		"name", providerName,
		"request_id", getRequestID(c),
	)

	// 检查供应商是否存在
	if !s.providerExists(providerType, providerName) {
		httpUtils.Response.NotFound(c, "供应商")
		return
	}

	httpUtils.Response.Success(c, gin.H{"provider": fmt.Sprintf("%s/%s", providerType, providerName)}, "供应商删除成功")
}



// ========== 模拟数据方法 ==========
// TODO: 实际实现中应该从数据库或配置中获取真实数据

func (s *SystemServiceV1) getMockSystemStatus() v1.SystemStatus {
	return v1.SystemStatus{
		Status:      "running",
		Version:     "1.0.0",
		StartTime:   time.Now().Add(-24 * time.Hour),
		Uptime:      86400, // 24小时
		Environment: "production",
		GoVersion:   "1.21.0",
		Services: []v1.ServiceStatus{
			{
				Name:      "web-server",
				Status:    "running",
				Health:    "healthy",
				Uptime:    86400,
				StartTime: time.Now().Add(-24 * time.Hour),
			},
			{
				Name:      "database",
				Status:    "running",
				Health:    "healthy",
				Uptime:    86400,
				StartTime: time.Now().Add(-24 * time.Hour),
			},
			{
				Name:      "redis",
				Status:    "running",
				Health:    "healthy",
				Uptime:    86400,
				StartTime: time.Now().Add(-24 * time.Hour),
			},
		},
		Database: &v1.DatabaseStatus{
			Status:         "connected",
			Connection:     "connected",
			Type:           "mysql",
			Host:           "localhost",
			Port:           3306,
			Database:       "xiaozhi_server",
			MaxConnections: 1000,
			OpenConnections: 5,
			SlowQueries:    0,
		},
		Statistics: v1.SystemStatistics{
			TotalRequests:    100000,
			SuccessRequests:  95000,
			ErrorRequests:    5000,
			ActiveSessions:   120,
			TotalUsers:       150,
			RegisteredUsers:  120,
			StorageUsed:      2147483648, // 2GB
			MemoryUsage:      134217728,  // 128MB
			CPUUsage:         25.5,
			DiskUsage:        65.2,
		},
	}
}

func (s *SystemServiceV1) providerExists(providerType, providerName string) bool {
	// 简单模拟供应商存在性检查
	existingProviders := []string{
		"llm/openai",
		"tts/azure-tts",
		"asr/whisper",
	}

	providerKey := fmt.Sprintf("%s/%s", providerType, providerName)
	for _, existing := range existingProviders {
		if existing == providerKey {
			return true
		}
	}
	return false
}

// ===== 统一系统API =====

// getUnifiedSystemInfo 获取统一系统信息
// @Summary 获取系统信息
// @Description 获取系统状态信息，根据用户权限返回不同级别的信息
// @Tags System
// @Produce json
// @Param Authorization header string false "Bearer token"
// @Success 200 {object} httptransport.APIResponse{data=v1.UnifiedSystemInfo}
// @Router /system [get]
func (s *SystemServiceV1) getUnifiedSystemInfo(c *gin.Context) {
	s.logger.InfoTag("API", "获取统一系统信息",
		"request_id", getRequestID(c),
	)

	// 检查用户权限（简化版本，实际项目中应该从token或session中获取）
	isAdmin := s.isAdminRequest(c)

	// 构建统一系统信息
	systemInfo := s.buildUnifiedSystemInfo(isAdmin)

	httpUtils.Response.Success(c, systemInfo, "获取系统信息成功")
}

// executeSystemOperation 执行系统操作
// @Summary 执行系统操作
// @Description 执行系统健康检查、重启等操作
// @Tags System
// @Accept json
// @Produce json
// @Param request body v1.SystemOperationRequest true "系统操作请求"
// @Param Authorization header string false "Bearer token"
// @Success 200 {object} httptransport.APIResponse{data=interface{}}
// @Router /system [post]
func (s *SystemServiceV1) executeSystemOperation(c *gin.Context) {
	var request v1.SystemOperationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "执行系统操作",
		"operation", request.Operation,
		"options", request.Options,
		"request_id", getRequestID(c),
	)

	// 检查权限（管理操作需要管理员权限）
	if s.requiresAdminPermission(request.Operation) {
		if !s.isAdminRequest(c) {
			httpUtils.Response.Error(c, "403", "需要管理员权限")
			return
		}
	}

	// 执行操作
	result, err := s.executeOperation(request)
	if err != nil {
		httpUtils.Response.Error(c, "500", fmt.Sprintf("执行操作失败: %v", err))
		return
	}

	httpUtils.Response.Success(c, result, "操作执行成功")
}

// buildUnifiedSystemInfo 构建统一系统信息
func (s *SystemServiceV1) buildUnifiedSystemInfo(isAdmin bool) v1.UnifiedSystemInfo {
	now := time.Now()
	uptime := time.Since(now.Add(-24 * time.Hour)).Seconds() // 模拟24小时运行时间

	// 基础信息（所有用户可见）
	basicInfo := v1.SystemBasicInfo{
		Status:    "running",
		Uptime:    int64(uptime),
		Version:   "1.0.0",
		Timestamp: now,
	}

	// 时间信息（所有用户可见）
	timeInfo := v1.SystemTimeInfo{
		CurrentTime: now,
		Timezone:    "UTC",
		Uptime:      int64(uptime),
	}

	// 健康检查信息（所有用户可见）
	healthInfo := s.getHealthInfo()

	var adminInfo v1.SystemAdminInfo
	if isAdmin {
		// 管理员专用信息
		adminInfo = v1.SystemAdminInfo{
			Memory:   "45%",
			CPU:      0.45,
			Disk:     "120GB/500GB",
			Services: []string{"web-server", "database", "redis", "websocket"},
			Load: v1.SystemLoadInfo{
				CPU:     0.45,
				Memory:  0.45,
				Disk:    0.24,
				Network: 0.05,
			},
			Logs: v1.SystemLogsInfo{
				Level:      "INFO",
				Count:      1250,
				LastLog:    "系统运行正常",
				TotalLines: 50000,
			},
			Config: v1.SystemConfigInfo{
				Initialized: true,
				NeedsSetup:  false,
				ConfigPath:  "/etc/xiaozhi-server/config.yaml",
				Database: v1.DatabaseConfigInfo{
					Type:     "mysql",
					Host:     "localhost",
					Port:     3306,
					Database: "xiaozhi_server",
					Status:   "connected",
				},
			},
		}
	}

	return v1.UnifiedSystemInfo{
		Basic:  basicInfo,
		Admin:  adminInfo,
		Health: healthInfo,
		Time:   timeInfo,
	}
}

// getHealthInfo 获取健康检查信息
func (s *SystemServiceV1) getHealthInfo() v1.SystemHealthInfo {
	components := []v1.HealthComponent{
		{
			Name:      "database",
			Status:    "healthy",
			Latency:   15,
		},
		{
			Name:      "redis",
			Status:    "healthy",
			Latency:   5,
		},
		{
			Name:      "websocket",
			Status:    "healthy",
			Latency:   2,
		},
		{
			Name:      "api-server",
			Status:    "healthy",
			Latency:   1,
		},
	}

	return v1.SystemHealthInfo{
		Overall:    "healthy",
		Components: components,
		Timestamp:  time.Now(),
	}
}

// isAdminRequest 检查是否为管理员请求
func (s *SystemServiceV1) isAdminRequest(c *gin.Context) bool {
	// 简化版本：检查是否有Authorization header
	// 实际项目中应该解析JWT token或session来获取用户角色
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// 这里应该解析token并检查用户角色
		// 暂时简化：有token就认为是管理员
		return true
	}
	return false
}

// requiresAdminPermission 检查操作是否需要管理员权限
func (s *SystemServiceV1) requiresAdminPermission(operation string) bool {
	switch operation {
	case v1.OperationRestart, v1.OperationCleanup:
		return true
	case v1.OperationHealthCheck, v1.OperationRefresh:
		return false
	default:
		return false
	}
}

// executeOperation 执行系统操作
func (s *SystemServiceV1) executeOperation(request v1.SystemOperationRequest) (interface{}, error) {
	switch request.Operation {
	case v1.OperationHealthCheck:
		return s.performHealthCheck(request.Options)
	case v1.OperationRefresh:
		return s.performRefresh(request.Options)
	case v1.OperationRestart:
		return s.performRestart(request.Options)
	case v1.OperationCleanup:
		return s.performCleanup(request.Options)
	default:
		return nil, fmt.Errorf("不支持的操作: %s", request.Operation)
	}
}

// performHealthCheck 执行健康检查
func (s *SystemServiceV1) performHealthCheck(options map[string]interface{}) (interface{}, error) {
	s.logger.InfoTag("System", "执行健康检查", "options", options)
	return map[string]interface{}{
		"status":    "completed",
		"timestamp": time.Now(),
		"result":    s.getHealthInfo(),
	}, nil
}

// performRefresh 执行刷新操作
func (s *SystemServiceV1) performRefresh(options map[string]interface{}) (interface{}, error) {
	s.logger.InfoTag("System", "执行刷新操作", "options", options)
	return map[string]interface{}{
		"status":    "completed",
		"timestamp": time.Now(),
		"message":   "系统配置已刷新",
	}, nil
}

// performRestart 执行重启操作
func (s *SystemServiceV1) performRestart(options map[string]interface{}) (interface{}, error) {
	s.logger.InfoTag("System", "执行重启操作", "options", options)
	return map[string]interface{}{
		"status":    "completed",
		"timestamp": time.Now(),
		"message":   "重启操作已执行",
	}, nil
}

// performCleanup 执行清理操作
func (s *SystemServiceV1) performCleanup(options map[string]interface{}) (interface{}, error) {
	s.logger.InfoTag("System", "执行清理操作", "options", options)
	return map[string]interface{}{
		"status":    "completed",
		"timestamp": time.Now(),
		"message":   "清理操作已完成",
	}, nil
}

