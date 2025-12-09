package openai

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"xiaozhi-server-go/internal/plugin/capability"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) GetCapabilities() []capability.Definition {
	return []capability.Definition{
		{
			ID:          "openai_chat",
			Type:        capability.TypeLLM,
			Name:        "OpenAI Chat",
			Description: "Chat completion using OpenAI API",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"api_key": {Type: "string", Secret: true, Description: "OpenAI API Key"},
					"base_url": {Type: "string", Description: "API Base URL (optional)"},
					"model": {Type: "string", Default: "gpt-3.5-turbo", Description: "Model name"},
				},
				Required: []string{"api_key"},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"messages": {
						Type: "array",
						Items: &capability.Schema{
							Type: "object",
							Properties: map[string]capability.Property{
								"role": {Type: "string"},
								"content": {Type: "string"},
							},
						},
					},
					"temperature": {Type: "number", Default: 0.7},
				},
				Required: []string{"messages"},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"content": {Type: "string"},
					"usage": {Type: "object"},
				},
			},
		},
	}
}

func (p *Provider) CreateExecutor(capabilityID string) (capability.Executor, error) {
	switch capabilityID {
	case "openai_chat":
		return &ChatExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

type ChatExecutor struct{}

func (e *ChatExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 1. Parse Config
	apiKey, _ := config["api_key"].(string)
	baseURL, _ := config["base_url"].(string)
	model, _ := config["model"].(string)
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	clientConfig := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		clientConfig.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(clientConfig)

	// 2. Parse Inputs
	msgsRaw, ok := inputs["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("messages input is required and must be an array")
	}

	var messages []openai.ChatCompletionMessage
	for _, m := range msgsRaw {
		msgMap, ok := m.(map[string]interface{})
		if !ok {
			// Try to handle if it's already the struct (unlikely in this architecture but good for safety)
			continue
		}
		role, _ := msgMap["role"].(string)
		content, _ := msgMap["content"].(string)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: content,
		})
	}
    
	temperature := 0.7
	if t, ok := inputs["temperature"].(float64); ok {
		temperature = t
	}

	// 3. Call API
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: float32(temperature),
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// 4. Format Output
	return map[string]interface{}{
		"content": resp.Choices[0].Message.Content,
		"usage": map[string]interface{}{
			"prompt_tokens":     resp.Usage.PromptTokens,
			"completion_tokens": resp.Usage.CompletionTokens,
			"total_tokens":      resp.Usage.TotalTokens,
		},
	}, nil
}
