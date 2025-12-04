package v1

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/transport/http/types/v1"
	httpUtils "xiaozhi-server-go/internal/transport/http/utils"
	"xiaozhi-server-go/internal/utils"
)

// DeviceServiceV1 V1版本设备服务
type DeviceServiceV1 struct {
	logger *utils.Logger
	config *config.Config
	// TODO: 添加实际的业务逻辑依赖
}

// NewDeviceServiceV1 创建设备服务V1实例
func NewDeviceServiceV1(config *config.Config, logger *utils.Logger) (*DeviceServiceV1, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &DeviceServiceV1{
		logger: logger,
		config: config,
	}, nil
}

// Register 注册设备API路由
func (s *DeviceServiceV1) Register(router *gin.RouterGroup) {
	// 设备管理
	devices := router.Group("/devices")
	{
		devices.POST("", s.registerDevice)           // 设备注册
		devices.GET("", s.listDevices)               // 获取设备列表
		devices.GET("/:id", s.getDevice)             // 获取设备详情
		devices.PUT("/:id", s.updateDevice)          // 更新设备信息
		devices.DELETE("/:id", s.deleteDevice)       // 删除设备
		devices.POST("/:id/activate", s.activateDevice) // 激活设备
	}

	// OTA固件更新
	ota := router.Group("/ota")
	{
		ota.GET("/firmware", s.listFirmware)         // 获取固件列表
		ota.POST("/updates", s.createOTAUpdate)      // 创建OTA更新
		ota.GET("/updates/:id", s.getOTAStatus)      // 获取OTA更新状态
		ota.POST("/updates/:id/cancel", s.cancelOTAUpdate) // 取消OTA更新
		ota.GET("/status", s.queryOTAStatus)         // 查询OTA状态
	}

	// 设备批量操作
	batch := router.Group("/devices/batch")
	{
		batch.POST("/ota", s.batchOTAUpdate)         // 批量OTA更新
		batch.POST("/command", s.batchCommand)       // 批量命令执行
		batch.POST("/config", s.batchConfig)         // 批量配置更新
	}

	// 连接信息
	connections := router.Group("/connections")
	{
		connections.GET("/websocket", s.getWebSocketInfo) // 获取WebSocket连接信息
		connections.GET("/mqtt", s.getMQTTInfo)           // 获取MQTT连接信息
	}
}

// registerDevice 设备注册
// @Summary 设备注册
// @Description 新设备注册到系统
// @Tags Devices
// @Accept json
// @Produce json
// @Param request body v1.DeviceRegistrationRequest true "设备注册信息"
// @Success 201 {object} httptransport.APIResponse{data=v1.DeviceInfo}
// @Failure 400 {object} httptransport.APIResponse
// @Router /devices [post]
func (s *DeviceServiceV1) registerDevice(c *gin.Context) {
	var request v1.DeviceRegistrationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "设备注册",
		"device_id", request.DeviceID,
		"device_name", request.DeviceName,
		"device_type", request.DeviceType,
		"request_id", getRequestID(c),
	)

	// 检查设备是否已存在
	if s.deviceExists(request.DeviceID) {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeDeviceExists, "设备已存在")
		return
	}

	// 创建设备记录
	device := v1.DeviceInfo{
		DeviceID:   request.DeviceID,
		DeviceName: request.DeviceName,
		DeviceType: request.DeviceType,
		Model:      request.Model,
		Version:    request.Version,
		Status:     "offline",
		Location:   request.Location,
		Metadata:   request.Metadata,
		Configuration: make(map[string]interface{}),
		IsActive:   false,
		IsActivated: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 分配数据库ID
	device.ID = time.Now().UnixNano()

	httpUtils.Response.Created(c, device, "设备注册成功")
}

