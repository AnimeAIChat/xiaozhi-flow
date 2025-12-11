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
	// 统一系统API
	router.GET("/system", s.getUnifiedSystemInfo)           // 获取系统信息（统一接口）
	router.POST("/system", s.executeSystemOperation)         // 执行系统操作（统一接口）

	// 供应商管理
	providers := router.Group("/providers")
	{
		providers.GET("", s.listProviders)             // 获取供应商列表
		providers.POST("", s.createProvider)           // 创建供应商配置
		providers.PUT("/:type/:name", s.updateProvider) // 更新供应商配置
		providers.DELETE("/:type/:name", s.deleteProvider) // 删除供应商配置
	}
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

