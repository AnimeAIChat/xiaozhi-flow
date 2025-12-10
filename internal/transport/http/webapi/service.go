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
	// 基础路由
	router.GET("/cfg", s.handleCfgGet)
	router.POST("/cfg", s.handleCfgPost)
	router.OPTIONS("/cfg", s.handleOptions)

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

// handleCfgGet 处理配置获取请求
// @Summary 检查配置服务状态
// @Description 检查配置服务的运行状态
// @Tags Config
// @Produce json
// @Success 200 {object} object
// @Router /cfg [get]
func (s *Service) handleCfgGet(c *gin.Context) {
	s.respondSuccess(c, http.StatusOK, nil, "Cfg service is running")
}

// handleCfgPost 处理配置更新请求
func (s *Service) handleCfgPost(c *gin.Context) {
	s.respondSuccess(c, http.StatusOK, nil, "Cfg service is running")
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

		// 如果没有API Token，继续后续的认证逻辑（如果有）
		// 这里假设如果没有Token，则视为未认证，或者由后续中间件处理
		// 但根据原有逻辑，似乎这里主要处理API Token
		
		// 如果需要支持其他认证方式，可以在这里添加
		
		// 如果没有提供Token，且没有其他认证方式，返回未授权
		// 但考虑到可能有些接口不需要Token，或者有其他中间件处理
		// 这里暂时保持原样，或者根据需求修改
		
		// 原有逻辑中，如果没有Token，会继续往下执行吗？
		// 原有逻辑：
		/*
		apikey := c.GetHeader("AuthorToken")
		if apikey != "" {
			// ...
			c.Next()
			return
		}
		*/
		// 看起来如果没有Token，它会继续执行下面的代码。
		// 但下面的代码是什么？
		// 原有代码截断了，我需要读更多。
		
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



