package v1

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"xiaozhi-server-go/internal/domain/plugin/config"
	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/transport/http/types"
	httpUtils "xiaozhi-server-go/internal/transport/http/utils"
)

// PluginConfigController 插件配置控制器
type PluginConfigController struct {
	logger  *logging.Logger
	service config.PluginConfigService
}

// NewPluginConfigController 创建插件配置控制器
func NewPluginConfigController(service config.PluginConfigService, logger *logging.Logger) *PluginConfigController {
	if service == nil {
		panic("plugin config service is required")
	}
	if logger == nil {
		panic("logger is required")
	}

	return &PluginConfigController{
		logger:  logger,
		service: service,
	}
}

// Register 注册插件配置API路由
func (c *PluginConfigController) Register(router *gin.RouterGroup) {
	plugin := router.Group("/plugin")
	{
		providers := plugin.Group("/providers")
		{
			providers.GET("", c.listProviderConfigs)              // 获取供应商配置列表
			providers.POST("", c.createProviderConfig)             // 创建供应商配置
			providers.GET("/available", c.getAvailableProviders)   // 获取可用供应商
			providers.GET("/stats", c.getPluginStats)             // 获取插件统计信息
			providers.POST("/test", c.testProviderConfig)          // 测试供应商配置
			providers.GET("/:id", c.getProviderConfig)             // 获取指定供应商配置
			providers.PUT("/:id", c.updateProviderConfig)          // 更新供应商配置
			providers.DELETE("/:id", c.deleteProviderConfig)       // 删除供应商配置

			// 快照管理
			providers.GET("/:id/snapshots", c.getConfigSnapshots)          // 获取配置快照列表
			providers.POST("/:id/snapshots", c.createConfigSnapshot)       // 创建配置快照
			providers.POST("/:id/snapshots/restore", c.restoreConfigSnapshot) // 恢复配置快照

			// 历史记录
			providers.GET("/:id/history", c.getConfigHistory)              // 获取配置变更历史
		}
	}
}

// listProviderConfigs 获取供应商配置列表
// @Summary 获取供应商配置列表
// @Description 分页获取所有插件供应商配置
// @Tags Plugin Config
// @Produce json
// @Param provider_type query string false "供应商类型过滤"
// @Param enabled query bool false "启用状态过滤"
// @Param health_status query string false "健康状态过滤"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} httptransport.APIResponse{data=types.PluginConfigListResponse}
// @Router /api/v1/plugin/providers [get]
func (c *PluginConfigController) listProviderConfigs(ctx *gin.Context) {
	c.logger.InfoTag("API", "获取供应商配置列表", "request_id", getRequestID(ctx))

	// 解析查询参数
	filter := &config.ProviderConfigFilter{
		ProviderType: config.ProviderType(ctx.Query("provider_type")),
		Page:         parseIntQuery(ctx, "page", 1),
		PageSize:     parseIntQuery(ctx, "page_size", 20),
	}

	if enabledStr := ctx.Query("enabled"); enabledStr != "" {
		enabled := enabledStr == "true"
		filter.Enabled = &enabled
	}

	filter.HealthStatus = config.HealthStatus(ctx.Query("health_status"))

	// 调用服务
	result, err := c.service.GetProviderConfigs(ctx, filter)
	if err != nil {
		c.logger.Error("获取供应商配置列表失败", "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "获取供应商配置列表失败")
		return
	}

	// 转换为响应格式
	response := &types.PluginConfigListResponse{
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
		Configs:    c.convertProviderConfigs(result.Configs),
	}

	httpUtils.Response.Success(ctx, response, "获取供应商配置列表成功")
}

// createProviderConfig 创建供应商配置
// @Summary 创建供应商配置
// @Description 创建新的插件供应商配置
// @Tags Plugin Config
// @Accept json
// @Produce json
// @Param request body types.PluginConfigRequest true "供应商配置"
// @Success 201 {object} httptransport.APIResponse{data=types.PluginConfigResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Router /api/v1/plugin/providers [post]
func (c *PluginConfigController) createProviderConfig(ctx *gin.Context) {
	c.logger.InfoTag("API", "创建供应商配置", "request_id", getRequestID(ctx))

	var req types.PluginConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("请求参数绑定失败", "error", err)
		httpUtils.Response.ValidationError(ctx, err)
		return
	}

	// 转换为服务请求
	serviceReq := &config.CreateProviderConfigRequest{
		ProviderType: req.ProviderType,
		ProviderName: req.ProviderName,
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		Config:       req.Config,
		Enabled:      req.Enabled,
		Priority:     req.Priority,
		CreatedBy:    getUserInfo(ctx),
		UserAgent:    ctx.GetHeader("User-Agent"),
		IPAddress:    ctx.ClientIP(),
	}

	// 调用服务
	providerConfig, err := c.service.CreateProviderConfig(ctx, serviceReq)
	if err != nil {
		c.logger.Error("创建供应商配置失败", "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "创建供应商配置失败")
		return
	}

	// 转换为响应格式
	response := c.convertProviderConfig(providerConfig)
	httpUtils.Response.Created(ctx, response, "供应商配置创建成功")
}

// getProviderConfig 获取指定供应商配置
// @Summary 获取供应商配置详情
// @Description 根据ID获取供应商配置详细信息
// @Tags Plugin Config
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} httptransport.APIResponse{data=types.PluginConfigResponse}
// @Failure 404 {object} httptransport.APIResponse
// @Router /api/v1/plugin/providers/{id} [get]
func (c *PluginConfigController) getProviderConfig(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		httpUtils.Response.BadRequest(ctx, "无效的配置ID")
		return
	}

	c.logger.InfoTag("API", "获取供应商配置详情", "id", id, "request_id", getRequestID(ctx))

	// 调用服务
	providerConfig, err := c.service.GetProviderConfig(ctx, id)
	if err != nil {
		c.logger.Error("获取供应商配置失败", "id", id, "error", err)
		httpUtils.Response.NotFound(ctx, "供应商配置")
		return
	}

	// 转换为响应格式
	response := c.convertProviderConfig(providerConfig)
	httpUtils.Response.Success(ctx, response, "获取供应商配置成功")
}

