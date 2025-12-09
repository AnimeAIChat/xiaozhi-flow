package httptransport

import (
	"xiaozhi-server-go/internal/platform/logging"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/observability"
	httpMiddleware "xiaozhi-server-go/internal/transport/http/middleware"
)

// Options configures the HTTP router builder.
type Options struct {
	Config         *config.Config
	Logger         *logging.Logger
	AuthMiddleware gin.HandlerFunc
	StaticRoot     string
}

// Router bundles together the gin engine and common route groups.
type Router struct {
	Engine   *gin.Engine
	API      *gin.RouterGroup
	Secured  *gin.RouterGroup
	V1       *gin.RouterGroup
	V1Secure *gin.RouterGroup
}

// Build constructs a gin engine pre-configured with logging, recovery, CORS and observability middlewares.
func Build(opts Options) (*Router, error) {
	if opts.Config == nil {
		return nil, fmt.Errorf("http router requires config")
	}
	logger := opts.Logger
	if logger == nil {
		logger = logging.DefaultLogger
	}

	if opts.Config.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// 使用新的中间件
	engine.Use(gin.Recovery())
	engine.Use(httpMiddleware.ErrorMiddleware(logger))
	engine.Use(httpMiddleware.ResponseMiddleware())
	engine.Use(httpMiddleware.LoggingMiddleware(logger))
	engine.Use(httpMiddleware.SecurityHeadersMiddleware())
	engine.Use(httpMiddleware.RequestSizeMiddleware(10 << 20)) // 10MB
	engine.Use(httpMiddleware.CORSMiddleware())
	engine.Use(loggingMiddleware(logger)) // 保留原有的日志中间件作为备份
	engine.Use(observabilityMiddleware())

	engine.SetTrustedProxies([]string{"0.0.0.0"})

	// 移除旧的CORS配置，使用新的统一CORS中间件

	api := engine.Group("/api")

	// 创建 V1 API 路由组
	v1 := api.Group("/v1")
	v1.Use(httpMiddleware.VersionMiddleware())

	var v1Secure *gin.RouterGroup
	if opts.AuthMiddleware != nil {
		v1Secure = v1.Group("")
		v1Secure.Use(opts.AuthMiddleware)
	}

	staticRoot := opts.StaticRoot
	if staticRoot == "" {
		// 默认使用 dist 目录
		staticRoot = "./web/dist"
	}

	// 为静态文件创建单独的组，避免与API冲突
	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 如果是API请求，继续处理
		if strings.HasPrefix(path, "/api") {
			c.Next()
			return
		}

		// 静态文件服务
		if _, err := os.Stat(staticRoot + path); err == nil {
			c.File(staticRoot + path)
			return
		}

		// SPA fallback
		if !strings.HasPrefix(path, "/static/") &&
		   !strings.HasPrefix(path, "/assets/") &&
		   path != "/favicon.ico" {
			c.File(staticRoot + "/index.html")
		} else {
			c.Status(404)
		}
	})
	var secured *gin.RouterGroup
	if opts.AuthMiddleware != nil {
		secured = api.Group("")
		secured.Use(opts.AuthMiddleware)
	}

	return &Router{
		Engine:   engine,
		API:      api,
		Secured:  secured,
		V1:       v1,
		V1Secure: v1Secure,
	}, nil
}

func loggingMiddleware(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		status := c.Writer.Status()

		if logger != nil {
			logger.Info(
				"[HTTP] %s %s -> %d (%s)",
				c.Request.Method,
				c.Request.URL.Path,
				status,
				duration,
			)
		}
	}
}

func observabilityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		reqCtx, spanEnd := observability.StartSpan(c.Request.Context(), "http.server", path)
		var spanErr error
		c.Request = c.Request.WithContext(reqCtx)

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		if len(c.Errors) > 0 {
			spanErr = c.Errors.Last().Err
		} else if status := c.Writer.Status(); status >= http.StatusInternalServerError {
			spanErr = fmt.Errorf("status %d", status)
		}
		spanEnd(spanErr)

		observability.RecordMetric(
			reqCtx,
			"http.requests",
			1,
			map[string]string{
				"component": "http.server",
				"method":    c.Request.Method,
				"path":      path,
				"status":    strconv.Itoa(c.Writer.Status()),
			},
		)
		observability.RecordMetric(
			reqCtx,
			"http.request.duration_ms",
			float64(duration.Milliseconds()),
			map[string]string{
				"component": "http.server",
				"method":    c.Request.Method,
				"path":      path,
			},
		)
	}
}



