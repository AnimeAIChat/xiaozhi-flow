package doubao

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
			ID:          "doubao_llm",
			Type:        capability.TypeLLM,
			Name:        "Doubao LLM",
			Description: "Doubao Large Language Model",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"api_key":   {Type: "string", Secret: true, Description: "API Key"},
					"base_url":  {Type: "string", Description: "API Base URL"},
					"model":     {Type: "string", Default: "doubao-pro-4k", Description: "Model ID (endpoint ID)"},
					"max_tokens": {Type: "number", Default: 2048},
				},
				Required: []string{"api_key", "model"},
			},
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"messages": {Type: "array"},
				},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"content": {Type: "string"},
				},
			},
		},
		{
			ID:          "doubao_tts",
			Type:        capability.TypeTTS,
			Name:        "Doubao TTS",
			Description: "Doubao Text to Speech",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"app_id":  {Type: "string", Description: "App ID"},
					"token":   {Type: "string", Secret: true, Description: "Access Token"},
					"cluster": {Type: "string", Description: "Cluster ID"},
					"voice":   {Type: "string", Default: "zh_female_shentong_mars_bigtts", Description: "Voice ID"},
				},
				Required: []string{"app_id", "token", "cluster"},
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
			ID:          "doubao_asr",
			Type:        capability.TypeASR,
			Name:        "Doubao ASR",
			Description: "Doubao Automatic Speech Recognition",
			ConfigSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"appid":        {Type: "string", Description: "App ID"},
					"access_token": {Type: "string", Secret: true, Description: "Access Token"},
					"cluster":      {Type: "string", Description: "Cluster ID"},
				},
				Required: []string{"appid", "access_token", "cluster"},
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
	case "doubao_llm":
		return &LLMExecutor{}, nil
	case "doubao_tts":
		return &TTSExecutor{}, nil
	case "doubao_asr":
		return &ASRExecutor{}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}

// --- LLM Executor ---

type LLMExecutor struct{}

func (e *LLMExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("doubao_llm only supports streaming via ExecuteStream")
}

func (e *LLMExecutor) ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error) {
	apiKey, _ := config["api_key"].(string)
	baseURL, _ := config["base_url"].(string)
	model, _ := config["model"].(string)
	maxTokens := 2048
	if mt, ok := config["max_tokens"].(float64); ok {
		maxTokens = int(mt)
	} else if mt, ok := config["max_tokens"].(int); ok {
		maxTokens = mt
	}

	llmConfig := &LLMConfig{
		APIKey:    apiKey,
		BaseURL:   baseURL,
		Model:     model,
		MaxTokens: maxTokens,
	}

	provider := NewLLMProvider(llmConfig)

	// Parse messages
	msgsRaw, ok := inputs["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("messages input is required")
	}

	var messages []Message
	for _, m := range msgsRaw {
		if msgMap, ok := m.(map[string]interface{}); ok {
			role, _ := msgMap["role"].(string)
			content, _ := msgMap["content"].(string)
			msg := Message{
				Role:    role,
				Content: content,
			}
			
			if tcID, ok := msgMap["tool_call_id"].(string); ok {
				msg.ToolCallID = tcID
			}
			
			if tcsRaw, ok := msgMap["tool_calls"].([]interface{}); ok {
				var tcs []ToolCall
				for _, tcRaw := range tcsRaw {
					if tcMap, ok := tcRaw.(map[string]interface{}); ok {
						tc := ToolCall{
							ID:   getString(tcMap, "id"),
							Type: getString(tcMap, "type"),
						}
						if fnMap, ok := tcMap["function"].(map[string]interface{}); ok {
							tc.Function = ToolCallFunction{
								Name:      getString(fnMap, "name"),
								Arguments: getString(fnMap, "arguments"),
							}
						}
						tcs = append(tcs, tc)
					}
				}
				msg.ToolCalls = tcs
			}
			
			messages = append(messages, msg)
		}
	}

	var tools []Tool
	if toolsRaw, ok := inputs["tools"].([]interface{}); ok {
		for _, t := range toolsRaw {
			if tMap, ok := t.(map[string]interface{}); ok {
				tool := Tool{
					Type: getString(tMap, "type"),
				}
				if fnMap, ok := tMap["function"].(map[string]interface{}); ok {
					tool.Function = ToolFunction{
						Name:        getString(fnMap, "name"),
						Description: getString(fnMap, "description"),
						Parameters:  fnMap["parameters"],
					}
				}
				tools = append(tools, tool)
			}
		}
	}

	stream, err := provider.Chat(ctx, messages, tools)
	if err != nil {
		return nil, err
	}

	outCh := make(chan map[string]interface{})
	go func() {
		defer close(outCh)
		for resp := range stream {
			if resp.Error != nil {
				// Optionally handle error
				continue
			}
			
			outMap := map[string]interface{}{}
			if resp.Content != "" {
				outMap["content"] = resp.Content
			}
			
			if len(resp.ToolCalls) > 0 {
				tcs := make([]interface{}, len(resp.ToolCalls))
				for i, tc := range resp.ToolCalls {
					tcs[i] = map[string]interface{}{
						"id":   tc.ID,
						"type": tc.Type,
						"function": map[string]interface{}{
							"name":      tc.Function.Name,
							"arguments": tc.Function.Arguments,
						},
					}
				}
				outMap["tool_calls"] = tcs
			}
			
			if len(outMap) > 0 {
				outCh <- outMap
			}
		}
	}()

	return outCh, nil
}



