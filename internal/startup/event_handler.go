package startup

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WebSocketEventHandler WebSocket事件处理器
type WebSocketEventHandler struct {
	broadcaster WebSocketEventBroadcaster
	logger      StartupLogger
}

// NewWebSocketEventHandler 创建WebSocket事件处理器
func NewWebSocketEventHandler(broadcaster WebSocketEventBroadcaster, logger StartupLogger) *WebSocketEventHandler {
	return &WebSocketEventHandler{
		broadcaster: broadcaster,
		logger:      logger,
	}
}

// OnExecutionStart 执行开始事件
func (h *WebSocketEventHandler) OnExecutionStart(ctx context.Context, execution *StartupWorkflowExecution) error {
	h.logger.Info("Workflow execution started", "execution_id", execution.ID, "workflow_id", execution.Workflow.ID)

	if h.broadcaster != nil {
		h.broadcaster.BroadcastExecutionEvent("execution_start", execution, map[string]interface{}{
			"action":          "started",
			"start_time":      execution.StartTime,
			"workflow_name":   execution.Workflow.Name,
			"workflow_version": execution.Workflow.Version,
		})
	}

	return nil
}

// OnExecutionEnd 执行结束事件
func (h *WebSocketEventHandler) OnExecutionEnd(ctx context.Context, execution *StartupWorkflowExecution) error {
	h.logger.Info("Workflow execution ended", "execution_id", execution.ID, "status", execution.Status, "duration", execution.Duration)

	if h.broadcaster != nil {
		eventData := map[string]interface{}{
			"action":     "completed",
			"end_time":   execution.EndTime,
			"duration":   execution.Duration.String(),
			"error":      execution.Error,
		}

		if execution.Status == WorkflowStatusFailed {
			eventData["failure_reason"] = execution.Error
		} else if execution.Status == WorkflowStatusCompleted {
			eventData["success"] = true
		}

		h.broadcaster.BroadcastExecutionEvent("execution_end", execution, eventData)
	}

	return nil
}

// OnNodeStart 节点开始事件
func (h *WebSocketEventHandler) OnNodeStart(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode) error {
	h.logger.Info("Node execution started", "execution_id", execution.ID, "node_id", node.ID, "node_name", node.Name)

	if h.broadcaster != nil {
		h.broadcaster.BroadcastNodeEvent("node_start", execution, node, map[string]interface{}{
			"action":      "started",
			"node_type":   string(node.Type),
			"description": node.Description,
			"timeout":     node.Timeout.String(),
			"critical":    node.Critical,
		})
	}

	return nil
}

// OnNodeProgress 节点进度事件
func (h *WebSocketEventHandler) OnNodeProgress(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, progress float64) error {
	h.logger.Debug("Node execution progress", "execution_id", execution.ID, "node_id", node.ID, "progress", progress)

	if h.broadcaster != nil {
		h.broadcaster.BroadcastNodeEvent("node_progress", execution, node, map[string]interface{}{
			"action":   "progress",
			"progress": progress,
		})
	}

	return nil
}

// OnNodeComplete 节点完成事件
func (h *WebSocketEventHandler) OnNodeComplete(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, result *StartupNodeResult) error {
	h.logger.Info("Node execution completed", "execution_id", execution.ID, "node_id", node.ID, "duration", result.Duration)

	if h.broadcaster != nil {
		h.broadcaster.BroadcastNodeEvent("node_complete", execution, node, map[string]interface{}{
			"action":     "completed",
			"duration":   result.Duration.String(),
			"outputs":    result.Outputs,
			"logs_count": len(result.Logs),
			"retry_count": result.RetryCount,
		})
	}

	return nil
}

// OnNodeError 节点错误事件
func (h *WebSocketEventHandler) OnNodeError(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, err error) error {
	h.logger.Error("Node execution failed", "execution_id", execution.ID, "node_id", node.ID, "error", err)

	if h.broadcaster != nil {
		h.broadcaster.BroadcastNodeEvent("node_error", execution, node, map[string]interface{}{
			"action":     "failed",
			"error":      err.Error(),
			"error_type": "execution_error",
		})
	}

	return nil
}

// OnNodeRetry 节点重试事件
func (h *WebSocketEventHandler) OnNodeRetry(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode) error {
	h.logger.Info("Node execution retry", "execution_id", execution.ID, "node_id", node.ID)

	if h.broadcaster != nil {
		h.broadcaster.BroadcastNodeEvent("node_retry", execution, node, map[string]interface{}{
			"action": "retry",
		})
	}

	return nil
}

// CompositeEventHandler 组合事件处理器，支持多个事件处理器
type CompositeEventHandler struct {
	handlers []StartupEventHandler
	logger   StartupLogger
}

