package stepfun

import (
	"context"
	"fmt"
	"xiaozhi-server-go/internal/core/providers/asr"
	stepfunasr "xiaozhi-server-go/internal/core/providers/asr/stepfun"
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
			ID:          "step_asr",
			Type:        capability.TypeASR,
			Name:        "StepFun ASR",
			Description: "StepFun Realtime ASR",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"api_key": {Type: "string", Description: "StepFun API Key"},
					"model":   {Type: "string", Default: "step-asr", Description: "Model name"},
					"voice":   {Type: "string", Default: "cixing", Description: "Voice ID"},
					"prompt":  {Type: "string", Description: "System prompt"},
				},
				Required: []string{"api_key"},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"audio_stream": {Type: "channel"}, // Stream of []byte
				},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"text":          {Type: "string"},
					"is_final":      {Type: "boolean"},
					"silence_count": {Type: "integer"},
				},
			},
		},
	}
}

func (p *Provider) CreateExecutor(capabilityID string) (capability.Executor, error) {
	switch capabilityID {
	case "step_asr":
		return &ASRExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

type ASRExecutor struct{}

func (e *ASRExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("step_asr only supports streaming execution")
}

type asrListener struct {
	outCh    chan<- map[string]interface{}
	provider *stepfunasr.Provider
}

func (l *asrListener) OnAsrResult(result string, isFinalResult bool) bool {
	select {
	case l.outCh <- map[string]interface{}{
		"text":          result,
		"is_final":      isFinalResult,
		"silence_count": l.provider.GetSilenceCount(),
	}:
	default:
	}
	return true
}

func (e *ASRExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	audioStream, ok := inputs["audio_stream"].(<-chan []byte)
	if !ok {
		return nil, fmt.Errorf("audio_stream input is required")
	}

	apiKey, _ := config["api_key"].(string)
	model, _ := config["model"].(string)
	if model == "" {
		model = "step-asr"
	}
	voice, _ := config["voice"].(string)
	if voice == "" {
		voice = "cixing"
	}
	prompt, _ := config["prompt"].(string)

	asrConfig := &asr.Config{
		Type: "stepfun",
		Data: map[string]interface{}{
			"api_key": apiKey,
			"model":   model,
			"voice":   voice,
			"prompt":  prompt,
		},
	}

	logger, _ := utils.NewLogger(&utils.LogCfg{
		LogLevel: "info",
	})
	provider, err := stepfunasr.NewProvider(asrConfig, false, logger)
	if err != nil {
		return nil, err
	}

	outCh := make(chan map[string]interface{})
	listener := &asrListener{
		outCh:    outCh,
		provider: provider,
	}
	provider.SetListener(listener)
	provider.EnableSilenceDetection(true)

	go func() {
		defer close(outCh)
		defer provider.Cleanup()

		// Start streaming explicitly if needed, or it might be lazy loaded in AddAudio
		// provider.StartStreaming(ctx) // stepfun provider handles this in AddAudioWithContext

		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-audioStream:
				if !ok {
					return
				}
				if err := provider.AddAudioWithContext(ctx, data); err != nil {
					// Log error or send error event?
					// For now just continue or break
				}
			}
		}
	}()

	return outCh, nil
}
