package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIError 标准API错误结构
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// APIResponse 标准API响应结构
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Version   string      `json:"version,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// generateRequestID 生成唯一的请求ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "req_" + hex.EncodeToString(bytes)
}

// ResponseMiddleware 统一响应格式中间件
func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID并设置到响应头
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		// 继续处理请求
		c.Next()

		// 如果响应已经被处理，则不再修改
		if c.Writer.Written() {
			return
		}
	}
}

// SuccessResponse 返回成功响应
func SuccessResponse(c *gin.Context, data interface{}, message string) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().Unix(),
		Version:   "v1", // 固定为 v1 版本
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusOK, response)
}

// ErrorResponse 返回错误响应
func ErrorResponse(c *gin.Context, errorCode, errorMessage string, details ...interface{}) {
	var detailsData interface{}
	if len(details) > 0 {
		detailsData = details[0]
	}

	response := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    errorCode,
			Message: errorMessage,
			Details: detailsData,
		},
		Timestamp: time.Now().Unix(),
		Version:   "v1", // 固定为 v1 版本
		RequestID: getRequestID(c),
	}

	// 根据错误码确定HTTP状态码
	statusCode := getStatusCodeFromErrorCode(errorCode)
	c.JSON(statusCode, response)
}

// ValidationError 返回验证错误响应
func ValidationError(c *gin.Context, err error) {
	ErrorResponse(c, "VALIDATION_FAILED", "请求参数验证失败", err.Error())
}

// NotFoundError 返回资源不存在错误
func NotFoundError(c *gin.Context, resource string) {
	ErrorResponse(c, "RESOURCE_NOT_FOUND", fmt.Sprintf("%s不存在", resource))
}

// UnauthorizedError 返回未授权错误
func UnauthorizedError(c *gin.Context, message string) {
	if message == "" {
		message = "未授权访问"
	}
	ErrorResponse(c, "UNAUTHORIZED", message)
}

// ForbiddenError 返回禁止访问错误
func ForbiddenError(c *gin.Context, message string) {
	if message == "" {
		message = "禁止访问"
	}
	ErrorResponse(c, "FORBIDDEN", message)
}

// InternalServerError 返回内部服务器错误
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "内部服务器错误"
	}
	ErrorResponse(c, "INTERNAL_SERVER_ERROR", message)
}

// getRequestID 从上下文中获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// getStatusCodeFromErrorCode 根据错误码获取对应的HTTP状态码
func getStatusCodeFromErrorCode(errorCode string) int {
	switch errorCode {
	case "VALIDATION_FAILED":
		return http.StatusBadRequest
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "RESOURCE_NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT":
		return http.StatusConflict
	case "UNSUPPORTED_API_VERSION":
		return http.StatusBadRequest
	case "WORKFLOW_NOT_FOUND", "EXECUTION_NOT_FOUND", "DEVICE_NOT_FOUND":
		return http.StatusNotFound
	case "WORKFLOW_EXECUTION_ERROR", "VISION_PROCESSING_FAILED":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}