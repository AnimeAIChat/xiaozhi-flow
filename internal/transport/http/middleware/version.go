package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VersionMiddleware API版本控制中间件
func VersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从URL参数获取版本，如果没有则默认为v1
		version := c.Param("version")
		if version == "" {
			version = "v1"
		}

		// 验证版本是否支持
		supportedVersions := []string{"v1"}
		isSupported := false
		for _, v := range supportedVersions {
			if v == version {
				isSupported = true
				break
			}
		}

		if !isSupported {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNSUPPORTED_API_VERSION",
					"message": fmt.Sprintf("API version %s is not supported", version),
				},
				"supported_versions": supportedVersions,
			})
			c.Abort()
			return
		}

		// 将版本信息存储到上下文中
		c.Set("api_version", version)
		c.Next()
	}
}

// GetAPIVersion 从上下文中获取API版本
func GetAPIVersion(c *gin.Context) string {
	if version, exists := c.Get("api_version"); exists {
		return version.(string)
	}
	return "v1" // 默认版本
}