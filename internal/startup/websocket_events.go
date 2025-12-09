package startup

import (
	"xiaozhi-server-go/internal/startup/model"
)

// WebSocketEventBroadcaster WebSocket事件广播器接口
type WebSocketEventBroadcaster interface {
	// BroadcastExecutionEvent 广播执行事件
	BroadcastExecutionEvent(eventType model.EventType, execution *model.StartupExecution, data map[string]interface{})

	// BroadcastNodeEvent 广播节点事件
	BroadcastNodeEvent(eventType model.EventType, execution *model.StartupExecution, node *model.StartupNode, data map[string]interface{})

	// GetConnectionStats 获取连接统计信息
	GetConnectionStats() map[string]interface{}
}

// WebSocketBroadcasterAdapter WebSocket广播器适配器
type WebSocketBroadcasterAdapter struct {
	broadcaster WebSocketEventBroadcaster
	logger      model.StartupLogger
}

// NewWebSocketBroadcasterAdapter 创建WebSocket广播器适配器
func NewWebSocketBroadcasterAdapter(broadcaster WebSocketEventBroadcaster, logger model.StartupLogger) *WebSocketBroadcasterAdapter {
	return &WebSocketBroadcasterAdapter{
		broadcaster: broadcaster,
		logger:      logger,
	}
}

// BroadcastExecutionEvent 广播执行事件
func (a *WebSocketBroadcasterAdapter) BroadcastExecutionEvent(eventType model.EventType, execution *model.StartupExecution, data map[string]interface{}) {
	if a.broadcaster != nil {
		a.broadcaster.BroadcastExecutionEvent(eventType, execution, data)
	}
}

// BroadcastNodeEvent 广播节点事件
func (a *WebSocketBroadcasterAdapter) BroadcastNodeEvent(eventType model.EventType, execution *model.StartupExecution, node *model.StartupNode, data map[string]interface{}) {
	if a.broadcaster != nil {
		a.broadcaster.BroadcastNodeEvent(eventType, execution, node, data)
	}
}

// GetConnectionStats 获取连接统计信息
func (a *WebSocketBroadcasterAdapter) GetConnectionStats() map[string]interface{} {
	if a.broadcaster != nil {
		return a.broadcaster.GetConnectionStats()
	}
	return map[string]interface{}{
		"total_connections": 0,
		"connections":       []interface{}{},
	}
}

// NoOpWebSocketBroadcaster 空操作WebSocket广播器（用于测试或无WebSocket环境）
type NoOpWebSocketBroadcaster struct{}

// NewNoOpWebSocketBroadcaster 创建空操作WebSocket广播器
func NewNoOpWebSocketBroadcaster() *NoOpWebSocketBroadcaster {
	return &NoOpWebSocketBroadcaster{}
}

// BroadcastExecutionEvent 广播执行事件（空操作）
func (n *NoOpWebSocketBroadcaster) BroadcastExecutionEvent(eventType model.EventType, execution *model.StartupExecution, data map[string]interface{}) {
	// 空操作 - 不执行任何广播
}

// BroadcastNodeEvent 广播节点事件（空操作）
func (n *NoOpWebSocketBroadcaster) BroadcastNodeEvent(eventType model.EventType, execution *model.StartupExecution, node *model.StartupNode, data map[string]interface{}) {
	// 空操作 - 不执行任何广播
}

// GetConnectionStats 获取连接统计信息（空操作）
func (n *NoOpWebSocketBroadcaster) GetConnectionStats() map[string]interface{} {
	return map[string]interface{}{
		"total_connections": 0,
		"connections":       []interface{}{},
		"type":              "noop",
	}
}


