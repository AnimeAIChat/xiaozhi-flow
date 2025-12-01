package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-hclog"
	v1 "xiaozhi-server-go/api/v1"
	sdk "xiaozhi-server-go/internal/plugin/sdk"
)

func main() {
	fmt.Println("=== LLM Plugin Test ===")

	// 创建日志记录器
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "test-llm",
		Level: hclog.Info,
	})

	// 创建插件实例
	plugin := &LLMPlugin{
		SimplePluginImpl: *sdk.NewSimplePlugin(&v1.PluginInfo{
			ID:      "llm-test",
			Name:    "LLM Test Plugin",
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

	// 测试2: 获取可用模型
	fmt.Println("\n2. Testing GetAvailableModels...")
	getModelsResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "get_available_models",
		Arguments: map[string]interface{}{},
	})
	if getModelsResp.Success {
		result := getModelsResp.Result.(map[string]interface{})
		fmt.Printf("Available models: %v\n", result)
	} else {
		fmt.Printf("GetAvailableModels failed: %s\n", getModelsResp.Error.Message)
	}

	// 测试3: 获取特定提供商的模型
	fmt.Println("\n3. Testing GetAvailableModels (OpenAI only)...")
	openaiModelsResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "get_available_models",
		Arguments: map[string]interface{}{
			"provider": "openai",
			"type":     "chat",
		},
	})
	if openaiModelsResp.Success {
		result := openaiModelsResp.Result.(map[string]interface{})
		fmt.Printf("OpenAI chat models: %v\n", result)
	} else {
		fmt.Printf("GetAvailableModels (OpenAI) failed: %s\n", openaiModelsResp.Error.Message)
	}

	// 测试4: 获取模型信息
	fmt.Println("\n4. Testing GetModelInfo...")
	modelInfoResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "get_model_info",
		Arguments: map[string]interface{}{
			"model_id": "gpt-3.5-turbo",
		},
	})
	if modelInfoResp.Success {
		result := modelInfoResp.Result.(map[string]interface{})
		fmt.Printf("Model info: %v\n", result)
	} else {
		fmt.Printf("GetModelInfo failed: %s\n", modelInfoResp.Error.Message)
	}

	// 测试5: 计算token数量
	fmt.Println("\n5. Testing CountTokens...")
	messages := []map[string]interface{}{
		{"role": "system", "content": "你是一个有用的AI助手。"},
		{"role": "user", "content": "你好，请介绍一下你自己。"},
	}
	countTokensResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "count_tokens",
		Arguments: map[string]interface{}{
			"messages": messages,
		},
	})
	if countTokensResp.Success {
		result := countTokensResp.Result.(map[string]interface{})
		fmt.Printf("Token count result: %v\n", result)
		fmt.Printf("Output: %s\n", countTokensResp.Output)
	} else {
		fmt.Printf("CountTokens failed: %s\n", countTokensResp.Error.Message)
	}

	// 测试6: 验证提示
	fmt.Println("\n6. Testing ValidatePrompt...")
	validatePromptResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "validate_prompt",
		Arguments: map[string]interface{}{
			"messages": messages,
			"model":    "gpt-3.5-turbo",
		},
	})
	if validatePromptResp.Success {
		result := validatePromptResp.Result.(map[string]interface{})
		fmt.Printf("Prompt validation result: %v\n", result)
		fmt.Printf("Output: %s\n", validatePromptResp.Output)
	} else {
		fmt.Printf("ValidatePrompt failed: %s\n", validatePromptResp.Error.Message)
	}

	// 测试7: 聊天完成
	fmt.Println("\n7. Testing ChatCompletion...")
	chatCompletionResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "chat_completion",
		Arguments: map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{"role": "system", "content": "你是一个友好的AI助手。"},
				{"role": "user", "content": "你好，请问你能帮我做什么？"},
			},
			"max_tokens":   500,
			"temperature": 0.7,
			"top_p":       1.0,
			"stream":      false,
		},
	})
	if chatCompletionResp.Success {
		result := chatCompletionResp.Result.(map[string]interface{})
		fmt.Printf("Chat completion result:\n")
		fmt.Printf("  ID: %v\n", result["id"])
		fmt.Printf("  Model: %v\n", result["model"])
		if choices, ok := result["choices"].([]map[string]interface{}); ok && len(choices) > 0 {
			fmt.Printf("  Response: %v\n", choices[0]["message"])
		}
		if usage, ok := result["usage"].(map[string]interface{}); ok {
			fmt.Printf("  Usage: %+v\n", usage)
		}
		fmt.Printf("Output: %s\n", chatCompletionResp.Output)
	} else {
		fmt.Printf("ChatCompletion failed: %s\n", chatCompletionResp.Error.Message)
	}

	// 测试8: 文本完成
	fmt.Println("\n8. Testing TextCompletion...")
	textCompletionResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "text_completion",
		Arguments: map[string]interface{}{
			"prompt":       "人工智能的发展历程可以追溯到",
			"model":        "text-davinci-003",
			"max_tokens":   200,
			"temperature":  0.7,
		},
	})
	if textCompletionResp.Success {
		result := textCompletionResp.Result.(map[string]interface{})
		fmt.Printf("Text completion result:\n")
		fmt.Printf("  Prompt: %v\n", result["prompt"])
		fmt.Printf("  Response: %v\n", result["response"])
		fmt.Printf("  Model: %v\n", result["model"])
		fmt.Printf("Output: %s\n", textCompletionResp.Output)
	} else {
		fmt.Printf("TextCompletion failed: %s\n", textCompletionResp.Error.Message)
	}

	// 测试9: 测试不同参数的聊天完成
	fmt.Println("\n9. Testing Different Parameters...")
	testCases := []map[string]interface{}{
		{
			"name":        "Low temperature",
			"temperature": 0.1,
			"prompt":      "请解释什么是机器学习",
		},
		{
			"name":        "High temperature",
			"temperature": 1.5,
			"prompt":      "请写一个关于未来的短故事",
		},
		{
			"name":        "Limited tokens",
			"max_tokens":  50,
			"prompt":      "请详细介绍量子计算的基本原理",
		},
		{
			"name": "With stop words",
			"stop":  []string{"\n", "。"},
			"prompt": "列出三个编程语言：",
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("Testing case %d: %s\n", i+1, testCase["name"])

		args := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{"role": "user", "content": testCase["prompt"]},
			},
		}

		if temp, ok := testCase["temperature"].(float64); ok {
			args["temperature"] = temp
		}
		if maxTokens, ok := testCase["max_tokens"].(int); ok {
			args["max_tokens"] = maxTokens
		}
		if stop, ok := testCase["stop"].([]string); ok {
			args["stop"] = stop
		}

		resp := plugin.CallTool(ctx, &v1.CallToolRequest{
			ToolName: "chat_completion",
			Arguments: args,
		})

		if resp.Success {
			result := resp.Result.(map[string]interface{})
			if choices, ok := result["choices"].([]map[string]interface{}); ok && len(choices) > 0 {
				message := choices[0]["message"].(map[string]interface{})
				fmt.Printf("  ✅ Response: %v\n", message["content"])
			}
		} else {
			fmt.Printf("  ❌ Failed: %s\n", resp.Error.Message)
		}
	}

	// 测试10: 测试多轮对话
	fmt.Println("\n10. Testing Multi-turn Conversation...")
	conversation := []map[string]interface{}{
		{"role": "system", "content": "你是一个专业的技术顾问。"},
		{"role": "user", "content": "我想学习Python编程，有什么建议吗？"},
		{"role": "assistant", "content": "学习Python是个很好的选择！建议你从基础语法开始，然后练习一些简单的项目。"},
		{"role": "user", "content": "你能推荐一些适合初学者的项目吗？"},
	}

	multiTurnResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "chat_completion",
		Arguments: map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": conversation,
			"max_tokens": 300,
		},
	})
	if multiTurnResp.Success {
		result := multiTurnResp.Result.(map[string]interface{})
		if choices, ok := result["choices"].([]map[string]interface{}); ok && len(choices) > 0 {
			message := choices[0]["message"].(map[string]interface{})
			fmt.Printf("✅ Multi-turn response: %v\n", message["content"])
		}
	} else {
		fmt.Printf("❌ Multi-turn failed: %s\n", multiTurnResp.Error.Message)
	}

	// 测试11: 错误处理测试
	fmt.Println("\n11. Testing Error Handling...")

	// 测试空消息
	emptyMessagesResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "chat_completion",
		Arguments: map[string]interface{}{
			"messages": []map[string]interface{}{},
		},
	})
	if !emptyMessagesResp.Success {
		fmt.Printf("Empty messages error (expected): %s\n", emptyMessagesResp.Error.Message)
	}

	// 测试无效模型
	invalidModelResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "get_model_info",
		Arguments: map[string]interface{}{
			"model_id": "invalid-model-123",
		},
	})
	if !invalidModelResp.Success {
		fmt.Printf("Invalid model error (expected): %s\n", invalidModelResp.Error.Message)
	}

	// 测试未知工具
	unknownToolResp := plugin.CallTool(ctx, &v1.CallToolRequest{
		ToolName: "unknown_tool",
		Arguments: map[string]interface{}{},
	})
	if !unknownToolResp.Success {
		fmt.Printf("Unknown tool error (expected): %s\n", unknownToolResp.Error.Message)
	}

	// 测试12: 获取指标
	fmt.Println("\n12. Testing GetMetrics...")
	metrics := plugin.GetMetrics(ctx)
	fmt.Printf("Plugin metrics: %+v\n", metrics)

	// 测试13: 健康检查
	fmt.Println("\n13. Testing HealthCheck...")
	health := plugin.HealthCheck(ctx)
	fmt.Printf("Health status: %+v\n", health)

	// 关闭插件
	fmt.Println("\n14. Shutting down plugin...")
	err = plugin.Shutdown(ctx)
	if err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	} else {
		fmt.Println("Plugin shutdown successfully")
	}

	fmt.Println("\n=== LLM Plugin Test Complete ===")
}