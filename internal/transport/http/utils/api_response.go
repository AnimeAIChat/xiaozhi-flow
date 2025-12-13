package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	httpMiddleware "xiaozhi-server-go/internal/transport/http/middleware"
)

// ResponseHelper API响应助手
type ResponseHelper struct{}

// NewResponseHelper 创建响应助手
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

// Success 成功响应
func (r *ResponseHelper) Success(c *gin.Context, data interface{}, message string) {
	httpMiddleware.SuccessResponse(c, data, message)
}

// Error 错误响应
func (r *ResponseHelper) Error(c *gin.Context, errorCode, errorMessage string, details ...interface{}) {
	httpMiddleware.ErrorResponse(c, errorCode, errorMessage, details...)
}

// ValidationError 验证错误
func (r *ResponseHelper) ValidationError(c *gin.Context, err error) {
	httpMiddleware.ValidationError(c, err)
}

// NotFound 资源不存在
func (r *ResponseHelper) NotFound(c *gin.Context, resource string) {
	httpMiddleware.NotFoundError(c, resource)
}

// Unauthorized 未授权
func (r *ResponseHelper) Unauthorized(c *gin.Context, message string) {
	httpMiddleware.UnauthorizedError(c, message)
}

// Forbidden 禁止访问
func (r *ResponseHelper) Forbidden(c *gin.Context, message string) {
	httpMiddleware.ForbiddenError(c, message)
}

// InternalError 内部服务器错误
func (r *ResponseHelper) InternalError(c *gin.Context, message string) {
	httpMiddleware.InternalServerError(c, message)
}

// BadRequest 请求错误
func (r *ResponseHelper) BadRequest(c *gin.Context, message string) {
	r.Error(c, ErrorCodeBadRequest, message)
}

// Conflict 资源冲突
func (r *ResponseHelper) Conflict(c *gin.Context, message string) {
	r.Error(c, ErrorCodeConflict, message)
}

// Created 资源创建成功
func (r *ResponseHelper) Created(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, httpMiddleware.APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: httpMiddleware.APIResponse{}.Timestamp, // 将在中间件中设置
		Version:   "v1", // 固定为 v1 版本
		RequestID: getRequestIDFromContext(c),
	})
}

// NoContent 无内容响应
func (r *ResponseHelper) NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// NotModified 未修改响应
func (r *ResponseHelper) NotModified(c *gin.Context) {
	c.Status(http.StatusNotModified)
}

// 响应状态码助手函数

// OK 200 OK
func (r *ResponseHelper) OK(c *gin.Context, data interface{}, message string) {
	r.Success(c, data, message)
}

// Accepted 202 Accepted
func (r *ResponseHelper) Accepted(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusAccepted, httpMiddleware.APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: httpMiddleware.APIResponse{}.Timestamp,
		Version:   "v1", // 固定为 v1 版本
		RequestID: getRequestIDFromContext(c),
	})
}

// MovedPermanently 301 Moved Permanently
func (r *ResponseHelper) MovedPermanently(c *gin.Context, location string) {
	c.Header("Location", location)
	c.Status(http.StatusMovedPermanently)
}

// MovedTemporarily 302 Moved Temporarily
func (r *ResponseHelper) MovedTemporarily(c *gin.Context, location string) {
	c.Header("Location", location)
	c.Status(http.StatusFound)
}

// 临时响应函数，等待中间件更新后替换
func getRequestIDFromContext(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// 全局响应助手实例
var Response = NewResponseHelper()