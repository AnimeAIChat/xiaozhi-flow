package gosherpa

import (
	"context"
	"fmt"

	"xiaozhi-server-go/internal/core/providers/asr"
	gosherpaasr "xiaozhi-server-go/internal/core/providers/asr/gosherpa"
	"xiaozhi-server-go/internal/core/providers/tts"
	gosherpatts "xiaozhi-server-go/internal/core/providers/tts/gosherpa"
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
			ID:          "gosherpa_tts",
			Type:        capability.TypeTTS,
			Name:        "GoSherpa TTS",
			Description: "GoSherpa Text to Speech",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"cluster": {Type: "string", Default: "ws://localhost:8888", Description: "WebSocket Address"},
				},
				Required: []string{"cluster"},
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
			ID:          "gosherpa_asr",
			Type:        capability.TypeASR,
			Name:        "GoSherpa ASR",
			Description: "GoSherpa Automatic Speech Recognition",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"addr": {Type: "string", Default: "ws://localhost:8889", Description: "WebSocket Address"},
				},
				Required: []string{"addr"},
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
	case "gosherpa_tts":
		return &TTSExecutor{}, nil
	case "gosherpa_asr":
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
		Type:    "gosherpa",
		Cluster: getString(config, "cluster"),
		OutputDir: "data/tmp",
	}
	if ttsConfig.Cluster == "" {
		ttsConfig.Cluster = "ws://localhost:8888"
	}

	provider, err := gosherpatts.NewProvider(ttsConfig, false)
	if err != nil {
		return nil, err
	}

	filepath, err := provider.ToTTS(text)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"file_path": filepath,
	}, nil
}

func (e *TTSExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	return nil, fmt.Errorf("gosherpa_tts does not support streaming in this wrapper yet")
}

// --- ASR Executor ---

type ASRExecutor struct{}

func (e *ASRExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("gosherpa_asr only supports streaming via ExecuteStream")
}

// Listener adapter
type asrListener struct {
	outputChan chan<- map[string]interface{}
	provider   *gosherpaasr.Provider
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

		addr := getString(config, "addr")
		if addr == "" {
			addr = "ws://localhost:8889"
		}

		// Map config
		asrConfig := &asr.Config{
			Type: "gosherpa",
			Data: map[string]interface{}{
				"addr": addr,
			},
		}

		// Create logger
		logger, _ := utils.NewLogger(&utils.LogCfg{
			LogLevel: "info",
		})

		// Create provider
		provider, err := gosherpaasr.NewProvider(asrConfig, false, logger)
		if err != nil {
			outputChan <- map[string]interface{}{"error": err.Error()}
			return
		}
		defer provider.CloseConnection()

		// Set listener
		// Note: gosherpa legacy provider uses PublishAsrResult which calls listener.OnAsrResult
		listener := &asrListener{
			outputChan: outputChan,
			provider:   provider,
		}
		provider.SetListener(listener)

		// Note: gosherpa legacy provider starts reading loop in NewProvider goroutine.
		// We just need to feed audio.

		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-audioStream:
				if !ok {
					return
				}
				if err := provider.AddAudio(data); err != nil {
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
