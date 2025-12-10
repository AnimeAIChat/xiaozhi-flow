package adapters

import (
	"context"
	"fmt"

	providers "xiaozhi-server-go/internal/domain/providers/types"
	"xiaozhi-server-go/internal/domain/providers/llm"
	"xiaozhi-server-go/internal/domain/llm/inter"
	"xiaozhi-server-go/internal/plugin/capability"
	"xiaozhi-server-go/internal/plugin/providers/coze"
	"xiaozhi-server-go/internal/plugin/providers/doubao"
	"xiaozhi-server-go/internal/plugin/providers/ollama"
	"xiaozhi-server-go/internal/plugin/providers/openai"
)

// RegisterLegacyAdapters registers plugin-based providers into the legacy system
func RegisterLegacyAdapters() {
	llm.Register("doubao", NewDoubaoAdapter)
	llm.Register("coze", NewCozeAdapter)
	llm.Register("ollama", NewOllamaAdapter)
	llm.Register("openai", NewOpenAIAdapter)
}

type PluginLLMAdapter struct {
	*llm.BaseProvider
	executor capability.Executor
}

func createAdapter(config *llm.Config, provider capability.Provider, capabilityID string) (llm.Provider, error) {
	executor, err := provider.CreateExecutor(capabilityID)
	if err != nil {
		return nil, err
	}
	return &PluginLLMAdapter{
		BaseProvider: llm.NewBaseProvider(config),
		executor:     executor,
	}, nil
}

func NewDoubaoAdapter(config *llm.Config) (llm.Provider, error) {
	return createAdapter(config, doubao.NewProvider(), "doubao_llm")
}

func NewCozeAdapter(config *llm.Config) (llm.Provider, error) {
	return createAdapter(config, coze.NewProvider(), "coze_llm")
}

func NewOllamaAdapter(config *llm.Config) (llm.Provider, error) {
	return createAdapter(config, ollama.NewProvider(), "ollama_llm")
}

func NewOpenAIAdapter(config *llm.Config) (llm.Provider, error) {
	return createAdapter(config, openai.NewProvider(), "openai_llm")
}

func (p *PluginLLMAdapter) Response(ctx context.Context, sessionID string, messages []providers.Message) (<-chan string, error) {
	cfg := p.prepareConfig()
	inputs := p.prepareInputs(messages, nil)

	streamExecutor, ok := p.executor.(capability.StreamExecutor)
	if !ok {
		return nil, fmt.Errorf("executor does not support streaming")
	}

	stream, err := streamExecutor.ExecuteStream(ctx, cfg, inputs)
	if err != nil {
		return nil, err
	}

	outChan := make(chan string)
	go func() {
		defer close(outChan)
		for item := range stream {
			if content, ok := item["content"].(string); ok {
				outChan <- content
			}
		}
	}()
	return outChan, nil
}

func (p *PluginLLMAdapter) ResponseWithFunctions(ctx context.Context, sessionID string, messages []providers.Message, tools []providers.Tool) (<-chan providers.Response, error) {
	cfg := p.prepareConfig()
	inputs := p.prepareInputs(messages, tools)

	streamExecutor, ok := p.executor.(capability.StreamExecutor)
	if !ok {
		return nil, fmt.Errorf("executor does not support streaming")
	}

	stream, err := streamExecutor.ExecuteStream(ctx, cfg, inputs)
	if err != nil {
		return nil, err
	}

	outChan := make(chan providers.Response)
	go func() {
		defer close(outChan)
		for item := range stream {
			resp := providers.Response{}
			if content, ok := item["content"].(string); ok {
				resp.Content = content
			}
			
			// Handle tool calls
			if tcs, ok := item["tool_calls"].([]interface{}); ok {
				resp.ToolCalls = make([]inter.ToolCall, len(tcs))
				for i, tcRaw := range tcs {
					if tc, ok := tcRaw.(map[string]interface{}); ok {
						resp.ToolCalls[i] = inter.ToolCall{
							ID:   fmt.Sprintf("%v", tc["id"]),
							Type: fmt.Sprintf("%v", tc["type"]),
						}
						if fn, ok := tc["function"].(map[string]interface{}); ok {
							resp.ToolCalls[i].Function = inter.ToolCallFunction{
								Name:      fmt.Sprintf("%v", fn["name"]),
								Arguments: fmt.Sprintf("%v", fn["arguments"]),
							}
						}
					}
				}
			}
			
			outChan <- resp
		}
	}()
	return outChan, nil
}

func (p *PluginLLMAdapter) prepareConfig() map[string]interface{} {
	cfg := map[string]interface{}{
		"api_key":    p.Config().APIKey,
		"base_url":   p.Config().BaseURL,
		"model":      p.Config().ModelName,
		"max_tokens": p.Config().MaxTokens,
	}
	for k, v := range p.Config().Extra {
		cfg[k] = v
	}
	return cfg
}

func (p *PluginLLMAdapter) prepareInputs(messages []providers.Message, tools []providers.Tool) map[string]interface{} {
	msgs := make([]interface{}, len(messages))
	for i, m := range messages {
		msgMap := map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
		if m.ToolCallID != "" {
			msgMap["tool_call_id"] = m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			tcs := make([]interface{}, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				tcs[j] = map[string]interface{}{
					"id":   tc.ID,
					"type": tc.Type,
					"function": map[string]interface{}{
						"name":      tc.Function.Name,
						"arguments": tc.Function.Arguments,
					},
				}
			}
			msgMap["tool_calls"] = tcs
		}
		msgs[i] = msgMap
	}

	inputs := map[string]interface{}{
		"messages": msgs,
	}
	
	if len(tools) > 0 {
		ts := make([]interface{}, len(tools))
		for i, t := range tools {
			ts[i] = map[string]interface{}{
				"type": t.Type,
				"function": map[string]interface{}{
					"name":        t.Function.Name,
					"description": t.Function.Description,
					"parameters":  t.Function.Parameters,
				},
			}
		}
		inputs["tools"] = ts
	}
	
	return inputs
}