// updateProviderConfig 更新供应商配置
// @Summary 更新供应商配置
// @Description 更新指定供应商的配置信息
// @Tags Plugin Config
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Param request body types.PluginConfigRequest true "供应商配置"
// @Success 200 {object} httptransport.APIResponse{data=types.PluginConfigResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Router /api/v1/plugin/providers/{id} [put]
func (c *PluginConfigController) updateProviderConfig(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		httpUtils.Response.BadRequest(ctx, "无效的配置ID")
		return
	}

	c.logger.InfoTag("API", "更新供应商配置", "id", id, "request_id", getRequestID(ctx))

	var req types.PluginConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("请求参数绑定失败", "error", err)
		httpUtils.Response.ValidationError(ctx, err)
		return
	}

	// 转换为服务请求
	serviceReq := &config.UpdateProviderConfigRequest{
		DisplayName: req.DisplayName,
		Description: req.Description,
		Config:      req.Config,
		Enabled:     &req.Enabled,
		Priority:    &req.Priority,
		UpdatedBy:   getUserInfo(ctx),
		UserAgent:   ctx.GetHeader("User-Agent"),
		IPAddress:   ctx.ClientIP(),
	}

	// 调用服务
	providerConfig, err := c.service.UpdateProviderConfig(ctx, id, serviceReq)
	if err != nil {
		c.logger.Error("更新供应商配置失败", "id", id, "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "更新供应商配置失败")
		return
	}

	// 转换为响应格式
	response := c.convertProviderConfig(providerConfig)
	httpUtils.Response.Success(ctx, response, "供应商配置更新成功")
}

// deleteProviderConfig 删除供应商配置
// @Summary 删除供应商配置
// @Description 删除指定的供应商配置
// @Tags Plugin Config
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Router /api/v1/plugin/providers/{id} [delete]
func (c *PluginConfigController) deleteProviderConfig(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		httpUtils.Response.BadRequest(ctx, "无效的配置ID")
		return
	}

	c.logger.InfoTag("API", "删除供应商配置", "id", id, "request_id", getRequestID(ctx))

	// 调用服务
	if err := c.service.DeleteProviderConfig(ctx, id); err != nil {
		c.logger.Error("删除供应商配置失败", "id", id, "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "删除供应商配置失败")
		return
	}

	httpUtils.Response.Success(ctx, gin.H{"id": id}, "供应商配置删除成功")
}

