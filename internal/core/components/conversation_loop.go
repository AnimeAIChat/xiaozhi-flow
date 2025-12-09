package components

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"xiaozhi-server-go/internal/domain/chat"
	domainimage "xiaozhi-server-go/internal/domain/image"
	domainllminter "xiaozhi-server-go/internal/domain/llm/inter"
	"xiaozhi-server-go/internal/domain/providers/llm"
	domaintts "xiaozhi-server-go/internal/domain/tts"
	domainttsinter "xiaozhi-server-go/internal/domain/tts/inter"
	providers "xiaozhi-server-go/internal/domain/providers/types"
	internallogging "xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/platform/config"
	internalutils "xiaozhi-server-go/internal/utils"

	"github.com/sashabaranov/go-openai"
)

// LLMManager defines the interface for LLM interactions
type LLMManager interface {
	Response(ctx context.Context, sessionID string, messages []domainllminter.Message, tools []domainllminter.Tool) (<-chan domainllminter.ResponseChunk, error)
}

type ConversationDialogueManager interface {
	Put(message chat.Message)
	GetLLMDialogue() []chat.Message
	SetSystemMessage(systemMessage string)
	KeepRecentMessages(maxMessages int)
}

// FunctionRegister defines the interface for tool management
type FunctionRegister interface {
	GetAllFunctions() []interface{}
}

type TTSTask struct {
	text      string
	round     int
	textIndex int
	filepath  string
}

type AudioTask struct {
	filepath  string
	text      string
	round     int
	textIndex int
}

type ConversationLoop struct {
	logger           *internallogging.Logger
	dialogueManager  ConversationDialogueManager
	llmManager       LLMManager
	llmProvider      providers.LLMProvider // Keep for event publisher access if needed
	ttsManager       *domaintts.Manager
	responseSender   *ResponseSender
	mcpDispatcher    *MCPDispatcher
	config           *config.Config
	functionRegister FunctionRegister
	audioSender      AudioSender

	// Queues
	ttsQueue           chan TTSTask
	audioMessagesQueue chan AudioTask
	stopChan           chan struct{}

	// State
	sessionID            string
	talkRound            int
	roundStartTime       time.Time
	serverVoiceStop      int32
	closeAfterChat       bool
	ttsPending           int32
	tts_last_audio_index int
	ttsProviderName      string
	mu                   sync.Mutex
}

func NewConversationLoop(
	logger *internallogging.Logger,
	sessionID string,
	dialogueManager ConversationDialogueManager,
	llmManager LLMManager,
	llmProvider providers.LLMProvider,
	ttsManager *domaintts.Manager,
	responseSender *ResponseSender,
	mcpDispatcher *MCPDispatcher,
	config *config.Config,
	functionRegister FunctionRegister,
	audioSender AudioSender,
	ttsProviderName string,
) *ConversationLoop {
	return &ConversationLoop{
		logger:             logger,
		sessionID:          sessionID,
		dialogueManager:    dialogueManager,
		llmManager:         llmManager,
		llmProvider:        llmProvider,
		ttsManager:         ttsManager,
		responseSender:     responseSender,
		mcpDispatcher:      mcpDispatcher,
		config:             config,
		functionRegister:   functionRegister,
		audioSender:        audioSender,
		ttsProviderName:    ttsProviderName,
		ttsQueue:           make(chan TTSTask, 100),
		audioMessagesQueue: make(chan AudioTask, 100),
		stopChan:           make(chan struct{}),
	}
}

func (c *ConversationLoop) Start() {
	go c.startTTSQueueHandler()
	go c.startAudioQueueHandler()
}

func (c *ConversationLoop) Stop() {
	close(c.stopChan)
}

func (c *ConversationLoop) startTTSQueueHandler() {
	c.logger.Info("[协程] [TTS队列] TTS队列处理协程启动")
	defer c.logger.Info("[协程] [TTS队列] TTS队列处理协程退出")

	for {
		select {
		case <-c.stopChan:
			c.logger.Debug("[协程] [TTS队列] 收到停止信号，退出协程")
			return
		case task := <-c.ttsQueue:
			c.processTTSTask(task.text, task.textIndex, task.round, task.filepath)
		}
	}
}

