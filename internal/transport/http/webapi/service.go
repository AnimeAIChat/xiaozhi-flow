package webapi

import (
	"xiaozhi-server-go/internal/platform/logging"
	"context"
	"net/http"
	"time"

	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/errors"

	"github.com/gin-gonic/gin"
)

// Service WebAPI服务的HTTP传输层实现
type Service struct {
	logger   *logging.Logger
	config   *config.Config
	startTime time.Time
}

// NewService 创建新的WebAPI服务实例
func NewService(config *config.Config, logger *logging.Logger) (*Service, error) {
	if config == nil {
		return nil, errors.Wrap(errors.KindConfig, "webapi.new", "config is required", nil)
	}
	if logger == nil {
		return nil, errors.Wrap(errors.KindConfig, "webapi.new", "logger is required", nil)
	}

	service := &Service{
		logger:   logger,
		config:   config,
		startTime: time.Now(),
	}

	return service, nil
}

// Register 注册WebAPI相关的HTTP路由
func (s *Service) Register(ctx context.Context, router *gin.RouterGroup) error {
	// 管理员路由
	s.registerAdminRoutes(router)

	s.logger.InfoTag("HTTP", "WebAPI服务路由注册完成")
	return nil
}

// registerAdminRoutes 注册管理员相关路由
func (s *Service) registerAdminRoutes(router *gin.RouterGroup) {
	adminGroup := router.Group("/admin")
	adminGroup.GET("", s.handleAdminGet)
}

// handleOptions 处理OPTIONS请求
func (s *Service) handleOptions(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, AuthorToken")
	c.Status(http.StatusNoContent)
}

// handleAdminGet 处理管理员服务状态检查
func (s *Service) handleAdminGet(c *gin.Context) {
	s.logger.InfoTag("HTTP", "Admin GET called: %s", c.Request.URL.Path)
	s.respondSuccess(c, http.StatusOK, nil, "Admin service is running")
}

// ServerConfig 服务器配置结构
type ServerConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol,omitempty"`
}

// ConnectionTestResult 连接测试结果
type ConnectionTestResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Latency  int64  `json:"latency,omitempty"`
	Version  string `json:"version,omitempty"`
	Services map[string]bool `json:"services,omitempty"`
}



// AuthMiddleware 认证中间件（公开方法）
func (s *Service) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apikey := c.GetHeader("AuthorToken")
		if apikey != "" {
			// 如果提供了API Token，直接验证
			if apikey != s.config.Server.Token {
				s.logger.Error("无效的API Token %s", apikey)
				s.respondError(c, http.StatusUnauthorized, "无效的API Token")
				c.Abort()
				return
			}
			s.logger.Info("API Token验证通过")
			c.Next()
			return
		}
		c.Next()
	}
}




// respondSuccess 返回成功响应
func (s *Service) respondSuccess(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    data,
		"message": message,
		"code":    statusCode,
	})
}

// respondError 返回错误响应
func (s *Service) respondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"data":    gin.H{"error": message},
		"message": message,
		"code":    statusCode,
	})
}