package core

import (
	"context"
	"fmt"
	providers "xiaozhi-server-go/internal/domain/providers/types"
)

type ASRExecutor struct {
	provider providers.ASRProvider
}

func (e *ASRExecutor) Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	audioData, ok := inputs["audio_data"].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid input: audio_data must be []byte")
	}

	text, err := e.provider.Transcribe(ctx, audioData)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"text": text,
	}, nil
}
