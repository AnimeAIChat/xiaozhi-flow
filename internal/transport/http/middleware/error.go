package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"xiaozhi-server-go/internal/utils"
)

// ErrorMiddleware 错误处理中间件
func ErrorMiddleware(logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误堆栈信息
				logger.ErrorTag("PANIC", "Recovering from panic",
					"error", err,
					"stack", string(debug.Stack()),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"request_id", getRequestID(c),
				)

				// 返回内部服务器错误
				ErrorResponse(c, "INTERNAL_SERVER_ERROR", "服务器内部错误")
				c.Abort()
			}
		}()

		// 继续处理请求
		c.Next()

		// 检查是否有错误发生
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last()
			handleError(c, err, logger)
		}
	}
}

// handleError 处理具体错误
func handleError(c *gin.Context, ginErr *gin.Error, logger *utils.Logger) {
	err := ginErr.Err

	// 记录错误日志
	logger.ErrorTag("API_ERROR", "Request processing error",
		"error", err.Error(),
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"request_id", getRequestID(c),
	)

	// 如果响应已经被写入，则不再处理错误
	if c.Writer.Written() {
		return
	}

	// 根据错误类型返回相应的响应
	if isValidationError(err) {
		ValidationError(c, err)
	} else if isNotFoundError(err) {
		NotFoundError(c, "资源")
	} else if isUnauthorizedError(err) {
		UnauthorizedError(c, err.Error())
	} else if isForbiddenError(err) {
		ForbiddenError(c, err.Error())
	} else {
		InternalServerError(c, err.Error())
	}
}

// isValidationError 判断是否为验证错误
func isValidationError(err error) bool {
	errorMsg := strings.ToLower(err.Error())
	validationKeywords := []string{
		"invalid", "required", "missing", "format", "parse",
		"bind", "validation", "参数", "格式", "验证",
	}

	for _, keyword := range validationKeywords {
		if strings.Contains(errorMsg, keyword) {
			return true
		}
	}
	return false
}

// isNotFoundError 判断是否为资源不存在错误
func isNotFoundError(err error) bool {
	errorMsg := strings.ToLower(err.Error())
	notFoundKeywords := []string{
		"not found", "not exist", "找不到", "不存在",
		"record not found", "no such", "nil",
	}

	for _, keyword := range notFoundKeywords {
		if strings.Contains(errorMsg, keyword) {
			return true
		}
	}
	return false
}

// isUnauthorizedError 判断是否为未授权错误
func isUnauthorizedError(err error) bool {
	errorMsg := strings.ToLower(err.Error())
	unauthorizedKeywords := []string{
		"unauthorized", "unauthenticated", "token", "jwt",
		"未授权", "未认证", "认证失败", "登录",
	}

	for _, keyword := range unauthorizedKeywords {
		if strings.Contains(errorMsg, keyword) {
			return true
		}
	}
	return false
}

// isForbiddenError 判断是否为禁止访问错误
func isForbiddenError(err error) bool {
	errorMsg := strings.ToLower(err.Error())
	forbiddenKeywords := []string{
		"forbidden", "permission", "access denied",
		"禁止", "权限", "拒绝访问",
	}

	for _, keyword := range forbiddenKeywords {
		if strings.Contains(errorMsg, keyword) {
			return true
		}
	}
	return false
}

// CORSMiddleware CORS处理中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, device-id, client-id")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}