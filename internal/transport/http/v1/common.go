package v1

import "github.com/gin-gonic/gin"

// getRequestID 获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	// 如果上下文中没有，尝试从Header获取
	return c.GetHeader("X-Request-ID")
}
