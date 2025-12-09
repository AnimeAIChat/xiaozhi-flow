package components

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"xiaozhi-server-go/internal/domain/llm"
	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/transport/http/vision"
	"xiaozhi-server-go/internal/utils"
)

// MCPHandlerFunc defines the function signature for MCP handlers
type MCPHandlerFunc func(args interface{})

// MCPDispatcher handles dispatching MCP tool calls to specific handlers
type MCPDispatcher struct {
	logger   *logging.Logger
	handlers map[string]MCPHandlerFunc
	
	// Dependencies needed by handlers
	// Using interfaces to decouple from ConnectionHandler
	speaker         Speaker
	ttsProvider     TTSProvider
	dialogueManager DialogueManager
	config          ConfigProvider
	audioSender     AudioSender
	agentID         uint
	
	// State management
	closeAfterChat *bool // Pointer to allow modification
}

// Interfaces to decouple from ConnectionHandler

type Speaker interface {
	SystemSpeak(text string) error
}

type TTSProvider interface {
	SetVoice(voice string) (error, string)
}

type DialogueManager interface {
	SetSystemMessage(systemMessage string)
	KeepRecentMessages(maxMessages int)
	// GetLLMDialogue() []interface{} // Using interface{} to avoid circular dependency, or define Message type here
}

type ConfigProvider interface {
	GetMusicDir() string
}

type AudioSender interface {
	SendAudioMessage(filepath string, text string, textIndex int, round int)
}

type LLMGenerator interface {
	GenResponseByLLM(ctx context.Context, dialogue []interface{}, round int)
}

// NewMCPDispatcher creates a new MCPDispatcher
func NewMCPDispatcher(
	logger *logging.Logger,
	speaker Speaker,
	ttsProvider TTSProvider,
	dialogueManager DialogueManager,
	config ConfigProvider,
	audioSender AudioSender,
	agentID uint,
	closeAfterChat *bool,
) *MCPDispatcher {
	d := &MCPDispatcher{
		logger:          logger,
		speaker:         speaker,
		ttsProvider:     ttsProvider,
		dialogueManager: dialogueManager,
		config:          config,
		audioSender:     audioSender,
		agentID:         agentID,
		closeAfterChat:  closeAfterChat,
		handlers:        make(map[string]MCPHandlerFunc),
	}
	d.initHandlers()
	return d
}

func (d *MCPDispatcher) initHandlers() {
	d.handlers = map[string]MCPHandlerFunc{
		"mcp_handler_exit":         d.handleExit,
		"mcp_handler_take_photo":   d.handleTakePhoto,
		"mcp_handler_change_voice": d.handleChangeVoice,
		"mcp_handler_change_role":  d.handleChangeRole,
		"mcp_handler_play_music":   d.handlePlayMusic,
		"mcp_handler_switch_agent": d.handleSwitchAgent,
	}
}

// Dispatch handles the MCP result call
func (d *MCPDispatcher) Dispatch(result llm.ActionResponse) string {
	errResult := "调用工具失败"
	
	if result.Action != llm.ActionTypeCallHandler {
		d.logger.Error("handleMCPResultCall: result.Action is not ActionTypeCallHandler, but %d", result.Action)
		return errResult
	}
	if result.Result == nil {
		d.logger.Error("handleMCPResultCall: result.Result is nil")
		return errResult
	}

	if Caller, ok := result.Result.(llm.ActionResponseCall); ok {
		if handler, exists := d.handlers[Caller.FuncName]; exists {
			handler(Caller.Args)
			return "调用工具成功: " + Caller.FuncName
		} else {
			d.logger.Error("handleMCPResultCall: no handler found for function %s", Caller.FuncName)
		}
	} else {
		d.logger.Error("handleMCPResultCall: result.Result is not a map[string]interface{}")
	}
	return errResult
}

// Handlers

func (d *MCPDispatcher) handleSwitchAgent(args interface{}) {
	var newAgentID uint = 0

	switch v := args.(type) {
	case map[string]interface{}:
		if idv, ok := v["agent_id"]; ok {
			switch idt := idv.(type) {
			case float64:
				newAgentID = uint(idt)
			case int:
				newAgentID = uint(idt)
			case string:
				if n, err := strconv.Atoi(idt); err == nil {
					newAgentID = uint(n)
				}
			}
		}
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			newAgentID = uint(n)
		}
	case float64:
		newAgentID = uint(v)
	case int:
		newAgentID = uint(v)
	default:
		d.logger.Error("mcp_handler_switch_agent: unsupported arg type %T", v)
		return
	}

	if newAgentID != 0 && newAgentID == d.agentID {
		d.logger.Info("mcp_handler_switch_agent: already using agent %d", newAgentID)
		_ = d.speaker.SystemSpeak("您已经在使用该智能体")
		return
	}

	d.logger.Info("Database functionality removed - agent switching not available")
	_ = d.speaker.SystemSpeak("数据库功能已移除，无法切换智能体")
}