// listDevices 获取设备列表
// @Summary 获取设备列表
// @Description 获取设备列表，支持分页和过滤
// @Tags Devices
// @Produce json
// @Param status query string false "按状态过滤"
// @Param device_type query string false "按设备类型过滤"
// @Param search query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段" default(created_at)
// @Param sort_order query string false "排序方向" default(desc)
// @Param location query bool false "是否返回位置信息"
// @Success 200 {object} httptransport.APIResponse{data=v1.DeviceListResponse}
// @Router /devices [get]
func (s *DeviceServiceV1) listDevices(c *gin.Context) {
	var query v1.DeviceQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "获取设备列表",
		"status", query.Status,
		"device_type", query.DeviceType,
		"search", query.Search,
		"page", query.Page,
		"limit", query.Limit,
		"request_id", getRequestID(c),
	)

	// 模拟获取设备列表
	devices, pagination := s.getMockDeviceList(query)

	response := v1.DeviceListResponse{
		Devices:    devices,
		Pagination: pagination,
	}

	httpUtils.Response.Success(c, response, "获取设备列表成功")
}

// getDevice 获取设备详情
// @Summary 获取设备详情
// @Description 根据ID获取设备的详细信息
// @Tags Devices
// @Produce json
// @Param id path string true "设备ID"
// @Success 200 {object} httptransport.APIResponse{data=v1.DeviceInfo}
// @Failure 404 {object} httptransport.APIResponse
// @Router /devices/{id} [get]
func (s *DeviceServiceV1) getDevice(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		httpUtils.Response.BadRequest(c, "设备ID不能为空")
		return
	}

	s.logger.InfoTag("API", "获取设备详情",
		"device_id", deviceID,
		"request_id", getRequestID(c),
	)

	// 模拟获取设备详情
	device := s.getMockDevice(deviceID)
	if device == nil {
		httpUtils.Response.NotFound(c, "设备")
		return
	}

	httpUtils.Response.Success(c, device, "获取设备详情成功")
}

// updateDevice 更新设备信息
// @Summary 更新设备信息
// @Description 更新指定设备的信息
// @Tags Devices
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Param request body v1.DeviceUpdateRequest true "设备更新信息"
// @Success 200 {object} httptransport.APIResponse{data=v1.DeviceInfo}
// @Failure 400 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Router /devices/{id} [put]
func (s *DeviceServiceV1) updateDevice(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		httpUtils.Response.BadRequest(c, "设备ID不能为空")
		return
	}

	var request v1.DeviceUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "更新设备信息",
		"device_id", deviceID,
		"device_name", request.DeviceName,
		"request_id", getRequestID(c),
	)

	// 获取设备并更新
	device := s.getMockDevice(deviceID)
	if device == nil {
		httpUtils.Response.NotFound(c, "设备")
		return
	}

	// 更新字段
	if request.DeviceName != "" {
		device.DeviceName = request.DeviceName
	}
	if request.Location != nil {
		device.Location = request.Location
	}
	if request.Configuration != nil {
		device.Configuration = request.Configuration
	}
	if request.Metadata != nil {
		device.Metadata = request.Metadata
	}
	if request.IsActive != nil {
		device.IsActive = *request.IsActive
	}

	device.UpdatedAt = time.Now()

	httpUtils.Response.Success(c, device, "设备信息更新成功")
}

// deleteDevice 删除设备
// @Summary 删除设备
// @Description 从系统中删除指定设备
// @Tags Devices
// @Produce json
// @Param id path string true "设备ID"
// @Success 200 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Router /devices/{id} [delete]
func (s *DeviceServiceV1) deleteDevice(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		httpUtils.Response.BadRequest(c, "设备ID不能为空")
		return
	}

	s.logger.InfoTag("API", "删除设备",
		"device_id", deviceID,
		"request_id", getRequestID(c),
	)

	// 检查设备是否存在
	device := s.getMockDevice(deviceID)
	if device == nil {
		httpUtils.Response.NotFound(c, "设备")
		return
	}

	// 检查设备是否可以删除（例如没有正在进行的OTA更新）
	if device.Status == "updating" {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeDeviceUpdating, "设备正在更新中，无法删除")
		return
	}

	httpUtils.Response.Success(c, map[string]interface{}{"device_id": deviceID}, "设备删除成功")
}

