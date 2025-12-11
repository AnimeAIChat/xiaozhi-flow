package chatglm

import (
	"context"
	"fmt"
	"sync"

	"github.com/sashabaranov/go-openai"
	"xiaozhi-server-go/internal/plugin/capability"
	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/plugin/grpc/server"
)

type Provider struct {
	logger        *logging.Logger
	grpcServer    *server.GRPCServer
	grpcService   *GRPCServer
	serviceAddress string
	mu           sync.RWMutex
}

func NewProvider() *Provider {
	return NewProviderWithLogger(nil)
}

func NewProviderWithLogger(logger *logging.Logger) *Provider {
	if logger == nil {
		logger = logging.DefaultLogger
	}
	return &Provider{
		logger: logger,
	}
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

// GetPluginID 返回插件ID
func (p *Provider) GetPluginID() string {
	return "chatglm"
}

// StartGRPCServer 启动ChatGLM插件的gRPC服务
func (p *Provider) StartGRPCServer(address string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.grpcServer != nil {
		return fmt.Errorf("ChatGLM gRPC server already started at %s", p.serviceAddress)
	}

	if p.logger != nil {
		p.logger.InfoTag("gRPC", "启动ChatGLM插件gRPC服务器",
			"address", address)
	}

	// 创建gRPC服务器
	p.grpcServer = server.NewGRPCServer(address, p.logger)

	// 创建gRPC服务实例
	p.grpcService = NewGRPCServer(p, p.logger)

	// 注册服务
	p.grpcServer.RegisterPluginService(p.grpcService)

	// 启用反射（用于调试）
	p.grpcServer.EnableReflection()

	// 启动服务器
	go func() {
		if err := p.grpcServer.Start(); err != nil {
			if p.logger != nil {
				p.logger.ErrorTag("gRPC", "ChatGLM gRPC服务器启动失败",
					"address", address,
					"error", err.Error())
			}
		} else {
			p.mu.Lock()
			p.serviceAddress = address
			p.mu.Unlock()
			if p.logger != nil {
				p.logger.InfoTag("gRPC", "ChatGLM插件gRPC服务器启动成功",
					"address", address)
			}
		}
	}()

	return nil
}

// StopGRPCServer 停止ChatGLM插件的gRPC服务器
func (p *Provider) StopGRPCServer() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.grpcServer == nil {
		return fmt.Errorf("ChatGLM gRPC server not started")
	}

	if p.logger != nil {
		p.logger.InfoTag("gRPC", "停止ChatGLM插件gRPC服务器",
			"address", p.serviceAddress)
	}

	p.grpcServer.Stop()

	p.grpcServer = nil
	p.grpcService = nil
	p.serviceAddress = ""

	return nil
}

// GetServiceAddress 获取gRPC服务地址
func (p *Provider) GetServiceAddress() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.serviceAddress
}
