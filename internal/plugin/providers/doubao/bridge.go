package doubao

import (
	"xiaozhi-server-go/internal/core/providers/asr"
	"xiaozhi-server-go/internal/core/providers/tts"
	"xiaozhi-server-go/internal/core/providers"
	"xiaozhi-server-go/internal/utils"
)

func init() {
	asr.Register("doubao", NewCoreASRProvider)
	tts.Register("doubao", NewCoreTTSProvider)
}

// CoreASRWrapper wraps the plugin ASR provider to adapt it to the core interface
type CoreASRWrapper struct {
	*ASRProvider
}

// SetListener adapts the core listener to the plugin listener
func (w *CoreASRWrapper) SetListener(listener providers.AsrEventListener) {
	if listener == nil {
		w.ASRProvider.SetListener(nil)
		return
	}
	w.ASRProvider.SetListener(&ListenerAdapter{coreListener: listener})
}

// ListenerAdapter adapts providers.AsrEventListener to doubao.AsrEventListener
type ListenerAdapter struct {
	coreListener providers.AsrEventListener
}

func (l *ListenerAdapter) OnAsrResult(result string, isFinalResult bool) bool {
	return l.coreListener.OnAsrResult(result, isFinalResult)
}

func NewCoreASRProvider(config *asr.Config, deleteFile bool, logger *utils.Logger) (asr.Provider, error) {
	localConfig := &ASRConfig{
		Name: config.Name,
		Type: config.Type,
		Data: config.Data,
	}
	
	// Create the plugin provider
	// Note: We pass nil for session as the core system handles connection differently or doesn't provide it here
	provider, err := NewASRProvider(localConfig, deleteFile, logger, nil)
	if err != nil {
		return nil, err
	}
	
	return &CoreASRWrapper{ASRProvider: provider}, nil
}

func NewCoreTTSProvider(config *tts.Config, deleteFile bool) (tts.Provider, error) {
	localConfig := &TTSConfig{
		Name:            config.Name,
		Type:            config.Type,
		OutputDir:       config.OutputDir,
		Voice:           config.Voice,
		Format:          config.Format,
		SampleRate:      config.SampleRate,
		AppID:           config.AppID,
		Token:           config.Token,
		Cluster:         config.Cluster,
		SupportedVoices: config.SupportedVoices,
	}
	
	// Create the plugin provider
	provider, err := NewTTSProvider(localConfig, deleteFile)
	if err != nil {
		return nil, err
	}
	
	return provider, nil
}
