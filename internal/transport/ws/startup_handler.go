package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ExecutionController 执行控制器接口（在ws包中定义以避免循环依赖）
type ExecutionController interface {
	ExecuteWorkflow(ctx context.Context, workflowID string, inputs map[string]interface{}) (*WorkflowExecution, error)
	GetExecution(executionID string) (*WorkflowExecution, bool)
	CancelExecution(executionID string) error
	PauseExecution(executionID string) error
	ResumeExecution(executionID string) error
}

// WorkflowExecution 工作流执行结构（简化版，用于WebSocket通信）
type WorkflowExecution struct {
	ID             string                 `json:"id"`
	WorkflowID     string                 `json:"workflow_id"`
	WorkflowName   string                 `json:"workflow_name"`
	Status         string                 `json:"status"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
	Duration       time.Duration          `json:"duration"`
	Progress       float64                `json:"progress"`
	TotalNodes     int                    `json:"total_nodes"`
	CompletedNodes int                    `json:"completed_nodes"`
	FailedNodes    int                    `json:"failed_nodes"`
	CurrentNodes   []string               `json:"current_nodes"`
	Error          string                 `json:"error,omitempty"`
	Context        map[string]interface{} `json:"context"`
}

// WorkflowNode 工作流节点结构（简化版，用于WebSocket通信）
type WorkflowNode struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Timeout     time.Duration          `json:"timeout"`
	Critical    bool                   `json:"critical"`
	Optional    bool                   `json:"optional"`
	Position    map[string]interface{} `json:"position"`
	Config      map[string]interface{} `json:"config"`
	Metadata    map[string]string      `json:"metadata"`
}

// StartupWebSocketHandler 启动工作流WebSocket处理器
type StartupWebSocketHandler struct {
	upgrader         websocket.Upgrader
	connections      map[string]*StartupConnection
	connectionsMutex sync.RWMutex
	executor         ExecutionController
	logger           StartupLogger
}

// StartupConnection 启动工作流WebSocket连接
type StartupConnection struct {
	Conn           *websocket.Conn
	ConnectionID   string
	ExecutionID    string
	UserID         string
	LastPing       time.Time
	SendChan       chan []byte
	CloseChan      chan struct{}
	Subscriptions  map[string]bool // 订阅的执行ID
	mutex          sync.RWMutex
}

// StartupMessage WebSocket消息类型
type StartupMessage struct {
	Type      string                 `json:"type"`
	EventID   string                 `json:"event_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// 消息类型常量
const (
	MsgTypeExecutionStart    = "execution_start"
	MsgTypeExecutionEnd      = "execution_end"
	MsgTypeExecutionProgress = "execution_progress"
	MsgTypeNodeStart         = "node_start"
	MsgTypeNodeProgress      = "node_progress"
	MsgTypeNodeComplete      = "node_complete"
	MsgTypeNodeError         = "node_error"
	MsgTypeExecutionCancel   = "execution_cancel"
	MsgTypeExecutionPause    = "execution_pause"
	MsgTypeExecutionResume   = "execution_resume"
	MsgTypeError             = "error"
	MsgTypePong              = "pong"
	MsgTypeSubscription      = "subscription"
)

// StartupLogger WebSocket日志接口
type StartupLogger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// NewStartupWebSocketHandler 创建启动工作流WebSocket处理器
func NewStartupWebSocketHandler(executor ExecutionController, logger StartupLogger) *StartupWebSocketHandler {
	return &StartupWebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该进行适当的来源检查
			},
		},
		connections: make(map[string]*StartupConnection),
		executor:    executor,
		logger:      logger,
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *StartupWebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接到WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return
	}

	// 创建连接对象
	connectionID := generateConnectionID()
	connection := &StartupConnection{
		Conn:          conn,
		ConnectionID:  connectionID,
		LastPing:      time.Now(),
		SendChan:      make(chan []byte, 256),
		CloseChan:     make(chan struct{}),
		Subscriptions: make(map[string]bool),
	}

	// 注册连接
	h.connectionsMutex.Lock()
	h.connections[connectionID] = connection
	h.connectionsMutex.Unlock()

	h.logger.Info("New WebSocket connection established", "connection_id", connectionID)

	// 启动消息处理协程
	go h.handleConnection(connection)

	// 发送连接确认消息
	h.sendMessage(connection, StartupMessage{
		Type:      "connection_established",
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"connection_id": connectionID,
			"server_time":   time.Now(),
		},
	})
}

