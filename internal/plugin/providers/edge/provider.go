package edge

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
			ID:          "edge_tts",
			Type:        capability.TypeTTS,
			Name:        "Edge TTS",
			Description: "Microsoft Edge Text to Speech",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"voice": {Type: "string", Default: "zh-CN-XiaoxiaoNeural", Description: "Voice ID"},
				},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"text": {Type: "string"},
				},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"file_path": {Type: "string"},
				},
			},
		},
	}
}

func (p *Provider) CreateExecutor(capabilityID string) (capability.Executor, error) {
	switch capabilityID {
	case "edge_tts":
		return &TTSExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

type TTSExecutor struct{}

func (e *TTSExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	text, ok := inputs["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text input is required")
	}

	voice, _ := config["voice"].(string)
	if voice == "" {
		voice = "zh-CN-XiaoxiaoNeural"
	}

	ttsConfig := &TTSConfig{
		Voice:     voice,
		OutputDir: "data/tmp",
	}

	filepath, err := synthesizeSpeech(ttsConfig, text)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"file_path": filepath,
	}, nil
}

func (e *TTSExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	return nil, fmt.Errorf("edge_tts does not support streaming in this wrapper yet")
}
