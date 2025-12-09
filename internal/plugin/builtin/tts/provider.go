package tts

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"xiaozhi-server-go/internal/core/providers/tts"
	"xiaozhi-server-go/internal/core/providers/tts/doubao"
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
					"app_id": {Type: "string", Description: "App ID"},
					"token":  {Type: "string", Secret: true, Description: "Access Token"},
					"cluster": {Type: "string", Description: "Cluster ID"},
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
					"file_path": {Type: "string", Description: "Path to the generated audio file"},
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
	// 1. Parse Config
	engine, _ := config["engine"].(string)
	if engine == "" {
		engine = "doubao"
	}

	// 2. Parse Inputs
	text, ok := inputs["text"].(string)
	if !ok || text == "" {
		return nil, fmt.Errorf("text input is required")
	}

	// 3. Execute TTS
	if engine == "doubao" {
		return e.executeDoubao(ctx, config, text)
	}

	return nil, fmt.Errorf("unsupported engine: %s", engine)
}

func (e *TTSExecutor) executeDoubao(ctx context.Context, config map[string]interface{}, text string) (map[string]interface{}, error) {
	// Map config to legacy struct
	ttsConfig := &tts.Config{
		Type:    "doubao",
		AppID:   getString(config, "app_id"),
		Token:   getString(config, "token"),
		Cluster: getString(config, "cluster"),
		Voice:   getString(config, "voice"),
		OutputDir: "data/tmp", // Default output dir
	}

	provider, err := doubao.NewProvider(ttsConfig, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create doubao provider: %w", err)
	}

	filePath, err := provider.ToTTS(text)
	if err != nil {
		return nil, fmt.Errorf("tts execution failed: %w", err)
	}

	// Read file content for base64 output (optional, but good for API)
	audioData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	return map[string]interface{}{
		"audio": base64.StdEncoding.EncodeToString(audioData),
		"file_path": filePath,
	}, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (e *TTSExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