// activateDevice 激活设备
// @Summary 激活设备
// @Description 激活已注册的设备
// @Tags Devices
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Param request body v1.DeviceActivationRequest true "设备激活信息"
// @Success 200 {object} httptransport.APIResponse{data=v1.DeviceActivationResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Router /devices/{id}/activate [post]
func (s *DeviceServiceV1) activateDevice(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		httpUtils.Response.BadRequest(c, "设备ID不能为空")
		return
	}

	var request v1.DeviceActivationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "激活设备",
		"device_id", deviceID,
		"activation_code", request.ActivationCode,
		"request_id", getRequestID(c),
	)

	// 获取设备
	device := s.getMockDevice(deviceID)
	if device == nil {
		httpUtils.Response.NotFound(c, "设备")
		return
	}

	// 检查设备是否已激活
	if device.IsActivated {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeDeviceActivated, "设备已激活")
		return
	}

	// 验证激活码
	if !s.validateActivationCode(request.ActivationCode, deviceID) {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInvalidActivationCode, "无效的激活码")
		return
	}

	// 激活设备
	device.IsActivated = true
	device.IsActive = true
	device.Status = "online"
	device.UpdatedAt = time.Now()

	// 生成设备令牌
	deviceToken := fmt.Sprintf("device_token_%d", time.Now().UnixNano())
	accessToken := fmt.Sprintf("access_token_%d", time.Now().UnixNano())

	response := v1.DeviceActivationResponse{
		Success:     true,
		DeviceToken: deviceToken,
		AccessToken: accessToken,
		ExpiresIn:   86400 * 30, // 30天
		Message:     "设备激活成功",
		DeviceInfo:  *device,
	}

	httpUtils.Response.Success(c, response, "设备激活成功")
}

// listFirmware 获取固件列表
// @Summary 获取固件列表
// @Description 获取可用的固件版本列表
// @Tags OTA
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=v1.FirmwareList}
// @Router /ota/firmware [get]
func (s *DeviceServiceV1) listFirmware(c *gin.Context) {
	s.logger.InfoTag("API", "获取固件列表",
		"request_id", getRequestID(c),
	)

	// 模拟固件列表数据
	firmwareList := v1.FirmwareList{
		Firmware: []v1.FirmwareInfo{
			{
				Version:       "1.0.0",
				URL:          "https://example.com/firmware/v1.0.0.bin",
				Checksum:     "sha256:abc123...",
				Size:         1024000,
				ReleaseDate:  time.Now().Add(-7 * 24 * time.Hour),
				DownloadCount: 150,
				Description:  "初始版本",
			},
			{
				Version:       "1.1.0",
				URL:          "https://example.com/firmware/v1.1.0.bin",
				Checksum:     "sha256:def456...",
				Size:         1050000,
				ReleaseDate:  time.Now().Add(-3 * 24 * time.Hour),
				DownloadCount: 85,
				Description:  "性能优化版本",
			},
			{
				Version:       "1.2.0",
				URL:          "https://example.com/firmware/v1.2.0.bin",
				Checksum:     "sha256:ghi789...",
				Size:         1080000,
				ReleaseDate:  time.Now().Add(-1 * 24 * time.Hour),
				DownloadCount: 25,
				Description:  "最新稳定版本",
			},
		},
	}

	httpUtils.Response.Success(c, firmwareList, "获取固件列表成功")
}

// createOTAUpdate 创建OTA更新
// @Summary 创建OTA更新
// @Description 为指定设备创建OTA固件更新任务
// @Tags OTA
// @Accept json
// @Produce json
// @Param request body v1.OTAUpdateRequest true "OTA更新请求"
// @Success 202 {object} httptransport.APIResponse{data=v1.OTAUpdateResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Router /ota/updates [post]
func (s *DeviceServiceV1) createOTAUpdate(c *gin.Context) {
	var request v1.OTAUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "创建OTA更新",
		"firmware_version", request.FirmwareVersion,
		"force_update", request.ForceUpdate,
		"request_id", getRequestID(c),
	)

	// 创建更新任务ID
	updateID := fmt.Sprintf("ota_%d", time.Now().UnixNano())

	response := v1.OTAUpdateResponse{
		UpdateID:     updateID,
		Status:       "pending",
		Progress:     0,
		Message:      "OTA更新任务已创建，等待设备响应",
		DownloadURL:  "https://example.com/firmware/" + request.FirmwareVersion + ".bin",
		FileSize:     1080000,
		StartedAt:    nil,
		CompletedAt:  nil,
	}

	httpUtils.Response.Accepted(c, response, "OTA更新任务创建成功")
}

