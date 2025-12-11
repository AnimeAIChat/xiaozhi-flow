package edge

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

// GetPluginID 返回插件ID
func (p *Provider) GetPluginID() string {
	return "edge"
}

// StartGRPCServer 启动Edge插件的gRPC服务
func (p *Provider) StartGRPCServer(address string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.grpcServer != nil {
		return fmt.Errorf("Edge gRPC server already started at %s", p.serviceAddress)
	}

	if p.logger != nil {
		p.logger.InfoTag("gRPC", "启动Edge插件gRPC服务器",
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
				p.logger.ErrorTag("gRPC", "Edge gRPC服务器启动失败",
					"address", address,
					"error", err.Error())
			}
		} else {
			p.mu.Lock()
			p.serviceAddress = address
			p.mu.Unlock()
			if p.logger != nil {
				p.logger.InfoTag("gRPC", "Edge插件gRPC服务器启动成功",
					"address", address)
			}
		}
	}()

	return nil
}

// StopGRPCServer 停止Edge插件的gRPC服务器
func (p *Provider) StopGRPCServer() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.grpcServer == nil {
		return fmt.Errorf("Edge gRPC server not started")
	}

	if p.logger != nil {
		p.logger.InfoTag("gRPC", "停止Edge插件gRPC服务器",
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
