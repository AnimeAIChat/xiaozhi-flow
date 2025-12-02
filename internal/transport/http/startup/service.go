package startup

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/transport/http"
	"xiaozhi-server-go/internal/transport/ws"
	"xiaozhi-server-go/internal/utils"
)

// Service 启动流程HTTP服务
type Service struct {
	config *config.Config
	logger *utils.Logger
	wsHandler *ws.StartupWebSocketHandler
}

// NewService 创建启动流程服务
func NewService(config *config.Config, logger *utils.Logger) (*Service, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		logger = utils.DefaultLogger
	}

	// 创建启动流程WebSocket处理器
	wsHandler := ws.NewStartupWebSocketHandler(nil, logger)

	return &Service{
		config: config,
		logger: logger,
		wsHandler: wsHandler,
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

		// 执行工作流
		startupGroup.POST("/workflows/execute", s.executeWorkflow)
	}

	// 执行相关
	executionGroup := router.Group("/startup/executions")
	{
		// 获取执行状态
		executionGroup.GET("/:id", s.getExecutionStatus)

		// 取消执行
		executionGroup.DELETE("/:id", s.cancelExecution)

		// 暂停执行
		executionGroup.POST("/:id/pause", s.pauseExecution)

		// 恢复执行
		executionGroup.POST("/:id/resume", s.resumeExecution)

		// 获取执行历史
		executionGroup.GET("", s.getExecutionHistory)
	}

	// WebSocket端点
	router.GET("/startup/ws", s.handleWebSocket)

	s.logger.Info("启动流程API路由注册完成")
}

// getWorkflows 获取可用的工作流列表
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

// executeWorkflow 执行工作流
func (s *Service) executeWorkflow(c *gin.Context) {
	var request struct {
		WorkflowID string                 `json:"workflow_id" binding:"required"`
		Inputs     map[string]interface{} `json:"inputs"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, httptransport.APIResponse{
			Success: false,
			Data:    nil,
			Message: "参数错误: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	s.logger.Info("执行启动工作流",
		"workflow_id", request.WorkflowID,
		"inputs", request.Inputs)

	// 创建模拟的执行对象
	execution := map[string]interface{}{
		"id":              fmt.Sprintf("exec_%d", time.Now().UnixNano()),
		"workflow_id":     request.WorkflowID,
		"workflow_name":   s.getWorkflowName(request.WorkflowID),
		"status":          "running",
		"start_time":      time.Now().Format(time.RFC3339),
		"progress":        0.0,
		"total_nodes":     11,
		"completed_nodes": 0,
		"failed_nodes":    0,
		"current_nodes":   []string{},
		"context":         request.Inputs,
		"nodes":           s.getWorkflowNodes(request.WorkflowID),
	}

	c.JSON(http.StatusOK, httptransport.APIResponse{
		Success: true,
		Data:    execution,
		Message: "执行已开始",
		Code:    http.StatusOK,
	})
}

// getExecutionStatus 获取执行状态
func (s *Service) getExecutionStatus(c *gin.Context) {
	executionID := c.Param("id")
	s.logger.Info("获取执行状态", "execution_id", executionID)

	// 模拟执行状态数据
	execution := map[string]interface{}{
		"id":              executionID,
		"workflow_id":     "xiaozhi-flow-default-startup",
		"workflow_name":   "XiaoZhi Flow 默认启动工作流",
		"status":          "completed",
		"start_time":      time.Now().Add(-5*time.Minute).Format(time.RFC3339),
		"end_time":        time.Now().Add(-1*time.Minute).Format(time.RFC3339),
		"duration":        240000000000, // 4分钟
		"progress":        1.0,
		"total_nodes":     11,
		"completed_nodes": 11,
		"failed_nodes":    0,
		"current_nodes":   []string{},
		"nodes":           s.getWorkflowNodes("xiaozhi-flow-default-startup"),
	}

	c.JSON(http.StatusOK, httptransport.APIResponse{
		Success: true,
		Data:    execution,
		Message: "获取成功",
		Code:    http.StatusOK,
	})
}

// cancelExecution 取消执行
func (s *Service) cancelExecution(c *gin.Context) {
	executionID := c.Param("id")
	s.logger.Info("取消执行", "execution_id", executionID)

	c.JSON(http.StatusOK, httptransport.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"execution_id": executionID, "action": "cancelled"},
		Message: "执行已取消",
		Code:    http.StatusOK,
	})
}

// pauseExecution 暂停执行
func (s *Service) pauseExecution(c *gin.Context) {
	executionID := c.Param("id")
	s.logger.Info("暂停执行", "execution_id", executionID)

	c.JSON(http.StatusOK, httptransport.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"execution_id": executionID, "action": "paused"},
		Message: "执行已暂停",
		Code:    http.StatusOK,
	})
}

// resumeExecution 恢复执行
func (s *Service) resumeExecution(c *gin.Context) {
	executionID := c.Param("id")
	s.logger.Info("恢复执行", "execution_id", executionID)

	c.JSON(http.StatusOK, httptransport.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"execution_id": executionID, "action": "resumed"},
		Message: "执行已恢复",
		Code:    http.StatusOK,
	})
}

// getExecutionHistory 获取执行历史
func (s *Service) getExecutionHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	s.logger.Info("获取执行历史", "limit", limit)

	// 模拟执行历史数据
	executions := []map[string]interface{}{
		{
			"id":              "exec_001",
			"workflow_id":     "xiaozhi-flow-default-startup",
			"workflow_name":   "XiaoZhi Flow 默认启动工作流",
			"status":          "completed",
			"start_time":      time.Now().Add(-30*time.Minute).Format(time.RFC3339),
			"end_time":        time.Now().Add(-25*time.Minute).Format(time.RFC3339),
			"duration":        300000000000, // 5分钟
			"progress":        1.0,
			"completed_nodes": 11,
			"failed_nodes":    0,
		},
		{
			"id":              "exec_002",
			"workflow_id":     "xiaozhi-flow-parallel-startup",
			"workflow_name":   "XiaoZhi Flow 并行启动工作流",
			"status":          "failed",
			"start_time":      time.Now().Add(-60*time.Minute).Format(time.RFC3339),
			"end_time":        time.Now().Add(-58*time.Minute).Format(time.RFC3339),
			"duration":        120000000000, // 2分钟
			"progress":        0.8,
			"completed_nodes": 8,
			"failed_nodes":    2,
		},
	}

	// 限制返回数量
	if limit > 0 && len(executions) > limit {
		executions = executions[:limit]
	}

	c.JSON(http.StatusOK, httptransport.APIResponse{
		Success: true,
		Data:    executions,
		Message: "获取成功",
		Code:    http.StatusOK,
	})
}

// handleWebSocket 处理WebSocket连接
func (s *Service) handleWebSocket(c *gin.Context) {
	s.logger.Info("处理启动流程WebSocket连接")
	s.wsHandler.HandleWebSocket(c.Writer, c.Request)
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