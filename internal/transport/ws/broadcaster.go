package ws

import (
	"time"
)

// WebSocketEventBroadcasterImpl WebSocket事件广播器实现
type WebSocketEventBroadcasterImpl struct {
	handler *StartupWebSocketHandler
}

// NewWebSocketEventBroadcasterImpl 创建WebSocket事件广播器实现
func NewWebSocketEventBroadcasterImpl(handler *StartupWebSocketHandler) *WebSocketEventBroadcasterImpl {
	return &WebSocketEventBroadcasterImpl{
		handler: handler,
	}
}

// BroadcastExecutionEvent 广播执行事件
func (w *WebSocketEventBroadcasterImpl) BroadcastExecutionEvent(eventType string, execution interface{}, data map[string]interface{}) {
	if w.handler != nil {
		// 将execution转换为ws包中的WorkflowExecution类型
		wsExecution := w.convertExecution(execution)
		w.handler.BroadcastExecutionEvent(eventType, wsExecution, data)
	}
}

// BroadcastNodeEvent 广播节点事件
func (w *WebSocketEventBroadcasterImpl) BroadcastNodeEvent(eventType string, execution interface{}, node interface{}, data map[string]interface{}) {
	if w.handler != nil {
		// 将execution和node转换为ws包中的类型
		wsExecution := w.convertExecution(execution)
		wsNode := w.convertNode(node)
		w.handler.BroadcastNodeEvent(eventType, wsExecution, wsNode, data)
	}
}

// GetConnectionStats 获取连接统计信息
func (w *WebSocketEventBroadcasterImpl) GetConnectionStats() map[string]interface{} {
	if w.handler != nil {
		return w.handler.GetConnectionStats()
	}
	return map[string]interface{}{
		"total_connections": 0,
		"connections":       []interface{}{},
	}
}

// convertExecution 转换执行对象（使用类型断言）
func (w *WebSocketEventBroadcasterImpl) convertExecution(execution interface{}) *WorkflowExecution {
	// 这里应该根据实际的执行对象类型进行转换
	// 暂时返回一个默认的执行对象
	return &WorkflowExecution{
		ID:             "converted-execution",
		WorkflowID:     "unknown-workflow",
		WorkflowName:   "Unknown Workflow",
		Status:         "unknown",
		StartTime:      time.Now(),
		Progress:       0.0,
		TotalNodes:     0,
		CompletedNodes: 0,
		FailedNodes:    0,
		CurrentNodes:   []string{},
		Context:        make(map[string]interface{}),
	}
}

// convertNode 转换节点对象（使用类型断言）
func (w *WebSocketEventBroadcasterImpl) convertNode(node interface{}) *WorkflowNode {
	// 这里应该根据实际的节点对象类型进行转换
	// 暂时返回一个默认的节点对象
	return &WorkflowNode{
		ID:          "converted-node",
		Name:        "Unknown Node",
		Type:        "unknown",
		Description: "Converted node",
		Status:      "unknown",
		Timeout:     0,
		Critical:    false,
		Optional:    false,
		Position:    make(map[string]interface{}),
		Config:      make(map[string]interface{}),
		Metadata:    make(map[string]string),
	}
}