// getOTAStatus 获取OTA更新状态
// @Summary 获取OTA更新状态
// @Description 查询指定OTA更新任务的状态
// @Tags OTA
// @Produce json
// @Param id path string true "更新任务ID"
// @Success 200 {object} httptransport.APIResponse{data=v1.OTAStatusResponse}
// @Failure 404 {object} httptransport.APIResponse
// @Router /ota/updates/{id} [get]
func (s *DeviceServiceV1) getOTAStatus(c *gin.Context) {
	updateID := c.Param("id")
	if updateID == "" {
		httpUtils.Response.BadRequest(c, "更新任务ID不能为空")
		return
	}

	s.logger.InfoTag("API", "获取OTA更新状态",
		"update_id", updateID,
		"request_id", getRequestID(c),
	)

	// 模拟获取OTA状态
	status := s.getMockOTAStatus(updateID)
	if status == nil {
		httpUtils.Response.NotFound(c, "OTA更新任务")
		return
	}

	httpUtils.Response.Success(c, status, "获取OTA更新状态成功")
}

// cancelOTAUpdate 取消OTA更新
// @Summary 取消OTA更新
// @Description 取消正在进行的OTA更新任务
// @Tags OTA
// @Produce json
// @Param id path string true "更新任务ID"
// @Success 200 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Failure 409 {object} httptransport.APIResponse
// @Router /ota/updates/{id}/cancel [post]
func (s *DeviceServiceV1) cancelOTAUpdate(c *gin.Context) {
	updateID := c.Param("id")
	if updateID == "" {
		httpUtils.Response.BadRequest(c, "更新任务ID不能为空")
		return
	}

	s.logger.InfoTag("API", "取消OTA更新",
		"update_id", updateID,
		"request_id", getRequestID(c),
	)

	// 获取OTA状态
	status := s.getMockOTAStatus(updateID)
	if status == nil {
		httpUtils.Response.NotFound(c, "OTA更新任务")
		return
	}

	// 检查是否可以取消
	if status.Status == "completed" {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeOTACompleted, "OTA更新已完成，无法取消")
		return
	}

	if status.Status == "failed" {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeOTAFailed, "OTA更新已失败")
		return
	}

	httpUtils.Response.Success(c, map[string]interface{}{"update_id": updateID}, "OTA更新已取消")
}

// queryOTAStatus 查询OTA状态
// @Summary 查询OTA状态
// @Description 批量查询OTA更新状态
// @Tags OTA
// @Produce json
// @Param update_id query string false "更新任务ID"
// @Param device_id query string false "设备ID"
// @Success 200 {object} httptransport.APIResponse{data=[]v1.OTAStatusResponse}
// @Router /ota/status [get]
func (s *DeviceServiceV1) queryOTAStatus(c *gin.Context) {
	updateID := c.Query("update_id")
	deviceID := c.Query("device_id")

	s.logger.InfoTag("API", "查询OTA状态",
		"update_id", updateID,
		"device_id", deviceID,
		"request_id", getRequestID(c),
	)

	var statuses []v1.OTAStatusResponse

	if updateID != "" {
		// 查询特定更新任务
		status := s.getMockOTAStatus(updateID)
		if status != nil {
			statuses = append(statuses, *status)
		}
	} else if deviceID != "" {
		// 查询设备的所有更新任务
		deviceStatuses := s.getMockDeviceOTAStatuses(deviceID)
		statuses = append(statuses, deviceStatuses...)
	} else {
		// 返回所有活跃的更新任务
		activeStatuses := s.getMockActiveOTAStatuses()
		statuses = append(statuses, activeStatuses...)
	}

	httpUtils.Response.Success(c, statuses, "查询OTA状态成功")
}

