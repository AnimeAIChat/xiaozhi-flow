package core

import (
	"context"
	"fmt"
	"strings"
	providers "xiaozhi-server-go/internal/domain/providers/types"
)

type LLMExecutor struct {
	provider providers.LLMProvider
}

func (e *LLMExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	prompt, ok := inputs["prompt"].(string)
	if !ok {
		// Try "text" as fallback
		prompt, ok = inputs["text"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid input: prompt or text must be string")
		}
	}

	// Construct messages
	messages := []providers.Message{
		{Role: "user", Content: prompt},
	}
	
	// Handle history if present
	// Note: The type assertion might fail if the input comes from JSON unmarshalling (which gives []interface{})
	// For now, we assume it's passed correctly or we might need to convert.
	if history, ok := inputs["history"].([]providers.Message); ok {
		// Prepend history? Or is history just the previous context?
		// Usually history comes before the new prompt.
		// But here we are constructing the full context.
		// If history is passed, we assume it's the list of previous messages.
		// We should append the new prompt to it.
		finalMessages := make([]providers.Message, len(history)+1)
		copy(finalMessages, history)
		finalMessages[len(history)] = messages[0]
		messages = finalMessages
	}

	// Call LLM
	// Note: Response returns a channel for streaming. We need to aggregate it.
	stream, err := e.provider.Response(ctx, e.provider.GetSessionID(), messages)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	for chunk := range stream {
		sb.WriteString(chunk)
	}

	return map[string]interface{}{
		"text": sb.String(),
	}, nil
}
