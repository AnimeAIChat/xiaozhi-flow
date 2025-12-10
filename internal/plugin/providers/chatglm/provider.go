package chatglm

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
			ID:          "chatglm_llm",
			Type:        capability.TypeLLM,
			Name:        "ChatGLM LLM",
			Description: "ChatGLM Large Language Model",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"api_key":   {Type: "string", Secret: true, Description: "API Key"},
					"base_url":  {Type: "string", Default: "https://open.bigmodel.cn/api/paas/v4/", Description: "API Base URL"},
					"model":     {Type: "string", Default: "glm-4", Description: "Model Name"},
					"max_tokens": {Type: "number", Default: 2048},
				},
				Required: []string{"api_key"},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"messages": {Type: "array"},
				},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"content": {Type: "string"},
				},
			},
		},
		{
			ID:          "chatglm_vllm",
			Type:        capability.TypeLLM, // VLLM is also LLM type in capability system usually, or maybe separate?
			// The user asked for "ChatGLMVLLM vllm". I should check if capability.TypeVLLM exists.
			// In capability/types.go I saw TypeLLM, TypeASR, TypeTTS, TypeTool.
			// I should probably use TypeLLM but with image input support, or add TypeVLLM if I can modify types.go.
			// For now I'll use TypeLLM and note it supports vision.
			Name:        "ChatGLM VLLM",
			Description: "ChatGLM Vision Language Model",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"api_key":   {Type: "string", Secret: true, Description: "API Key"},
					"base_url":  {Type: "string", Default: "https://open.bigmodel.cn/api/paas/v4/", Description: "API Base URL"},
					"model":     {Type: "string", Default: "glm-4v", Description: "Model Name"},
					"max_tokens": {Type: "number", Default: 2048},
				},
				Required: []string{"api_key"},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"messages": {Type: "array"},
					"images":   {Type: "array", Description: "List of base64 encoded images"},
				},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"content": {Type: "string"},
				},
			},
		},
	}
}

func (p *Provider) CreateExecutor(capabilityID string) (capability.Executor, error) {
	switch capabilityID {
	case "chatglm_llm", "chatglm_vllm":
		return &ChatExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

type ChatExecutor struct{}

func (e *ChatExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("chatglm only supports streaming via ExecuteStream")
}

func (e *ChatExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	apiKey, _ := config["api_key"].(string)
	baseURL, _ := config["base_url"].(string)
	model, _ := config["model"].(string)
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4/"
	}
	if model == "" {
		model = "glm-4"
	}

	clientConfig := openai.DefaultConfig(apiKey)
	clientConfig.BaseURL = baseURL
	client := openai.NewClientWithConfig(clientConfig)

	// Parse messages
	msgsRaw, ok := inputs["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("messages input is required")
	}

	var messages []openai.ChatCompletionMessage
	for _, m := range msgsRaw {
		if msgMap, ok := m.(map[string]interface{}); ok {
			role, _ := msgMap["role"].(string)
			content, _ := msgMap["content"].(string)
			
			// Handle multi-modal content if present (for VLLM)
			// This is a simplified handling. Real VLLM might need more complex parsing.
			// If inputs["images"] is present, we might need to attach it to the last user message.
			
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    role,
				Content: content,
			})
		}
	}

	// Handle images for VLLM
	if imagesRaw, ok := inputs["images"].([]interface{}); ok && len(imagesRaw) > 0 {
		// Find the last user message
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == openai.ChatMessageRoleUser {
				// Convert content to multi-part
				contentParts := []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: messages[i].Content,
					},
				}
				
				for _, img := range imagesRaw {
					if imgStr, ok := img.(string); ok {
						contentParts = append(contentParts, openai.ChatMessagePart{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL: fmt.Sprintf("data:image/jpeg;base64,%s", imgStr),
							},
						})
					}
				}
				
				messages[i].Content = ""
				messages[i].MultiContent = contentParts
				break
			}
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	outCh := make(chan map[string]interface{})
	go func() {
		defer close(outCh)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				return
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					outCh <- map[string]interface{}{
						"content": content,
						"done":    false,
					}
				}
				if response.Choices[0].FinishReason != "" {
					outCh <- map[string]interface{}{
						"content": "",
						"done":    true,
					}
				}
			}
		}
	}()

	return outCh, nil
}