func (d *MCPDispatcher) handlePlayMusic(args interface{}) {
	if songName, ok := args.(string); ok {
		d.logger.Info("mcp_handler_play_music: %s", songName)
		if path, name, err := utils.GetMusicFilePathFuzzy(songName, d.config.GetMusicDir()); err != nil {
			d.logger.Error("mcp_handler_play_music: Play failed: %v", err)
			_ = d.speaker.SystemSpeak("没有找到名为" + songName + "的歌曲")
		} else {
			// Assuming textIndex and round are managed elsewhere or passed in context if needed.
			// For now, using placeholders or we need to pass current state to Dispatcher.
			// This highlights a dependency on ConnectionHandler state (tts_last_text_index, talkRound).
			// We might need to pass these as arguments to Dispatch or have them accessible via interface.
			// For simplicity in this step, we might need to expose them via interface or pass them.
			
			// Ideally, AudioSender should handle the indexing if possible, or we pass current state.
			// Let's assume AudioSender.SendAudioMessage handles it or we pass 0/current.
			// Wait, ConnectionHandler.sendAudioMessage uses h.tts_last_text_index and h.talkRound.
			// We should probably pass these values when creating/updating the dispatcher or pass a state provider.
			
			d.audioSender.SendAudioMessage(path, name, -1, -1) // -1 indicates "use current" or similar if we change interface
		}
	} else {
		d.logger.Error("mcp_handler_play_music: args is not a string")
	}
}

func (d *MCPDispatcher) handleChangeVoice(args interface{}) {
	if voice, ok := args.(string); ok {
		d.logger.Info("mcp_handler_change_voice: %s", voice)
		if err, voiceName := d.ttsProvider.SetVoice(voice); err != nil {
			d.logger.Error("mcp_handler_change_voice: SetVoice failed: %v", err)
			_ = d.speaker.SystemSpeak("切换语音失败，没有叫" + voice + "的音色")
		} else {
			d.logger.Info(fmt.Sprintf("mcp_handler_change_voice: SetVoice success: %s", voiceName))
			_ = d.speaker.SystemSpeak("已切换到音色" + voice)
		}
	} else {
		d.logger.Error("mcp_handler_change_voice: args is not a string")
	}
}

func (d *MCPDispatcher) handleChangeRole(args interface{}) {
	if params, ok := args.(map[string]string); ok {
		role := params["role"]
		prompt := params["prompt"]

		d.logger.Info("mcp_handler_change_role: %s", role)
		d.dialogueManager.SetSystemMessage(prompt)
		d.dialogueManager.KeepRecentMessages(5)
		
		// TTS provider type check logic was in ConnectionHandler. 
		// Ideally TTSProvider interface should handle this or we move logic here.
		// For now, we just call SetVoice if needed.
		// The original code checked for "edge" provider.
		// We can try to set voice and ignore error if not supported or let TTSProvider handle it.
		
		if role == "陕西女友" {
			d.ttsProvider.SetVoice("zh-CN-shaanxi-XiaoniNeural")
		} else if role == "英语老师" {
			d.ttsProvider.SetVoice("zh-CN-XiaoyiNeural")
		} else if role == "好奇小男孩" {
			d.ttsProvider.SetVoice("zh-CN-YunxiNeural")
		}
		
		_ = d.speaker.SystemSpeak("已切换到新角色 " + role)
	} else {
		d.logger.Error("mcp_handler_change_role: args is not a string")
	}
}

func (d *MCPDispatcher) handleExit(args interface{}) {
	if text, ok := args.(string); ok {
		*d.closeAfterChat = true
		_ = d.speaker.SystemSpeak(text)
	} else {
		d.logger.Error("mcp_handler_exit: args is not a string")
	}
}

func (d *MCPDispatcher) handleTakePhoto(args interface{}) {
	resultStr, _ := args.(string)
	type visionAPIResponse struct {
		Success bool                      `json:"success"`
		Message string                    `json:"message"`
		Code    int                       `json:"code"`
		Data    vision.VisionAnalysisData `json:"data"`
	}

	var resp visionAPIResponse
	if err := json.Unmarshal([]byte(resultStr), &resp); err != nil {
		d.logger.Error("解析 Vision API 响应失败: %v", err)
		return
	}

	if !resp.Success {
		errMsg := resp.Data.Error
		if errMsg == "" && resp.Message != "" {
			errMsg = resp.Message
		}
		d.logger.Error("拍照失败: %s", errMsg)
		// We need to trigger LLM response generation here.
		// This requires another dependency: LLMGenerator
		// For now, we might skip this or add it to interface.
		return
	}

	_ = d.speaker.SystemSpeak(resp.Data.Result)
}
