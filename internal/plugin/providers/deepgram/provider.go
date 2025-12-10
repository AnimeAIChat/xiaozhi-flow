package deepgram

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
			ID:          "deepgram_tts",
			Type:        capability.TypeTTS,
			Name:        "Deepgram TTS",
			Description: "Deepgram Text to Speech",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"token":   {Type: "string", Secret: true, Description: "API Token"},
					"voice":   {Type: "string", Default: "aura-asteria-en", Description: "Voice ID"},
					"cluster": {Type: "string", Default: "wss://api.deepgram.com/v1/speak", Description: "API Endpoint"},
				},
				Required: []string{"token"},
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
		{
			ID:          "deepgram_asr",
			Type:        capability.TypeASR,
			Name:        "Deepgram ASR",
			Description: "Deepgram Automatic Speech Recognition",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"api_key": {Type: "string", Secret: true, Description: "API Key"},
					"lang":    {Type: "string", Default: "en", Description: "Language Code"},
				},
				Required: []string{"api_key"},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"audio_stream": {Type: "object"},
				},
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
	case "deepgram_tts":
		return &TTSExecutor{}, nil
	case "deepgram_asr":
		return &ASRExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

// --- TTS Executor ---

type TTSExecutor struct{}

func (e *TTSExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	text, ok := inputs["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text input is required")
	}

	ttsConfig := &TTSConfig{
		Token:     getString(config, "token"),
		Voice:     getString(config, "voice"),
		Cluster:   getString(config, "cluster"),
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
	return nil, fmt.Errorf("deepgram_tts does not support streaming in this wrapper yet")
}

// --- ASR Executor ---

type ASRExecutor struct{}

func (e *ASRExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("deepgram_asr only supports streaming via ExecuteStream")
}

func (e *ASRExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	// Get audio stream
	audioStream, ok := inputs["audio_stream"].(<-chan []byte)
	if !ok {
		return nil, fmt.Errorf("audio_stream input is required and must be <-chan []byte")
	}

	// Create output channel
	outputChan := make(chan map[string]interface{}, 10)

	// Start processing in goroutine
	go func() {
		defer close(outputChan)

		// Map config
		asrConfig := &ASRConfig{
			APIKey:   getString(config, "api_key"),
			Language: getString(config, "lang"),
		}

		// Create provider
		provider := NewASRProvider(asrConfig, outputChan)

		// Start streaming
		if err := provider.Start(ctx, audioStream); err != nil {
			outputChan <- map[string]interface{}{"error": err.Error()}
			return
		}

		// Wait for context done (Start runs in background goroutines)
		<-ctx.Done()
	}()

	return outputChan, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
