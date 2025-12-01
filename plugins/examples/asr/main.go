package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	v1 "xiaozhi-server-go/api/v1"
	sdk "xiaozhi-server-go/internal/plugin/sdk"
)

// ASRPlugin 语音识别插件
type ASRPlugin struct {
	sdk.SimplePluginImpl
	logger hclog.Logger
}

// NewASRPlugin 创建ASR插件实例
func NewASRPlugin(logger hclog.Logger) *ASRPlugin {
	info := &v1.PluginInfo{
		ID:          "asr-plugin",
		Name:        "ASR Plugin",
		Version:     "1.0.0",
		Description: "语音识别插件，支持多种音频格式转文字",
		Author:      "XiaoZhi Flow Team",
		Type:        v1.PluginTypeAudio,
		Tags:        []string{"asr", "speech", "recognition", "audio"},
		Capabilities: []string{"speech_to_text", "audio_format_detect", "batch_processing"},
		Dependencies: []string{"ffmpeg", "whisper-api"},
	}

	return &ASRPlugin{
		SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
		logger:          logger.Named("asr-plugin"),
	}
}

// CallTool 实现工具调用
func (p *ASRPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	startTime := time.Now()
	p.IncrementCounter("asr.calls.total")

	switch req.ToolName {
	case "speech_to_text":
		return p.handleSpeechToText(ctx, req)
	case "detect_audio_format":
		return p.handleDetectAudioFormat(ctx, req)
	case "batch_transcribe":
		return p.handleBatchTranscribe(ctx, req)
	case "get_supported_formats":
		return p.handleGetSupportedFormats(ctx, req)
	default:
		p.IncrementCounter("asr.calls.unknown")
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "UNKNOWN_TOOL",
				Message: fmt.Sprintf("未知工具: %s", req.ToolName),
			},
		}
	}
}

// handleSpeechToText 处理语音转文字
func (p *ASRPlugin) handleSpeechToText(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	// 解析参数
	audioData, ok := req.Arguments["audio_data"].(string)
	if !ok {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 audio_data 参数",
			},
		}
	}

	format, _ := req.Arguments["format"].(string)
	if format == "" {
		format = "wav"
	}

	language, _ := req.Arguments["language"].(string)
	if language == "" {
		language = "zh-CN"
	}

	p.logger.Info("Processing speech to text",
		"format", format,
		"language", language,
		"data_size", len(audioData))

	// 解码音频数据
	audioBytes, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		p.IncrementCounter("asr.errors.decode")
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "DECODE_ERROR",
				Message: fmt.Sprintf("音频数据解码失败: %v", err),
			},
		}
	}

	// 模拟ASR处理（实际应用中这里会调用真实的ASR引擎）
	result, confidence, duration, err := p.processAudio(audioBytes, format, language)
	if err != nil {
		p.IncrementCounter("asr.errors.processing")
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "PROCESSING_ERROR",
				Message: fmt.Sprintf("语音处理失败: %v", err),
			},
		}
	}

	// 记录成功指标
	p.IncrementCounter("asr.calls.success")
	p.RecordHistogram("asr.processing_duration", float64(duration.Milliseconds()))
	p.SetGauge("asr.confidence", confidence)

	p.logger.Info("Speech to text completed",
		"text_length", len(result),
		"confidence", confidence,
		"duration_ms", duration.Milliseconds())

	return &v1.CallToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"text":       result,
			"confidence": confidence,
			"duration":   duration.Milliseconds(),
			"language":   language,
			"format":     format,
			"timestamp":  time.Now().Unix(),
		},
		Output: fmt.Sprintf("识别结果: %s (置信度: %.2f%%)", result, confidence*100),
	}
}

// handleDetectAudioFormat 检测音频格式
func (p *ASRPlugin) handleDetectAudioFormat(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	audioData, ok := req.Arguments["audio_data"].(string)
	if !ok {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 audio_data 参数",
			},
		}
	}

	audioBytes, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "DECODE_ERROR",
				Message: fmt.Sprintf("音频数据解码失败: %v", err),
			},
		}
	}

	// 简单的格式检测
	format := p.detectFormat(audioBytes)
	duration := p.estimateDuration(audioBytes, format)

	p.IncrementCounter("asr.format_detect.calls")

	return &v1.CallToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"format":   format,
			"duration": duration.Milliseconds(),
			"size":     len(audioBytes),
		},
		Output: fmt.Sprintf("检测到音频格式: %s, 预计时长: %.2f秒", format, duration.Seconds()),
	}
}