func (c *ConversationLoop) startAudioQueueHandler() {
	c.logger.Info("[协程] [音频发送] 音频消息发送协程启动")
	defer c.logger.Info("[协程] [音频发送] 音频消息发送协程退出")

	for {
		select {
		case <-c.stopChan:
			c.logger.Debug("[协程] [音频发送] 收到停止信号，退出协程")
			return
		case task := <-c.audioMessagesQueue:
			c.audioSender.SendAudioMessage(task.filepath, task.text, task.textIndex, task.round)
		}
	}
}


func (c *ConversationLoop) HandleChatMessage(ctx context.Context, text string) error {
	if text == "" {
		c.logger.Warn("收到空聊天消息，忽略")
		// c.clientAbortChat() // TODO: Callback or event for abort
		return fmt.Errorf("聊天消息为空")
	}

	if c.QuitIntent(text) {
		return nil
	}

	// 新的一轮对话开始，确保允许继续流式识别
	c.closeAfterChat = false

	// 检测是否是唤醒词，实现快速响应
	if internalutils.IsWakeUpWord(text) {
		c.logger.Info(fmt.Sprintf("[唤醒] [检测成功] 文本 '%s' 匹配唤醒词模式", text))
		return c.handleWakeUpMessage(ctx, text)
	} else {
		c.logger.Info(fmt.Sprintf("[唤醒] [检测失败] 文本 '%s' 不匹配唤醒词模式", text))
	}

	// 记录正在处理对话的状态
	c.logger.Info(fmt.Sprintf("[对话] [开始处理] 文本: %s", internalutils.SanitizeForLog(text)))

	// TODO: Clear audio queue logic needs to be handled by ConnectionHandler or passed in as a callback
	// For now, we assume the caller handles audio queue clearing or we provide a method to do it.

	// 增加对话轮次
	c.mu.Lock()
	c.talkRound++
	c.roundStartTime = time.Now()
	currentRound := c.talkRound
	c.mu.Unlock()
	
	c.logger.Info(fmt.Sprintf("[对话] [轮次 %d] 开始新的对话轮次", currentRound))

	// 普通文本消息处理流程
	// 立即发送 stt 消息
	err := c.responseSender.SendSTT(text)
	if err != nil {
		c.logger.Error(fmt.Sprintf("发送STT消息失败: %v", err))
		return fmt.Errorf("发送STT消息失败: %v", err)
	}

	c.logger.Info(fmt.Sprintf("[聊天] [消息 %s]", internalutils.SanitizeForLog(text)))

	// 发送tts start状态
	if err := c.responseSender.SendTTSState("start", "", 0); err != nil {
		c.logger.Error(fmt.Sprintf("发送TTS开始状态失败: %v", err))
		return fmt.Errorf("发送TTS开始状态失败: %v", err)
	}

	// 发送思考状态的情绪
	if err := c.responseSender.SendEmotion("thinking"); err != nil {
		c.logger.Error(fmt.Sprintf("发送思考状态情绪消息失败: %v", err))
		return fmt.Errorf("发送情绪消息失败: %v", err)
	}

	// 添加用户消息到对话历史
	c.dialogueManager.Put(chat.Message{
		Role:    "user",
		Content: text,
	})

	// Get messages for LLM
	llmMessages := make([]providers.Message, 0)
	for _, msg := range c.dialogueManager.GetLLMDialogue() {
		llmMessages = append(llmMessages, providers.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
			ToolCalls:  msg.ToolCalls,
		})
	}

	return c.GenResponseByLLM(ctx, llmMessages, currentRound)
}

func (c *ConversationLoop) QuitIntent(text string) bool {
	if text == "退出" || text == "再见" {
		c.logger.Info("检测到退出意图")
		c.responseSender.SendRawText("再见！")
		// TODO: Signal connection close
		return true
	}
	return false
}

