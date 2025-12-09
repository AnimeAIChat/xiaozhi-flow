package main

import (
	"context"
	"encoding/json"
	"fmt"

	"xiaozhi-server-go/internal/plugin/builtin/openai"
	"xiaozhi-server-go/internal/plugin/capability"
)

func main() {
	// 1. Initialize Registry
	registry := capability.NewRegistry()

	// 2. Register OpenAI Provider
	openaiProvider := openai.NewProvider()
	registry.Register("openai", openaiProvider)

	fmt.Println("Registered Capabilities:")
	for _, cap := range registry.ListCapabilities() {
		fmt.Printf("- %s (%s): %s\n", cap.ID, cap.Type, cap.Description)
	}

	// 3. Get Executor
	capID := "openai_chat"
	executor, err := registry.GetExecutor(capID)
	if err != nil {
		fmt.Printf("Error getting executor: %v\n", err)
		return
	}

	// 4. Prepare Config and Input
	// Using Doubao config from user's environment
	apiKey := "2658d315-1f2f-440d-83b8-ccf8d070a208"
	baseURL := "https://ark.cn-beijing.volces.com/api/v3"
	model := "doubao-seed-1-6-251015"

	config := map[string]interface{}{
		"api_key":  apiKey,
		"base_url": baseURL,
		"model":    model,
	}

	inputs := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Hello, tell me a short joke.",
			},
		},
		"temperature": 0.7,
	}

	// 5. Execute
	fmt.Println("\nExecuting OpenAI Chat...")
	output, err := executor.Execute(context.Background(), config, inputs)
	if err != nil {
		fmt.Printf("Execution failed: %v\n", err)
		return
	}

	// 6. Print Result
	jsonOutput, _ := json.MarshalIndent(output, "", "  ")
	fmt.Printf("Output:\n%s\n", string(jsonOutput))
}
