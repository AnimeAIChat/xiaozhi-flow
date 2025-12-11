package deepgram

import (
	"context"
	"fmt"
	"sync"

	"xiaozhi-server-go/internal/plugin/capability"
	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/plugin/grpc/server"
)

type Provider struct {
	logger        *logging.Logger
	grpcServer    *server.GRPCServer
	grpcService   *GRPCServer
	serviceAddress string
	mu           sync.RWMutex
}

func NewProvider() *Provider {
	return NewProviderWithLogger(nil)
}

func NewProviderWithLogger(logger *logging.Logger) *Provider {
	if logger == nil {
		logger = logging.DefaultLogger
	}
	return &Provider{
		logger: logger,
	}
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

// GetPluginID 返回插件ID
func (p *Provider) GetPluginID() string {
	return "deepgram"
}

// StartGRPCServer 启动Deepgram插件的gRPC服务
func (p *Provider) StartGRPCServer(address string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.grpcServer != nil {
		return fmt.Errorf("Deepgram gRPC server already started at %s", p.serviceAddress)
	}

	if p.logger != nil {
		p.logger.InfoTag("gRPC", "启动Deepgram插件gRPC服务器",
			"address", address)
	}

	// 创建gRPC服务器
	p.grpcServer = server.NewGRPCServer(address, p.logger)

	// 创建gRPC服务实例
	p.grpcService = NewGRPCServer(p, p.logger)

	// 注册服务
	p.grpcServer.RegisterPluginService(p.grpcService)

	// 启用反射（用于调试）
	p.grpcServer.EnableReflection()

	// 启动服务器
	go func() {
		if err := p.grpcServer.Start(); err != nil {
			if p.logger != nil {
				p.logger.ErrorTag("gRPC", "Deepgram gRPC服务器启动失败",
					"address", address,
					"error", err.Error())
			}
		} else {
			p.mu.Lock()
			p.serviceAddress = address
			p.mu.Unlock()
			if p.logger != nil {
				p.logger.InfoTag("gRPC", "Deepgram插件gRPC服务器启动成功",
					"address", address)
			}
		}
	}()

	return nil
}

// StopGRPCServer 停止Deepgram插件的gRPC服务器
func (p *Provider) StopGRPCServer() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.grpcServer == nil {
		return fmt.Errorf("Deepgram gRPC server not started")
	}

	if p.logger != nil {
		p.logger.InfoTag("gRPC", "停止Deepgram插件gRPC服务器",
			"address", p.serviceAddress)
	}

	p.grpcServer.Stop()

	p.grpcServer = nil
	p.grpcService = nil
	p.serviceAddress = ""

	return nil
}

// GetServiceAddress 获取gRPC服务地址
func (p *Provider) GetServiceAddress() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.serviceAddress
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
