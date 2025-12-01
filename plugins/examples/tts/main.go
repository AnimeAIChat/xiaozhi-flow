package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	v1 "xiaozhi-server-go/api/v1"
	sdk "xiaozhi-server-go/internal/plugin/sdk"
)

// TTSPlugin 文本转语音插件
type TTSPlugin struct {
	sdk.SimplePluginImpl
	logger hclog.Logger
}

// VoiceInfo 语音信息
type VoiceInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Language    string `json:"language"`
	Gender      string `json:"gender"`
	Accent      string `json:"accent"`
	Description string `json:"description"`
}

// NewTTSPlugin 创建TTS插件实例
func NewTTSPlugin(logger hclog.Logger) *TTSPlugin {
	info := &v1.PluginInfo{
		ID:          "tts-plugin",
		Name:        "TTS Plugin",
		Version:     "1.0.0",
		Description: "文本转语音插件，支持多种语音和语言",
		Author:      "XiaoZhi Flow Team",
		Type:        v1.PluginTypeAudio,
		Tags:        []string{"tts", "speech", "synthesis", "voice"},
		Capabilities: []string{"text_to_speech", "voice_management", "ssml_support", "batch_synthesis"},
		Dependencies: []string{"ffmpeg", "tts-engine"},
	}

	return &TTSPlugin{
		SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
		logger:          logger.Named("tts-plugin"),
	}
}

// CallTool 实现工具调用
func (p *TTSPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	startTime := time.Now()
	p.IncrementCounter("tts.calls.total")

	switch req.ToolName {
	case "text_to_speech":
		return p.handleTextToSpeech(ctx, req)
	case "get_available_voices":
		return p.handleGetAvailableVoices(ctx, req)
	case "synthesize_batch":
		return p.handleSynthesizeBatch(ctx, req)
	case "validate_text":
		return p.handleValidateText(ctx, req)
	case "get_supported_formats":
		return p.handleGetSupportedFormats(ctx, req)
	default:
		p.IncrementCounter("tts.calls.unknown")
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "UNKNOWN_TOOL",
				Message: fmt.Sprintf("未知工具: %s", req.ToolName),
			},
		}
	}
}

// handleTextToSpeech 处理文本转语音
func (p *TTSPlugin) handleTextToSpeech(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	// 解析参数
	text, ok := req.Arguments["text"].(string)
	if !ok || text == "" {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 text 参数或参数为空",
			},
		}
	}

	voice, _ := req.Arguments["voice"].(string)
	if voice == "" {
		voice = "default-zh-CN-female"
	}

	format, _ := req.Arguments["format"].(string)
	if format == "" {
		format = "mp3"
	}

	rate, _ := req.Arguments["rate"].(float64)
	if rate == 0 {
		rate = 1.0
	}

	pitch, _ := req.Arguments["pitch"].(float64)
	if pitch == 0 {
		pitch = 1.0
	}

	volume, _ := req.Arguments["volume"].(float64)
	if volume == 0 {
		volume = 1.0
	}

	// 验证文本长度
	if len(text) > 10000 {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "TEXT_TOO_LONG",
				Message: "文本过长，最大支持10000字符",
			},
		}
	}

	p.logger.Info("Synthesizing speech",
		"text_length", len(text),
		"voice", voice,
		"format", format,
		"rate", rate,
		"pitch", pitch,
		"volume", volume)

	// 合成语音
	audioData, duration, err := p.synthesizeSpeech(text, voice, format, rate, pitch, volume)
	if err != nil {
		p.IncrementCounter("tts.errors.synthesis")
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "SYNTHESIS_ERROR",
				Message: fmt.Sprintf("语音合成失败: %v", err),
			},
		}
	}

	// 编码音频数据
	audioDataBase64 := base64.StdEncoding.EncodeToString(audioData)

	// 记录成功指标
	p.IncrementCounter("tts.calls.success")
	p.RecordHistogram("tts.synthesis_duration", float64(duration.Milliseconds()))
	p.SetGauge("tts.audio_size", float64(len(audioData)))

	p.logger.Info("Speech synthesis completed",
		"audio_size", len(audioData),
		"duration_ms", duration.Milliseconds())

	return &v1.CallToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"audio_data": audioDataBase64,
			"format":     format,
			"duration":   duration.Milliseconds(),
			"voice":      voice,
			"size":       len(audioData),
			"timestamp":  time.Now().Unix(),
		},
		Output: fmt.Sprintf("语音合成完成，时长 %.2f 秒", duration.Seconds()),
	}
}