func (c *ConversationLoop) handleWakeUpMessage(ctx context.Context, text string) error {
	c.logger.Info(fmt.Sprintf("[唤醒] [快速响应] 检测到唤醒词: %s", text))
	
	// 停止当前的语音播放
	// TODO: Stop audio playback logic
	c.responseSender.SendTTSState("stop", "", 0)

	// 增加对话轮次
	c.mu.Lock()
	c.talkRound++
	currentRound := c.talkRound
	c.mu.Unlock()

	c.logger.Info(fmt.Sprintf("[对话] [轮次 %d] 唤醒响应", currentRound))

	// 立即发送 stt 消息
	err := c.responseSender.SendSTT(text)
	if err != nil {
		c.logger.Error(fmt.Sprintf("发送STT消息失败: %v", err))
	}

	// 检查是否开启了快速回复
	// TODO: Check config for fast reply
	// For now assume false or handle later

	c.logger.Info("[唤醒] [LLM回复] 快速回复已关闭，使用LLM生成回复")
	
	// 添加用户消息到对话历史
	c.dialogueManager.Put(chat.Message{
		Role:    "user",
		Content: text,
	})

	// Get messages for LLM
	llmMessages := make([]providers.Message, 0)
	for _, msg := range c.dialogueManager.GetLLMDialogue() {
		llmMessages = append(llmMessages, providers.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
			ToolCalls:  msg.ToolCalls,
		})
	}

	return c.GenResponseByLLM(ctx, llmMessages, currentRound)
}

func (c *ConversationLoop) GenResponseByLLM(ctx context.Context, messages []providers.Message, round int) error {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(fmt.Sprintf("GenResponseByLLM发生panic: %v", r))
			errorMsg := "抱歉，处理您的请求时发生了错误"
			c.SpeakAndPlay(errorMsg, 1, round)
		}
	}()

	// 发布LLM开始事件
	if publisher := llm.GetEventPublisher(c.llmProvider); publisher != nil {
		publisher.SetSessionID(c.sessionID)
		publisher.PublishLLMResponse("", false, round, nil, 0, "") // 开始事件
	}

	// 使用LLM生成回复
	tools := c.functionRegister.GetAllFunctions()
	
	// 转换消息格式
	interMessages := make([]domainllminter.Message, len(messages))
	for i, msg := range messages {
		interMsg := domainllminter.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}

		// 转换ToolCalls
		if len(msg.ToolCalls) > 0 {
			interMsg.ToolCalls = make([]domainllminter.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				interMsg.ToolCalls[j] = domainllminter.ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: domainllminter.ToolCallFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}

		interMessages[i] = interMsg
	}

	// 转换工具格式
	interTools := make([]domainllminter.Tool, 0, len(tools))
	for i, toolInterface := range tools {
		tool, ok := toolInterface.(openai.Tool)
		if !ok {
			c.logger.Error(fmt.Sprintf("工具类型转换失败 [%d]: %T", i, toolInterface))
			continue
		}
		interTool := domainllminter.Tool{
			Type: string(tool.Type),
			Function: domainllminter.ToolFunction{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
		interTools = append(interTools, interTool)
	}

	responses, err := c.llmManager.Response(ctx, c.sessionID, interMessages, interTools)
	if err != nil {
		// 发布LLM错误事件
		if publisher := llm.GetEventPublisher(c.llmProvider); publisher != nil {
			publisher.PublishLLMError(err, round)
		}
		return fmt.Errorf("LLM生成回复失败: %v", err)
	}

	// 处理回复
	var responseMessage []string
	processedChars := 0
	textIndex := 0

	atomic.StoreInt32(&c.serverVoiceStop, 0)

	// 处理流式响应
	toolCallFlag := false
	functionName := ""
	functionID := ""
	functionArguments := ""
	contentArguments := ""

	for response := range responses {
		content := response.Content
		toolCall := response.ToolCalls

		if response.Error != nil {
			c.logger.Error(fmt.Sprintf("LLM响应错误: %s", response.Error.Error()))
			errorMsg := "抱歉，服务暂时不可用，请稍后再试"
			c.SpeakAndPlay(errorMsg, 1, round)
			return fmt.Errorf("LLM响应错误: %s", response.Error)
		}

		if content != "" {
			contentArguments += content
		}

		if !toolCallFlag && strings.HasPrefix(contentArguments, "<tool_call>") {
			toolCallFlag = true
		}

		if len(toolCall) > 0 {
			toolCallFlag = true
			if toolCall[0].ID != "" {
				functionID = toolCall[0].ID
			}
			if toolCall[0].Function.Name != "" {
				functionName = toolCall[0].Function.Name
			}
			if toolCall[0].Function.Arguments != "" {
				functionArguments += toolCall[0].Function.Arguments
			}
		}

		if toolCallFlag {
			continue
		}

		// 累积回复内容
		responseMessage = append(responseMessage, content)
		
		// 实时处理文本用于TTS
		fullText := strings.Join(responseMessage, "")
		if len(fullText) > processedChars {
			newText := fullText[processedChars:]
			
			// 简单的分句逻辑，实际可能需要更复杂的处理
			if strings.ContainsAny(newText, "，。！？；：,.!?;:") {
				// 找到最后一个标点符号
				lastPunct := -1
				for i, r := range newText {
					if strings.ContainsRune("，。！？；：,.!?;:", r) {
						lastPunct = i
					}
				}
				
				if lastPunct != -1 {
					sentence := newText[:lastPunct+1]
					processedChars += len(sentence)
					textIndex++
					
					// 异步发送TTS
					go c.SpeakAndPlay(sentence, textIndex, round)
				}
			}
		}
	}

	// 处理剩余文本
	fullText := strings.Join(responseMessage, "")
	if len(fullText) > processedChars {
		remainingText := fullText[processedChars:]
		if strings.TrimSpace(remainingText) != "" {
			textIndex++
			go c.SpeakAndPlay(remainingText, textIndex, round)
		}
	}

	// 处理工具调用
	if toolCallFlag {
		// TODO: Handle tool calls using MCPDispatcher
		// This part needs to be implemented to call mcpDispatcher.HandleToolCall
		// and then recursively call GenResponseByLLM with the result
		
		// For now, just log
		c.logger.Info(fmt.Sprintf("Tool call detected: %s %s", functionName, functionArguments))
		
		// Construct tool call message
		c.dialogueManager.Put(chat.Message{
			Role: "assistant",
			ToolCalls: []domainllminter.ToolCall{
				{
					ID:   functionID,
					Type: "function",
					Function: domainllminter.ToolCallFunction{
						Name:      functionName,
						Arguments: functionArguments,
					},
				},
			},
		})

		// Execute tool
		// This requires MCPDispatcher to return the result, which we then feed back to LLM
		// Since MCPDispatcher logic is currently in ConnectionHandler (partially) or split,
		// we need to make sure we can call it here.
		
		// Assuming MCPDispatcher has a method to execute tool and return result string
		// But MCPDispatcher currently handles specific MCP tools.
		// We need a general way to handle tools.
		
		// Let's assume we can use mcpDispatcher to handle it if it's an MCP tool.
		// Or we might need to move the tool execution logic here or to a ToolManager.
	} else {
		// Add assistant response to history
		c.dialogueManager.Put(chat.Message{
			Role:    "assistant",
			Content: fullText,
		})
	}

	return nil
}

func (c *ConversationLoop) GetTalkRound() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.talkRound
}