// handleConnection 处理WebSocket连接
func (h *StartupWebSocketHandler) handleConnection(conn *StartupConnection) {
	defer func() {
		h.closeConnection(conn)
	}()

	// 启动发送协程
	go h.handleSending(conn)

	// 处理接收消息
	for {
		select {
		case <-conn.CloseChan:
			return
		default:
			// 设置读取超时
			conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			var msg map[string]interface{}
			err := conn.Conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					h.logger.Info("WebSocket connection closed normally", "connection_id", conn.ConnectionID)
				} else {
					h.logger.Error("WebSocket read error", "connection_id", conn.ConnectionID, "error", err)
				}
				return
			}

			// 处理接收到的消息
			h.handleIncomingMessage(conn, msg)
		}
	}
}

// handleIncomingMessage 处理接收到的消息
func (h *StartupWebSocketHandler) handleIncomingMessage(conn *StartupConnection, msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		h.sendError(conn, "missing message type")
		return
	}

	switch msgType {
	case "ping":
		h.handlePing(conn)
	case "subscribe":
		h.handleSubscribe(conn, msg)
	case "unsubscribe":
		h.handleUnsubscribe(conn, msg)
	case "execute_workflow":
		h.handleExecuteWorkflow(conn, msg)
	case "cancel_execution":
		h.handleCancelExecution(conn, msg)
	case "pause_execution":
		h.handlePauseExecution(conn, msg)
	case "resume_execution":
		h.handleResumeExecution(conn, msg)
	case "get_execution_status":
		h.handleGetExecutionStatus(conn, msg)
	default:
		h.sendError(conn, fmt.Sprintf("unknown message type: %s", msgType))
	}
}

// handlePing 处理ping消息
func (h *StartupWebSocketHandler) handlePing(conn *StartupConnection) {
	conn.LastPing = time.Now()
	h.sendMessage(conn, StartupMessage{
		Type:      MsgTypePong,
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"timestamp": time.Now(),
		},
	})
}

// handleSubscribe 处理订阅消息
func (h *StartupWebSocketHandler) handleSubscribe(conn *StartupConnection, msg map[string]interface{}) {
	executionID, ok := msg["execution_id"].(string)
	if !ok {
		h.sendError(conn, "missing execution_id in subscription message")
		return
	}

	conn.mutex.Lock()
	conn.Subscriptions[executionID] = true
	conn.mutex.Unlock()

	h.sendMessage(conn, StartupMessage{
		Type:      MsgTypeSubscription,
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"action":       "subscribed",
			"execution_id": executionID,
		},
	})

	h.logger.Info("Client subscribed to execution", "connection_id", conn.ConnectionID, "execution_id", executionID)
}

// handleUnsubscribe 处理取消订阅消息
func (h *StartupWebSocketHandler) handleUnsubscribe(conn *StartupConnection, msg map[string]interface{}) {
	executionID, ok := msg["execution_id"].(string)
	if !ok {
		h.sendError(conn, "missing execution_id in unsubscribe message")
		return
	}

	conn.mutex.Lock()
	delete(conn.Subscriptions, executionID)
	conn.mutex.Unlock()

	h.sendMessage(conn, StartupMessage{
		Type:      MsgTypeSubscription,
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"action":       "unsubscribed",
			"execution_id": executionID,
		},
	})

	h.logger.Info("Client unsubscribed from execution", "connection_id", conn.ConnectionID, "execution_id", executionID)
}

