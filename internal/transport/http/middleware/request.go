package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"xiaozhi-server-go/internal/utils"
)

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware(logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 记录请求信息
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		// 读取请求体（用于日志记录）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，以便后续处理器可以读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 记录请求开始
		logger.InfoTag("HTTP", "请求开始",
			"method", c.Request.Method,
			"path", path,
			"remote_addr", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"content_type", c.GetHeader("Content-Type"),
			"request_id", getRequestID(c),
		)

		// 如果有请求体且不是敏感信息，则记录（限制大小）
		if len(requestBody) > 0 && len(requestBody) < 1024 {
			contentType := c.GetHeader("Content-Type")
			// 只记录非敏感内容类型的请求体
			if !isSensitiveContentType(contentType) {
				logger.DebugTag("HTTP", "请求体",
					"body", string(requestBody),
					"request_id", getRequestID(c),
				)
			}
		}

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 记录响应信息
		logger.InfoTag("HTTP", "请求完成",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"response_size", c.Writer.Size(),
			"request_id", getRequestID(c),
		)

		// 如果是错误响应，记录详细信息
		if c.Writer.Status() >= 400 {
			logger.WarnTag("HTTP", "请求返回错误状态",
				"status", c.Writer.Status(),
				"method", c.Request.Method,
				"path", path,
				"request_id", getRequestID(c),
			)
		}
	}
}

// RequestSizeMiddleware 请求大小限制中间件
func RequestSizeMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			ErrorResponse(c, "REQUEST_TOO_LARGE",
				"请求体过大",
				map[string]interface{}{
					"max_size": maxSize,
					"actual_size": c.Request.ContentLength,
				})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SecurityHeadersMiddleware 安全头部中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全相关的HTTP头部
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 在生产环境中设置HSTS
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// HealthCheckMiddleware 健康检查中间件（跳过日志记录）
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/ping" {
			c.Next()
			return
		}
		c.Next()
	}
}

// isSensitiveContentType 判断是否为敏感内容类型
func isSensitiveContentType(contentType string) bool {
	sensitiveTypes := []string{
		"multipart/form-data",
		"application/x-www-form-urlencoded",
	}

	for _, sensitiveType := range sensitiveTypes {
		if contentType == sensitiveType {
			return true
		}
	}
	return false
}