// handleGetAvailableVoices 获取可用语音
func (p *TTSPlugin) handleGetAvailableVoices(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	language, _ := req.Arguments["language"].(string)
	gender, _ := req.Arguments["gender"].(string)

	voices := p.getAvailableVoices(language, gender)

	p.IncrementCounter("tts.voices_list.calls")

	return &v1.CallToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"voices": voices,
			"count":  len(voices),
		},
		Output: fmt.Sprintf("找到 %d 个可用语音", len(voices)),
	}
}

// handleSynthesizeBatch 批量合成
func (p *TTSPlugin) handleSynthesizeBatch(ctx context.Context, req *v1.CallToolResponse) *v1.CallToolRequest {
	texts, ok := req.Arguments["texts"].([]interface{})
	if !ok {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 texts 参数",
			},
		}
	}

	voice, _ := req.Arguments["voice"].(string)
	if voice == "" {
		voice = "default-zh-CN-female"
	}

	format, _ := req.Arguments["format"].(string)
	if format == "" {
		format = "mp3"
	}

	p.logger.Info("Starting batch synthesis", "texts_count", len(texts))

	results := make([]map[string]interface{}, 0, len(texts))
	totalDuration := time.Duration(0)
	totalSize := 0

	for i, textInterface := range texts {
		text, ok := textInterface.(string)
		if !ok {
			continue
		}

		if text == "" {
			continue
		}

		// 合成单个文本
		audioData, duration, err := p.synthesizeSpeech(text, voice, format, 1.0, 1.0, 1.0)
		if err != nil {
			p.logger.Warn("Failed to synthesize text", "index", i, "error", err)
			results = append(results, map[string]interface{}{
				"index":     i,
				"text":      text,
				"success":   false,
				"error":     err.Error(),
				"timestamp": time.Now().Unix(),
			})
			continue
		}

		totalDuration += duration
		totalSize += len(audioData)

		results = append(results, map[string]interface{}{
			"index":      i,
			"text":       text,
			"success":    true,
			"audio_data": base64.StdEncoding.EncodeToString(audioData),
			"duration":   duration.Milliseconds(),
			"size":       len(audioData),
			"format":     format,
			"voice":      voice,
			"timestamp":  time.Now().Unix(),
		})
	}

	p.IncrementCounter("tts.batch.calls")
	p.RecordHistogram("tts.batch.texts_count", float64(len(texts)))
	p.RecordHistogram("tts.batch.total_duration", float64(totalDuration.Milliseconds()))
	p.RecordHistogram("tts.batch.total_size", float64(totalSize))

	return &v1.CallToolRequest{
		Success: true,
		Result: map[string]interface{}{
			"results":        results,
			"processed":      len(results),
			"total_duration": totalDuration.Milliseconds(),
			"total_size":     totalSize,
			"voice":          voice,
			"format":         format,
		},
		Output: fmt.Sprintf("批量合成完成，处理了 %d 个文本", len(results)),
	}
}

// handleValidateText 验证文本
func (p *TTSPlugin) handleValidateText(ctx context.Context, req *v1.CallToolResponse) *v1.CallToolRequest {
	text, ok := req.Arguments["text"].(string)
	if !ok {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "缺少 text 参数",
			},
		}
	}

	validation := p.validateText(text)

	p.IncrementCounter("tts.validate_text.calls")

	return &v1.CallToolRequest{
		Success: true,
		Result: map[string]interface{}{
			"valid":      validation.Valid,
			"issues":     validation.Issues,
			"warnings":   validation.Warnings,
			"char_count": validation.CharCount,
			"word_count": validation.WordCount,
			"estimated_duration": validation.EstimatedDuration.Milliseconds(),
		},
		Output: validation.Summary,
	}
}

// handleGetSupportedFormats 获取支持的格式
func (p *TTSPlugin) handleGetSupportedFormats(ctx context.Context, req *v1.CallToolResponse) *v1.CallToolRequest {
	supportedFormats := []string{
		"mp3", "wav", "flac", "aac", "ogg",
	}

	voiceFeatures := []string{
		"rate_adjustment",
		"pitch_adjustment",
		"volume_control",
		"ssml_support",
		"emotional_speech",
		"voice_cloning",
	}

	languages := []string{
		"zh-CN", "zh-TW", "en-US", "en-GB", "ja-JP", "ko-KR",
		"es-ES", "fr-FR", "de-DE", "it-IT", "pt-BR", "ru-RU",
	}

	return &v1.CallToolRequest{
		Success: true,
		Result: map[string]interface{}{
			"formats":        supportedFormats,
			"voice_features": voiceFeatures,
			"languages":      languages,
		},
		Output: "支持的音频格式: " + strings.Join(supportedFormats, ", "),
	}
}

