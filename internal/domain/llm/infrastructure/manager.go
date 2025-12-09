package infrastructure

import (
	"context"
	"fmt"
	"sync"

	"xiaozhi-server-go/internal/domain/llm/aggregate"
	"xiaozhi-server-go/internal/domain/llm/provider"
	"xiaozhi-server-go/internal/domain/llm/repository"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/errors"
)

type LLMManager struct {
	providers map[string]provider.Provider
	mu        sync.RWMutex
}

func NewLLMManager(cfg *config.Config) (repository.LLMRepository, error) {
	m := &LLMManager{
		providers: make(map[string]provider.Provider),
	}

	// Initialize providers from config
	for id, llmCfg := range cfg.LLM {
		p, err := createProvider(id, llmCfg)
		if err != nil {
			// Log error but continue? Or fail?
			// For now, let's just print it and continue
			fmt.Printf("Failed to create LLM provider %s: %v\n", id, err)
			continue
		}
		m.providers[id] = p
	}

	return m, nil
}

func createProvider(id string, cfg config.LLMConfig) (provider.Provider, error) {
	switch cfg.Type {
	case "openai", "doubao", "ollama":
		return NewOpenAIProvider(id, cfg), nil
	// Add other types here
	default:
		return nil, fmt.Errorf("unsupported LLM provider type: %s", cfg.Type)
	}
}

func (m *LLMManager) Generate(ctx context.Context, req repository.GenerateRequest) (*repository.GenerateResult, error) {
	p, err := m.getProvider(req.Config.Provider)
	if err != nil {
		return nil, err
	}
	return p.Generate(ctx, req)
}

func (m *LLMManager) Stream(ctx context.Context, req repository.GenerateRequest) (<-chan repository.ResponseChunk, error) {
	p, err := m.getProvider(req.Config.Provider)
	if err != nil {
		return nil, err
	}
	return p.Stream(ctx, req)
}

func (m *LLMManager) ValidateConnection(ctx context.Context, config aggregate.Config) error {
	// This might need to be adjusted. 
	// If config.Provider is set, we validate that provider.
	// Or we create a temporary provider to validate the connection parameters?
	if config.Provider != "" {
		_, err := m.getProvider(config.Provider)
		return err
	}
	return nil
}

func (m *LLMManager) GetProviderInfo(providerID string) (*repository.ProviderInfo, error) {
	p, err := m.getProvider(providerID)
	if err != nil {
		return nil, err
	}
	return &repository.ProviderInfo{
		Name: p.Type(),
	}, nil
}

func (m *LLMManager) getProvider(id string) (provider.Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.providers[id]
	if !ok {
		return nil, errors.New(errors.KindDomain, "llm_manager", fmt.Sprintf("provider not found: %s", id))
	}
	return p, nil
}
