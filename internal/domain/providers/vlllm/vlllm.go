package vlllm

import (
	"context"
	"fmt"
	providers "xiaozhi-server-go/internal/domain/providers/types"
	domainimage "xiaozhi-server-go/internal/domain/image"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/utils"
)

type Provider struct {
	config *config.VLLLMConfig
	logger *utils.Logger
}

func Create(providerType string, config *config.VLLLMConfig, logger *utils.Logger) (*Provider, error) {
	return &Provider{
		config: config,
		logger: logger,
	}, nil
}

func (p *Provider) Cleanup() error {
	return nil
}

func (p *Provider) Initialize() error {
	return nil
}

func (p *Provider) ResponseWithImage(ctx context.Context, sessionID string, messages []providers.Message, imageData domainimage.ImageData, text string) (<-chan string, error) {
	return nil, fmt.Errorf("VLLLM provider is migrated to plugins. Please update configuration to use LLM manager.")
}

func (p *Provider) stats() map[string]int64 {
	return map[string]int64{}
}