// ListTools 列出可用工具
func (p *TTSPlugin) ListTools(ctx context.Context) *v1.ListToolsResponse {
	tools := []*v1.ToolInfo{
		{
			Name:        "text_to_speech",
			Description: "将文本转换为语音",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "要转换的文本",
					},
					"voice": map[string]interface{}{
						"type":        "string",
						"description": "语音ID",
						"default":     "default-zh-CN-female",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "音频格式",
						"default":     "mp3",
					},
					"rate": map[string]interface{}{
						"type":        "number",
						"description": "语速 (0.1-2.0)",
						"default":     1.0,
					},
					"pitch": map[string]interface{}{
						"type":        "number",
						"description": "音调 (0.1-2.0)",
						"default":     1.0,
					},
					"volume": map[string]interface{}{
						"type":        "number",
						"description": "音量 (0.1-2.0)",
						"default":     1.0,
					},
				},
				"required": []string{"text"},
			},
		},
		{
			Name:        "get_available_voices",
			Description: "获取可用语音列表",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"language": map[string]interface{}{
						"type":        "string",
						"description": "语言代码 (可选)",
					},
					"gender": map[string]interface{}{
						"type":        "string",
						"description": "性别 (male/female/neutral)",
					},
				},
			},
		},
		{
			Name:        "synthesize_batch",
			Description: "批量合成多个文本",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"texts": map[string]interface{}{
						"type":        "array",
						"description": "文本列表",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"voice": map[string]interface{}{
						"type":        "string",
						"description": "语音ID",
						"default":     "default-zh-CN-female",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "音频格式",
						"default":     "mp3",
					},
				},
				"required": []string{"texts"},
			},
		},
		{
			Name:        "validate_text",
			Description: "验证文本是否适合合成",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "要验证的文本",
					},
				},
				"required": []string{"text"},
			},
		},
		{
			Name:        "get_supported_formats",
			Description: "获取支持的格式和功能",
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
func (p *TTSPlugin) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
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

// TextValidationResult 文本验证结果
type TextValidationResult struct {
	Valid             bool
	Issues            []string
	Warnings          []string
	CharCount         int
	WordCount         int
	EstimatedDuration time.Duration
	Summary           string
}

// 模拟语音合成函数（实际应用中替换为真实的TTS引擎调用）
func (p *TTSPlugin) synthesizeSpeech(text, voice, format string, rate, pitch, volume float64) ([]byte, time.Duration, error) {
	// 模拟处理时间（基于文本长度）
	processingTime := time.Duration(len(text)*10) * time.Millisecond
	time.Sleep(processingTime)

	// 估算音频时长（基于字数，每秒约4个中文字或8个英文词）
	duration := time.Duration(float64(len(text))*250) * time.Millisecond

	// 生成模拟音频数据
	var audioSize int
	switch format {
	case "mp3":
		audioSize = len(text) * 1024 // MP3压缩后
	case "wav":
		audioSize = len(text) * 4096 // WAV无损
	case "flac":
		audioSize = len(text) * 2048 // FLAC压缩
	default:
		audioSize = len(text) * 1024
	}

	// 应用音质参数调整
	audioSize = int(float64(audioSize) * rate * pitch * volume)

	// 生成随机音频数据
	audioData := make([]byte, audioSize)
	rand.Read(audioData)

	return audioData, duration, nil
}

// getAvailableVoices 获取可用语音列表
func (p *TTSPlugin) getAvailableVoices(language, gender string) []VoiceInfo {
	allVoices := []VoiceInfo{
		// 中文语音
		{ID: "zh-CN-female-1", Name: "小美", Language: "zh-CN", Gender: "female", Accent: "standard", Description: "标准普通话女声"},
		{ID: "zh-CN-male-1", Name: "小明", Language: "zh-CN", Gender: "male", Accent: "standard", Description: "标准普通话男声"},
		{ID: "zh-CN-female-2", Name: "小雅", Language: "zh-CN", Gender: "female", Accent: "sweet", Description: "甜美女声"},
		{ID: "zh-TW-female-1", Name: "小婷", Language: "zh-TW", Gender: "female", Accent: "taiwan", Description: "台湾女声"},

		// 英文语音
		{ID: "en-US-female-1", Name: "Emma", Language: "en-US", Gender: "female", Accent: "american", Description: "美国女声"},
		{ID: "en-US-male-1", Name: "John", Language: "en-US", Gender: "male", Accent: "american", Description: "美国男声"},
		{ID: "en-GB-female-1", Name: "Sophie", Language: "en-GB", Gender: "female", Accent: "british", Description: "英国女声"},
		{ID: "en-GB-male-1", Name: "James", Language: "en-GB", Gender: "male", Accent: "british", Description: "英国男声"},

		// 日文语音
		{ID: "ja-JP-female-1", Name: "由美", Language: "ja-JP", Gender: "female", Accent: "standard", Description: "标准日本女声"},
		{ID: "ja-JP-male-1", Name: "健一", Language: "ja-JP", Gender: "male", Accent: "standard", Description: "标准日本男声"},

		// 韩文语音
		{ID: "ko-KR-female-1", Name: "지수", Language: "ko-KR", Gender: "female", Accent: "standard", Description: "标准韩国女声"},
		{ID: "ko-KR-male-1", Name: "민준", Language: "ko-KR", Gender: "male", Accent: "standard", Description: "标准韩国男声"},
	}

	// 过滤语音
	var voices []VoiceInfo
	for _, voice := range allVoices {
		if language != "" && voice.Language != language {
			continue
		}
		if gender != "" && voice.Gender != gender {
			continue
		}
		voices = append(voices, voice)
	}

	return voices
}

// validateText 验证文本
func (p *TTSPlugin) validateText(text string) TextValidationResult {
	result := TextValidationResult{
		Valid:             true,
		Issues:            []string{},
		Warnings:          []string{},
		CharCount:         len(text),
		WordCount:         len(strings.Fields(text)),
		EstimatedDuration: time.Duration(float64(len(text))*250) * time.Millisecond,
	}

	// 检查长度
	if len(text) == 0 {
		result.Valid = false
		result.Issues = append(result.Issues, "文本为空")
		result.Summary = "文本验证失败：文本为空"
		return result
	}

	if len(text) > 10000 {
		result.Valid = false
		result.Issues = append(result.Issues, fmt.Sprintf("文本过长 (%d 字符)，最大支持 10000 字符", len(text)))
	}

	// 检查字符
	for i, char := range text {
		if char < 32 && char != '\n' && char != '\t' && char != '\r' {
			result.Issues = append(result.Issues, fmt.Sprintf("位置 %d 包含无效控制字符", i))
		}
	}

	// 检查可能的问题
	if strings.Contains(text, "http://") || strings.Contains(text, "https://") {
		result.Warnings = append(result.Warnings, "文本包含URL，可能影响语音质量")
	}

	if strings.Count(text, ".") > 20 || strings.Count(text, "。") > 20 {
		result.Warnings = append(result.Warnings, "标点符号较多，可能影响语音流畅度")
	}

	// 检查特殊字符
	specialChars := "!@#$%^&*()_+-=[]{}|;:'\",.<>?/`~"
	specialCount := 0
	for _, char := range specialChars {
		specialCount += strings.Count(text, string(char))
	}
	if specialCount > len(text)/10 {
		result.Warnings = append(result.Warnings, "特殊字符较多，可能影响语音质量")
	}

	// 生成摘要
	if len(result.Issues) == 0 {
		if len(result.Warnings) == 0 {
			result.Summary = "文本验证通过，适合语音合成"
		} else {
			result.Summary = fmt.Sprintf("文本验证通过，但有 %d 个警告", len(result.Warnings))
		}
	} else {
		result.Summary = fmt.Sprintf("文本验证失败，发现 %d 个问题", len(result.Issues))
	}

	return result
}

func main() {
	// 创建日志记录器
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "tts-plugin",
		Level:  hclog.Info,
		Output: hclog.DefaultOutput,
	})

	// 创建插件实例
	plugin := NewTTSPlugin(logger)

	logger.Info("Starting TTS Plugin")

	// 服务插件
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.SimpleHandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"plugin": &sdk.SimplePluginRPC{Impl: plugin},
		},
	})
}