// getAvailableProviders 获取可用供应商列表
// @Summary 获取可用供应商列表
// @Description 获取系统中可用的供应商类型和配置模板
// @Tags Plugin Config
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=[]types.AvailableProviderResponse}
// @Router /api/v1/plugin/providers/available [get]
func (c *PluginConfigController) getAvailableProviders(ctx *gin.Context) {
	c.logger.InfoTag("API", "获取可用供应商列表", "request_id", getRequestID(ctx))

	// 调用服务
	providers, err := c.service.GetAvailableProviders(ctx)
	if err != nil {
		c.logger.Error("获取可用供应商列表失败", "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "获取可用供应商列表失败")
		return
	}

	// 转换为响应格式
	response := c.convertAvailableProviders(providers)
	httpUtils.Response.Success(ctx, response, "获取可用供应商列表成功")
}

// testProviderConfig 测试供应商配置
// @Summary 测试供应商配置
// @Description 测试供应商配置的连接性
// @Tags Plugin Config
// @Accept json
// @Produce json
// @Param request body types.HealthTestRequest true "测试配置"
// @Success 200 {object} httptransport.APIResponse{data=types.HealthTestResponse}
// @Router /api/v1/plugin/providers/test [post]
func (c *PluginConfigController) testProviderConfig(ctx *gin.Context) {
	c.logger.InfoTag("API", "测试供应商配置", "request_id", getRequestID(ctx))

	var req types.HealthTestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("请求参数绑定失败", "error", err)
		httpUtils.Response.ValidationError(ctx, err)
		return
	}

	// 从请求中提取供应商类型（这里简化处理，实际可能需要根据配置内容判断）
	providerType := config.ProviderTypeOpenAI // 默认值，实际应该从配置中解析

	// 转换为服务请求
	serviceReq := &config.TestProviderConfigRequest{
		ProviderType: providerType,
		Config:       req.Config,
	}

	// 调用服务
	result, err := c.service.TestProviderConfig(ctx, serviceReq)
	if err != nil {
		c.logger.Error("测试供应商配置失败", "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "测试供应商配置失败")
		return
	}

	// 转换为响应格式
	response := &types.HealthTestResponse{
		Success:   result.Success,
		Message:   result.Message,
		Latency:   result.Latency,
		Details:   result.Details,
		Timestamp: result.Timestamp,
	}

	httpUtils.Response.Success(ctx, response, "测试完成")
}

// getPluginStats 获取插件统计信息
// @Summary 获取插件统计信息
// @Description 获取插件配置的统计信息
// @Tags Plugin Config
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=types.PluginStatsResponse}
// @Router /api/v1/plugin/providers/stats [get]
func (c *PluginConfigController) getPluginStats(ctx *gin.Context) {
	c.logger.InfoTag("API", "获取插件统计信息", "request_id", getRequestID(ctx))

	// 调用服务
	stats, err := c.service.GetPluginStats(ctx)
	if err != nil {
		c.logger.Error("获取插件统计信息失败", "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "获取插件统计信息失败")
		return
	}

	// 转换为响应格式
	response := c.convertPluginStats(stats)
	httpUtils.Response.Success(ctx, response, "获取插件统计信息成功")
}

// getConfigSnapshots 获取配置快照列表
// @Summary 获取配置快照列表
// @Description 获取指定供应商配置的快照列表
// @Tags Plugin Config
// @Produce json
// @Param id path int true "配置ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} httptransport.APIResponse{data=types.ConfigSnapshotListResponse}
// @Router /api/v1/plugin/providers/{id}/snapshots [get]
func (c *PluginConfigController) getConfigSnapshots(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		httpUtils.Response.BadRequest(ctx, "无效的配置ID")
		return
	}

	c.logger.InfoTag("API", "获取配置快照列表", "id", id, "request_id", getRequestID(ctx))

	filter := &config.SnapshotFilter{
		Page:     parseIntQuery(ctx, "page", 1),
		PageSize: parseIntQuery(ctx, "page_size", 20),
	}

	// 调用服务
	result, err := c.service.GetConfigSnapshots(ctx, id, filter)
	if err != nil {
		c.logger.Error("获取配置快照列表失败", "id", id, "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "获取配置快照列表失败")
		return
	}

	// 转换为响应格式
	response := &types.ConfigSnapshotListResponse{
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
		Snapshots:  c.convertSnapshots(result.Snapshots),
	}

	httpUtils.Response.Success(ctx, response, "获取配置快照列表成功")
}

// createConfigSnapshot 创建配置快照
// @Summary 创建配置快照
// @Description 为指定供应商配置创建快照
// @Tags Plugin Config
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Param request body types.ConfigSnapshotRequest true "快照信息"
// @Success 201 {object} httptransport.APIResponse{data=types.ConfigSnapshotResponse}
// @Router /api/v1/plugin/providers/{id}/snapshots [post]
func (c *PluginConfigController) createConfigSnapshot(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		httpUtils.Response.BadRequest(ctx, "无效的配置ID")
		return
	}

	c.logger.InfoTag("API", "创建配置快照", "id", id, "request_id", getRequestID(ctx))

	var req types.ConfigSnapshotRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("请求参数绑定失败", "error", err)
		httpUtils.Response.ValidationError(ctx, err)
		return
	}

	// 转换为服务请求
	serviceReq := &config.CreateSnapshotRequest{
		Version:     req.Version,
		SnapshotName: req.SnapshotName,
		Description: req.Description,
		CreatedBy:   getUserInfo(ctx),
	}

	// 调用服务
	snapshot, err := c.service.CreateConfigSnapshot(ctx, id, serviceReq)
	if err != nil {
		c.logger.Error("创建配置快照失败", "id", id, "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "创建配置快照失败")
		return
	}

	// 转换为响应格式
	response := c.convertSnapshot(snapshot)
	httpUtils.Response.Created(ctx, response, "配置快照创建成功")
}

