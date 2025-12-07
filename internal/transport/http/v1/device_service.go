package v1

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"xiaozhi-server-go/internal/domain/device/aggregate"
	"xiaozhi-server-go/internal/domain/device/repository"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/storage"
	"xiaozhi-server-go/internal/transport/http/types/v1"
	httpUtils "xiaozhi-server-go/internal/transport/http/utils"
	"xiaozhi-server-go/internal/utils"
	"gorm.io/gorm"
)

// DeviceConnectionManager 设备连接管理器接口
type DeviceConnectionManager interface {
	CloseDeviceConnection(deviceID string) error
}

// DeviceServiceV1 V1版本设备服务
type DeviceServiceV1 struct {
	logger            *utils.Logger
	config            *config.Config
	db                *gorm.DB
	deviceRepo        repository.DeviceRepository
	connManager       DeviceConnectionManager
}

// NewDeviceServiceV1 创建设备服务V1实例
func NewDeviceServiceV1(config *config.Config, logger *utils.Logger, connManager DeviceConnectionManager) (*DeviceServiceV1, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	logger.InfoTag("DeviceService", "开始初始化设备服务")

	// 获取数据库连接
	db := storage.GetDB()
	if db == nil {
		logger.ErrorTag("DeviceService", "数据库未初始化")
		return nil, fmt.Errorf("database not initialized")
	}

	logger.InfoTag("DeviceService", "数据库连接成功")

	// 测试数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		logger.ErrorTag("DeviceService", "获取底层SQL数据库连接失败", "error", err)
		return nil, fmt.Errorf("failed to get underlying database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		logger.ErrorTag("DeviceService", "数据库连接测试失败", "error", err)
		return nil, fmt.Errorf("database connection test failed: %w", err)
	}

	logger.InfoTag("DeviceService", "数据库连接测试成功")

	// 创建设备仓库
	deviceRepo := storage.NewDeviceRepository(db)
	logger.InfoTag("DeviceService", "设备仓库创建成功")

	service := &DeviceServiceV1{
		logger:      logger,
		config:      config,
		db:          db,
		deviceRepo:  deviceRepo,
		connManager: connManager,
	}

	logger.InfoTag("DeviceService", "设备服务初始化完成")
	return service, nil
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
	ctx := context.Background()
	existingDevice, err := s.deviceRepo.FindByDeviceID(ctx, request.DeviceID)
	if err != nil {
		s.logger.ErrorTag("API", "检查设备是否存在失败", "error", err, "device_id", request.DeviceID, "request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "检查设备失败")
		return
	}
	if existingDevice != nil {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeDeviceExists, "设备已存在")
		return
	}

	// 创建设备聚合根
	now := time.Now()
	newDevice := &aggregate.Device{
		DeviceID:       request.DeviceID,
		ClientID:       fmt.Sprintf("client_%s", request.DeviceID),
		Name:           request.DeviceName,
		BoardType:      request.DeviceType,
		ChipModelName:  request.Model,
		Version:        request.Version,
		Online:         false,
		AuthStatus:     aggregate.DeviceStatusPending,
		RegisterTime:   now,
		LastActiveTime: now,
	}

	// 保存到数据库
	if err := s.deviceRepo.Save(ctx, newDevice); err != nil {
		s.logger.ErrorTag("API", "保存设备失败", "error", err, "device_id", request.DeviceID, "request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "设备注册失败")
		return
	}

	// 转换为API响应格式
	device := s.convertAggregateToAPI(newDevice)
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

	// 从数据库获取设备列表
	s.logger.InfoTag("API", "开始从数据库获取设备列表", "request_id", getRequestID(c))
	devices, total, err := s.getDeviceListFromDB(query)
	if err != nil {
		s.logger.ErrorTag("API", "获取设备列表失败",
			"error", err,
			"error_details", err.Error(),
			"request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, fmt.Sprintf("获取设备列表失败: %v", err))
		return
	}

	// 计算分页信息
	totalPages := (total + int64(query.Limit) - 1) / int64(query.Limit)
	pagination := v1.Pagination{
		Page:       int64(query.Page),
		Limit:      int64(query.Limit),
		Total:      total,
		TotalPages: totalPages,
		HasNext:    int64(query.Page) < totalPages,
		HasPrev:    query.Page > 1,
	}

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

	// 从数据库获取设备详情
	device, err := s.getDeviceFromDB(deviceID)
	if err != nil {
		s.logger.ErrorTag("API", "获取设备详情失败", "error", err, "device_id", deviceID, "request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "获取设备详情失败")
		return
	}
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

	// 从数据库获取设备
	ctx := context.Background()
	device, err := s.deviceRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		s.logger.ErrorTag("API", "获取设备失败", "error", err, "device_id", deviceID, "request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "获取设备失败")
		return
	}
	if device == nil {
		httpUtils.Response.NotFound(c, "设备")
		return
	}

	// 更新字段
	updated := false
	if request.DeviceName != "" {
		device.Name = request.DeviceName
		updated = true
	}
	if request.IsActive != nil {
		device.Online = *request.IsActive
		// 同时更新认证状态
		if *request.IsActive {
			device.AuthStatus = aggregate.DeviceStatusApproved
		} else {
			device.AuthStatus = aggregate.DeviceStatusRejected
			// 如果禁用设备，强制断开连接
			if s.connManager != nil {
				if err := s.connManager.CloseDeviceConnection(deviceID); err != nil {
					s.logger.WarnTag("API", "断开设备连接失败: %v", err)
				} else {
					s.logger.InfoTag("API", "已强制断开设备连接", "device_id", deviceID)
				}
			}
		}
		updated = true
	}

	// 更新数据库
	if updated {
		device.LastActiveTime = time.Now()
		if err := s.deviceRepo.Update(ctx, device); err != nil {
			s.logger.ErrorTag("API", "更新设备失败", "error", err, "device_id", deviceID, "request_id", getRequestID(c))
			httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "更新设备失败")
			return
		}
	}

	httpUtils.Response.Success(c, s.convertAggregateToAPI(device), "设备信息更新成功")
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
	device, err := s.getDeviceFromDB(deviceID)
	if err != nil {
		s.logger.ErrorTag("API", "获取设备失败", "error", err, "device_id", deviceID, "request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "获取设备失败")
		return
	}
	if device == nil {
		httpUtils.Response.NotFound(c, "设备")
		return
	}

	// 检查设备是否可以删除（例如没有正在进行的OTA更新）
	if device.Status == "updating" {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeDeviceUpdating, "设备正在更新中，无法删除")
		return
	}

	// 从数据库删除设备
	ctx := context.Background()
	if err := s.deviceRepo.Delete(ctx, deviceID); err != nil {
		s.logger.ErrorTag("API", "删除设备失败", "error", err, "device_id", deviceID, "request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "删除设备失败")
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

	// 从数据库获取设备
	ctx := context.Background()
	device, err := s.deviceRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		s.logger.ErrorTag("API", "获取设备失败", "error", err, "device_id", deviceID, "request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "获取设备失败")
		return
	}
	if device == nil {
		httpUtils.Response.NotFound(c, "设备")
		return
	}

	// 检查设备是否已激活
	if device.AuthStatus == aggregate.DeviceStatusApproved {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeDeviceActivated, "设备已激活")
		return
	}

	// 验证激活码
	if !s.validateActivationCode(request.ActivationCode, deviceID) {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInvalidActivationCode, "无效的激活码")
		return
	}

	// 激活设备 - 更新数据库
	if err := s.deviceRepo.UpdateDeviceStatus(ctx, deviceID, true); err != nil {
		s.logger.ErrorTag("API", "激活设备失败", "error", err, "device_id", deviceID, "request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "激活设备失败")
		return
	}

	// 更新本地对象状态以返回
	device.AuthStatus = aggregate.DeviceStatusApproved
	device.Online = true
	device.LastActiveTime = time.Now()

	// 生成设备令牌
	deviceToken := fmt.Sprintf("device_token_%d", time.Now().UnixNano())
	accessToken := fmt.Sprintf("access_token_%d", time.Now().UnixNano())

	response := v1.DeviceActivationResponse{
		Success:     true,
		DeviceToken: deviceToken,
		AccessToken: accessToken,
		ExpiresIn:   86400 * 30, // 30天
		Message:     "设备激活成功",
		DeviceInfo:  *s.convertAggregateToAPI(device),
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
		"is_active", *request.IsActive,
		"request_id", getRequestID(c),
	)

	// 从数据库获取设备
	ctx := context.Background()
	device, err := s.deviceRepo.FindByDeviceID(ctx, request.DeviceID)
	if err != nil {
		s.logger.ErrorTag("API", "获取设备失败", "error", err, "device_id", request.DeviceID, "request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "获取设备失败")
		return
	}
	if device == nil {
		httpUtils.Response.NotFound(c, "设备")
		return
	}

	// 获取旧的认证状态
	oldAuthStatus := device.AuthStatus

	// 更新数据库中的设备认证状态
	if err := s.deviceRepo.UpdateDeviceStatus(ctx, request.DeviceID, *request.IsActive); err != nil {
		s.logger.ErrorTag("API", "更新设备状态失败", "error", err, "device_id", request.DeviceID, "request_id", getRequestID(c))
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInternalServer, "更新设备状态失败")
		return
	}

	// 更新本地对象状态
	if *request.IsActive {
		device.AuthStatus = aggregate.DeviceStatusApproved
		device.Online = true
	} else {
		device.AuthStatus = aggregate.DeviceStatusRejected
		device.Online = false
		// 如果禁用设备，强制断开连接
		if s.connManager != nil {
			if err := s.connManager.CloseDeviceConnection(request.DeviceID); err != nil {
				s.logger.WarnTag("API", "断开设备连接失败: %v", err)
			} else {
				s.logger.InfoTag("API", "已强制断开设备连接", "device_id", request.DeviceID)
			}
		}
	}
	device.LastActiveTime = time.Now()

	// 构建响应消息
	var message string
	if *request.IsActive {
		message = "设备激活成功"
	} else {
		message = "设备禁用成功"
	}

	// 记录状态变化
	s.logger.InfoTag("API", "设备认证状态已更新",
		"device_id", request.DeviceID,
		"old_auth_status", oldAuthStatus,
		"new_auth_status", device.AuthStatus,
		"request_id", getRequestID(c),
	)

	response := v1.DeviceStatusResponse{
		Success:    true,
		Message:    message,
		DeviceInfo: *s.convertAggregateToAPI(device),
	}

	httpUtils.Response.Success(c, response, message)
}




// ========== 数据转换方法 ==========
// convertStorageToAPI 将数据库Device模型转换为API类型
func (s *DeviceServiceV1) convertStorageToAPI(device *storage.Device) *v1.DeviceInfo {
	if device == nil {
		return nil
	}

	// 转换在线状态
	status := "offline"
	if device.Online {
		status = "online"
	}

	// 如果auth_status存在，使用它作为状态
	if device.AuthStatus != "" {
		status = device.AuthStatus
	}

	deviceInfo := &v1.DeviceInfo{
		ID:            int64(device.ID),
		DeviceID:      device.DeviceID,
		DeviceName:    device.Name,
		DeviceType:    device.BoardType, // 使用BoardType作为设备类型
		Model:         device.ChipModelName,
		Version:       device.Version,
		Status:        status,
		Configuration: make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
		IsActive:      device.Online,
		IsActivated:   device.AuthStatus == "approved",
		CreatedAt:     device.RegisterTimeV2,
		UpdatedAt:     device.LastActiveTimeV2,
	}

	// 如果有最后活跃时间，设置LastSeen
	if !device.LastActiveTimeV2.IsZero() {
		deviceInfo.LastSeen = &device.LastActiveTimeV2
	}

	// 添加一些基础元数据
	if device.ChipModelName != "" {
		deviceInfo.Metadata["chip_model"] = device.ChipModelName
	}
	if device.SSID != "" {
		deviceInfo.Metadata["ssid"] = device.SSID
	}
	if device.LastIP != "" {
		deviceInfo.Metadata["last_ip"] = device.LastIP
	}

	return deviceInfo
}

// convertAggregateToAPI 将领域聚合Device模型转换为API类型
func (s *DeviceServiceV1) convertAggregateToAPI(device *aggregate.Device) *v1.DeviceInfo {
	if device == nil {
		return nil
	}

	// 转换在线状态
	status := "offline"
	if device.Online {
		status = "online"
	}

	// 如果auth_status存在，使用它作为状态
	if string(device.AuthStatus) != "" {
		status = string(device.AuthStatus)
	}

	deviceInfo := &v1.DeviceInfo{
		ID:            int64(device.ID),
		DeviceID:      device.DeviceID,
		DeviceName:    device.Name,
		DeviceType:    device.BoardType,
		Model:         device.ChipModelName,
		Version:       device.Version,
		Status:        status,
		Configuration: make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
		IsActive:      device.Online,
		IsActivated:   device.AuthStatus == aggregate.DeviceStatusApproved,
		CreatedAt:     device.RegisterTime,
		UpdatedAt:     device.LastActiveTime,
	}

	// 如果有最后活跃时间，设置LastSeen
	if !device.LastActiveTime.IsZero() {
		deviceInfo.LastSeen = &device.LastActiveTime
	}

	// 添加一些基础元数据
	if device.ChipModelName != "" {
		deviceInfo.Metadata["chip_model"] = device.ChipModelName
	}
	if device.SSID != "" {
		deviceInfo.Metadata["ssid"] = device.SSID
	}
	if device.LastIP != "" {
		deviceInfo.Metadata["last_ip"] = device.LastIP
	}

	return deviceInfo
}

func (s *DeviceServiceV1) validateActivationCode(code, deviceID string) bool {
	// 简单的激活码验证逻辑
	return code == fmt.Sprintf("ACT_%s", deviceID)
}

// ========== 数据库查询方法 ==========

// getDeviceListFromDB 从数据库获取设备列表
func (s *DeviceServiceV1) getDeviceListFromDB(query v1.DeviceQuery) ([]v1.DeviceInfo, int64, error) {
	// 检查数据库连接
	if s.db == nil {
		return nil, 0, fmt.Errorf("database connection is nil")
	}

	s.logger.DebugTag("API", "getDeviceListFromDB: 开始查询",
		"status", query.Status,
		"device_type", query.DeviceType,
		"search", query.Search,
		"page", query.Page,
		"limit", query.Limit)

	// 构建查询
	db := s.db.Model(&storage.Device{})
	if db == nil {
		return nil, 0, fmt.Errorf("failed to create database model")
	}

	// 添加过滤条件
	if query.Status != "" {
		db = db.Where("auth_status = ?", query.Status)
	}
	if query.DeviceType != "" {
		db = db.Where("board_type = ?", query.DeviceType)
	}
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		db = db.Where("device_id LIKE ? OR name LIKE ?", searchPattern, searchPattern)
	}

	// 获取总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count devices: %w", err)
	}

	s.logger.DebugTag("API", "getDeviceListFromDB: 设备总数", "total", total)

	// 添加排序
	orderBy := "register_time_v2 DESC"
	if query.SortBy != "" {
		direction := "ASC"
		if query.SortOrder == "desc" {
			direction = "DESC"
		}
		// 映射API字段名到数据库字段名
		dbField := query.SortBy
		switch query.SortBy {
		case "created_at":
			dbField = "register_time_v2"
		case "updated_at":
			dbField = "last_active_time_v2"
		case "device_name":
			dbField = "name"
		case "device_type":
			dbField = "board_type"
		}
		orderBy = fmt.Sprintf("%s %s", dbField, direction)
	}
	db = db.Order(orderBy)

	// 分页
	offset := (query.Page - 1) * query.Limit

	// 查询数据
	var devices []storage.Device
	if err := db.Offset(offset).Limit(query.Limit).Find(&devices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch devices: %w", err)
	}

	s.logger.DebugTag("API", "getDeviceListFromDB: 查询到设备数量", "count", len(devices))

	// 转换为API类型
	var deviceInfos []v1.DeviceInfo
	for _, device := range devices {
		deviceInfo := s.convertStorageToAPI(&device)

		// 如果不要求位置信息，清空位置数据
		if !query.Location {
			deviceInfo.Location = nil
		}

		deviceInfos = append(deviceInfos, *deviceInfo)
	}

	return deviceInfos, total, nil
}

// getDeviceFromDB 从数据库获取单个设备
func (s *DeviceServiceV1) getDeviceFromDB(deviceID string) (*v1.DeviceInfo, error) {
	ctx := context.Background()

	// 从领域层获取设备
	deviceAggregate, err := s.deviceRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find device: %w", err)
	}
	if deviceAggregate == nil {
		return nil, nil // 设备不存在
	}

	// 转换为API类型
	deviceInfo := s.convertAggregateToAPI(deviceAggregate)
	return deviceInfo, nil
}