// NewCompositeEventHandler 创建组合事件处理器
func NewCompositeEventHandler(logger StartupLogger) *CompositeEventHandler {
	return &CompositeEventHandler{
		handlers: make([]StartupEventHandler, 0),
		logger:   logger,
	}
}

// AddHandler 添加事件处理器
func (h *CompositeEventHandler) AddHandler(handler StartupEventHandler) {
	h.handlers = append(h.handlers, handler)
	h.logger.Info("Added event handler", "handler_type", fmt.Sprintf("%T", handler))
}

// OnExecutionStart 执行开始事件
func (h *CompositeEventHandler) OnExecutionStart(ctx context.Context, execution *StartupWorkflowExecution) error {
	for _, handler := range h.handlers {
		if err := handler.OnExecutionStart(ctx, execution); err != nil {
			h.logger.Error("Event handler error in OnExecutionStart", "handler", fmt.Sprintf("%T", handler), "error", err)
		}
	}
	return nil
}

// OnExecutionEnd 执行结束事件
func (h *CompositeEventHandler) OnExecutionEnd(ctx context.Context, execution *StartupWorkflowExecution) error {
	for _, handler := range h.handlers {
		if err := handler.OnExecutionEnd(ctx, execution); err != nil {
			h.logger.Error("Event handler error in OnExecutionEnd", "handler", fmt.Sprintf("%T", handler), "error", err)
		}
	}
	return nil
}

// OnNodeStart 节点开始事件
func (h *CompositeEventHandler) OnNodeStart(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode) error {
	for _, handler := range h.handlers {
		if err := handler.OnNodeStart(ctx, execution, node); err != nil {
			h.logger.Error("Event handler error in OnNodeStart", "handler", fmt.Sprintf("%T", handler), "error", err)
		}
	}
	return nil
}

// OnNodeProgress 节点进度事件
func (h *CompositeEventHandler) OnNodeProgress(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, progress float64) error {
	for _, handler := range h.handlers {
		if err := handler.OnNodeProgress(ctx, execution, node, progress); err != nil {
			h.logger.Error("Event handler error in OnNodeProgress", "handler", fmt.Sprintf("%T", handler), "error", err)
		}
	}
	return nil
}

// OnNodeComplete 节点完成事件
func (h *CompositeEventHandler) OnNodeComplete(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, result *StartupNodeResult) error {
	for _, handler := range h.handlers {
		if err := handler.OnNodeComplete(ctx, execution, node, result); err != nil {
			h.logger.Error("Event handler error in OnNodeComplete", "handler", fmt.Sprintf("%T", handler), "error", err)
		}
	}
	return nil
}

// OnNodeError 节点错误事件
func (h *CompositeEventHandler) OnNodeError(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, err error) error {
	for _, handler := range h.handlers {
		if err := handler.OnNodeError(ctx, execution, node, err); err != nil {
			h.logger.Error("Event handler error in OnNodeError", "handler", fmt.Sprintf("%T", handler), "error", err)
		}
	}
	return nil
}

// LoggingEventHandler 纯日志事件处理器
type LoggingEventHandler struct {
	logger StartupLogger
}

// NewLoggingEventHandler 创建日志事件处理器
func NewLoggingEventHandler(logger StartupLogger) *LoggingEventHandler {
	return &LoggingEventHandler{
		logger: logger,
	}
}

// OnExecutionStart 执行开始事件
func (h *LoggingEventHandler) OnExecutionStart(ctx context.Context, execution *StartupWorkflowExecution) error {
	h.logger.Info("🚀 Workflow execution started",
		"execution_id", execution.ID,
		"workflow_id", execution.Workflow.ID,
		"workflow_name", execution.Workflow.Name,
		"total_nodes", execution.TotalNodes)
	return nil
}

// OnExecutionEnd 执行结束事件
func (h *LoggingEventHandler) OnExecutionEnd(ctx context.Context, execution *StartupWorkflowExecution) error {
	if execution.Status == WorkflowStatusCompleted {
		h.logger.Info("✅ Workflow execution completed successfully",
			"execution_id", execution.ID,
			"duration", execution.Duration,
			"completed_nodes", execution.CompletedNodes)
	} else if execution.Status == WorkflowStatusFailed {
		h.logger.Error("❌ Workflow execution failed",
			"execution_id", execution.ID,
			"duration", execution.Duration,
			"error", execution.Error,
			"completed_nodes", execution.CompletedNodes)
	} else {
		h.logger.Info("⏹️ Workflow execution ended",
			"execution_id", execution.ID,
			"status", execution.Status,
			"duration", execution.Duration)
	}
	return nil
}