func (c *ConversationLoop) SpeakAndPlay(text string, textIndex int, round int) {
	// 暂停将客户端音频发送到ASR（避免TTS播放期间触发ASR导致服务端sequence冲突）
	// TODO: Need a way to pause ASR. Maybe callback or interface?
	// For now, we skip this or assume AudioProcessor handles it?
	// ConnectionHandler had atomic.StoreInt32(&h.asrPause, 1)
	// We might need to expose this control.

	defer func() {
		c.ttsQueue <- TTSTask{
			text:      text,
			round:     round,
			textIndex: textIndex,
			filepath:  "",
		}
	}()

	originText := text
	text = internalutils.RemoveAllEmoji(text)
	text = internalutils.RemoveMarkdownSyntax(text)
	if text == "" {
		c.logger.Debug(fmt.Sprintf("SpeakAndPlay 跳过空文本分段，原始文本: %s, 索引: %d", internalutils.SanitizeForLog(originText), textIndex))
		return
	}

	if atomic.LoadInt32(&c.serverVoiceStop) == 1 {
		c.logger.Info(fmt.Sprintf("speakAndPlay 服务端语音停止, 不再发送音频数据：%s", internalutils.SanitizeForLog(text)))
		return
	}

	if len(text) > 255 {
		c.logger.Warn(fmt.Sprintf("文本过长，超过255字符限制，截断合成语音: %s", internalutils.SanitizeForLog(text)))
		text = text[:255]
	}
}