// restoreConfigSnapshot 恢复配置快照
// @Summary 恢复配置快照
// @Description 将供应商配置恢复到指定快照状态
// @Tags Plugin Config
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Param request body types.ConfigRestoreRequest true "恢复请求"
// @Success 200 {object} httptransport.APIResponse
// @Router /api/v1/plugin/providers/{id}/snapshots/restore [post]
func (c *PluginConfigController) restoreConfigSnapshot(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		httpUtils.Response.BadRequest(ctx, "无效的配置ID")
		return
	}

	c.logger.InfoTag("API", "恢复配置快照", "id", id, "request_id", getRequestID(ctx))

	var req types.ConfigRestoreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("请求参数绑定失败", "error", err)
		httpUtils.Response.ValidationError(ctx, err)
		return
	}

	// 调用服务
	if err := c.service.RestoreConfigSnapshot(ctx, id, req.SnapshotID); err != nil {
		c.logger.Error("恢复配置快照失败", "id", id, "snapshot_id", req.SnapshotID, "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "恢复配置快照失败")
		return
	}

	httpUtils.Response.Success(ctx, gin.H{"id": id, "snapshot_id": req.SnapshotID}, "配置快照恢复成功")
}

// getConfigHistory 获取配置变更历史
// @Summary 获取配置变更历史
// @Description 获取指定供应商配置的变更历史记录
// @Tags Plugin Config
// @Produce json
// @Param id path int true "配置ID"
// @Param operation query string false "操作类型过滤"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} httptransport.APIResponse{data=types.ConfigHistoryListResponse}
// @Router /api/v1/plugin/providers/{id}/history [get]
func (c *PluginConfigController) getConfigHistory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		httpUtils.Response.BadRequest(ctx, "无效的配置ID")
		return
	}

	c.logger.InfoTag("API", "获取配置变更历史", "id", id, "request_id", getRequestID(ctx))

	filter := &config.HistoryFilter{
		Operation: config.HistoryOperation(ctx.Query("operation")),
		Page:      parseIntQuery(ctx, "page", 1),
		PageSize:  parseIntQuery(ctx, "page_size", 20),
	}

	// 调用服务
	result, err := c.service.GetConfigHistory(ctx, id, filter)
	if err != nil {
		c.logger.Error("获取配置变更历史失败", "id", id, "error", err)
		httpUtils.Response.Error(ctx, httpUtils.ErrorCodeInternalServer, "获取配置变更历史失败")
		return
	}

	// 转换为响应格式
	response := &types.ConfigHistoryListResponse{
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
		History:    c.convertHistory(result.History),
	}

	httpUtils.Response.Success(ctx, response, "获取配置变更历史成功")
}

