package webrtc

import (
	"context"
	"fmt"

	"xiaozhi-server-go/internal/domain/vad/webrtc_vad"
	"xiaozhi-server-go/internal/plugin/capability"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) GetCapabilities() []capability.Definition {
	return []capability.Definition{
		{
			ID:          "webrtc_vad",
			Type:        capability.TypeVAD,
			Name:        "WebRTC VAD",
			Description: "WebRTC Voice Activity Detection",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"sample_rate": {Type: "number", Default: 16000},
					"channels":    {Type: "number", Default: 1},
					"mode":        {Type: "number", Default: 3, Description: "VAD Mode (0-3)"},
				},
				Required: []string{"sample_rate"},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"audio_chunk": {Type: "object"}, // []byte
				},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"is_speech": {Type: "boolean"},
				},
			},
		},
	}
}

func (p *Provider) CreateExecutor(capabilityID string) (capability.Executor, error) {
	switch capabilityID {
	case "webrtc_vad":
		return &VADExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

type VADExecutor struct{}

func (e *VADExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	sampleRate := 16000
	if sr, ok := config["sample_rate"].(float64); ok {
		sampleRate = int(sr)
	} else if sr, ok := config["sample_rate"].(int); ok {
		sampleRate = sr
	}

	channels := 1
	if ch, ok := config["channels"].(float64); ok {
		channels = int(ch)
	} else if ch, ok := config["channels"].(int); ok {
		channels = ch
	}
    
    vadConfig := map[string]interface{}{
        "sample_rate": float64(sampleRate),
        "channels":    float64(channels),
    }
    // Add other config
    for k, v := range config {
        vadConfig[k] = v
    }

    vadInstance, err := webrtc_vad.AcquireVAD(vadConfig)
    if err != nil {
        return nil, err
    }
    
    audioChunk, ok := inputs["audio_chunk"].([]byte)
    if !ok {
        return nil, fmt.Errorf("audio_chunk input is required")
    }

    isSpeech, err := vadInstance.ProcessAudio(audioChunk)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "is_speech": isSpeech,
    }, nil
}

func (e *VADExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
    return nil, fmt.Errorf("webrtc_vad does not support streaming")
}