// batchOTAUpdate 批量OTA更新
// @Summary 批量OTA更新
// @Description 为多个设备批量创建OTA更新任务
// @Tags Devices
// @Accept json
// @Produce json
// @Param request body v1.BatchOTAUpdateRequest true "批量OTA更新参数"
// @Success 202 {object} httptransport.APIResponse{data=v1.BatchOTAUpdateResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Router /devices/batch/ota [post]
func (s *DeviceServiceV1) batchOTAUpdate(c *gin.Context) {
	var request v1.BatchOTAUpdateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "批量OTA更新",
		"device_count", len(request.DeviceIDs),
		"firmware_version", request.FirmwareVersion,
		"force_update", request.ForceUpdate,
		"request_id", getRequestID(c),
	)

	// 模拟批量创建OTA更新
	updateTasks := make([]v1.OTATask, 0, len(request.DeviceIDs))
	for _, deviceID := range request.DeviceIDs {
		task := v1.OTATask{
			DeviceID:        deviceID,
			UpdateID:        fmt.Sprintf("ota_%d_%s", time.Now().UnixNano(), deviceID),
			FirmwareVersion:  request.FirmwareVersion,
			Status:          "pending",
			CreatedAt:       time.Now(),
		}
		updateTasks = append(updateTasks, task)
	}

	response := v1.BatchOTAUpdateResponse{
		BatchID:    fmt.Sprintf("batch_%d", time.Now().UnixNano()),
		TotalTasks: len(updateTasks),
		Tasks:      updateTasks,
		CreatedAt:  time.Now(),
	}

	httpUtils.Response.Accepted(c, response, "批量OTA更新任务创建成功")
}

// batchCommand 批量命令执行
// @Summary 批量命令执行
// @Description 向多个设备批量执行命令
// @Tags Devices
// @Accept json
// @Produce json
// @Param request body v1.BatchCommandRequest true "批量命令参数"
// @Success 202 {object} httptransport.APIResponse{data=v1.BatchCommandResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Router /devices/batch/command [post]
func (s *DeviceServiceV1) batchCommand(c *gin.Context) {
	var request v1.BatchCommandRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "批量命令执行",
		"device_count", len(request.DeviceIDs),
		"command", request.Command,
		"request_id", getRequestID(c),
	)

	// 模拟批量命令执行
	commands := make([]v1.CommandTask, 0, len(request.DeviceIDs))
	for _, deviceID := range request.DeviceIDs {
		cmd := v1.CommandTask{
			DeviceID:  deviceID,
			CommandID: fmt.Sprintf("cmd_%d_%s", time.Now().UnixNano(), deviceID),
			Command:   request.Command,
			Params:    request.Params,
			Status:    "sent",
			SentAt:    time.Now(),
		}
		commands = append(commands, cmd)
	}

	response := v1.BatchCommandResponse{
		BatchID:   fmt.Sprintf("batch_%d", time.Now().UnixNano()),
		TotalCmds: len(commands),
		Commands:  commands,
		CreatedAt: time.Now(),
	}

	httpUtils.Response.Accepted(c, response, "批量命令发送成功")
}

// batchConfig 批量配置更新
// @Summary 批量配置更新
// @Description 批量更新多个设备的配置
// @Tags Devices
// @Accept json
// @Produce json
// @Param request body v1.BatchConfigRequest true "批量配置参数"
// @Success 202 {object} httptransport.APIResponse{data=v1.BatchConfigResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Router /devices/batch/config [post]
func (s *DeviceServiceV1) batchConfig(c *gin.Context) {
	var request v1.BatchConfigRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "批量配置更新",
		"device_count", len(request.DeviceIDs),
		"config_keys", len(request.Configuration),
		"merge", request.Merge,
		"restart", request.Restart,
		"request_id", getRequestID(c),
	)

	// 模拟批量配置更新
	configs := make([]v1.ConfigTask, 0, len(request.DeviceIDs))
	for _, deviceID := range request.DeviceIDs {
		cfg := v1.ConfigTask{
			DeviceID:     deviceID,
			ConfigID:     fmt.Sprintf("cfg_%d_%s", time.Now().UnixNano(), deviceID),
			Configuration: request.Configuration,
			Merge:        request.Merge,
			Restart:      request.Restart,
			Status:       "pending",
			CreatedAt:    time.Now(),
		}
		configs = append(configs, cfg)
	}

	response := v1.BatchConfigResponse{
		BatchID:      fmt.Sprintf("batch_%d", time.Now().UnixNano()),
		TotalConfigs: len(configs),
		Configs:      configs,
		CreatedAt:    time.Now(),
	}

	httpUtils.Response.Accepted(c, response, "批量配置更新任务创建成功")
}