func (c *ConversationLoop) processTTSTask(text string, textIndex int, round int, filepath string) {
	hasAudio := false
	defer func() {
		if hasAudio {
			c.tts_last_audio_index = textIndex
			c.audioSender.AddTTSPending(1)
			c.audioMessagesQueue <- AudioTask{
				filepath:  filepath,
				text:      text,
				round:     round,
				textIndex: textIndex,
			}
		} else {
			c.logger.DebugTag("TTS", "跳过音频任务，样本索引=%d，暂无可播放内容", textIndex)
		}
	}()

	if filepath != "" {
		hasAudio = true
		return
	}

	// ttsStartTime := time.Now()
	cleanText := internalutils.RemoveAllEmoji(text)
	cleanText = internalutils.RemoveParentheses(cleanText)
	cleanText = regexp.MustCompile(`[~]`).ReplaceAllString(cleanText, "")

	if cleanText == "" {
		c.logger.DebugTag("TTS", "跳过空文本任务，样本索引=%d", textIndex)
		return
	}

	logText := cleanText
	if len(logText) > 20 {
		logText = logText[:20] + "..."
	}

	if atomic.LoadInt32(&c.serverVoiceStop) == 1 {
		c.logger.Info(fmt.Sprintf("processTTSTask 服务端语音停止, 不再生成音频：%s", logText))
		return
	}

	var generatedFile string
	var err error

	// Check for plugin TTS
	ttsProviderName := c.ttsProviderName
	
	// Try to use specific TTS provider configuration if available
	if ttsProviderName != "" && c.config != nil && c.config.TTS != nil {
		if cfg, ok := c.config.TTS[ttsProviderName]; ok {
			ttsConfig := domainttsinter.TTSConfig{
				Provider:   cfg.Type,
				Voice:      cfg.Voice,
				Format:     cfg.Format,
			}
			if ttsConfig.Provider == "" {
				ttsConfig.Provider = ttsProviderName
			}
			
			generatedFile, err = c.ttsManager.ToTTSWithConfig(text, ttsConfig, c.config)
			if err != nil {
				c.logger.Error(fmt.Sprintf("TTS转换失败(%s): %v，尝试默认配置", ttsProviderName, err))
				generatedFile = "" // Reset to try default
			}
		}
	}

	if generatedFile == "" {
		generatedFile, err = c.ttsManager.ToTTS(text)
		if err != nil {
			c.logger.Error(fmt.Sprintf("TTS转换失败:text(%s) %v", logText, err))
			return
		}
	}
	
	filepath = generatedFile
	hasAudio = true
	c.logger.DebugTag("TTS", "转换成功 text=%s index=%d 文件=%s", logText, textIndex, filepath)

	if atomic.LoadInt32(&c.serverVoiceStop) == 1 {
		c.logger.Info(fmt.Sprintf("processTTSTask 服务端语音停止, 不再发送音频数据：%s", logText))
		c.deleteAudioFileIfNeeded(filepath, "服务端语音停止时")
		hasAudio = false
		filepath = ""
		return
	}
}

func (c *ConversationLoop) deleteAudioFileIfNeeded(filepath string, reason string) {
	if c.config != nil && !c.config.Audio.DeleteAudio {
		return
	}
	if filepath == "" {
		return
	}

	if internalutils.IsMusicFile(filepath) {
		c.logger.Info(fmt.Sprintf(reason+" 跳过删除音乐文件: %s", filepath))
		return
	}

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.logger.Debug(fmt.Sprintf(reason+" 文件不存在，无需删除: %s", filepath))
		return
	}

	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			c.logger.Debug(fmt.Sprintf(reason+" 文件已被删除: %s", filepath))
		} else {
			c.logger.Error(fmt.Sprintf(reason+" 删除音频文件失败: %v", err))
		}
	} else {
		c.logger.Debug(fmt.Sprintf("%s 已删除音频文件: %s", reason, filepath))
	}
}

func (c *ConversationLoop) GenResponseByVLLM(ctx context.Context, messages []providers.Message, imageData domainimage.ImageData, text string, round int) error {
	// TODO: Implement VLLM logic
	return nil
}