// 辅助方法：转换数据结构
func (c *PluginConfigController) convertProviderConfig(pc *config.ProviderConfig) *types.PluginConfigResponse {
	// 解密配置数据
	var configData map[string]interface{}
	var configSchema map[string]interface{}

	if pc.ConfigData != "" {
		// 这里应该使用加密器解密，暂时简化处理
		json.Unmarshal([]byte(pc.ConfigData), &configData)
	}

	if pc.ConfigSchema != "" {
		json.Unmarshal([]byte(pc.ConfigSchema), &configSchema)
	}

	return &types.PluginConfigResponse{
		ID:              pc.ID,
		ProviderType:    pc.ProviderType,
		ProviderName:    pc.ProviderName,
		DisplayName:     pc.DisplayName,
		Description:     pc.Description,
		Config:          configData,
		ConfigSchema:    configSchema,
		Enabled:         pc.Enabled,
		Priority:        pc.Priority,
		HealthStatus:    pc.HealthStatus,
		LastHealthCheck: pc.LastHealthCheck,
		CreatedAt:       pc.CreatedAt,
		UpdatedAt:       pc.UpdatedAt,
		Capabilities:    c.convertCapabilities(pc.Capabilities),
	}
}

func (c *PluginConfigController) convertProviderConfigs(configs []config.ProviderConfig) []types.PluginConfigResponse {
	result := make([]types.PluginConfigResponse, len(configs))
	for i, pc := range configs {
		result[i] = *c.convertProviderConfig(&pc)
	}
	return result
}

func (c *PluginConfigController) convertCapability(cap config.Capability) types.CapabilityResponse {
	var inputSchema, outputSchema map[string]interface{}

	if cap.InputSchema != "" {
		json.Unmarshal([]byte(cap.InputSchema), &inputSchema)
	}
	if cap.OutputSchema != "" {
		json.Unmarshal([]byte(cap.OutputSchema), &outputSchema)
	}

	return types.CapabilityResponse{
		ID:                    cap.ID,
		ProviderConfigID:      cap.ProviderConfigID,
		CapabilityID:          cap.CapabilityID,
		CapabilityType:        cap.CapabilityType,
		CapabilityName:        cap.CapabilityName,
		CapabilityDescription: cap.CapabilityDescription,
		InputSchema:           inputSchema,
		OutputSchema:          outputSchema,
		Enabled:               cap.Enabled,
		CreatedAt:             cap.CreatedAt,
		UpdatedAt:             cap.UpdatedAt,
	}
}

func (c *PluginConfigController) convertCapabilities(caps []config.Capability) []types.CapabilityResponse {
	result := make([]types.CapabilityResponse, len(caps))
	for i, cap := range caps {
		result[i] = c.convertCapability(cap)
	}
	return result
}

func (c *PluginConfigController) convertAvailableProviders(providers []config.AvailableProvider) []types.AvailableProviderResponse {
	result := make([]types.AvailableProviderResponse, len(providers))
	for i, p := range providers {
		capabilities := make([]types.CapabilityTemplate, len(p.Capabilities))
		for j, cap := range p.Capabilities {
			capabilities[j] = types.CapabilityTemplate{
				CapabilityID:          cap.CapabilityID,
				CapabilityType:        cap.CapabilityType,
				CapabilityName:        cap.CapabilityName,
				CapabilityDescription: cap.CapabilityDescription,
				InputSchema:           cap.InputSchema,
				OutputSchema:          cap.OutputSchema,
			}
		}

		result[i] = types.AvailableProviderResponse{
			ProviderType:   p.ProviderType,
			ProviderName:   p.ProviderName,
			DisplayName:    p.DisplayName,
			Description:    p.Description,
			ConfigTemplate: p.ConfigTemplate,
			ConfigSchema:   p.ConfigSchema,
			Capabilities:   capabilities,
		}
	}
	return result
}

