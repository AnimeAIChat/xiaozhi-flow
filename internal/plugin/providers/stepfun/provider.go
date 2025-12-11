package stepfun

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
	p.BaseGRPCProvider = server.NewBaseGRPCProvider("stepfun", logger, func() pluginpb.PluginServiceServer {
		return NewGRPCServer(p, logger)
	})
	return p
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

func (e *ASRExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	audioStream, ok := inputs["audio_stream"].(<-chan []byte)
	if !ok {
		return nil, fmt.Errorf("audio_stream input is required")
	}

	apiKey, _ := config["api_key"].(string)
	model, _ := config["model"].(string)
	voice, _ := config["voice"].(string)
	prompt, _ := config["prompt"].(string)

	asrConfig := &ASRConfig{
		APIKey: apiKey,
		Model:  model,
		Voice:  voice,
		Prompt: prompt,
	}

	outCh := make(chan map[string]interface{}, 10)

	go func() {
		defer close(outCh)

		provider := NewASRProvider(asrConfig, outCh)
		if err := provider.Start(ctx, audioStream); err != nil {
			outCh <- map[string]interface{}{"error": err.Error()}
			return
		}

		<-ctx.Done()
	}()

	return outCh, nil
}
