package core

import (
	"context"
	"fmt"
	"os"
	providers "xiaozhi-server-go/internal/domain/providers/types"
)

type TTSExecutor struct {
	provider providers.TTSProvider
}

func (e *TTSExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	text, ok := inputs["text"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid input: text must be string")
	}

	filePath, err := e.provider.ToTTS(text)
	if err != nil {
		return nil, err
	}

	// Read file to bytes
	audioData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TTS output file: %w", err)
	}

	return map[string]interface{}{
		"audio_file": filePath,
		"audio_data": audioData,
	}, nil
}
