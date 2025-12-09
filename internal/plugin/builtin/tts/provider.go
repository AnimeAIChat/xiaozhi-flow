package tts

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
			ID:          "builtin_tts",
			Type:        capability.TypeTTS,
			Name:        "Builtin TTS",
			Description: "Text to Speech",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"engine": {Type: "string", Default: "doubao", Description: "TTS Engine"},
					"voice":  {Type: "string", Description: "Voice ID"},
				},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"text": {Type: "string", Description: "Text to speak"},
				},
				Required: []string{"text"},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"audio": {Type: "string", Description: "Base64 encoded audio data"},
				},
			},
		},
	}
}

func (p *Provider) CreateExecutor(capabilityID string) (capability.Executor, error) {
	switch capabilityID {
	case "builtin_tts":
		return &TTSExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

type TTSExecutor struct{}

func (e *TTSExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	// TODO: Implement actual TTS logic
	return map[string]interface{}{
		"audio": "", // Placeholder
	}, nil
}

func (e *TTSExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