// OnNodeStart 节点开始事件
func (h *LoggingEventHandler) OnNodeStart(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode) error {
	h.logger.Info("▶️ Node started",
		"execution_id", execution.ID,
		"node_id", node.ID,
		"node_name", node.Name,
		"node_type", string(node.Type))
	return nil
}

// OnNodeProgress 节点进度事件
func (h *LoggingEventHandler) OnNodeProgress(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, progress float64) error {
	h.logger.Debug("📊 Node progress",
		"execution_id", execution.ID,
		"node_id", node.ID,
		"progress", fmt.Sprintf("%.1f%%", progress))
	return nil
}

// OnNodeComplete 节点完成事件
func (h *LoggingEventHandler) OnNodeComplete(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, result *StartupNodeResult) error {
	h.logger.Info("✅ Node completed",
		"execution_id", execution.ID,
		"node_id", node.ID,
		"duration", result.Duration,
		"retry_count", result.RetryCount)
	return nil
}

// OnNodeError 节点错误事件
func (h *LoggingEventHandler) OnNodeError(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, err error) error {
	h.logger.Error("❌ Node failed",
		"execution_id", execution.ID,
		"node_id", node.ID,
		"error", err)
	return nil
}

// MetricsEventHandler 指标事件处理器
type MetricsEventHandler struct {
	logger  StartupLogger
	metrics map[string]interface{}
	mutex   sync.RWMutex
}

// NewMetricsEventHandler 创建指标事件处理器
func NewMetricsEventHandler(logger StartupLogger) *MetricsEventHandler {
	return &MetricsEventHandler{
		logger:  logger,
		metrics: make(map[string]interface{}),
	}
}

// OnExecutionStart 执行开始事件
func (h *MetricsEventHandler) OnExecutionStart(ctx context.Context, execution *StartupWorkflowExecution) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.metrics["total_executions"] = h.getIntMetric("total_executions") + 1
	h.metrics["last_execution_time"] = time.Now()

	return nil
}

// OnExecutionEnd 执行结束事件
func (h *MetricsEventHandler) OnExecutionEnd(ctx context.Context, execution *StartupWorkflowExecution) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if execution.Status == WorkflowStatusCompleted {
		h.metrics["successful_executions"] = h.getIntMetric("successful_executions") + 1
	} else if execution.Status == WorkflowStatusFailed {
		h.metrics["failed_executions"] = h.getIntMetric("failed_executions") + 1
	}

	// 更新平均执行时间
	totalSuccessful := h.getIntMetric("successful_executions")
	if totalSuccessful > 0 {
		avgDuration := h.getFloatMetric("average_execution_time")
		newAvg := ((avgDuration * float64(totalSuccessful-1)) + execution.Duration.Seconds()) / float64(totalSuccessful)
		h.metrics["average_execution_time"] = newAvg
	}

	return nil
}

// OnNodeStart 节点开始事件
func (h *MetricsEventHandler) OnNodeStart(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	nodeKey := fmt.Sprintf("node_executions_%s", node.Type)
	h.metrics[nodeKey] = h.getIntMetric(nodeKey) + 1

	return nil
}

// OnNodeProgress 节点进度事件
func (h *MetricsEventHandler) OnNodeProgress(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, progress float64) error {
	// 进度事件通常不记录指标
	return nil
}

// OnNodeComplete 节点完成事件
func (h *MetricsEventHandler) OnNodeComplete(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, result *StartupNodeResult) error {
	return nil
}

// OnNodeError 节点错误事件
func (h *MetricsEventHandler) OnNodeError(ctx context.Context, execution *StartupWorkflowExecution, node *StartupNode, err error) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	errorKey := fmt.Sprintf("node_errors_%s", node.Type)
	h.metrics[errorKey] = h.getIntMetric(errorKey) + 1

	return nil
}

// GetMetrics 获取指标
func (h *MetricsEventHandler) GetMetrics() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	metrics := make(map[string]interface{})
	for k, v := range h.metrics {
		metrics[k] = v
	}

	return metrics
}

// ResetMetrics 重置指标
func (h *MetricsEventHandler) ResetMetrics() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.metrics = make(map[string]interface{})
	h.logger.Info("Metrics reset")
}

// 辅助方法

func (h *MetricsEventHandler) getIntMetric(key string) int {
	if value, exists := h.metrics[key]; exists {
		if intVal, ok := value.(int); ok {
			return intVal
		}
	}
	return 0
}

func (h *MetricsEventHandler) getFloatMetric(key string) float64 {
	if value, exists := h.metrics[key]; exists {
		if floatVal, ok := value.(float64); ok {
			return floatVal
		}
	}
	return 0.0
}