// handleExecuteWorkflow 处理执行工作流消息
func (h *StartupWebSocketHandler) handleExecuteWorkflow(conn *StartupConnection, msg map[string]interface{}) {
	workflowID, ok := msg["workflow_id"].(string)
	if !ok {
		h.sendError(conn, "missing workflow_id")
		return
	}

	inputs, ok := msg["inputs"].(map[string]interface{})
	if !ok {
		inputs = make(map[string]interface{})
	}

	// 执行工作流
	execution, err := h.executor.ExecuteWorkflow(context.Background(), workflowID, inputs)
	if err != nil {
		h.sendError(conn, fmt.Sprintf("Failed to execute workflow: %s", err.Error()))
		return
	}

	// 自动订阅此执行
	conn.mutex.Lock()
	conn.Subscriptions[execution.ID] = true
	conn.mutex.Unlock()

	h.sendMessage(conn, StartupMessage{
		Type:      MsgTypeExecutionStart,
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"execution_id":   execution.ID,
			"workflow_id":    workflowID,
			"status":         string(execution.Status),
			"total_nodes":    execution.TotalNodes,
			"completed_nodes": execution.CompletedNodes,
		},
	})
}

// handleCancelExecution 处理取消执行消息
func (h *StartupWebSocketHandler) handleCancelExecution(conn *StartupConnection, msg map[string]interface{}) {
	executionID, ok := msg["execution_id"].(string)
	if !ok {
		h.sendError(conn, "missing execution_id")
		return
	}

	err := h.executor.CancelExecution(executionID)
	if err != nil {
		h.sendError(conn, fmt.Sprintf("Failed to cancel execution: %s", err.Error()))
		return
	}

	h.sendMessage(conn, StartupMessage{
		Type:      MsgTypeExecutionCancel,
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"execution_id": executionID,
			"action":       "cancelled",
		},
	})
}

// handlePauseExecution 处理暂停执行消息
func (h *StartupWebSocketHandler) handlePauseExecution(conn *StartupConnection, msg map[string]interface{}) {
	executionID, ok := msg["execution_id"].(string)
	if !ok {
		h.sendError(conn, "missing execution_id")
		return
	}

	err := h.executor.PauseExecution(executionID)
	if err != nil {
		h.sendError(conn, fmt.Sprintf("Failed to pause execution: %s", err.Error()))
		return
	}

	h.sendMessage(conn, StartupMessage{
		Type:      MsgTypeExecutionPause,
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"execution_id": executionID,
			"action":       "paused",
		},
	})
}

// handleResumeExecution 处理恢复执行消息
func (h *StartupWebSocketHandler) handleResumeExecution(conn *StartupConnection, msg map[string]interface{}) {
	executionID, ok := msg["execution_id"].(string)
	if !ok {
		h.sendError(conn, "missing execution_id")
		return
	}

	err := h.executor.ResumeExecution(executionID)
	if err != nil {
		h.sendError(conn, fmt.Sprintf("Failed to resume execution: %s", err.Error()))
		return
	}

	h.sendMessage(conn, StartupMessage{
		Type:      MsgTypeExecutionResume,
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"execution_id": executionID,
			"action":       "resumed",
		},
	})
}

// handleGetExecutionStatus 处理获取执行状态消息
func (h *StartupWebSocketHandler) handleGetExecutionStatus(conn *StartupConnection, msg map[string]interface{}) {
	executionID, ok := msg["execution_id"].(string)
	if !ok {
		h.sendError(conn, "missing execution_id")
		return
	}

	execution, exists := h.executor.GetExecution(executionID)
	if !exists {
		h.sendError(conn, "execution not found")
		return
	}

	h.sendMessage(conn, StartupMessage{
		Type:      "execution_status",
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"execution": execution,
		},
	})
}

// handleSending 处理发送消息
func (h *StartupWebSocketHandler) handleSending(conn *StartupConnection) {
	ticker := time.NewTicker(54 * time.Second) // WebSocket ping interval
	defer ticker.Stop()

	for {
		select {
		case <-conn.CloseChan:
			return
		case message, ok := <-conn.SendChan:
			if !ok {
				return
			}

			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				h.logger.Error("Failed to send WebSocket message", "connection_id", conn.ConnectionID, "error", err)
				return
			}
		case <-ticker.C:
			// 发送ping消息
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Error("Failed to send WebSocket ping", "connection_id", conn.ConnectionID, "error", err)
				return
			}
		}
	}
}

