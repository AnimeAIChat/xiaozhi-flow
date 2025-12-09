package coze

import (
	"context"
	"fmt"
	"time"

	"xiaozhi-server-go/internal/core/providers"
	"xiaozhi-server-go/internal/core/providers/llm"
	cozellm "xiaozhi-server-go/internal/core/providers/llm/coze"
	"xiaozhi-server-go/internal/plugin/capability"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
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
	
	extra := make(map[string]interface{})
	extra["bot_id"] = botID
	extra["user_id"] = userID
	
	if pat, ok := config["personal_access_token"].(string); ok {
		extra["personal_access_token"] = pat
	}
	if cid, ok := config["client_id"].(string); ok {
		extra["client_id"] = cid
	}
	if pk, ok := config["public_key"].(string); ok {
		extra["public_key"] = pk
	}
	if prk, ok := config["private_key"].(string); ok {
		extra["private_key"] = prk
	}
	// Legacy provider uses "url" in extra for base url fallback, but also BaseURL field
	extra["url"] = baseURL

	llmConfig := &llm.Config{
		Type:    "coze",
		BaseURL: baseURL,
		Extra:   extra,
	}

	provider, err := cozellm.NewProvider(llmConfig)
	if err != nil {
		return nil, err
	}
	if err := provider.Initialize(); err != nil {
		return nil, err
	}

	// Parse messages
	msgsRaw, ok := inputs["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("messages input is required")
	}

	var messages []providers.Message
	for _, m := range msgsRaw {
		if msgMap, ok := m.(map[string]interface{}); ok {
			role, _ := msgMap["role"].(string)
			content, _ := msgMap["content"].(string)
			messages = append(messages, providers.Message{
				Role:    role,
				Content: content,
			})
		}
	}

	sessionID := fmt.Sprintf("plugin-%d", time.Now().UnixNano())
	stream, err := provider.Response(ctx, sessionID, messages)
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
