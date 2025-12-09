package infrastructure

import (
	"context"
	"fmt"

	"xiaozhi-server-go/internal/domain/llm/aggregate"
	"xiaozhi-server-go/internal/domain/llm/repository"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/errors"
	"xiaozhi-server-go/internal/plugin/capability"
)

type LLMManager struct {
	registry *capability.Registry
	config   *config.Config
}

func NewLLMManager(cfg *config.Config, registry *capability.Registry) (repository.LLMRepository, error) {
	return &LLMManager{
		registry: registry,
		config:   cfg,
	}, nil
}

func (m *LLMManager) Generate(ctx context.Context, req repository.GenerateRequest) (*repository.GenerateResult, error) {
	// 1. Get Provider Config
	providerID := req.Config.Provider
	llmCfg, ok := m.config.LLM[providerID]
	if !ok {
		return nil, errors.New(errors.KindDomain, "llm_manager", fmt.Sprintf("provider config not found: %s", providerID))
	}

	// 2. Map Config to Plugin Config
	pluginConfig := map[string]interface{}{
		"api_key":  llmCfg.APIKey,
		"base_url": llmCfg.BaseURL,
		"model":    llmCfg.ModelName,
	}
	// Override model if specified in request
	if req.Config.Model != "" {
		pluginConfig["model"] = req.Config.Model
	}

	// 3. Map Request to Plugin Inputs
	inputs := map[string]interface{}{
		"messages":    convertMessagesToPlugin(req.Messages),
		"temperature": float64(req.Config.Temperature),
	}

	// 4. Get Executor
	capabilityID := m.resolveCapabilityID(llmCfg.Type)
	
	executor, err := m.registry.GetExecutor(capabilityID)
	if err != nil {
		return nil, errors.Wrap(errors.KindDomain, "llm_manager", fmt.Sprintf("failed to get executor for capability %s (type: %s)", capabilityID, llmCfg.Type), err)
	}

	// 5. Execute
	output, err := executor.Execute(ctx, pluginConfig, inputs)
	if err != nil {
		return nil, errors.Wrap(errors.KindDomain, "llm_manager", "plugin execution failed", err)
	}

	// 6. Map Output to Result
	content, _ := output["content"].(string)
	usageMap, _ := output["usage"].(map[string]interface{})
	
	usage := &aggregate.Usage{}
	if usageMap != nil {
		if pt, ok := usageMap["prompt_tokens"].(int); ok {
			usage.PromptTokens = pt
		}
		if ct, ok := usageMap["completion_tokens"].(int); ok {
			usage.CompletionTokens = ct
		}
		if tt, ok := usageMap["total_tokens"].(int); ok {
			usage.TotalTokens = tt
		}
	}

	return &repository.GenerateResult{
		Content: content,
		Usage:   usage,
	}, nil
}

func (m *LLMManager) Stream(ctx context.Context, req repository.GenerateRequest) (<-chan repository.ResponseChunk, error) {
	// 1. Get Provider Config
	providerID := req.Config.Provider
	llmCfg, ok := m.config.LLM[providerID]
	if !ok {
		return nil, errors.New(errors.KindDomain, "llm_manager", fmt.Sprintf("provider config not found: %s", providerID))
	}

	// 2. Map Config
	pluginConfig := map[string]interface{}{
		"api_key":  llmCfg.APIKey,
		"base_url": llmCfg.BaseURL,
		"model":    llmCfg.ModelName,
	}
	if req.Config.Model != "" {
		pluginConfig["model"] = req.Config.Model
	}

	// 3. Map Inputs
	inputs := map[string]interface{}{
		"messages":    convertMessagesToPlugin(req.Messages),
		"temperature": float64(req.Config.Temperature),
	}

	// 4. Get Executor
	capabilityID := m.resolveCapabilityID(llmCfg.Type)
	executor, err := m.registry.GetExecutor(capabilityID)
	if err != nil {
		return nil, errors.Wrap(errors.KindDomain, "llm_manager", fmt.Sprintf("failed to get executor for capability %s (type: %s)", capabilityID, llmCfg.Type), err)
	}

	streamExecutor, ok := executor.(capability.StreamExecutor)
	if !ok {
		return nil, errors.New(errors.KindDomain, "llm_manager", "executor does not support streaming")
	}

	// 5. Execute Stream
	pluginStream, err := streamExecutor.ExecuteStream(ctx, pluginConfig, inputs)
	if err != nil {
		return nil, errors.Wrap(errors.KindDomain, "llm_manager", "plugin stream execution failed", err)
	}

	// 6. Map Output Stream
	outCh := make(chan repository.ResponseChunk)
	go func() {
		defer close(outCh)
		for output := range pluginStream {
			content, _ := output["content"].(string)
			done, _ := output["done"].(bool)
			
			chunk := repository.ResponseChunk{
				Content: content,
				Done:    done,
			}
			outCh <- chunk
		}
	}()

	return outCh, nil
}

func (m *LLMManager) ValidateConnection(ctx context.Context, config aggregate.Config) error {
	return nil
}

func (m *LLMManager) GetProviderInfo(providerID string) (*repository.ProviderInfo, error) {
	// This is a bit fake now, as we don't have provider instances anymore.
	// We just return the type from config.
	llmCfg, ok := m.config.LLM[providerID]
	if !ok {
		return nil, errors.New(errors.KindDomain, "llm_manager", fmt.Sprintf("provider config not found: %s", providerID))
	}
	return &repository.ProviderInfo{
		Name: llmCfg.Type,
	}, nil
}

func convertMessagesToPlugin(msgs []repository.Message) []interface{} {
	result := make([]interface{}, len(msgs))
	for i, m := range msgs {
		result[i] = map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
	}
	return result
}

func (m *LLMManager) resolveCapabilityID(providerType string) string {
	switch providerType {
	case "openai", "doubao", "ollama":
		return "openai_chat"
	// Future: case "coze": return "coze_chat"
	default:
		// Fallback or return type as is if we assume 1:1 mapping
		return "openai_chat"
	}
}
