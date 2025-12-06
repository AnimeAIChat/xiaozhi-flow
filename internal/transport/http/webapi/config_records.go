package webapi

import (
	"net/http"
	"strconv"
	"time"

	"xiaozhi-server-go/internal/platform/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// handleGetConfigRecords 获取配置记录列表
// @Summary 获取配置记录列表
// @Description 获取所有配置记录，支持分页和过滤
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param category query string false "按分类过滤"
// @Param active query bool false "按活跃状态过滤"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} APIResponse{data=ConfigRecordResponse}
// @Failure 401 {object} object
// @Router /admin/config/records [get]
func (s *Service) handleGetConfigRecords(c *gin.Context) {
	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "Database not available")
		return
	}

	// 获取查询参数
	category := c.Query("category")
	activeStr := c.Query("active")
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	// 设置分页默认值
	page := 1
	limit := 20
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// 构建查询
	query := db.Model(&storage.ConfigRecord{})

	// 应用过滤条件
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if activeStr != "" {
		active := activeStr == "true"
		query = query.Where("is_active = ?", active)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var records []storage.ConfigRecord
	offset := (page - 1) * limit
	result := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&records)

	if result.Error != nil {
		s.logger.ErrorTag("Config", "Failed to get config records: %v", result.Error)
		s.respondError(c, http.StatusInternalServerError, "Failed to get config records")
		return
	}

	s.respondSuccess(c, http.StatusOK, gin.H{
		"data": gin.H{
			"records": records,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	}, "Config records retrieved successfully")
}

// handleGetConfigRecord 获取单个配置记录
// @Summary 获取单个配置记录
// @Description 根据ID获取配置记录详情
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path uint true "配置记录ID"
// @Success 200 {object} APIResponse{data=ConfigRecord}
// @Failure 404 {object} object
// @Router /admin/config/records/{id} [get]
func (s *Service) handleGetConfigRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "Database not available")
		return
	}

	var record storage.ConfigRecord
	result := db.First(&record, uint(id))

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			s.respondError(c, http.StatusNotFound, "Config record not found")
		} else {
			s.logger.ErrorTag("Config", "Failed to get config record: %v", result.Error)
			s.respondError(c, http.StatusInternalServerError, "Failed to get config record")
		}
		return
	}

	s.respondSuccess(c, http.StatusOK, gin.H{
		"data": record,
	}, "Config record retrieved successfully")
}

// handleCreateConfigRecord 创建配置记录
// @Summary 创建配置记录
// @Description 创建新的配置记录
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param record body CreateConfigRecordRequest true "配置记录数据"
// @Success 201 {object} APIResponse{data=ConfigRecord}
// @Failure 400 {object} object
// @Router /admin/config/records [post]
func (s *Service) handleCreateConfigRecord(c *gin.Context) {
	var request CreateConfigRecordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "Database not available")
		return
	}

	// 验证必填字段
	if request.Key == "" {
		s.respondError(c, http.StatusBadRequest, "Config key is required")
		return
	}

	// 检查键是否已存在
	var existingCount int64
	db.Model(&storage.ConfigRecord{}).Where("key = ?", request.Key).Count(&existingCount)
	if existingCount > 0 {
		s.respondError(c, http.StatusConflict, "Config key already exists")
		return
	}

	record := storage.ConfigRecord{
		Key:         request.Key,
		Value:       storage.FlexibleJSON{Data: request.Value},
		Description: request.Description,
		Category:    request.Category,
		Version:     1,
		IsActive:    request.IsActive,
	}

	result := db.Create(&record)
	if result.Error != nil {
		s.logger.ErrorTag("Config", "Failed to create config record: %v", result.Error)
		s.respondError(c, http.StatusInternalServerError, "Failed to create config record")
		return
	}

	s.respondSuccess(c, http.StatusCreated, gin.H{
		"data": record,
	}, "Config record created successfully")
}

// handleUpdateConfigRecord 更新配置记录
// @Summary 更新配置记录
// @Description 根据ID更新配置记录
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "配置记录ID"
// @Param record body UpdateConfigRecordRequest true "配置记录更新数据"
// @Success 200 {object} APIResponse{data=ConfigRecord}
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Router /admin/config/records/{id} [put]
func (s *Service) handleUpdateConfigRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var request UpdateConfigRecordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "Database not available")
		return
	}

	var record storage.ConfigRecord
	result := db.First(&record, uint(id))
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			s.respondError(c, http.StatusNotFound, "Config record not found")
		} else {
			s.respondError(c, http.StatusInternalServerError, "Failed to get config record")
		}
		return
	}

	// 更新字段
	if request.Description != nil {
		record.Description = *request.Description
	}
	if request.Category != nil {
		record.Category = *request.Category
	}
	if request.IsActive != nil {
		record.IsActive = *request.IsActive
	}
	if request.Value != nil {
		record.Value = storage.FlexibleJSON{Data: *request.Value}
		record.Version++ // 版本号递增
	}

	result = db.Save(&record)
	if result.Error != nil {
		s.logger.ErrorTag("Config", "Failed to update config record: %v", result.Error)
		s.respondError(c, http.StatusInternalServerError, "Failed to update config record")
		return
	}

	s.respondSuccess(c, http.StatusOK, gin.H{
		"data": record,
	}, "Config record updated successfully")
}

// handleDeleteConfigRecord 删除配置记录
// @Summary 删除配置记录
// @Description 根据ID删除配置记录
// @Tags Admin
// @Security BearerAuth
// @Param id path uint true "配置记录ID"
// @Success 200 {object} APIResponse
// @Failure 404 {object} object
// @Router /admin/config/records/{id} [delete]
func (s *Service) handleDeleteConfigRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "Database not available")
		return
	}

	result := db.Delete(&storage.ConfigRecord{}, uint(id))
	if result.Error != nil {
		s.logger.ErrorTag("Config", "Failed to delete config record: %v", result.Error)
		s.respondError(c, http.StatusInternalServerError, "Failed to delete config record")
		return
	}

	if result.RowsAffected == 0 {
		s.respondError(c, http.StatusNotFound, "Config record not found")
		return
	}

	s.respondSuccess(c, http.StatusOK, gin.H{
		"data": nil,
	}, "Config record deleted successfully")
}

// Request/Response structures
type CreateConfigRecordRequest struct {
	Key         string      `json:"key" binding:"required"`
	Value       interface{} `json:"value" binding:"required"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	IsActive    bool        `json:"is_active"`
}

type UpdateConfigRecordRequest struct {
	Value       *interface{} `json:"value"`
	Description *string     `json:"description"`
	Category    *string     `json:"category"`
	IsActive    *bool       `json:"is_active"`
}

// ConfigRecordResponse 配置记录响应
type ConfigRecordResponse struct {
	Records    []ConfigRecord `json:"records"`
	Pagination Pagination     `json:"pagination"`
}

// ConfigRecord 配置记录（用于Swagger文档）
type ConfigRecord struct {
	ID          uint                   `json:"id"`
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Version     int                    `json:"version"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type Pagination struct {
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
	Total int64 `json:"total"`
	Pages int64 `json:"pages"`
}

// APIResponse 通用API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}