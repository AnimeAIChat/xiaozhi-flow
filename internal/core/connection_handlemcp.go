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


func (h *ConnectionHandler) handleMCPResultCall(result llm.ActionResponse) string {
	if h.mcpDispatcher != nil {
		return h.mcpDispatcher.Dispatch(result)
	}
	h.logger.Error("handleMCPResultCall: mcpDispatcher is nil")
	return "调用工具失败: MCP分发器未初始化"
}
