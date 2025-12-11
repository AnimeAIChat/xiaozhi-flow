package coze

import (
	"context"
	"fmt"
	"sync"
	"time"

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
			ID:          "coze_llm",
			Type:        capability.TypeLLM,
			Name:        "Coze LLM",
			Description: "Coze Bot Platform",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"base_url":              {Type: "string", Default: "https://api.coze.cn", Description: "API Base URL"},
					"bot_id":                {Type: "string", Description: "Bot ID"},
					"user_id":               {Type: "string", Description: "User ID"},
					"personal_access_token": {Type: "string", Secret: true, Description: "Personal Access Token"},
					"client_id":             {Type: "string", Description: "Client ID (for JWT Auth)"},
					"public_key":            {Type: "string", Description: "Public Key (for JWT Auth)"},
					"private_key":           {Type: "string", Secret: true, Description: "Private Key (for JWT Auth)"},
				},
				Required: []string{"base_url", "bot_id", "user_id"},
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
	}
}

func (p *Provider) CreateExecutor(capabilityID string) (capability.Executor, error) {
	switch capabilityID {
	case "coze_llm":
		return &ChatExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

type ChatExecutor struct{}

func (e *ChatExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("coze only supports streaming via ExecuteStream")
}

func (e *ChatExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	baseURL, _ := config["base_url"].(string)
	botID, _ := config["bot_id"].(string)
	userID, _ := config["user_id"].(string)
	
	llmConfig := &LLMConfig{
		BaseURL: baseURL,
		BotID:   botID,
		UserID:  userID,
	}

	if pat, ok := config["personal_access_token"].(string); ok {
		llmConfig.AccessToken = pat
	}
	if cid, ok := config["client_id"].(string); ok {
		llmConfig.ClientID = cid
	}
	if pk, ok := config["public_key"].(string); ok {
		llmConfig.PublicKey = pk
	}
	if prk, ok := config["private_key"].(string); ok {
		llmConfig.PrivateKey = prk
	}

	provider, err := NewLLMProvider(llmConfig)
	if err != nil {
		return nil, err
	}

	// Parse messages
	msgsRaw, ok := inputs["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("messages input is required")
	}

	var messages []Message
	for _, m := range msgsRaw {
		if msgMap, ok := m.(map[string]interface{}); ok {
			role, _ := msgMap["role"].(string)
			content, _ := msgMap["content"].(string)
			messages = append(messages, Message{
				Role:    role,
				Content: content,
			})
		}
	}

	sessionID := fmt.Sprintf("plugin-%d", time.Now().UnixNano())
	stream, err := provider.Chat(ctx, sessionID, messages)
	if err != nil {
		return nil, err
	}

	outCh := make(chan map[string]interface{})
	go func() {
		defer close(outCh)
		for chunk := range stream {
			outCh <- map[string]interface{}{
				"content": chunk,
				"done":    false,
			}
		}
		outCh <- map[string]interface{}{
			"content": "",
			"done":    true,
		}
	}()

	return outCh, nil
}

// GetPluginID 返回插件ID
func (p *Provider) GetPluginID() string {
	return "coze"
}

// StartGRPCServer 启动Coze插件的gRPC服务
func (p *Provider) StartGRPCServer(address string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.grpcServer != nil {
		return fmt.Errorf("Coze gRPC server already started at %s", p.serviceAddress)
	}

	if p.logger != nil {
		p.logger.InfoTag("gRPC", "启动Coze插件gRPC服务器",
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
				p.logger.ErrorTag("gRPC", "Coze gRPC服务器启动失败",
					"address", address,
					"error", err.Error())
			}
		} else {
			p.mu.Lock()
			p.serviceAddress = address
			p.mu.Unlock()
			if p.logger != nil {
				p.logger.InfoTag("gRPC", "Coze插件gRPC服务器启动成功",
					"address", address)
			}
		}
	}()

	return nil
}

// StopGRPCServer 停止Coze插件的gRPC服务器
func (p *Provider) StopGRPCServer() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.grpcServer == nil {
		return fmt.Errorf("Coze gRPC server not started")
	}

	if p.logger != nil {
		p.logger.InfoTag("gRPC", "停止Coze插件gRPC服务器",
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
