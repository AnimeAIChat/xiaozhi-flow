package openai

import (
	"xiaozhi-server-go/internal/platform/logging"
	"fmt"
	"time"

	contractProviders "xiaozhi-server-go/internal/contracts/providers"
	"xiaozhi-server-go/internal/platform/config"
)

// OpenAILLMFactory OpenAI LLM提供者工厂
type OpenAILLMFactory struct {
	providerName string
}

// NewOpenAILLMFactory 创建OpenAI LLM工厂
func NewOpenAILLMFactory() contractProviders.LLMProviderFactory {
	return &OpenAILLMFactory{
		providerName: "openai",
	}
}

// GetProviderName 获取提供者名称
func (f *OpenAILLMFactory) GetProviderName() string {
	return f.providerName
}

// ValidateConfig 验证配置
func (f *OpenAILLMFactory) ValidateConfig(cfg interface{}) error {
	c, ok := cfg.(Config)
	if !ok {
		return fmt.Errorf("invalid config type, expected openai.Config")
	}

	if c.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	if c.Model == "" {
		return fmt.Errorf("model is required")
	}

	// 验证温度参数范围
	if c.Temperature < 0 || c.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	// 验证最大token数
	if c.MaxTokens < 1 || c.MaxTokens > 8192 {
		return fmt.Errorf("max_tokens must be between 1 and 8192")
	}

	return nil
}

// CreateProvider 创建LLM提供者实例
func (f *OpenAILLMFactory) CreateProvider(cfg interface{}, options map[string]interface{}) (contractProviders.LLMProvider, error) {
	// 解析配置
	var c Config

	// 尝试从platform.Config转换为OpenAI配置
	if platformConfig, ok := cfg.(*config.Config); ok {
		c = f.extractFromPlatformConfig(platformConfig)
	} else if openaiConfig, ok := cfg.(Config); ok {
		c = openaiConfig
	} else {
		return nil, fmt.Errorf("unsupported config type for openai LLM provider")
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
	provider := NewOpenAILLMProvider(c, logger)

	// 初始化提供者
	if err := provider.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	return provider, nil
}

// extractFromPlatformConfig 从平台配置提取OpenAI配置
func (f *OpenAILLMFactory) extractFromPlatformConfig(platformConfig *config.Config) Config {
	llmConfig := platformConfig.LLM["openai"]
	return Config{
		APIKey:      llmConfig.APIKey,
		BaseURL:     llmConfig.BaseURL,
		Model:       llmConfig.ModelName,
		MaxTokens:   llmConfig.MaxTokens,
		Temperature: float32(llmConfig.Temperature),
		Timeout:     30 * time.Second,
	}
}


