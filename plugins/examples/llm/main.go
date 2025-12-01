package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	v1 "xiaozhi-server-go/api/v1"
	sdk "xiaozhi-server-go/internal/plugin/sdk"
)

// LLMPlugin 大语言模型插件
type LLMPlugin struct {
	sdk.SimplePluginImpl
	logger hclog.Logger
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Provider    string   `json:"provider"`
	Type        string   `json:"type"`
	MaxTokens   int      `json:"max_tokens"`
	ContextSize int      `json:"context_size"`
	Capabilities []string `json:"capabilities"`
	Description string   `json:"description"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// CompletionRequest 完成请求
type CompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
}

// CompletionResponse 完成响应
type CompletionResponse struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []ChatChoice  `json:"choices"`
	Usage   TokenUsage    `json:"usage"`
}

// ChatChoice 聊天选择
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// TokenUsage Token使用情况
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewLLMPlugin 创建LLM插件实例
func NewLLMPlugin(logger hclog.Logger) *LLMPlugin {
	info := &v1.PluginInfo{
		ID:          "llm-plugin",
		Name:        "LLM Plugin",
		Version:     "1.0.0",
		Description: "大语言模型集成插件，支持多种LLM提供商",
		Author:      "XiaoZhi Flow Team",
		Type:        v1.PluginTypeLLM,
		Tags:        []string{"llm", "ai", "chat", "completion"},
		Capabilities: []string{"chat_completion", "text_completion", "embedding", "fine_tuning"},
		Dependencies: []string{"openai-sdk", "anthropic-sdk", "azure-openai"},
	}

	return &LLMPlugin{
		SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
		logger:          logger.Named("llm-plugin"),
	}
}

// CallTool 实现工具调用
func (p *LLMPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	startTime := time.Now()
	p.IncrementCounter("llm.calls.total")

	switch req.ToolName {
	case "chat_completion":
		return p.handleChatCompletion(ctx, req)
	case "text_completion":
		return p.handleTextCompletion(ctx, req)
	case "get_available_models":
		return p.handleGetAvailableModels(ctx, req)
	case "count_tokens":
		return p.handleCountTokens(ctx, req)
	case "validate_prompt":
		return p.handleValidatePrompt(ctx, req)
	case "get_model_info":
		return p.handleGetModelInfo(ctx, req)
	default:
		p.IncrementCounter("llm.calls.unknown")
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "UNKNOWN_TOOL",
				Message: fmt.Sprintf("未知工具: %s", req.ToolName),
			},
		}
	}
}

// handleChatCompletion 处理聊天完成
func (p *LLMPlugin) handleChatCompletion(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	// 解析参数
	model, _ := req.Arguments["model"].(string)
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	messagesInterface, ok := req.Arguments["messages"].([]interface{})
	if !ok || len(messagesInterface) == 0 {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 messages 参数或参数为空",
			},
		}
	}

	maxTokens, _ := req.Arguments["max_tokens"].(int)
	if maxTokens == 0 {
		maxTokens = 1000
	}

	temperature, _ := req.Arguments["temperature"].(float64)
	if temperature == 0 {
		temperature = 0.7
	}

	topP, _ := req.Arguments["top_p"].(float64)
	if topP == 0 {
		topP = 1.0
	}

	stream, _ := req.Arguments["stream"].(bool)

	stopInterface, _ := req.Arguments["stop"].([]interface{})
	var stop []string
	for _, s := range stopInterface {
		if stopStr, ok := s.(string); ok {
			stop = append(stop, stopStr)
		}
	}

	// 解析消息
	var messages []ChatMessage
	for _, msgInterface := range messagesInterface {
		msgMap, ok := msgInterface.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msgMap["role"].(string)
		content, _ := msgMap["content"].(string)
		name, _ := msgMap["name"].(string)

		if role != "" && content != "" {
			messages = append(messages, ChatMessage{
				Role:    role,
				Content: content,
				Name:    name,
			})
		}
	}

	if len(messages) == 0 {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "没有有效的消息",
			},
		}
	}

	// 检查token数量
	totalTokens := p.countTokens(messages)
	if totalTokens > 4000 {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "CONTEXT_TOO_LONG",
				Message: fmt.Sprintf("上下文过长 (%d tokens)，请减少消息长度", totalTokens),
			},
		}
	}

	p.logger.Info("Processing chat completion",
		"model", model,
		"messages", len(messages),
		"max_tokens", maxTokens,
		"temperature", temperature,
		"stream", stream)

	// 生成完成
	completion, err := p.generateChatCompletion(model, messages, maxTokens, temperature, topP, stop)
	if err != nil {
		p.IncrementCounter("llm.errors.completion")
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "COMPLETION_ERROR",
				Message: fmt.Sprintf("生成完成失败: %v", err),
			},
		}
	}

	// 记录成功指标
	p.IncrementCounter("llm.calls.success")
	p.RecordHistogram("llm.completion_duration", float64(time.Since(startTime).Milliseconds()))
	p.IncrementCounter("llm.tokens.input", float64(completion.Usage.PromptTokens))
	p.IncrementCounter("llm.tokens.output", float64(completion.Usage.CompletionTokens))

	p.logger.Info("Chat completion completed",
		"prompt_tokens", completion.Usage.PromptTokens,
		"completion_tokens", completion.Usage.CompletionTokens,
		"total_tokens", completion.Usage.TotalTokens)

	return &v1.CallToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"id":      completion.ID,
			"object":  completion.Object,
			"created": completion.Created,
			"model":   completion.Model,
			"choices": completion.Choices,
			"usage":   completion.Usage,
		},
		Output: fmt.Sprintf("生成完成，使用了 %d tokens", completion.Usage.TotalTokens),
	}
}

// handleTextCompletion 处理文本完成
func (p *LLMPlugin) handleTextCompletion(ctx context.Context, req *v1.CallToolResponse) *v1.CallToolRequest {
	prompt, ok := req.Arguments["prompt"].(string)
	if !ok || prompt == "" {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 prompt 参数或参数为空",
			},
		}
	}

	model, _ := req.Arguments["model"].(string)
	if model == "" {
		model = "text-davinci-003"
	}

	maxTokens, _ := req.Arguments["max_tokens"].(int)
	if maxTokens == 0 {
		maxTokens = 500
	}

	temperature, _ := req.Arguments["temperature"].(float64)
	if temperature == 0 {
		temperature = 0.7
	}

	// 构造消息
	messages := []ChatMessage{
		{Role: "user", Content: prompt},
	}

	// 生成完成
	completion, err := p.generateChatCompletion(model, messages, maxTokens, temperature, 1.0, []string{})
	if err != nil {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "COMPLETION_ERROR",
				Message: fmt.Sprintf("文本完成失败: %v", err),
			},
		}
	}

	p.IncrementCounter("llm.text_completion.calls")

	return &v1.CallToolRequest{
		Success: true,
		Result: map[string]interface{}{
			"prompt":   prompt,
			"response": completion.Choices[0].Message.Content,
			"model":    completion.Model,
			"usage":    completion.Usage,
		},
		Output: fmt.Sprintf("文本完成，生成了 %d tokens", completion.Usage.CompletionTokens),
	}
}

// handleGetAvailableModels 获取可用模型
func (p *LLMPlugin) handleGetAvailableModels(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolRequest {
	provider, _ := req.Arguments["provider"].(string)
	modelType, _ := req.Arguments["type"].(string)

	models := p.getAvailableModels(provider, modelType)

	p.IncrementCounter("llm.models_list.calls")

	return &v1.CallToolRequest{
		Success: true,
		Result: map[string]interface{}{
			"models": models,
			"count":  len(models),
		},
		Output: fmt.Sprintf("找到 %d 个可用模型", len(models)),
	}
}

// handleCountTokens 计算token数量
func (p *LLMPlugin) handleCountTokens(ctx context.Context, req *v1.CallToolResponse) *v1.CallToolRequest {
	messagesInterface, ok := req.Arguments["messages"].([]interface{})
	if !ok {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 messages 参数",
			},
		}
	}

	// 解析消息
	var messages []ChatMessage
	for _, msgInterface := range messagesInterface {
		msgMap, ok := msgInterface.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msgMap["role"].(string)
		content, _ := msgMap["content"].(string)

		if role != "" && content != "" {
			messages = append(messages, ChatMessage{
				Role:    role,
				Content: content,
			})
		}
	}

	tokenCount := p.countTokens(messages)
	estimatedCost := p.estimateCost(messages, len(messages))

	p.IncrementCounter("llm.count_tokens.calls")

	return &v1.CallToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"token_count":     tokenCount,
			"estimated_cost":  estimatedCost,
			"message_count":   len(messages),
			"characters":      p.countCharacters(messages),
		},
		Output: fmt.Sprintf("总计 %d tokens，预计成本 $%.6f", tokenCount, estimatedCost),
	}
}

// handleValidatePrompt 验证提示
func (p *LLMPlugin) handleValidatePrompt(ctx context.Context, req *v1.CallToolResponse) *v1.CallToolRequest {
	messagesInterface, ok := req.Arguments["messages"].([]interface{})
	if !ok {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 messages 参数",
			},
		}
	}

	model, _ := req.Arguments["model"].(string)
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	// 解析消息
	var messages []ChatMessage
	for _, msgInterface := range messagesInterface {
		msgMap, ok := msgInterface.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msgMap["role"].(string)
		content, _ := msgMap["content"].(string)

		if role != "" && content != "" {
			messages = append(messages, ChatMessage{
				Role:    role,
				Content: content,
			})
		}
	}

	validation := p.validatePrompt(messages, model)

	p.IncrementCounter("llm.validate_prompt.calls")

	return &v1.CallToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"valid":           validation.Valid,
			"issues":          validation.Issues,
			"warnings":        validation.Warnings,
			"suggestions":     validation.Suggestions,
			"token_count":     validation.TokenCount,
			"estimated_cost":  validation.EstimatedCost,
			"estimated_time":  validation.EstimatedTime.Milliseconds(),
		},
		Output: validation.Summary,
	}
}

// handleGetModelInfo 获取模型信息
func (p *LLMPlugin) handleGetModelInfo(ctx context.Context, req *v1.CallToolResponse) *v1.CallToolRequest {
	modelID, ok := req.Arguments["model_id"].(string)
	if !ok {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 model_id 参数",
			},
		}
	}

	modelInfo := p.getModelInfo(modelID)

	if modelInfo == nil {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "MODEL_NOT_FOUND",
				Message: fmt.Sprintf("模型 %s 未找到", modelID),
			},
		}
	}

	p.IncrementCounter("llm.model_info.calls")

	return &v1.CallToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"model": modelInfo,
		},
		Output: fmt.Sprintf("模型信息：%s (%s)", modelInfo.Name, modelInfo.Provider),
	}
}

// PromptValidationResult 提示验证结果
type PromptValidationResult struct {
	Valid          bool
	Issues         []string
	Warnings       []string
	Suggestions    []string
	TokenCount     int
	EstimatedCost  float64
	EstimatedTime  time.Duration
	Summary        string
}

// 模拟聊天完成生成
func (p *LLMPlugin) generateChatCompletion(model string, messages []ChatMessage, maxTokens int, temperature float64, topP float64, stop []string) (*CompletionResponse, error) {
	// 模拟处理时间
	processingTime := time.Duration(200+len(p.countTokens(messages))*50) * time.Millisecond
	time.Sleep(processingTime)

	// 生成模拟响应
	response := ""
	if len(messages) > 0 && strings.Contains(messages[len(messages)-1].Content, "你好") {
		response = "你好！我是AI助手，很高兴为您服务。有什么我可以帮助您的吗？"
	} else if len(messages) > 0 && strings.Contains(messages[len(messages)-1].Content, "天气") {
		response = "对不起，我无法获取实时天气信息。建议您查看天气应用或网站获取最新的天气预报。"
	} else {
		responses := []string{
			"这是一个很好的问题。让我来详细解释一下...",
			"我理解您的疑问。基于我的分析...",
			"感谢您的提问。这涉及到多个方面...",
			"这个问题很有意思。从不同角度来看...",
		}
		response = responses[rand.Intn(len(responses))]
	}

	// 估算token使用
	promptTokens := p.countTokens(messages)
	completionTokens := len(strings.Fields(response))
	totalTokens := promptTokens + completionTokens

	return &CompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: response,
				},
				FinishReason: "stop",
			},
		},
		Usage: TokenUsage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		},
	}, nil
}

// getAvailableModels 获取可用模型列表
func (p *LLMPlugin) getAvailableModels(provider, modelType string) []ModelInfo {
	allModels := []ModelInfo{
		// OpenAI 模型
		{ID: "gpt-4", Name: "GPT-4", Provider: "openai", Type: "chat", MaxTokens: 8192, ContextSize: 8192, Capabilities: []string{"chat", "completion", "analysis"}, Description: "最强大的语言模型"},
		{ID: "gpt-4-turbo", Name: "GPT-4 Turbo", Provider: "openai", Type: "chat", MaxTokens: 4096, ContextSize: 128000, Capabilities: []string{"chat", "completion", "vision", "analysis"}, Description: "高性能多模态模型"},
		{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", Provider: "openai", Type: "chat", MaxTokens: 4096, ContextSize: 16384, Capabilities: []string{"chat", "completion"}, Description: "快速高效的对话模型"},
		{ID: "text-davinci-003", Name: "Text Davinci 003", Provider: "openai", Type: "completion", MaxTokens: 4096, ContextSize: 4096, Capabilities: []string{"completion", "analysis"}, Description: "强大的文本生成模型"},

		// Anthropic 模型
		{ID: "claude-3-opus", Name: "Claude 3 Opus", Provider: "anthropic", Type: "chat", MaxTokens: 4096, ContextSize: 200000, Capabilities: []string{"chat", "analysis", "vision"}, Description: "最强大的Claude模型"},
		{ID: "claude-3-sonnet", Name: "Claude 3 Sonnet", Provider: "anthropic", Type: "chat", MaxTokens: 4096, ContextSize: 200000, Capabilities: []string{"chat", "analysis", "vision"}, Description: "平衡性能的Claude模型"},
		{ID: "claude-3-haiku", Name: "Claude 3 Haiku", Provider: "anthropic", Type: "chat", MaxTokens: 4096, ContextSize: 200000, Capabilities: []string{"chat", "fast_response"}, Description: "快速响应的Claude模型"},

		// Azure OpenAI 模型
		{ID: "azure-gpt-4", Name: "Azure GPT-4", Provider: "azure", Type: "chat", MaxTokens: 8192, ContextSize: 8192, Capabilities: []string{"chat", "completion", "enterprise"}, Description: "企业级GPT-4部署"},
		{ID: "azure-gpt-35-turbo", Name: "Azure GPT-3.5 Turbo", Provider: "azure", Type: "chat", MaxTokens: 4096, ContextSize: 16384, Capabilities: []string{"chat", "completion", "enterprise"}, Description: "企业级GPT-3.5部署"},

		// 本地模型
		{ID: "llama-2-7b", Name: "LLaMA 2 7B", Provider: "local", Type: "chat", MaxTokens: 2048, ContextSize: 4096, Capabilities: []string{"chat", "local", "privacy"}, Description: "本地部署7B参数模型"},
		{ID: "llama-2-13b", Name: "LLaMA 2 13B", Provider: "local", Type: "chat", MaxTokens: 2048, ContextSize: 4096, Capabilities: []string{"chat", "local", "privacy"}, Description: "本地部署13B参数模型"},
	}

	// 过滤模型
	var models []ModelInfo
	for _, model := range allModels {
		if provider != "" && model.Provider != provider {
			continue
		}
		if modelType != "" && model.Type != modelType {
			continue
		}
		models = append(models, model)
	}

	return models
}

// countTokens 计算token数量（简化实现）
func (p *LLMPlugin) countTokens(messages []ChatMessage) int {
	// 简化的token计算：按字符数/4估算
	charCount := 0
	for _, msg := range messages {
		charCount += len(msg.Content) + len(msg.Role)
		if msg.Name != "" {
			charCount += len(msg.Name)
		}
	}
	return charCount / 4
}

// countCharacters 计算字符数量
func (p *LLMPlugin) countCharacters(messages []ChatMessage) int {
	count := 0
	for _, msg := range messages {
		count += len(msg.Content)
	}
	return count
}

// estimateCost 估算成本
func (p *LLMPlugin) estimateCost(messages []ChatMessage, messageCount int) float64 {
	// 简化的成本计算
	tokenCount := p.countTokens(messages)
	// 假设每1000 tokens花费 $0.002
	return float64(tokenCount) * 0.002 / 1000
}

// validatePrompt 验证提示
func (p *LLMPlugin) validatePrompt(messages []ChatMessage, model string) PromptValidationResult {
	result := PromptValidationResult{
		Valid:          true,
		Issues:         []string{},
		Warnings:       []string{},
		Suggestions:    []string{},
		TokenCount:     p.countTokens(messages),
		EstimatedCost:  p.estimateCost(messages, len(messages)),
		EstimatedTime:  time.Duration(p.countTokens(messages)*50) * time.Millisecond,
	}

	// 检查消息数量
	if len(messages) == 0 {
		result.Valid = false
		result.Issues = append(result.Issues, "没有提供消息")
		result.Summary = "提示验证失败：没有消息"
		return result
	}

	if len(messages) > 50 {
		result.Valid = false
		result.Issues = append(result.Issues, fmt.Sprintf("消息过多 (%d)，最多支持 50 条", len(messages)))
	}

	// 检查消息结构
	hasUserMessage := false
	hasSystemMessage := false
	for i, msg := range messages {
		// 检查角色
		if msg.Role == "" {
			result.Issues = append(result.Issues, fmt.Sprintf("消息 %d 缺少角色", i))
		} else if msg.Role != "system" && msg.Role != "user" && msg.Role != "assistant" {
			result.Issues = append(result.Issues, fmt.Sprintf("消息 %d 角色无效: %s", i, msg.Role))
		}

		if msg.Role == "user" {
			hasUserMessage = true
		}
		if msg.Role == "system" {
			hasSystemMessage = true
		}

		// 检查内容
		if msg.Content == "" {
			result.Issues = append(result.Issues, fmt.Sprintf("消息 %d 内容为空", i))
		} else if len(msg.Content) > 32000 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("消息 %d 内容过长 (%d 字符)", i, len(msg.Content)))
		}
	}

	if !hasUserMessage {
		result.Issues = append(result.Issues, "缺少用户消息")
	}

	// 检查token限制
	if result.TokenCount > 8000 {
		result.Valid = false
		result.Issues = append(result.Issues, fmt.Sprintf("Token数量过多 (%d)，超过模型限制", result.TokenCount))
	} else if result.TokenCount > 6000 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Token数量较多 (%d)，接近模型限制", result.TokenCount))
	}

	// 生成建议
	if !hasSystemMessage {
		result.Suggestions = append(result.Suggestions, "建议添加系统消息以定义AI的角色和行为")
	}

	if len(messages) == 1 && messages[0].Role == "user" {
		result.Suggestions = append(result.Suggestions, "考虑添加系统消息来改善回复质量")
	}

	for i, msg := range messages {
		if len(strings.Fields(msg.Content)) < 3 {
			result.Suggestions = append(result.Suggestions, fmt.Sprintf("消息 %d 内容过短，可能影响理解", i))
		}
	}

	// 生成摘要
	if len(result.Issues) == 0 {
		if len(result.Warnings) == 0 {
			result.Summary = "提示验证通过，适合发送到LLM"
		} else {
			result.Summary = fmt.Sprintf("提示验证通过，但有 %d 个警告", len(result.Warnings))
		}
	} else {
		result.Summary = fmt.Sprintf("提示验证失败，发现 %d 个问题", len(result.Issues))
	}

	return result
}

// getModelInfo 获取模型信息
func (p *LLMPlugin) getModelInfo(modelID string) *ModelInfo {
	models := p.getAvailableModels("", "")
	for _, model := range models {
		if model.ID == modelID {
			return &model
		}
	}
	return nil
}

// ListTools 列出可用工具
func (p *LLMPlugin) ListTools(ctx context.Context) *v1.ListToolsResponse {
	tools := []*v1.ToolInfo{
		{
			Name:        "chat_completion",
			Description: "进行聊天对话完成",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"model": map[string]interface{}{
						"type":        "string",
						"description": "模型ID",
						"default":     "gpt-3.5-turbo",
					},
					"messages": map[string]interface{}{
						"type":        "array",
						"description": "聊天消息列表",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"role": map[string]interface{}{
									"type": "string",
									"enum": []string{"system", "user", "assistant"},
								},
								"content": map[string]interface{}{
									"type": "string",
								},
								"name": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
					"max_tokens": map[string]interface{}{
						"type":        "integer",
						"description": "最大生成token数",
						"default":     1000,
					},
					"temperature": map[string]interface{}{
						"type":        "number",
						"description": "随机性控制 (0-2)",
						"default":     0.7,
					},
					"top_p": map[string]interface{}{
						"type":        "number",
						"description": "核采样参数",
						"default":     1.0,
					},
					"stream": map[string]interface{}{
						"type":        "boolean",
						"description": "是否流式返回",
						"default":     false,
					},
					"stop": map[string]interface{}{
						"type":        "array",
						"description": "停止词列表",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"messages"},
			},
		},
		{
			Name:        "text_completion",
			Description: "进行文本补全",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "提示文本",
					},
					"model": map[string]interface{}{
						"type":        "string",
						"description": "模型ID",
						"default":     "text-davinci-003",
					},
					"max_tokens": map[string]interface{}{
						"type":        "integer",
						"description": "最大生成token数",
						"default":     500,
					},
					"temperature": map[string]interface{}{
						"type":        "number",
						"description": "随机性控制 (0-2)",
						"default":     0.7,
					},
				},
				"required": []string{"prompt"},
			},
		},
		{
			Name:        "get_available_models",
			Description: "获取可用模型列表",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"provider": map[string]interface{}{
						"type":        "string",
						"description": "提供商筛选 (openai, anthropic, azure, local)",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "模型类型筛选 (chat, completion)",
					},
				},
			},
		},
		{
			Name:        "count_tokens",
			Description: "计算消息的token数量和成本",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"messages": map[string]interface{}{
						"type":        "array",
						"description": "消息列表",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"role": map[string]interface{}{
									"type": "string",
								},
								"content": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
				"required": []string{"messages"},
			},
		},
		{
			Name:        "validate_prompt",
			Description: "验证提示是否适合发送",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"messages": map[string]interface{}{
						"type":        "array",
						"description": "消息列表",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"role": map[string]interface{}{
									"type": "string",
								},
								"content": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
					"model": map[string]interface{}{
						"type":        "string",
						"description": "模型ID",
						"default":     "gpt-3.5-turbo",
					},
				},
				"required": []string{"messages"},
			},
		},
		{
			Name:        "get_model_info",
			Description: "获取特定模型的详细信息",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"model_id": map[string]interface{}{
						"type":        "string",
						"description": "模型ID",
					},
				},
				"required": []string{"model_id"},
			},
		},
	}

	return &v1.ListToolsResponse{
		Success: true,
		Tools:   tools,
	}
}

// GetToolSchema 获取工具模式
func (p *LLMPlugin) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
	return &v1.GetToolSchemaResponse{
		Success: true,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tool_name": map[string]interface{}{
					"type": "string",
				},
				"description": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}
}

func main() {
	// 创建日志记录器
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "llm-plugin",
		Level:  hclog.Info,
		Output: hclog.DefaultOutput,
	})

	// 创建插件实例
	plugin := NewLLMPlugin(logger)

	logger.Info("Starting LLM Plugin")

	// 服务插件
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.SimpleHandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"plugin": &sdk.SimplePluginRPC{Impl: plugin},
		},
	})
}