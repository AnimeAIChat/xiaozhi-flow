package asr

import (
	"context"
	"fmt"
	"xiaozhi-server-go/internal/plugin/capability"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) GetCapabilities() []capability.Definition {
	return []capability.Definition{
		{
			ID:          "builtin_asr",
			Type:        capability.TypeASR,
			Name:        "Builtin ASR",
			Description: "Automatic Speech Recognition",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"engine": {Type: "string", Default: "doubao", Description: "ASR Engine (doubao, openai, etc.)"},
				},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"audio": {Type: "string", Description: "Base64 encoded audio data"},
				},
				Required: []string{"audio"},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"text": {Type: "string"},
				},
			},
		},
	}
}

func (p *Provider) CreateExecutor(capabilityID string) (capability.Executor, error) {
	switch capabilityID {
	case "builtin_asr":
		return &ASRExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

type ASRExecutor struct{}

func (e *ASRExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	// TODO: Implement actual ASR logic
	return map[string]interface{}{
		"text": "ASR Placeholder",
	}, nil
}

func (e *ASRExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