// sendMessage 发送消息到连接
func (h *StartupWebSocketHandler) sendMessage(conn *StartupConnection, msg StartupMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Failed to marshal WebSocket message", "error", err)
		return
	}

	select {
	case conn.SendChan <- data:
	case <-conn.CloseChan:
		// 连接已关闭
	default:
		// 发送通道已满，丢弃消息
		h.logger.Warn("WebSocket send channel full, dropping message", "connection_id", conn.ConnectionID)
	}
}

// sendError 发送错误消息
func (h *StartupWebSocketHandler) sendError(conn *StartupConnection, errorMsg string) {
	h.sendMessage(conn, StartupMessage{
		Type:      MsgTypeError,
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"error": errorMsg,
		},
	})
}

// closeConnection 关闭连接
func (h *StartupWebSocketHandler) closeConnection(conn *StartupConnection) {
	h.connectionsMutex.Lock()
	delete(h.connections, conn.ConnectionID)
	h.connectionsMutex.Unlock()

	close(conn.CloseChan)
	close(conn.SendChan)
	conn.Conn.Close()

	h.logger.Info("WebSocket connection closed", "connection_id", conn.ConnectionID)
}

// BroadcastExecutionEvent 广播执行事件
func (h *StartupWebSocketHandler) BroadcastExecutionEvent(eventType string, execution *WorkflowExecution, data map[string]interface{}) {
	message := StartupMessage{
		Type:      eventType,
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"execution_id":   execution.ID,
			"workflow_id":    execution.WorkflowID,
			"status":         execution.Status,
			"progress":       execution.Progress,
			"completed_nodes": execution.CompletedNodes,
			"total_nodes":    execution.TotalNodes,
		},
	}

	// 合并额外数据
	for k, v := range data {
		message.Data[k] = v
	}

	h.broadcastToSubscribers(execution.ID, message)
}

// BroadcastNodeEvent 广播节点事件
func (h *StartupWebSocketHandler) BroadcastNodeEvent(eventType string, execution *WorkflowExecution, node *WorkflowNode, data map[string]interface{}) {
	message := StartupMessage{
		Type:      eventType,
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"execution_id": execution.ID,
			"workflow_id":  execution.WorkflowID,
			"node_id":      node.ID,
			"node_name":    node.Name,
			"node_type":    node.Type,
		},
	}

	// 合并额外数据
	for k, v := range data {
		message.Data[k] = v
	}

	h.broadcastToSubscribers(execution.ID, message)
}

// broadcastToSubscribers 向订阅者广播消息
func (h *StartupWebSocketHandler) broadcastToSubscribers(executionID string, message StartupMessage) {
	h.connectionsMutex.RLock()
	defer h.connectionsMutex.RUnlock()

	for _, conn := range h.connections {
		conn.mutex.RLock()
		subscribed := conn.Subscriptions[executionID]
		conn.mutex.RUnlock()

		if subscribed {
			h.sendMessage(conn, message)
		}
	}
}

// GetConnectionStats 获取连接统计信息
func (h *StartupWebSocketHandler) GetConnectionStats() map[string]interface{} {
	h.connectionsMutex.RLock()
	defer h.connectionsMutex.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(h.connections),
		"connections":       make([]interface{}, 0),
	}

	for _, conn := range h.connections {
		conn.mutex.RLock()
		subscriptionCount := len(conn.Subscriptions)
		conn.mutex.RUnlock()

		connStats := map[string]interface{}{
			"connection_id":       conn.ConnectionID,
			"execution_id":        conn.ExecutionID,
			"last_ping":          conn.LastPing,
			"subscription_count": subscriptionCount,
			"uptime":            time.Since(conn.LastPing).String(),
		}
		stats["connections"] = append(stats["connections"].([]interface{}), connStats)
	}

	return stats
}

// 辅助函数

// generateConnectionID 生成连接ID
func generateConnectionID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}

// generateEventID 生成事件ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}