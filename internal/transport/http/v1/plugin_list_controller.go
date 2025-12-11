package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/plugin/status"
)

// API 错误代码常量
const (
	ValidationFailed     = "VALIDATION_FAILED"
	InternalServerError  = "INTERNAL_SERVER_ERROR"
	ResourceNotFound     = "RESOURCE_NOT_FOUND"
)

// APIResponse 标准API响应结构
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Version   string      `json:"version"`
	RequestID string      `json:"request_id,omitempty"`
}

// APIError API错误结构
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GetRequestID 获取请求ID
func GetRequestID(ctx *gin.Context) string {
	if requestID := ctx.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	// 如果没有请求ID，返回空字符串，避免为每个请求生成唯一ID的开销
	return ""
}

// PluginListController 插件列表API控制器
type PluginListController struct {
	logger         *logging.Logger
	statusManager  *status.PluginStatusManager
}

// NewPluginListController 创建插件列表控制器
func NewPluginListController(
	statusManager *status.PluginStatusManager,
	logger *logging.Logger,
) *PluginListController {
	if logger == nil {
		logger = logging.DefaultLogger
	}

	return &PluginListController{
		logger:        logger,
		statusManager: statusManager,
	}
}

// Register 注册路由
func (c *PluginListController) Register(router *gin.RouterGroup) {
	plugins := router.Group("/plugins")
	{
		plugins.GET("/", c.ListPlugins)
		plugins.GET("/stats", c.GetPluginStats)
		plugins.GET("/ports", c.GetPortStats)
		plugins.GET("/:id", c.GetPlugin)
	plugins.POST("/:id/control", c.ControlPlugin)
		plugins.POST("/:id/health", c.CheckPluginHealth)
		plugins.POST("/:id/reallocate-port", c.ReallocatePort)
		plugins.GET("/capabilities", c.GetCapabilities)
		plugins.GET("/capabilities/:type", c.GetCapabilitiesByType)
	}
}

// ListPlugins 获取插件列表
// @Summary 获取插件列表
// @Description 获取所有插件的信息，支持分页、筛选和排序
// @Tags plugins
// @Param type query string false "插件类型" Enums(LLM,TTS,ASR,Tool)
// @Param status query string false "插件状态" Enums(installed,enabled,disabled,running,stopped,error)
// @Param health_status query string false "健康状态" Enums(healthy,unhealthy,unknown)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param sort_by query string false "排序字段" default(updated_at)
// @Param sort_order query string false "排序方向" Enums(asc,desc) default(desc)
// @Param search query string false "搜索关键词"
// @Produce json
// @Success 200 {object} APIResponse{data=status.PluginListResponse}
// @Router /api/v1/plugins [get]
func (c *PluginListController) ListPlugins(ctx *gin.Context) {
	// 解析查询参数
	filter := status.DefaultPluginFilter()
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    ValidationFailed,
				Message: "请求参数验证失败: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	// 验证筛选条件
	if err := filter.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    ValidationFailed,
				Message: "筛选条件验证失败: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	// 获取插件列表
	response, err := c.statusManager.ListPlugins(filter)
	if err != nil {
		c.logger.ErrorTag("plugin_list", "获取插件列表失败",
			"error", err.Error(),
			"request_id", GetRequestID(ctx))

		ctx.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    InternalServerError,
				Message: "获取插件列表失败: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	// 添加时间戳到每个插件
	for i := range response.Plugins {
		response.Plugins[i].LastHealthCheck = time.Now()
	}

	if c.logger != nil {
		c.logger.InfoTag("plugin_list", "获取插件列表成功",
			"total", response.Total,
			"page", response.Page,
			"plugins_count", len(response.Plugins),
			"request_id", GetRequestID(ctx))
	}

	ctx.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      response,
		Message:   "获取插件列表成功",
		Timestamp: time.Now().Unix(),
		Version:   "v1",
		RequestID: GetRequestID(ctx),
	})
}

// GetPluginStats 获取插件统计信息
// @Summary 获取插件统计信息
// @Description 获取插件的数量、状态分布、健康状态等统计信息
// @Tags plugins
// @Produce json
// @Success 200 {object} APIResponse{data=status.PluginStats}
// @Router /api/v1/plugins/stats [get]
func (c *PluginListController) GetPluginStats(ctx *gin.Context) {
	stats := c.statusManager.GetStats()

	if c.logger != nil {
		c.logger.InfoTag("plugin_stats", "获取插件统计信息",
			"total_plugins", stats.TotalPlugins,
			"running_plugins", stats.RunningPlugins,
			"request_id", GetRequestID(ctx))
	}

	ctx.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      stats,
		Message:   "获取插件统计信息成功",
		Timestamp: time.Now().Unix(),
		Version:   "v1",
		RequestID: GetRequestID(ctx),
	})
}