// handleBatchTranscribe 批量转录
func (p *ASRPlugin) handleBatchTranscribe(ctx context.Context, req *v1.CallToolResponse) *v1.CallToolRequest {
	audioFiles, ok := req.Arguments["audio_files"].([]interface{})
	if !ok {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 audio_files 参数",
			},
		}
	}

	language, _ := req.Arguments["language"].(string)
	if language == "" {
		language = "zh-CN"
	}

	p.logger.Info("Starting batch transcription", "files_count", len(audioFiles))

	results := make([]map[string]interface{}, 0, len(audioFiles))
	totalDuration := time.Duration(0)

	for i, fileInterface := range audioFiles {
		fileMap, ok := fileInterface.(map[string]interface{})
		if !ok {
			continue
		}

		filename, _ := fileMap["filename"].(string)
		audioData, _ := fileMap["audio_data"].(string)

		if filename == "" || audioData == "" {
			continue
		}

		// 处理单个文件
		audioBytes, err := base64.StdEncoding.DecodeString(audioData)
		if err != nil {
			p.logger.Warn("Failed to decode audio file", "filename", filename, "error", err)
			continue
		}

		// 获取文件扩展名作为格式
		format := strings.TrimPrefix(filepath.Ext(filename), ".")
		if format == "" {
			format = p.detectFormat(audioBytes)
		}

		result, confidence, duration, err := p.processAudio(audioBytes, format, language)
		if err != nil {
			p.logger.Warn("Failed to process audio file", "filename", filename, "error", err)
			results = append(results, map[string]interface{}{
				"filename":  filename,
				"success":   false,
				"error":     err.Error(),
				"timestamp": time.Now().Unix(),
			})
			continue
		}

		totalDuration += duration
		results = append(results, map[string]interface{}{
			"filename":   filename,
			"success":    true,
			"text":       result,
			"confidence": confidence,
			"duration":   duration.Milliseconds(),
			"timestamp":  time.Now().Unix(),
		})
	}

	p.IncrementCounter("asr.batch.calls")
	p.RecordHistogram("asr.batch.files_count", float64(len(audioFiles)))
	p.RecordHistogram("asr.batch.total_duration", float64(totalDuration.Milliseconds()))

	return &v1.CallToolRequest{
		Success: true,
		Result: map[string]interface{}{
			"results":        results,
			"processed":      len(results),
			"total_duration": totalDuration.Milliseconds(),
			"language":       language,
		},
		Output: fmt.Sprintf("批量转录完成，处理了 %d 个文件", len(results)),
	}
}

// handleGetSupportedFormats 获取支持的格式
func (p *ASRPlugin) handleGetSupportedFormats(ctx context.Context, req *v1.CallToolResponse) *v1.CallToolRequest {
	supportedFormats := []string{
		"wav", "mp3", "flac", "aac", "ogg", "m4a", "wma",
	}

	supportedLanguages := []string{
		"zh-CN", "zh-TW", "en-US", "en-GB", "ja-JP", "ko-KR",
	}

	features := []string{
		"speaker_diarization",
		"punctuation",
		"timestamp",
		"confidence_score",
		"batch_processing",
	}

	return &v1.CallToolRequest{
		Success: true,
		Result: map[string]interface{}{
			"formats":   supportedFormats,
			"languages": supportedLanguages,
			"features":  features,
		},
		Output: "支持的音频格式: " + strings.Join(supportedFormats, ", "),
	}
}

