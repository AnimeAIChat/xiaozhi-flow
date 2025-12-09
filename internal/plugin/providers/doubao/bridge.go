package doubao

import (
	"xiaozhi-server-go/internal/core/providers/asr"
	"xiaozhi-server-go/internal/core/providers/tts"
	"xiaozhi-server-go/internal/utils"
)

func init() {
	asr.Register("doubao", NewCoreASRProvider)
	tts.Register("doubao", NewCoreTTSProvider)
}

func NewCoreASRProvider(config *asr.Config, deleteFile bool, logger *utils.Logger) (asr.Provider, error) {
	localConfig := &ASRConfig{
		Name: config.Name,
		Type: config.Type,
		Data: config.Data,
	}
	
	// Create the plugin provider
	// Note: We pass nil for session as the core system handles connection differently or doesn't provide it here
	return NewASRProvider(localConfig, deleteFile, logger, nil)
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
