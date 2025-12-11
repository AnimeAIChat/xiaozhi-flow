package gosherpa

import (
	"context"
	"fmt"

	pluginpb "xiaozhi-server-go/gen/go/api/proto"
	"xiaozhi-server-go/internal/plugin/capability"
	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/plugin/grpc/server"
)

type Provider struct {
	*server.BaseGRPCProvider
	logger *logging.Logger
}

func NewProvider() *Provider {
	return NewProviderWithLogger(nil)
}

func NewProviderWithLogger(logger *logging.Logger) *Provider {
	if logger == nil {
		logger = logging.DefaultLogger
	}
	p := &Provider{
		logger: logger,
	}
	p.BaseGRPCProvider = server.NewBaseGRPCProvider("gosherpa", logger, func() pluginpb.PluginServiceServer {
		return NewGRPCServer(p, logger)
	})
	return p
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

	ttsConfig := &TTSConfig{
		Cluster:   getString(config, "cluster"),
		OutputDir: "data/tmp",
	}
	if ttsConfig.Cluster == "" {
		ttsConfig.Cluster = "ws://localhost:8888"
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
	return nil, fmt.Errorf("gosherpa_tts does not support streaming in this wrapper yet")
}

// --- ASR Executor ---

type ASRExecutor struct{}

func (e *ASRExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("gosherpa_asr only supports streaming via ExecuteStream")
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
		asrConfig := &ASRConfig{
			Cluster: addr,
		}

		// Create provider
		provider := NewASRProvider(asrConfig, outputChan)

		// Start streaming
		if err := provider.Start(ctx, audioStream); err != nil {
			outputChan <- map[string]interface{}{"error": err.Error()}
			return
		}

		// Wait for context done
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