func (c *PluginConfigController) convertSnapshot(snapshot *config.ConfigSnapshot) *types.ConfigSnapshotResponse {
	var snapshotData map[string]interface{}
	if snapshot.SnapshotData != "" {
		json.Unmarshal([]byte(snapshot.SnapshotData), &snapshotData)
	}

	return &types.ConfigSnapshotResponse{
		ID:            snapshot.ID,
		ProviderConfigID: snapshot.ProviderConfigID,
		Version:       snapshot.Version,
		SnapshotName:  snapshot.SnapshotName,
		Description:   snapshot.Description,
		SnapshotData:  snapshotData,
		IsActive:      snapshot.IsActive,
		CreatedBy:     snapshot.CreatedBy,
		CreatedAt:     snapshot.CreatedAt,
	}
}

func (c *PluginConfigController) convertSnapshots(snapshots []config.ConfigSnapshot) []types.ConfigSnapshotResponse {
	result := make([]types.ConfigSnapshotResponse, len(snapshots))
	for i, snapshot := range snapshots {
		result[i] = *c.convertSnapshot(&snapshot)
	}
	return result
}

func (c *PluginConfigController) convertHistory(history []config.ConfigHistory) []types.ConfigHistoryResponse {
	result := make([]types.ConfigHistoryResponse, len(history))
	for i, h := range history {
		var oldData, newData map[string]interface{}
		var changedFields []string

		if h.OldData != "" {
			json.Unmarshal([]byte(h.OldData), &oldData)
		}
		if h.NewData != "" {
			json.Unmarshal([]byte(h.NewData), &newData)
		}
		if h.ChangedFields != "" {
			json.Unmarshal([]byte(h.ChangedFields), &changedFields)
		}

		result[i] = types.ConfigHistoryResponse{
			ID:              h.ID,
			ProviderConfigID: h.ProviderConfigID,
			Operation:       h.Operation,
			OldData:         oldData,
			NewData:         newData,
			ChangeSummary:   h.ChangeSummary,
			ChangedFields:   changedFields,
			CreatedBy:       h.CreatedBy,
			UserAgent:       h.UserAgent,
			IPAddress:       h.IPAddress,
			CreatedAt:       h.CreatedAt,
		}
	}
	return result
}

func (c *PluginConfigController) convertPluginStats(stats *config.PluginStats) *types.PluginStatsResponse {
	providerStats := make(map[string]types.ProviderStats)
	for k, v := range stats.ProviderStats {
		providerStats[k] = types.ProviderStats{
			Type:         v.Type,
			Count:        v.Count,
			EnabledCount: v.EnabledCount,
			HealthyCount: v.HealthyCount,
		}
	}

	capabilityStats := make(map[string]types.CapabilityStats)
	for k, v := range stats.CapabilityStats {
		capabilityStats[k] = types.CapabilityStats{
			Type:         v.Type,
			Count:        v.Count,
			EnabledCount: v.EnabledCount,
		}
	}

	return &types.PluginStatsResponse{
		TotalProviders:      stats.TotalProviders,
		EnabledProviders:    stats.EnabledProviders,
		HealthyProviders:    stats.HealthyProviders,
		TotalCapabilities:   stats.TotalCapabilities,
		EnabledCapabilities: stats.EnabledCapabilities,
		ProviderStats:       providerStats,
		CapabilityStats:     capabilityStats,
	}
}

// 辅助函数
func getUserInfo(c *gin.Context) string {
	// 从JWT token或session中获取用户信息
	// 暂时返回简单标识
	return fmt.Sprintf("user_%s", c.ClientIP())
}

func parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultValue
}