// GetPortStats 获取端口统计信息
// @Summary 获取端口统计信息
// @Description 获取端口使用情况统计
// @Tags plugins
// @Produce json
// @Success 200 {object} APIResponse{data=ports.PortStats}
// @Router /api/v1/plugins/ports [get]
func (c *PluginListController) GetPortStats(ctx *gin.Context) {
	// 这里需要访问PortManager，需要扩展StatusManager
	// 暂时返回模拟数据
	stats := map[string]interface{}{
		"total_ports":     10000,
		"allocated_ports": 5,
		"available_ports": 9995,
		"reserved_ports":  0,
		"usage_percent":   0.05,
	}

	if c.logger != nil {
		c.logger.InfoTag("port_stats", "获取端口统计信息",
			"request_id", GetRequestID(ctx))
	}

	ctx.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      stats,
		Message:   "获取端口统计信息成功",
		Timestamp: time.Now().Unix(),
		Version:   "v1",
		RequestID: GetRequestID(ctx),
	})
}

// GetPlugin 获取单个插件详情
// @Summary 获取插件详情
// @Description 根据插件ID获取详细信息
// @Tags plugins
// @Param id path string true "插件ID"
// @Produce json
// @Success 200 {object} APIResponse{data=status.PluginStatus}
// @Failure 404 {object} APIResponse
// @Router /api/v1/plugins/{id} [get]
func (c *PluginListController) GetPlugin(ctx *gin.Context) {
	pluginID := ctx.Param("id")
	if pluginID == "" {
		ctx.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    ValidationFailed,
				Message: "插件ID不能为空",
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	plugin, err := c.statusManager.GetPluginStatus(pluginID)
	if err != nil {
		if c.logger != nil {
			c.logger.ErrorTag("plugin_get", "获取插件详情失败",
				"plugin_id", pluginID,
				"error", err.Error(),
				"request_id", GetRequestID(ctx))
		}

		ctx.JSON(http.StatusNotFound, APIResponse{
			Success:   false,
			Error: &APIError{
				Code:    ResourceNotFound,
				Message: "插件不存在: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	if c.logger != nil {
		c.logger.InfoTag("plugin_get", "获取插件详情成功",
			"plugin_id", pluginID,
			"plugin_name", plugin.Name,
			"status", plugin.Status,
			"request_id", GetRequestID(ctx))
	}

	ctx.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      plugin,
		Message:   "获取插件详情成功",
		Timestamp: time.Now().Unix(),
		Version:   "v1",
		RequestID: GetRequestID(ctx),
	})
}

// ControlPlugin 控制插件
// @Summary 控制插件
// @Description 对插件进行启动、停止、重启、重新分配端口等操作
// @Tags plugins
// @Param id path string true "插件ID"
// @Param body body status.PluginControlRequest true "控制请求"
// @Produce json
// @Success 200 {object} APIResponse{data=status.PluginControlResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /api/v1/plugins/{id}/control [post]
func (c *PluginListController) ControlPlugin(ctx *gin.Context) {
	pluginID := ctx.Param("id")
	if pluginID == "" {
		ctx.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    ValidationFailed,
				Message: "插件ID不能为空",
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	var req status.PluginControlRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    ValidationFailed,
				Message: "请求体格式错误: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	// 验证操作类型
	validActions := map[string]bool{
		"start":           true,
		"stop":            true,
		"restart":         true,
		"reallocate_port":  true,
	}

	if !validActions[req.Action] {
		ctx.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    ValidationFailed,
				Message: "不支持的操作类型: " + req.Action,
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	// 获取当前插件状态
	currentPlugin, err := c.statusManager.GetPluginStatus(pluginID)
	if err != nil {
		if c.logger != nil {
			c.logger.ErrorTag("plugin_control", "获取插件状态失败",
				"plugin_id", pluginID,
				"action", req.Action,
				"error", err.Error(),
				"request_id", GetRequestID(ctx))
		}

		ctx.JSON(http.StatusNotFound, APIResponse{
			Success:   false,
			Error: &APIError{
				Code:    ResourceNotFound,
				Message: "插件不存在: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	startTime := time.Now()
	oldStatus := string(currentPlugin.Status)
	oldPort := currentPlugin.Port

	var response status.PluginControlResponse
	var controlErr error

	switch req.Action {
	case "start":
		if req.Config != nil {
			controlErr = c.statusManager.StartPluginWithConfig(pluginID, req.Config)
		} else {
			controlErr = c.statusManager.StartPlugin(pluginID)
		}
	case "stop":
		controlErr = c.statusManager.StopPlugin(pluginID)
	case "restart":
		controlErr = c.statusManager.RestartPlugin(pluginID)
	case "reallocate_port":
		controlErr = c.statusManager.ReallocatePort(pluginID)
	}

	processTime := time.Since(startTime).String()

	if controlErr != nil {
		response = status.PluginControlResponse{
			Success:     false,
			Message:     "操作失败: " + controlErr.Error(),
			OldStatus:   oldStatus,
			ProcessTime: processTime,
		}

		if c.logger != nil {
			c.logger.ErrorTag("plugin_control", "插件控制操作失败",
				"plugin_id", pluginID,
				"action", req.Action,
				"error", controlErr.Error(),
				"request_id", GetRequestID(ctx))
		}

		ctx.JSON(http.StatusInternalServerError, APIResponse{
			Success:   false,
			Error: &APIError{
				Code:    InternalServerError,
				Message: "插件控制失败: " + controlErr.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	// 获取更新后的插件状态
	updatedPlugin, err := c.statusManager.GetPluginStatus(pluginID)
	if err == nil {
		response = status.PluginControlResponse{
			Success:     true,
			Message:     "操作成功",
			OldStatus:   oldStatus,
			NewStatus:   string(updatedPlugin.Status),
			OldPort:     oldPort,
			NewPort:     updatedPlugin.Port,
			ProcessTime:  processTime,
		}
	} else {
		response = status.PluginControlResponse{
			Success:     true,
			Message:     "操作成功",
			OldStatus:   oldStatus,
			ProcessTime:  processTime,
		}
	}

	if c.logger != nil {
		c.logger.InfoTag("plugin_control", "插件控制操作成功",
			"plugin_id", pluginID,
			"action", req.Action,
			"old_status", response.OldStatus,
			"new_status", response.NewStatus,
			"old_port", response.OldPort,
			"new_port", response.NewPort,
			"process_time", response.ProcessTime,
			"request_id", GetRequestID(ctx))
	}

	ctx.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      response,
		Message:   "插件控制操作完成",
		Timestamp: time.Now().Unix(),
		Version:   "v1",
		RequestID: GetRequestID(ctx),
	})
}

// CheckPluginHealth 检查插件健康状态
// @Summary 检查插件健康状态
// @Description 手动触发插件健康检查
// @Tags plugins
// @Param id path string true "插件ID"
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /api/v1/plugins/{id}/health [post]
func (c *PluginListController) CheckPluginHealth(ctx *gin.Context) {
	pluginID := ctx.Param("id")
	if pluginID == "" {
		ctx.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    ValidationFailed,
				Message: "插件ID不能为空",
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	plugin, err := c.statusManager.GetPluginStatus(pluginID)
	if err != nil {
		if c.logger != nil {
			c.logger.ErrorTag("plugin_health", "获取插件状态失败",
				"plugin_id", pluginID,
				"error", err.Error(),
				"request_id", GetRequestID(ctx))
		}

		ctx.JSON(http.StatusNotFound, APIResponse{
			Success:   false,
			Error: &APIError{
				Code:    ResourceNotFound,
				Message: "插件不存在: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	// 手动触发健康检查
	c.statusManager.UpdatePluginHealth(pluginID, status.HealthStatusHealthy, "手动健康检查")

	// 获取更新后的状态
	updatedPlugin, err := c.statusManager.GetPluginStatus(pluginID)
	if err == nil {
		plugin = updatedPlugin
	}

	if c.logger != nil {
		c.logger.InfoTag("plugin_health", "插件健康检查完成",
			"plugin_id", pluginID,
			"health_status", plugin.HealthStatus,
			"request_id", GetRequestID(ctx))
	}

	ctx.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data: map[string]interface{}{
			"plugin_id":      pluginID,
			"health_status": plugin.HealthStatus,
			"last_check":    plugin.LastHealthCheck.Format(time.RFC3339),
		},
		Message:   "健康检查完成",
		Timestamp: time.Now().Unix(),
		Version:   "v1",
		RequestID: GetRequestID(ctx),
	})
}

// ReallocatePort 重新分配插件端口
// @Summary 重新分配插件端口
// @Description 为插件分配新的端口
// @Tags plugins
// @Param id path string true "插件ID"
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /api/v1/plugins/{id}/reallocate-port [post]
func (c *PluginListController) ReallocatePort(ctx *gin.Context) {
	pluginID := ctx.Param("id")
	if pluginID == "" {
		ctx.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    ValidationFailed,
				Message: "插件ID不能为空",
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	err := c.statusManager.ReallocatePort(pluginID)
	if err != nil {
		if c.logger != nil {
			c.logger.ErrorTag("plugin_reallocate_port", "重新分配端口失败",
				"plugin_id", pluginID,
				"error", err.Error(),
				"request_id", GetRequestID(ctx))
		}

		ctx.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    InternalServerError,
				Message: "重新分配端口失败: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	// 获取更新后的插件状态
	plugin, err := c.statusManager.GetPluginStatus(pluginID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    InternalServerError,
				Message: "获取更新后插件状态失败: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	if c.logger != nil {
		c.logger.InfoTag("plugin_reallocate_port", "插件端口重新分配成功",
			"plugin_id", pluginID,
			"new_port", plugin.Port,
			"new_address", plugin.Address,
			"request_id", GetRequestID(ctx))
	}

	ctx.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"plugin_id": pluginID,
			"port":      plugin.Port,
			"address":   plugin.Address,
		},
		Message: "端口重新分配成功",
		Timestamp: time.Now().Unix(),
		Version:   "v1",
		RequestID: GetRequestID(ctx),
	})
}

// GetCapabilities 获取所有插件能力
// @Summary 获取所有插件能力
// @Description 获取所有插件的能力定义
// @Tags plugins
// @Produce json
// @Success 200 {object} APIResponse
// @Router /api/v1/plugins/capabilities [get]
func (c *PluginListController) GetCapabilities(ctx *gin.Context) {
	// 获取所有插件
	response, err := c.statusManager.ListPlugins(status.DefaultPluginFilter())
	if err != nil {
		if c.logger != nil {
			c.logger.ErrorTag("plugin_capabilities", "获取插件列表失败",
				"error", err.Error(),
				"request_id", GetRequestID(ctx))
		}

		ctx.JSON(http.StatusInternalServerError, APIResponse{
			Success:   false,
			Error: &APIError{
				Code:    InternalServerError,
				Message: "获取插件列表失败: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	// 收集所有能力
	capabilities := make([]map[string]interface{}, 0)
	for _, plugin := range response.Plugins {
		for _, cap := range plugin.Capabilities {
			capabilities = append(capabilities, map[string]interface{}{
				"plugin_id":   plugin.ID,
				"plugin_name": plugin.Name,
				"capability": cap,
			})
		}
	}

	if c.logger != nil {
		c.logger.InfoTag("plugin_capabilities", "获取插件能力列表成功",
			"total_capabilities", len(capabilities),
			"request_id", GetRequestID(ctx))
	}

	ctx.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      capabilities,
		Message:   "获取插件能力列表成功",
		Timestamp: time.Now().Unix(),
		Version:   "v1",
		RequestID: GetRequestID(ctx),
	})
}

// GetCapabilitiesByType 按类型获取插件能力
// @Summary 按类型获取插件能力
// @Description 根据类型筛选插件能力
// @Tags plugins
// @Param type path string true "能力类型"
// @Produce json
// @Success 200 {object} APIResponse
// @Router /api/v1/plugins/capabilities/{type} [get]
func (c *PluginListController) GetCapabilitiesByType(ctx *gin.Context) {
	capabilityType := ctx.Param("type")
	if capabilityType == "" {
		ctx.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    ValidationFailed,
				Message: "能力类型不能为空",
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
		RequestID: GetRequestID(ctx),
		})
		return
	}

	// 获取所有插件
	response, err := c.statusManager.ListPlugins(status.DefaultPluginFilter())
	if err != nil {
		if c.logger != nil {
			c.logger.ErrorTag("plugin_capabilities_type", "获取插件列表失败",
				"type", capabilityType,
				"error", err.Error(),
				"request_id", GetRequestID(ctx))
		}

		ctx.JSON(http.StatusInternalServerError, APIResponse{
			Success:   false,
			Error: &APIError{
				Code:    InternalServerError,
				Message: "获取插件列表失败: " + err.Error(),
			},
			Timestamp: time.Now().Unix(),
			Version:   "v1",
			RequestID: GetRequestID(ctx),
		})
		return
	}

	// 筛选指定类型的能力
	capabilities := make([]map[string]interface{}, 0)
	for _, plugin := range response.Plugins {
		for _, cap := range plugin.Capabilities {
			if cap.Type == capabilityType {
				capabilities = append(capabilities, map[string]interface{}{
					"plugin_id":   plugin.ID,
					"plugin_name": plugin.Name,
					"capability": cap,
				})
			}
		}
	}

	if c.logger != nil {
		c.logger.InfoTag("plugin_capabilities_type", "按类型获取插件能力成功",
			"type", capabilityType,
			"total_capabilities", len(capabilities),
			"request_id", GetRequestID(ctx))
	}

	ctx.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      capabilities,
		Message:   "获取插件能力列表成功",
		Timestamp: time.Now().Unix(),
		Version:   "v1",
		RequestID: GetRequestID(ctx),
	})
}