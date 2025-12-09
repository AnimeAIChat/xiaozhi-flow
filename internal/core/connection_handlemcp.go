package core

import (
	"xiaozhi-server-go/internal/domain/llm"
)

func (h *ConnectionHandler) initMCPResultHandlers() {
	// 初始化MCP结果处理器
	// 使用 MCPDispatcher 处理
	// We keep this map for backward compatibility if other parts use it, 
	// but ideally we should use mcpDispatcher.Dispatch directly.
	// However, handleMCPResultCall uses this map.
	// Let's update handleMCPResultCall to use dispatcher.
}

// mcp_handler_switch_agent is deprecated, use MCPDispatcher
func (h *ConnectionHandler) mcp_handler_switch_agent(args interface{}) {
	// Deprecated: Logic moved to MCPDispatcher
}

func (h *ConnectionHandler) handleMCPResultCall(result llm.ActionResponse) string {
	if h.mcpDispatcher != nil {
		return h.mcpDispatcher.Dispatch(result)
	}
	h.logger.Error("handleMCPResultCall: mcpDispatcher is nil")
	return "调用工具失败: MCP分发器未初始化"
}

// mcp_handler_play_music is deprecated, use MCPDispatcher
func (h *ConnectionHandler) mcp_handler_play_music(args interface{}) {
	// Deprecated: Logic moved to MCPDispatcher
}

// mcp_handler_change_voice is deprecated, use MCPDispatcher
func (h *ConnectionHandler) mcp_handler_change_voice(args interface{}) {
	// Deprecated: Logic moved to MCPDispatcher
}

// mcp_handler_change_role is deprecated, use MCPDispatcher
func (h *ConnectionHandler) mcp_handler_change_role(args interface{}) {
	// Deprecated: Logic moved to MCPDispatcher
}

// mcp_handler_exit is deprecated, use MCPDispatcher
func (h *ConnectionHandler) mcp_handler_exit(args interface{}) {
	// Deprecated: Logic moved to MCPDispatcher
}

// mcp_handler_take_photo is deprecated, use MCPDispatcher
func (h *ConnectionHandler) mcp_handler_take_photo(args interface{}) {
	// Deprecated: Logic moved to MCPDispatcher
}
