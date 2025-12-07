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
		devices.POST("/status", s.updateDeviceStatus) // 管理员激活/禁用设备
	}

	// 注意：OTA接口已移除，使用主服务的 /api/ota/ 接口
	// 这样避免了重复的OTA接口，设备统一使用 /api/ota/ 进行通信
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
// @Router /v1/devices [post]
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
// @Router /v1/devices [get]
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
// @Router /v1/devices/{id} [get]
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
// @Router /v1/devices/{id} [put]
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
// @Router /v1/devices/{id} [delete]
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

// 注意：handleOTARequest 函数已移除，避免与主OTA服务 (/api/ota/) 冲突
// 设备应统一使用主服务的 /api/ota/ 接口进行OTA操作

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
// @Router /v1/devices/{id}/activate [post]
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

// updateDeviceStatus 管理员激活/禁用设备
// @Summary 管理员激活/禁用设备
// @Description 管理员快速激活或禁用设备，通过设备MAC地址和激活状态
// @Tags Devices
// @Accept json
// @Produce json
// @Param request body v1.DeviceStatusRequest true "设备状态管理请求"
// @Success 200 {object} httptransport.APIResponse{data=v1.DeviceStatusResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Router /v1/devices/status [post]
func (s *DeviceServiceV1) updateDeviceStatus(c *gin.Context) {
	var request v1.DeviceStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		s.logger.ErrorTag("API", "JSON绑定失败",
			"error", err,
			"request_id", getRequestID(c),
		)
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "管理员更新设备状态",
		"device_id", request.DeviceID,
		"is_active", request.IsActive,
		"request_id", getRequestID(c),
	)

	// 获取设备
	device := s.getMockDevice(request.DeviceID)
	if device == nil {
		httpUtils.Response.NotFound(c, "设备")
		return
	}

	// 更新设备认证状态
	oldAuthStatus := device.Status // 这里复用Status字段来表示auth_status

	if request.IsActive {
		// 激活设备：设置为已认证状态
		device.Status = "approved"
		device.IsActivated = true
	} else {
		// 禁用设备：设置为已拒绝状态
		device.Status = "rejected"
		device.IsActivated = false
	}

	device.IsActive = request.IsActive
	device.UpdatedAt = time.Now()

	// 构建响应消息
	var message string
	if request.IsActive {
		message = "设备激活成功"
	} else {
		message = "设备禁用成功"
	}

	// 记录状态变化
	s.logger.InfoTag("API", "设备认证状态已更新",
		"device_id", request.DeviceID,
		"old_auth_status", oldAuthStatus,
		"new_auth_status", device.Status,
		"request_id", getRequestID(c),
	)

	response := v1.DeviceStatusResponse{
		Success:   true,
		Message:   message,
		DeviceInfo: *device,
	}

	httpUtils.Response.Success(c, response, message)
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
	mockDevices := map[string]*v1.DeviceInfo{
		"device_001": {
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
		},
		"device_002": {
			ID:             2,
			DeviceID:       "device_002",
			DeviceName:     "温湿度传感器",
			DeviceType:     "sensor",
			Model:          "XZ-T200",
			Version:        "1.1.0",
			Status:         "offline",
			Location: &v1.DeviceLocation{
				Latitude:  31.2304,
				Longitude: 121.4737,
				Address:   "上海市浦东新区",
				City:      "上海",
				Province:  "上海",
				Country:   "中国",
			},
			Configuration: map[string]interface{}{
				"temperature_unit": "celsius",
				"report_interval":  300,
			},
			Metadata: map[string]interface{}{
				"manufacturer": "XiaoZhi Tech",
				"serial_number": "SN002",
			},
			IsActive:    false,
			IsActivated: true,
			CreatedAt:   time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
		"device_003": {
			ID:             3,
			DeviceID:       "device_003",
			DeviceName:     "智能摄像头",
			DeviceType:     "camera",
			Model:          "XZ-C300",
			Version:        "1.0.0",
			Status:         "updating",
			Location: &v1.DeviceLocation{
				Latitude:  22.5431,
				Longitude: 114.0579,
				Address:   "深圳市南山区",
				City:      "深圳",
				Province:  "广东",
				Country:   "中国",
			},
			Configuration: map[string]interface{}{
				"resolution": "1080p",
				"night_vision": true,
			},
			Metadata: map[string]interface{}{
				"manufacturer": "XiaoZhi Tech",
				"serial_number": "SN003",
			},
			IsActive:    true,
			IsActivated: true,
			CreatedAt:   time.Now().Add(-3 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * time.Minute),
		},
	}

	device, exists := mockDevices[deviceID]
	if !exists {
		return nil
	}

	// 返回设备的副本以避免修改原始数据
	deviceCopy := *device
	return &deviceCopy
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