// getWebSocketInfo 获取WebSocket连接信息
// @Summary 获取WebSocket连接信息
// @Description 获取WebSocket服务连接信息
// @Tags Connections
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=v1.WebSocketInfo}
// @Router /connections/websocket [get]
func (s *DeviceServiceV1) getWebSocketInfo(c *gin.Context) {
	s.logger.InfoTag("API", "获取WebSocket连接信息",
		"request_id", getRequestID(c),
	)

	wsInfo := v1.WebSocketInfo{
		URL:       "ws://localhost:8080/ws",
		Path:      "/ws",
		Protocol:  "websocket",
		Status:    "running",
		Connected: true,
		Clients:   25,
		StartTime: time.Now().Add(-2 * time.Hour),
	}

	httpUtils.Response.Success(c, wsInfo, "获取WebSocket连接信息成功")
}

// getMQTTInfo 获取MQTT连接信息
// @Summary 获取MQTT连接信息
// @Description 获取MQTT服务连接信息
// @Tags Connections
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=v1.MQTTInfo}
// @Router /connections/mqtt [get]
func (s *DeviceServiceV1) getMQTTInfo(c *gin.Context) {
	s.logger.InfoTag("API", "获取MQTT连接信息",
		"request_id", getRequestID(c),
	)

	mqttInfo := v1.MQTTInfo{
		Broker:    "mqtt.example.com",
		Port:      1883,
		Protocol:  "mqtt",
		Status:    "running",
		Connected: true,
		Clients:   150,
		Uptime:    7200, // 2小时
	}

	httpUtils.Response.Success(c, mqttInfo, "获取MQTT连接信息成功")
}

// ========== 模拟数据方法 ==========
// TODO: 实际实现中应该从数据库或配置中获取真实数据

func (s *DeviceServiceV1) deviceExists(deviceID string) bool {
	// 简单模拟设备存在性检查
	existingDevices := []string{"device_001", "device_002", "device_003"}
	for _, existing := range existingDevices {
		if existing == deviceID {
			return true
		}
	}
	return false
}

func (s *DeviceServiceV1) getMockDevice(deviceID string) *v1.DeviceInfo {
	// 模拟设备数据
	if deviceID == "device_001" {
		return &v1.DeviceInfo{
			ID:             1,
			DeviceID:       "device_001",
			DeviceName:     "智能门锁",
			DeviceType:     "smart_lock",
			Model:          "XZ-L100",
			Version:        "1.0.0",
			Status:         "online",
			Location: &v1.DeviceLocation{
				Latitude:  39.9042,
				Longitude: 116.4074,
				Address:   "北京市朝阳区",
				City:      "北京",
				Province:  "北京",
				Country:   "中国",
			},
			Configuration: map[string]interface{}{
				"auto_lock": true,
				"lock_delay": 30,
			},
			Metadata: map[string]interface{}{
				"manufacturer": "XiaoZhi Tech",
				"serial_number": "SN001",
			},
			IsActive:    true,
			IsActivated: true,
			CreatedAt:   time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		}
	}
	return nil
}

