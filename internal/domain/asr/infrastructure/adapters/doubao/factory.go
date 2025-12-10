package doubao

import (
	"xiaozhi-server-go/internal/platform/logging"
	"fmt"

	contractProviders "xiaozhi-server-go/internal/contracts/providers"
	"xiaozhi-server-go/internal/platform/config"
)

// DoubaoASRFactory Doubao ASR提供者工厂
type DoubaoASRFactory struct {
	providerName string
}

// NewDoubaoASRFactory 创建Doubao ASR工厂
func NewDoubaoASRFactory() contractProviders.ASRProviderFactory {
	return &DoubaoASRFactory{
		providerName: "doubao",
	}
}

// GetProviderName 获取提供者名称
func (f *DoubaoASRFactory) GetProviderName() string {
	return f.providerName
}

// ValidateConfig 验证配置
func (f *DoubaoASRFactory) ValidateConfig(cfg interface{}) error {
	c, ok := cfg.(Config)
	if !ok {
		return fmt.Errorf("invalid config type, expected doubao.Config")
	}

	if c.AppID == "" {
		return fmt.Errorf("app_id is required")
	}

	if c.AccessToken == "" {
		return fmt.Errorf("access_token is required")
	}

	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	if c.WSURL == "" {
		return fmt.Errorf("ws_url is required")
	}

	return nil
}

// CreateProvider 创建ASR提供者实例
func (f *DoubaoASRFactory) CreateProvider(cfg interface{}, options map[string]interface{}) (contractProviders.ASRProvider, error) {
	// 解析配置
	var c Config

	// 尝试从platform.Config转换为Doubao配置
	if platformConfig, ok := cfg.(*config.Config); ok {
		c = f.extractFromPlatformConfig(platformConfig)
	} else if doubaoConfig, ok := cfg.(Config); ok {
		c = doubaoConfig
	} else {
		return nil, fmt.Errorf("unsupported config type for doubao ASR provider")
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
	provider := NewDoubaoASRProvider(c, logger)

	// 初始化提供者
	if err := provider.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	return provider, nil
}

// extractFromPlatformConfig 从平台配置提取Doubao配置
func (f *DoubaoASRFactory) extractFromPlatformConfig(platformConfig *config.Config) Config {
	doubaoConfig, ok := platformConfig.ASR["doubao"].(map[string]interface{})
	if !ok {
		return Config{}
	}

	getString := func(key string) string {
		if v, ok := doubaoConfig[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	getInt := func(key string) int {
		if v, ok := doubaoConfig[key]; ok {
			if i, ok := v.(float64); ok {
				return int(i)
			}
			if i, ok := v.(int); ok {
				return i
			}
		}
		return 0
	}

	getBool := func(key string) bool {
		if v, ok := doubaoConfig[key]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
		return false
	}

	return Config{
		AppID:         getString("app_id"),
		AccessToken:   getString("access_token"),
		Host:          getString("host"),
		WSURL:         getString("ws_url"),
		ChunkDuration: getInt("chunk_duration"),
		ModelName:     getString("model"),
		EndWindowSize: getInt("end_window_size"),
		EnablePunc:    getBool("enable_punc"),
		EnableITN:     getBool("enable_itn"),
		EnableDDC:     getBool("enable_ddc"),
		OutputDir:     getString("output_dir"),
	}
}


