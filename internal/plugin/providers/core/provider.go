package core

import (
	"fmt"
	"xiaozhi-server-go/internal/plugin/capability"
	providers "xiaozhi-server-go/internal/domain/providers/types"
)

type CoreProvider struct {
	asr providers.ASRProvider
	llm providers.LLMProvider
	tts providers.TTSProvider
}

func NewCoreProvider(asr providers.ASRProvider, llm providers.LLMProvider, tts providers.TTSProvider) *CoreProvider {
	return &CoreProvider{
		asr: asr,
		llm: llm,
		tts: tts,
	}
}

func (p *CoreProvider) GetCapabilities() []capability.Definition {
	return []capability.Definition{
		{
			ID:   "core.asr",
			Type: capability.TypeASR,
			Name: "Core ASR",
			Description: "Standard ASR capability",
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"audio_data": {Type: "string", Description: "Audio data bytes"},
				},
				Required: []string{"audio_data"},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"text": {Type: "string"},
				},
			},
		},
		{
			ID:   "core.llm",
			Type: capability.TypeLLM,
			Name: "Core LLM",
			Description: "Standard LLM capability",
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"prompt": {Type: "string"},
					"text":   {Type: "string"},
					"history": {Type: "array"},
				},
				Required: []string{"prompt"},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"text": {Type: "string"},
				},
			},
		},
		{
			ID:   "core.tts",
			Type: capability.TypeTTS,
			Name: "Core TTS",
			Description: "Standard TTS capability",
			InputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"text": {Type: "string"},
				},
				Required: []string{"text"},
			},
			OutputSchema: capability.Schema{
				Type: "object",
				Properties: map[string]capability.Property{
					"audio_file": {Type: "string"},
					"audio_data": {Type: "string"},
				},
			},
		},
	}
}

func (p *CoreProvider) CreateExecutor(capabilityID string) (capability.Executor, error) {
	switch capabilityID {
	case "core.asr":
		return &ASRExecutor{provider: p.asr}, nil
	case "core.llm":
		return &LLMExecutor{provider: p.llm}, nil
	case "core.tts":
		return &TTSExecutor{provider: p.tts}, nil
	default:
		return nil, fmt.Errorf("unknown capability: %s", capabilityID)
	}
}
