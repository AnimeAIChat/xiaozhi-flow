package edge

import (
	"xiaozhi-server-go/internal/platform/logging"
	"fmt"

	contractProviders "xiaozhi-server-go/internal/contracts/providers"
	"xiaozhi-server-go/internal/platform/config"
)

// EdgeTTSFactory Edge TTS提供者工厂
type EdgeTTSFactory struct {
	providerName string
}

// NewEdgeTTSFactory 创建Edge TTS工厂
func NewEdgeTTSFactory() contractProviders.TTSProviderFactory {
	return &EdgeTTSFactory{
		providerName: "edge",
	}
}

// GetProviderName 获取提供者名称
func (f *EdgeTTSFactory) GetProviderName() string {
	return f.providerName
}

// ValidateConfig 验证配置
func (f *EdgeTTSFactory) ValidateConfig(cfg interface{}) error {
	c, ok := cfg.(Config)
	if !ok {
		return fmt.Errorf("invalid config type, expected edge.Config")
	}

	// 验证采样率
	if c.SampleRate < 8000 || c.SampleRate > 48000 {
		return fmt.Errorf("sample_rate must be between 8000 and 48000")
	}

	// 验证速度参数
	if c.Speed < 0.25 || c.Speed > 3.0 {
		return fmt.Errorf("speed must be between 0.25 and 3.0")
	}

	// 验证音调参数
	if c.Pitch < -20.0 || c.Pitch > 20.0 {
		return fmt.Errorf("pitch must be between -20.0 and 20.0")
	}

	// 验证音量参数
	if c.Volume < 0.0 || c.Volume > 1.0 {
		return fmt.Errorf("volume must be between 0.0 and 1.0")
	}

	return nil
}

// CreateProvider 创建TTS提供者实例
func (f *EdgeTTSFactory) CreateProvider(cfg interface{}, options map[string]interface{}) (contractProviders.TTSProvider, error) {
	// 解析配置
	var c Config

	// 尝试从platform.Config转换为Edge配置
	if platformConfig, ok := cfg.(*config.Config); ok {
		c = f.extractFromPlatformConfig(platformConfig)
	} else if edgeConfig, ok := cfg.(Config); ok {
		c = edgeConfig
	} else {
		return nil, fmt.Errorf("unsupported config type for edge TTS provider")
	}

	// 验证配置
	if err := f.ValidateConfig(c); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// 获取日志记录器
	var logger *logging.Logger
	if loggerVal, ok := options["logger"]; ok {
		if logger, ok = loggerVal.(*logging.Logger); !ok {
			return nil, fmt.Errorf("logger option must be *logging.Logger")
		}
	} else {
		logger = logging.DefaultLogger
	}

	// 创建提供者实例
	provider := NewEdgeTTSProvider(c, logger)

	// 初始化提供者
	if err := provider.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	return provider, nil
}

// extractFromPlatformConfig 从平台配置提取Edge配置
func (f *EdgeTTSFactory) extractFromPlatformConfig(platformConfig *config.Config) Config {
	ttsConfig := platformConfig.TTS["edge"]

	getFloat := func(key string, def float32) float32 {
		if v, ok := ttsConfig.Extra[key]; ok {
			if f, ok := v.(float64); ok {
				return float32(f)
			}
			if f, ok := v.(float32); ok {
				return f
			}
		}
		return def
	}

	getBool := func(key string, def bool) bool {
		if v, ok := ttsConfig.Extra[key]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
		return def
	}

	getInt := func(key string, def int) int {
		if v, ok := ttsConfig.Extra[key]; ok {
			if i, ok := v.(float64); ok {
				return int(i)
			}
			if i, ok := v.(int); ok {
				return i
			}
		}
		return def
	}

	return Config{
		Voice:      ttsConfig.Voice,
		OutputDir:  ttsConfig.OutputDir,
		DeleteFile: getBool("delete_file", false),
		SampleRate: getInt("sample_rate", 16000),
		Format:     ttsConfig.Format,
		Speed:      getFloat("speed", 1.0),
		Pitch:      getFloat("pitch", 0.0),
		Volume:     getFloat("volume", 1.0),
	}
}


