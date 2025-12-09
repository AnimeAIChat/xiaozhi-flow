package deepgram

import (
	"context"
	"fmt"

	"xiaozhi-server-go/internal/core/providers/asr"
	deepgramasr "xiaozhi-server-go/internal/core/providers/asr/deepgram"
	"xiaozhi-server-go/internal/core/providers/tts"
	deepgramtts "xiaozhi-server-go/internal/core/providers/tts/deepgram"
	"xiaozhi-server-go/internal/plugin/capability"
	"xiaozhi-server-go/internal/utils"
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

	ttsConfig := &tts.Config{
		Type:    "deepgram",
		Token:   getString(config, "token"),
		Voice:   getString(config, "voice"),
		Cluster: getString(config, "cluster"),
		OutputDir: "data/tmp",
	}
	if ttsConfig.Cluster == "" {
		ttsConfig.Cluster = "wss://api.deepgram.com/v1/speak"
	}

	provider, err := deepgramtts.NewProvider(ttsConfig, false)
	if err != nil {
		return nil, err
	}

	// The legacy provider's ToTTS method returns file path
	filepath, err := provider.ToTTS(text)
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

// Listener adapter
type asrListener struct {
	outputChan chan<- map[string]interface{}
	provider   *deepgramasr.Provider
}

func (l *asrListener) OnAsrResult(result string, isFinalResult bool) bool {
	silenceCount := 0
	if l.provider != nil {
		silenceCount = l.provider.GetSilenceCount()
	}
	l.outputChan <- map[string]interface{}{
		"text":          result,
		"is_final":      isFinalResult,
		"silence_count": silenceCount,
	}
	return false // Continue recognition
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
		asrConfig := &asr.Config{
			Type: "deepgram",
			Data: map[string]interface{}{
				"api_key": getString(config, "api_key"),
				"lang":    getString(config, "lang"),
			},
		}

		// Create logger
		logger, _ := utils.NewLogger(&utils.LogCfg{
			LogLevel: "info",
		})

		// Create provider
		provider, err := deepgramasr.NewProvider(asrConfig, false, logger)
		if err != nil {
			outputChan <- map[string]interface{}{"error": err.Error()}
			return
		}
		defer provider.Cleanup()

		// Set listener
		listener := &asrListener{
			outputChan: outputChan,
			provider:   provider,
		}
		provider.SetListener(listener)

		// Start streaming
		if err := provider.StartStreaming(ctx); err != nil {
			outputChan <- map[string]interface{}{"error": err.Error()}
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-audioStream:
				if !ok {
					// provider.StopStreaming() // Not available
					return
				}
				if err := provider.AddAudioWithContext(ctx, data); err != nil {
					outputChan <- map[string]interface{}{"error": err.Error()}
				}
			}
		}
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
