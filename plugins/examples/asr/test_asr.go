package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	v1 "xiaozhi-server-go/api/v1"
	sdk "xiaozhi-server-go/internal/plugin/sdk"
)

// 创建模拟的WAV音频数据
func createMockWAVData() []byte {
	// 简单的WAV文件头
	wavHeader := []byte{
		// RIFF header
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x24, 0x08, 0x00, 0x00, // file size - 8
		0x57, 0x41, 0x56, 0x45, // "WAVE"

		// fmt chunk
		0x66, 0x6d, 0x74, 0x20, // "fmt "
		0x10, 0x00, 0x00, 0x00, // chunk size
		0x01, 0x00,             // audio format (PCM)
		0x01, 0x00,             // channels
		0x40, 0x1f, 0x00, 0x00, // sample rate (8000)
		0x40, 0x1f, 0x00, 0x00, // byte rate
		0x01, 0x00,             // block align
		0x08, 0x00,             // bits per sample

		// data chunk
		0x64, 0x61, 0x74, 0x61, // "data"
		0x00, 0x08, 0x00, 0x00, // data size
	}

	// 模拟音频数据
	audioData := make([]byte, 2048)
	for i := range audioData {
		audioData[i] = byte(i % 256)
	}

	return append(wavHeader, audioData...)
}

func main() {
	fmt.Println("=== ASR Plugin Test ===")

	// 创建日志记录器（使用标准日志）
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "test-asr",
		Level: hclog.Info,
	})

	// 创建插件实例
	plugin := &ASRPlugin{
		SimplePluginImpl: *sdk.NewSimplePlugin(&v1.PluginInfo{
			ID:      "asr-test",
			Name:    "ASR Test Plugin",
			Version: "1.0.0",
		}, logger),
		logger: logger,
	}

	// 初始化插件
	ctx := context.Background()
	err := plugin.Initialize(ctx, &sdk.InitializeConfig{
		Config: make(map[string]interface{}),
	})
	if err != nil {
		log.Fatalf("Failed to initialize plugin: %v", err)
	}

	// 测试1: 列出工具
	fmt.Println("\n1. Testing ListTools...")
	listToolsResp := plugin.ListTools(ctx)
	if listToolsResp.Success {
		fmt.Printf("Available tools: %d\n", len(listToolsResp.Tools))
		for _, tool := range listToolsResp.Tools {
			fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
		}
	} else {
		fmt.Printf("ListTools failed: %s\n", listToolsResp.Error.Message)
	}

	// 测试2: 获取支持的格式
	fmt.Println("\n2. Testing GetSupportedFormats...")
	supportedFormatsResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "get_supported_formats",
		Arguments: map[string]interface{}{},
	})
	if supportedFormatsResp.Success {
		fmt.Printf("Supported formats: %v\n", supportedFormatsResp.Result)
	} else {
		fmt.Printf("GetSupportedFormats failed: %s\n", supportedFormatsResp.Error.Message)
	}

	// 测试3: 检测音频格式
	fmt.Println("\n3. Testing DetectAudioFormat...")
	mockWavData := createMockWAVData()
	audioDataBase64 := base64.StdEncoding.EncodeToString(mockWavData)

	detectFormatResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "detect_audio_format",
		Arguments: map[string]interface{}{
			"audio_data": audioDataBase64,
		},
	})
	if detectFormatResp.Success {
		fmt.Printf("Detected format: %v\n", detectFormatResp.Result)
		fmt.Printf("Output: %s\n", detectFormatResp.Output)
	} else {
		fmt.Printf("DetectAudioFormat failed: %s\n", detectFormatResp.Error.Message)
	}

	// 测试4: 语音转文字
	fmt.Println("\n4. Testing SpeechToText...")
	speechToTextResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "speech_to_text",
		Arguments: map[string]interface{}{
			"audio_data": audioDataBase64,
			"format":     "wav",
			"language":   "zh-CN",
		},
	})
	if speechToTextResp.Success {
		fmt.Printf("Recognition result: %v\n", speechToTextResp.Result)
		fmt.Printf("Output: %s\n", speechToTextResp.Output)
	} else {
		fmt.Printf("SpeechToText failed: %s\n", speechToTextResp.Error.Message)
	}

	// 测试5: 批量转录
	fmt.Println("\n5. Testing BatchTranscribe...")
	batchFiles := []map[string]interface{}{
		{
			"filename":   "test1.wav",
			"audio_data": audioDataBase64,
		},
		{
			"filename":   "test2.wav",
			"audio_data": audioDataBase64,
		},
	}

	batchTranscribeResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "batch_transcribe",
		Arguments: map[string]interface{}{
			"audio_files": batchFiles,
			"language":    "zh-CN",
		},
	})
	if batchTranscribeResp.Success {
		fmt.Printf("Batch result: %v\n", batchTranscribeResp.Result)
		fmt.Printf("Output: %s\n", batchTranscribeResp.Output)
	} else {
		fmt.Printf("BatchTranscribe failed: %s\n", batchTranscribeResp.Error.Message)
	}

	// 测试6: 获取指标
	fmt.Println("\n6. Testing GetMetrics...")
	metrics := plugin.GetMetrics(ctx)
	fmt.Printf("Plugin metrics: %+v\n", metrics)

	// 测试7: 健康检查
	fmt.Println("\n7. Testing HealthCheck...")
	health := plugin.HealthCheck(ctx)
	fmt.Printf("Health status: %+v\n", health)

	// 测试8: 测试错误处理
	fmt.Println("\n8. Testing Error Handling...")
	errorResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "unknown_tool",
		Arguments: map[string]interface{}{},
	})
	if !errorResp.Success {
		fmt.Printf("Expected error: %s - %s\n", errorResp.Error.Code, errorResp.Error.Message)
	}

	// 关闭插件
	fmt.Println("\n9. Shutting down plugin...")
	err = plugin.Shutdown(ctx)
	if err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	} else {
		fmt.Println("Plugin shutdown successfully")
	}

	fmt.Println("\n=== ASR Plugin Test Complete ===")
}