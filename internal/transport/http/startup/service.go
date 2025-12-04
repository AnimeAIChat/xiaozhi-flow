package startup

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"xiaozhi-server-go/internal/platform/config"
	httptransport "xiaozhi-server-go/internal/transport/http"
	"xiaozhi-server-go/internal/utils"
)

// Service 启动流程HTTP服务
type Service struct {
	config *config.Config
	logger *utils.Logger
}

// NewService 创建启动流程服务
func NewService(config *config.Config, logger *utils.Logger) (*Service, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		logger = utils.DefaultLogger
	}

	return &Service{
		config: config,
		logger: logger,
	}, nil
}

// Register 注册路由
func (s *Service) Register(ctx context.Context, router *gin.RouterGroup) {
	s.logger.Info("注册启动流程API路由")

	// 启动流程工作流相关
	startupGroup := router.Group("/startup")
	{
		// 获取可用的工作流列表
		startupGroup.GET("/workflows", s.getWorkflows)

		// 获取特定工作流详情
		startupGroup.GET("/workflows/:id", s.getWorkflow)
	}

	s.logger.Info("启动流程API路由注册完成")
}

// getWorkflows 获取可用的工作流列表
// @Summary 获取工作流列表
// @Description 获取所有可用的启动工作流列表
// @Tags Workflow
// @Produce json
// @Success 200 {object} httptransport.APIResponse
// @Router /startup/workflows [get]
func (s *Service) getWorkflows(c *gin.Context) {
	s.logger.Info("获取启动工作流列表")

	workflows := []map[string]interface{}{
		{
			"id":          "xiaozhi-flow-default-startup",
			"name":        "XiaoZhi Flow 默认启动工作流",
			"description": "将现有的bootstrap启动步骤转换为可视化工作流",
			"version":     "1.0.0",
			"tags":        []string{"default", "system", "startup"},
			"created_at":  time.Now().Format(time.RFC3339),
			"updated_at":  time.Now().Format(time.RFC3339),
		},
		{
			"id":          "xiaozhi-flow-parallel-startup",
			"name":        "XiaoZhi Flow 并行启动工作流",
			"description": "优化的并行启动工作流，支持并行执行无依赖的步骤",
			"version":     "1.0.0",
			"tags":        []string{"parallel", "optimized", "system", "startup"},
			"created_at":  time.Now().Format(time.RFC3339),
			"updated_at":  time.Now().Format(time.RFC3339),
		},
		{
			"id":          "xiaozhi-flow-minimal-startup",
			"name":        "XiaoZhi Flow 最小启动工作流",
			"description": "仅包含必要组件的最小化启动工作流",
			"version":     "1.0.0",
			"tags":        []string{"minimal", "basic", "startup"},
			"created_at":  time.Now().Format(time.RFC3339),
			"updated_at":  time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, httptransport.APIResponse{
		Success: true,
		Data:    workflows,
		Message: "获取成功",
		Code:    http.StatusOK,
	})
}

// getWorkflow 获取特定工作流详情
// @Summary 获取工作流详情
// @Description 根据ID获取特定工作流的详细信息
// @Tags Workflow
// @Produce json
// @Param id path string true "工作流ID"
// @Success 200 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Router /startup/workflows/{id} [get]
func (s *Service) getWorkflow(c *gin.Context) {
	workflowID := c.Param("id")
	s.logger.Info("获取启动工作流详情", "workflow_id", workflowID)

	// 根据ID获取对应的工作流定义
	var workflow map[string]interface{}

	switch workflowID {
	case "xiaozhi-flow-default-startup":
		workflow = s.getDefaultWorkflow()
	case "xiaozhi-flow-parallel-startup":
		workflow = s.getParallelWorkflow()
	case "xiaozhi-flow-minimal-startup":
		workflow = s.getMinimalWorkflow()
	default:
		c.JSON(http.StatusNotFound, httptransport.APIResponse{
			Success: false,
			Data:    nil,
			Message: fmt.Sprintf("工作流 %s 不存在", workflowID),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, httptransport.APIResponse{
		Success: true,
		Data:    workflow,
		Message: "获取成功",
		Code:    http.StatusOK,
	})
}



// 辅助方法

func (s *Service) getWorkflowName(workflowID string) string {
	names := map[string]string{
		"xiaozhi-flow-default-startup":   "XiaoZhi Flow 默认启动工作流",
		"xiaozhi-flow-parallel-startup": "XiaoZhi Flow 并行启动工作流",
		"xiaozhi-flow-minimal-startup":  "XiaoZhi Flow 最小启动工作流",
	}

	if name, ok := names[workflowID]; ok {
		return name
	}
	return "未知工作流"
}

func (s *Service) getWorkflowNodes(workflowID string) []map[string]interface{} {
	// 返回完整的11个节点数据
	return []map[string]interface{}{
		{
			"id":          "storage:init-config-store",
			"name":        "初始化配置存储",
			"type":        "storage",
			"description": "初始化配置存储系统，用于持久化应用配置",
			"status":      "completed",
			"position":    map[string]float64{"x": 100, "y": 100},
			"start_time":  time.Now().Add(-5*time.Minute).Format(time.RFC3339),
			"end_time":    time.Now().Add(-4*time.Minute).Format(time.RFC3339),
			"duration":    60000000000,
			"critical":    true,
			"optional":    false,
			"depends_on":  []string{},
		},
		{
			"id":          "storage:init-database",
			"name":        "初始化数据库",
			"type":        "storage",
			"description": "初始化数据库连接，支持SQLite、MySQL、PostgreSQL",
			"status":      "completed",
			"position":    map[string]float64{"x": 300, "y": 100},
			"start_time":  time.Now().Add(-4*time.Minute).Format(time.RFC3339),
			"end_time":    time.Now().Add(-3*time.Minute).Format(time.RFC3339),
			"duration":    60000000000,
			"critical":    true,
			"optional":    false,
			"depends_on":  []string{"storage:init-config-store"},
		},
		{
			"id":          "config:load-default",
			"name":        "加载默认配置",
			"type":        "config",
			"description": "从数据库加载默认配置，如果没有则使用内置默认值",
			"status":      "completed",
			"position":    map[string]float64{"x": 500, "y": 100},
			"start_time":  time.Now().Add(-3*time.Minute).Format(time.RFC3339),
			"end_time":    time.Now().Add(-2*time.Minute).Format(time.RFC3339),
			"duration":    60000000000,
			"critical":    true,
			"optional":    false,
			"depends_on":  []string{"storage:init-config-store", "storage:init-database"},
		},
		{
			"id":          "logging:init-provider",
			"name":        "初始化日志系统",
			"type":        "service",
			"description": "初始化日志提供者，设置日志级别和输出格式",
			"status":      "completed",
			"position":    map[string]float64{"x": 700, "y": 100},
			"start_time":  time.Now().Add(-2*time.Minute).Format(time.RFC3339),
			"end_time":    time.Now().Add(-1*time.Minute).Format(time.RFC3339),
			"duration":    60000000000,
			"critical":    false,
			"optional":    false,
			"depends_on":  []string{"config:load-default"},
		},
		{
			"id":          "components:init-container",
			"name":        "初始化组件容器",
			"type":        "service",
			"description": "初始化组件容器，管理所有依赖注入",
			"status":      "completed",
			"position":    map[string]float64{"x": 900, "y": 100},
			"start_time":  time.Now().Add(-1*time.Minute).Format(time.RFC3339),
			"end_time":    time.Now().Add(-30*time.Second).Format(time.RFC3339),
			"duration":    30000000000,
			"critical":    false,
			"optional":    false,
			"depends_on":  []string{"logging:init-provider"},
		},
		{
			"id":          "config:init-integrator",
			"name":        "初始化配置集成器",
			"type":        "service",
			"description": "初始化配置集成器，统一配置管理",
			"status":      "completed",
			"position":    map[string]float64{"x": 1100, "y": 100},
			"start_time":  time.Now().Add(-30*time.Second).Format(time.RFC3339),
			"end_time":    time.Now().Add(-20*time.Second).Format(time.RFC3339),
			"duration":    10000000000,
			"critical":    false,
			"optional":    false,
			"depends_on":  []string{"components:init-container", "logging:init-provider"},
		},
		{
			"id":          "auth:init-manager",
			"name":        "初始化认证管理器",
			"type":        "auth",
			"description": "初始化认证管理器，设置用户认证和会话管理",
			"status":      "completed",
			"position":    map[string]float64{"x": 1300, "y": 100},
			"start_time":  time.Now().Add(-20*time.Second).Format(time.RFC3339),
			"end_time":    time.Now().Add(-10*time.Second).Format(time.RFC3339),
			"duration":    10000000000,
			"critical":    true,
			"optional":    false,
			"depends_on":  []string{"components:init-container"},
		},
		{
			"id":          "mcp:init-manager",
			"name":        "初始化MCP管理器",
			"type":        "service",
			"description": "初始化MCP（Model Context Protocol）管理器",
			"status":      "completed",
			"position":    map[string]float64{"x": 1500, "y": 100},
			"start_time":  time.Now().Add(-10*time.Second).Format(time.RFC3339),
			"end_time":    time.Now().Add(-8*time.Second).Format(time.RFC3339),
			"duration":    2000000000,
			"critical":    false,
			"optional":    true,
			"depends_on":  []string{"logging:init-provider"},
		},
		{
			"id":          "observability:setup-hooks",
			"name":        "设置可观测性钩子",
			"type":        "service",
			"description": "设置可观测性钩子，包括指标收集和分布式追踪",
			"status":      "completed",
			"position":    map[string]float64{"x": 1700, "y": 100},
			"start_time":  time.Now().Add(-8*time.Second).Format(time.RFC3339),
			"end_time":    time.Now().Add(-6*time.Second).Format(time.RFC3339),
			"duration":    2000000000,
			"critical":    false,
			"optional":    true,
			"depends_on":  []string{"logging:init-provider"},
		},
		{
			"id":          "plugin:init-manager",
			"name":        "初始化插件管理器",
			"type":        "plugin",
			"description": "初始化插件管理器，设置插件发现和管理功能",
			"status":      "completed",
			"position":    map[string]float64{"x": 1900, "y": 100},
			"start_time":  time.Now().Add(-6*time.Second).Format(time.RFC3339),
			"end_time":    time.Now().Add(-4*time.Second).Format(time.RFC3339),
			"duration":    2000000000,
			"critical":    false,
			"optional":    true,
			"depends_on":  []string{"logging:init-provider"},
		},
		{
			"id":          "start-services",
			"name":        "启动系统服务",
			"type":        "service",
			"description": "启动HTTP、WebSocket、OTA等系统服务",
			"status":      "completed",
			"position":    map[string]float64{"x": 2100, "y": 100},
			"start_time":  time.Now().Add(-4*time.Second).Format(time.RFC3339),
			"end_time":    time.Now().Add(-2*time.Second).Format(time.RFC3339),
			"duration":    2000000000,
			"critical":    true,
			"optional":    false,
			"depends_on":  []string{"storage:init-database", "config:load-default", "auth:init-manager"},
		},
	}
}

func (s *Service) getDefaultWorkflow() map[string]interface{} {
	// 返回默认工作流的完整定义
	return map[string]interface{}{
		"id":          "xiaozhi-flow-default-startup",
		"name":        "XiaoZhi Flow 默认启动工作流",
		"description": "将现有的bootstrap启动步骤转换为可视化工作流",
		"version":     "1.0.0",
		"created_at":  time.Now().Format(time.RFC3339),
		"updated_at":  time.Now().Format(time.RFC3339),
		"tags":        []string{"default", "system", "startup"},
		"config": map[string]interface{}{
			"timeout":       600000000000, // 10分钟
			"max_retries":    3,
			"parallel_limit": 4,
			"enable_log":     true,
			"environment":    map[string]interface{}{},
			"variables":      map[string]interface{}{},
			"on_failure":     "stop",
		},
		"nodes": s.getWorkflowNodes("xiaozhi-flow-default-startup"),
		"edges": []map[string]interface{}{
			{"id": "e1", "from": "storage:init-config-store", "to": "storage:init-database", "label": "依赖"},
			{"id": "e2", "from": "storage:init-database", "to": "config:load-default", "label": "依赖"},
			{"id": "e3", "from": "config:load-default", "to": "logging:init-provider", "label": "依赖"},
			{"id": "e4", "from": "logging:init-provider", "to": "components:init-container", "label": "依赖"},
			{"id": "e5", "from": "components:init-container", "to": "config:init-integrator", "label": "依赖"},
			{"id": "e6", "from": "components:init-container", "to": "auth:init-manager", "label": "依赖"},
			{"id": "e7", "from": "logging:init-provider", "to": "mcp:init-manager", "label": "依赖"},
			{"id": "e8", "from": "logging:init-provider", "to": "observability:setup-hooks", "label": "依赖"},
			{"id": "e9", "from": "logging:init-provider", "to": "plugin:init-manager", "label": "依赖"},
			{"id": "e10", "from": "config:load-default", "to": "start-services", "label": "依赖"},
			{"id": "e11", "from": "auth:init-manager", "to": "start-services", "label": "依赖"},
		},
	}
}

func (s *Service) getParallelWorkflow() map[string]interface{} {
	// 返回并行工作流定义
	workflow := s.getDefaultWorkflow()
	workflow["id"] = "xiaozhi-flow-parallel-startup"
	workflow["name"] = "XiaoZhi Flow 并行启动工作流"
	workflow["description"] = "优化的并行启动工作流，支持并行执行无依赖的步骤"
	workflow["tags"] = []string{"parallel", "optimized", "system", "startup"}

	if config, ok := workflow["config"].(map[string]interface{}); ok {
		config["parallel_limit"] = 8
	}

	return workflow
}

func (s *Service) getMinimalWorkflow() map[string]interface{} {
	// 返回最小工作流定义
	workflow := s.getDefaultWorkflow()
	workflow["id"] = "xiaozhi-flow-minimal-startup"
	workflow["name"] = "XiaoZhi Flow 最小启动工作流"
	workflow["description"] = "仅包含必要组件的最小化启动工作流"
	workflow["tags"] = []string{"minimal", "basic", "startup"}

	if config, ok := workflow["config"].(map[string]interface{}); ok {
		config["timeout"] = 300000000000 // 5分钟
		config["max_retries"] = 1
		config["parallel_limit"] = 2
	}

	return workflow
}