// --- TTS Executor ---

type TTSExecutor struct{}

func (e *TTSExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	text, ok := inputs["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text input is required")
	}

	ttsConfig := &TTSConfig{
		Type:      "doubao",
		AppID:     getString(config, "app_id"),
		Token:     getString(config, "token"),
		Cluster:   getString(config, "cluster"),
		Voice:     getString(config, "voice"),
		OutputDir: "data/tmp",
	}

	provider, err := NewTTSProvider(ttsConfig, false)
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

// --- ASR Executor ---

type ASRExecutor struct{}

func (e *ASRExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("doubao_asr only supports streaming via ExecuteStream")
}

// Listener adapter
type asrListener struct {
	outputChan chan<- map[string]interface{}
	provider   *ASRProvider
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
		asrConfig := &ASRConfig{
			Type: "doubao",
			Data: map[string]interface{}{
				"appid":        getString(config, "appid"),
				"access_token": getString(config, "access_token"),
				"cluster":      getString(config, "cluster"),
			},
		}

		// Create logger
		logger, err := logging.New(logging.Config{
			Level:    "info",
			Dir:      "data/logs",
			Filename: "doubao_asr.log",
		})
		if err != nil {
			outputChan <- map[string]interface{}{"error": fmt.Sprintf("failed to initialize logger: %v", err)}
			return
		}

		// Create provider
		provider, err := NewASRProvider(asrConfig, false, logger, nil)
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

// GetPluginID 返回插件ID
func (p *Provider) GetPluginID() string {
	return "doubao"
}

// StartGRPCServer 启动Doubao插件的gRPC服务
func (p *Provider) StartGRPCServer(address string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.grpcServer != nil {
		return fmt.Errorf("Doubao gRPC server already started at %s", p.serviceAddress)
	}

	if p.logger != nil {
		p.logger.InfoTag("gRPC", "启动Doubao插件gRPC服务器",
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
				p.logger.ErrorTag("gRPC", "Doubao gRPC服务器启动失败",
					"address", address,
					"error", err.Error())
			}
		} else {
			p.mu.Lock()
			p.serviceAddress = address
			p.mu.Unlock()
			if p.logger != nil {
				p.logger.InfoTag("gRPC", "Doubao插件gRPC服务器启动成功",
					"address", address)
			}
		}
	}()

	return nil
}

// StopGRPCServer 停止Doubao插件的gRPC服务器
func (p *Provider) StopGRPCServer() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.grpcServer == nil {
		return fmt.Errorf("Doubao gRPC server not started")
	}

	if p.logger != nil {
		p.logger.InfoTag("gRPC", "停止Doubao插件gRPC服务器",
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