// ListTools 列出可用工具
func (p *ASRPlugin) ListTools(ctx context.Context) *v1.ListToolsResponse {
	tools := []*v1.ToolInfo{
		{
			Name:        "speech_to_text",
			Description: "将语音转换为文字",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"audio_data": map[string]interface{}{
						"type":        "string",
						"description": "Base64编码的音频数据",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "音频格式 (wav, mp3, flac等)",
						"default":     "wav",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "语言代码 (zh-CN, en-US等)",
						"default":     "zh-CN",
					},
				},
				"required": []string{"audio_data"},
			},
		},
		{
			Name:        "detect_audio_format",
			Description: "检测音频格式和信息",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"audio_data": map[string]interface{}{
						"type":        "string",
						"description": "Base64编码的音频数据",
					},
				},
				"required": []string{"audio_data"},
			},
		},
		{
			Name:        "batch_transcribe",
			Description: "批量转录多个音频文件",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"audio_files": map[string]interface{}{
						"type":        "array",
						"description": "音频文件列表，每个文件包含filename和audio_data",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"filename": map[string]interface{}{
									"type": "string",
								},
								"audio_data": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "语言代码",
						"default":     "zh-CN",
					},
				},
				"required": []string{"audio_files"},
			},
		},
		{
			Name:        "get_supported_formats",
			Description: "获取支持的音频格式和语言",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	return &v1.ListToolsResponse{
		Success: true,
		Tools:   tools,
	}
}

// GetToolSchema 获取工具模式
func (p *ASRPlugin) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
	// 这里可以根据 req.ToolName 返回对应的工具模式
	return &v1.GetToolSchemaResponse{
		Success: true,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tool_name": map[string]interface{}{
					"type": "string",
				},
				"description": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}
}

// 模拟ASR处理函数（实际应用中替换为真实的ASR引擎调用）
func (p *ASRPlugin) processAudio(audioBytes []byte, format, language string) (string, float64, time.Duration, error) {
	// 模拟处理时间
	duration := time.Duration(len(audioBytes)/1000) * time.Millisecond
	time.Sleep(duration)

	// 模拟识别结果
	var result string
	var confidence float64

	switch language {
	case "zh-CN":
		result = "这是一段中文语音识别的示例结果"
		confidence = 0.95
	case "en-US":
		result = "This is an example English speech recognition result"
		confidence = 0.92
	case "ja-JP":
		result = "これは日本語音声認識のサンプル結果です"
		confidence = 0.88
	default:
		result = "未知的语言，但仍然可以尝试识别"
		confidence = 0.75
	}

	// 根据音频大小调整置信度
	if len(audioBytes) < 1000 {
		confidence *= 0.8 // 音频太短，降低置信度
	}

	return result, confidence, duration, nil
}

// detectFormat 检测音频格式
func (p *ASRPlugin) detectFormat(audioBytes []byte) string {
	if len(audioBytes) < 12 {
		return "unknown"
	}

	// WAV 格式检查
	if string(audioBytes[0:4]) == "RIFF" && string(audioBytes[8:12]) == "WAVE" {
		return "wav"
	}

	// MP3 格式检查 (ID3v2)
	if string(audioBytes[0:3]) == "ID3" {
		return "mp3"
	}

	// FLAC 格式检查
	if string(audioBytes[0:4]) == "fLaC" {
		return "flac"
	}

	// OGG 格式检查
	if string(audioBytes[0:4]) == "OggS" {
		return "ogg"
	}

	// 根据文件扩展名猜测
	if len(audioBytes) > 0 {
		return "unknown"
	}

	return "unknown"
}

// estimateFormat 估算音频时长
func (p *ASRPlugin) estimateDuration(audioBytes []byte, format string) time.Duration {
	// 简单的时长估算（实际应用中应该解析音频头信息）
	var bitrate int
	switch format {
	case "wav":
		bitrate = 1411 // kbps
	case "mp3":
		bitrate = 128
	case "flac":
		bitrate = 800
	case "aac":
		bitrate = 256
	default:
		bitrate = 128
	}

	// 时长 = 文件大小 * 8 / 比特率
	seconds := float64(len(audioBytes)*8) / float64(bitrate*1000)
	return time.Duration(seconds*1000) * time.Millisecond
}

func main() {
	// 创建日志记录器
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "asr-plugin",
		Level:  hclog.Info,
		Output: hclog.DefaultOutput,
	})

	// 创建插件实例
	plugin := NewASRPlugin(logger)

	logger.Info("Starting ASR Plugin")

	// 服务插件
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.SimpleHandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"plugin": &sdk.SimplePluginRPC{Impl: plugin},
		},
	})
}