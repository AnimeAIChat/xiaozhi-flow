package ollama

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
			ID:          "ollama_llm",
			Type:        capability.TypeLLM,
			Name:        "Ollama LLM",
			Description: "Ollama Local Language Model",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"base_url":  {Type: "string", Default: "http://localhost:11434/v1", Description: "API Base URL"},
					"model":     {Type: "string", Default: "llama3", Description: "Model Name"},
					"api_key":   {Type: "string", Default: "ollama", Description: "API Key (ignored by Ollama but required by client)"},
				},
				Required: []string{"base_url", "model"},
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
			ID:          "ollama_vllm",
			Type:        capability.TypeLLM,
			Name:        "Ollama VLLM",
			Description: "Ollama Vision Language Model",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"base_url":  {Type: "string", Default: "http://localhost:11434/v1", Description: "API Base URL"},
					"model":     {Type: "string", Default: "llava", Description: "Model Name"},
					"api_key":   {Type: "string", Default: "ollama", Description: "API Key"},
				},
				Required: []string{"base_url", "model"},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"messages": {Type: "array"},
					"images":   {Type: "array"},
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
	case "ollama_llm", "ollama_vllm":
		return &ChatExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

type ChatExecutor struct{}

func (e *ChatExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("ollama only supports streaming via ExecuteStream")
}

func (e *ChatExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	baseURL, _ := config["base_url"].(string)
	model, _ := config["model"].(string)
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" {
		apiKey = "ollama"
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
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    role,
				Content: content,
			})
		}
	}

	// Handle images for VLLM
	if imagesRaw, ok := inputs["images"].([]interface{}); ok && len(imagesRaw) > 0 {
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == openai.ChatMessageRoleUser {
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