func (s *DeviceServiceV1) getMockDeviceList(query v1.DeviceQuery) ([]v1.DeviceInfo, v1.Pagination) {
	// 模拟设备列表数据
	devices := []v1.DeviceInfo{
		{
			ID:             1,
			DeviceID:       "device_001",
			DeviceName:     "智能门锁",
			DeviceType:     "smart_lock",
			Model:          "XZ-L100",
			Version:        "1.0.0",
			Status:         "online",
			IsActive:       true,
			IsActivated:    true,
			CreatedAt:      time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             2,
			DeviceID:       "device_002",
			DeviceName:     "温湿度传感器",
			DeviceType:     "sensor",
			Model:          "XZ-T200",
			Version:        "1.1.0",
			Status:         "offline",
			IsActive:       false,
			IsActivated:    true,
			CreatedAt:      time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt:      time.Now().Add(-1 * time.Hour),
		},
		{
			ID:             3,
			DeviceID:       "device_003",
			DeviceName:     "智能摄像头",
			DeviceType:     "camera",
			Model:          "XZ-C300",
			Version:        "1.0.0",
			Status:         "updating",
			IsActive:       true,
			IsActivated:    true,
			CreatedAt:      time.Now().Add(-3 * 24 * time.Hour),
			UpdatedAt:      time.Now().Add(-30 * time.Minute),
		},
	}

	// 简单的过滤逻辑
	var filtered []v1.DeviceInfo
	for _, device := range devices {
		if query.Status != "" && device.Status != query.Status {
			continue
		}
		if query.DeviceType != "" && device.DeviceType != query.DeviceType {
			continue
		}
		if query.Search != "" {
			// 简单搜索逻辑
			match := false
			for _, field := range []string{device.DeviceName, device.DeviceID} {
				if len(field) >= len(query.Search) {
					if field[:len(query.Search)] == query.Search {
						match = true
						break
					}
				}
			}
			if !match {
				continue
			}
		}
		filtered = append(filtered, device)
	}

	// 分页逻辑
	total := int64(len(filtered))
	totalPages := (total + int64(query.Limit) - 1) / int64(query.Limit)
	start := (query.Page - 1) * query.Limit
	end := start + query.Limit
	if end > len(filtered) {
		end = len(filtered)
	}
	if start >= len(filtered) {
		return []v1.DeviceInfo{}, v1.Pagination{
			Page:       int64(query.Page),
			Limit:      int64(query.Limit),
			Total:      total,
			TotalPages: totalPages,
			HasNext:    false,
			HasPrev:    query.Page > 1,
		}
	}

	// 处理位置信息
	pagedDevices := filtered[start:end]
	if !query.Location {
		for i := range pagedDevices {
			pagedDevices[i].Location = nil
		}
	}

	pagination := v1.Pagination{
		Page:       int64(query.Page),
		Limit:      int64(query.Limit),
		Total:      total,
		TotalPages: totalPages,
		HasNext:    int64(query.Page) < totalPages,
		HasPrev:    query.Page > 1,
	}

	return pagedDevices, pagination
}

func (s *DeviceServiceV1) validateActivationCode(code, deviceID string) bool {
	// 简单的激活码验证逻辑
	return code == fmt.Sprintf("ACT_%s", deviceID)
}

func (s *DeviceServiceV1) getMockOTAStatus(updateID string) *v1.OTAStatusResponse {
	// 模拟OTA状态数据
	return &v1.OTAStatusResponse{
		UpdateID: updateID,
		DeviceID: "device_001",
		Status:   "downloading",
		Progress: 65,
		Message:  "正在下载固件...",
		FirmwareInfo: &v1.FirmwareInfo{
			Version:      "1.2.0",
			URL:          "https://example.com/firmware/v1.2.0.bin",
			Checksum:     "sha256:ghi789...",
			Size:         1080000,
			ReleaseDate:  time.Now().Add(-1 * 24 * time.Hour),
			DownloadCount: 25,
			Description:  "最新稳定版本",
		},
		StartedAt: &[]time.Time{time.Now().Add(-10 * time.Minute)}[0],
	}
}

func (s *DeviceServiceV1) getMockDeviceOTAStatuses(deviceID string) []v1.OTAStatusResponse {
	// 模拟设备的OTA状态列表
	return []v1.OTAStatusResponse{
		{
			UpdateID: fmt.Sprintf("ota_%d_%s", time.Now().UnixNano(), deviceID),
			DeviceID: deviceID,
			Status:   "completed",
			Progress: 100,
			Message:  "更新完成",
		},
	}
}

func (s *DeviceServiceV1) getMockActiveOTAStatuses() []v1.OTAStatusResponse {
	// 模拟所有活跃的OTA状态
	return []v1.OTAStatusResponse{
		{
			UpdateID: "ota_active_1",
			DeviceID: "device_001",
			Status:   "downloading",
			Progress: 45,
			Message:  "下载中...",
		},
		{
			UpdateID: "ota_active_2",
			DeviceID: "device_002",
			Status:   "installing",
			Progress: 85,
			Message:  "安装中...",
		},
	}
}