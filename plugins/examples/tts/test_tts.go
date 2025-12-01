package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-hclog"
	v1 "xiaozhi-server-go/api/v1"
	sdk "xiaozhi-server-go/internal/plugin/sdk"
)

func main() {
	fmt.Println("=== TTS Plugin Test ===")

	// 创建日志记录器
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "test-tts",
		Level: hclog.Info,
	})

	// 创建插件实例
	plugin := &TTSPlugin{
		SimplePluginImpl: *sdk.NewSimplePlugin(&v1.PluginInfo{
			ID:      "tts-test",
			Name:    "TTS Test Plugin",
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

	// 测试3: 获取可用语音
	fmt.Println("\n3. Testing GetAvailableVoices...")
	getVoicesResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "get_available_voices",
		Arguments: map[string]interface{}{
			"language": "zh-CN",
			"gender":   "female",
		},
	})
	if getVoicesResp.Success {
		fmt.Printf("Available voices: %v\n", getVoicesResp.Result)
	} else {
		fmt.Printf("GetAvailableVoices failed: %s\n", getVoicesResp.Error.Message)
	}

	// 测试4: 验证文本
	fmt.Println("\n4. Testing ValidateText...")
	validateTextResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "validate_text",
		Arguments: map[string]interface{}{
			"text": "这是一个测试文本，用于验证文本转语音功能。",
		},
	})
	if validateTextResp.Success {
		fmt.Printf("Text validation result: %v\n", validateTextResp.Result)
		fmt.Printf("Output: %s\n", validateTextResp.Output)
	} else {
		fmt.Printf("ValidateText failed: %s\n", validateTextResp.Error.Message)
	}

	// 测试5: 文本转语音
	fmt.Println("\n5. Testing TextToSpeech...")
	textToSpeechResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "text_to_speech",
		Arguments: map[string]interface{}{
			"text":    "你好，这是一个语音合成测试。",
			"voice":   "zh-CN-female-1",
			"format":  "mp3",
			"rate":    1.0,
			"pitch":   1.0,
			"volume":  1.0,
		},
	})
	if textToSpeechResp.Success {
		result := textToSpeechResp.Result.(map[string]interface{})
		fmt.Printf("Synthesis completed:\n")
		fmt.Printf("  - Audio size: %v bytes\n", result["size"])
		fmt.Printf("  - Duration: %v ms\n", result["duration"])
		fmt.Printf("  - Format: %v\n", result["format"])
		fmt.Printf("  - Voice: %v\n", result["voice"])
		fmt.Printf("  - Audio data length: %d\n", len(result["audio_data"].(string)))
		fmt.Printf("Output: %s\n", textToSpeechResp.Output)
	} else {
		fmt.Printf("TextToSpeech failed: %s\n", textToSpeechResp.Error.Message)
	}

	// 测试6: 批量合成
	fmt.Println("\n6. Testing SynthesizeBatch...")
	texts := []interface{}{
		"第一段文本：这是批量测试的第一段。",
		"第二段文本：这是批量测试的第二段。",
		"第三段文本：这是批量测试的第三段。",
		"Hello, this is the fourth text.",
		"こんにちは、これは5番目のテキストです。",
	}

	batchSynthesizeResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "synthesize_batch",
		Arguments: map[string]interface{}{
			"texts":  texts,
			"voice":  "zh-CN-female-1",
			"format": "mp3",
		},
	})
	if batchSynthesizeResp.Success {
		result := batchSynthesizeResp.Result.(map[string]interface{})
		fmt.Printf("Batch synthesis completed:\n")
		fmt.Printf("  - Processed texts: %v\n", result["processed"])
		fmt.Printf("  - Total duration: %v ms\n", result["total_duration"])
		fmt.Printf("  - Total size: %v bytes\n", result["total_size"])
		fmt.Printf("  - Voice: %v\n", result["voice"])
		fmt.Printf("  - Format: %v\n", result["format"])

		// 显示每个文本的处理结果
		if results, ok := result["results"].([]map[string]interface{}); ok {
			for i, res := range results {
				if success, ok := res["success"].(bool); ok && success {
					fmt.Printf("  Text %d: SUCCESS (size: %v, duration: %v ms)\n",
						i, res["size"], res["duration"])
				} else {
					fmt.Printf("  Text %d: FAILED (%v)\n", i, res["error"])
				}
			}
		}
		fmt.Printf("Output: %s\n", batchSynthesizeResp.Output)
	} else {
		fmt.Printf("BatchSynthesize failed: %s\n", batchSynthesizeResp.Error.Message)
	}

	// 测试7: 测试不同语音
	fmt.Println("\n7. Testing Different Voices...")
	voices := []string{"zh-CN-female-1", "zh-CN-male-1", "en-US-female-1", "ja-JP-female-1"}

	for _, voice := range voices {
		fmt.Printf("Testing voice: %s\n", voice)
		resp := plugin.CallTool(ctx, &v1.CallToolRequest{
			ToolName: "text_to_speech",
			Arguments: map[string]interface{}{
				"text":   "Hello, this is a test for " + voice,
				"voice":  voice,
				"format": "mp3",
			},
		})
		if resp.Success {
			result := resp.Result.(map[string]interface{})
			fmt.Printf("  ✅ Success - Size: %v, Duration: %v ms\n",
				result["size"], result["duration"])
		} else {
			fmt.Printf("  ❌ Failed - %s\n", resp.Error.Message)
		}
	}

	// 测试8: 测试不同音频格式
	fmt.Println("\n8. Testing Different Formats...")
	formats := []string{"mp3", "wav", "flac", "aac"}

	for _, format := range formats {
		fmt.Printf("Testing format: %s\n", format)
		resp := plugin.CallTool(ctx, &v1.CallToolRequest{
			ToolName: "text_to_speech",
			Arguments: map[string]interface{}{
				"text":   "This is a test for " + format + " format.",
				"format": format,
			},
		})
		if resp.Success {
			result := resp.Result.(map[string]interface{})
			fmt.Printf("  ✅ Success - Size: %v, Duration: %v ms\n",
				result["size"], result["duration"])
		} else {
			fmt.Printf("  ❌ Failed - %s\n", resp.Error.Message)
		}
	}

	// 测试9: 测试参数调整
	fmt.Println("\n9. Testing Parameter Adjustments...")
	testParams := []map[string]interface{}{
		{"name": "Fast speech", "rate": 1.5, "pitch": 1.0, "volume": 1.0},
		{"name": "Slow speech", "rate": 0.5, "pitch": 1.0, "volume": 1.0},
		{"name": "High pitch", "rate": 1.0, "pitch": 1.5, "volume": 1.0},
		{"name": "Low pitch", "rate": 1.0, "pitch": 0.5, "volume": 1.0},
		{"name": "Low volume", "rate": 1.0, "pitch": 1.0, "volume": 0.5},
		{"name": "High volume", "rate": 1.0, "pitch": 1.0, "volume": 1.5},
	}

	for _, params := range testParams {
		name := params["name"].(string)
		fmt.Printf("Testing %s:\n", name)
		resp := plugin.CallTool(ctx, &v1.CallToolRequest{
			ToolName: "text_to_speech",
			Arguments: map[string]interface{}{
				"text":   "This is a test for parameter adjustment.",
				"rate":   params["rate"],
				"pitch":  params["pitch"],
				"volume": params["volume"],
			},
		})
		if resp.Success {
			result := resp.Result.(map[string]interface{})
			fmt.Printf("  ✅ Success - Size: %v, Duration: %v ms\n",
				result["size"], result["duration"])
		} else {
			fmt.Printf("  ❌ Failed - %s\n", resp.Error.Message)
		}
	}

	// 测试10: 获取指标
	fmt.Println("\n10. Testing GetMetrics...")
	metrics := plugin.GetMetrics(ctx)
	fmt.Printf("Plugin metrics: %+v\n", metrics)

	// 测试11: 健康检查
	fmt.Println("\n11. Testing HealthCheck...")
	health := plugin.HealthCheck(ctx)
	fmt.Printf("Health status: %+v\n", health)

	// 测试12: 错误处理测试
	fmt.Println("\n12. Testing Error Handling...")

	// 测试空文本
	emptyTextResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "text_to_speech",
		Arguments: map[string]interface{}{
			"text": "",
		},
	})
	if !emptyTextResp.Success {
		fmt.Printf("Empty text error (expected): %s\n", emptyTextResp.Error.Message)
	}

	// 测试过长文本
	longText := ""
	for i := 0; i < 11000; i++ {
		longText += "a"
	}
	longTextResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "text_to_speech",
		Arguments: map[string]interface{}{
			"text": longText,
		},
	})
	if !longTextResp.Success {
		fmt.Printf("Long text error (expected): %s\n", longTextResp.Error.Message)
	}

	// 测试未知工具
	unknownToolResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "unknown_tool",
		Arguments: map[string]interface{}{},
	})
	if !unknownToolResp.Success {
		fmt.Printf("Unknown tool error (expected): %s\n", unknownToolResp.Error.Message)
	}

	// 关闭插件
	fmt.Println("\n13. Shutting down plugin...")
	err = plugin.Shutdown(ctx)
	if err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	} else {
		fmt.Println("Plugin shutdown successfully")
	}

	fmt.Println("\n=== TTS Plugin Test Complete ===")
}