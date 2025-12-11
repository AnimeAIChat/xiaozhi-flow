package config

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"xiaozhi-server-go/internal/platform/errors"
)

// ConfigValidator 配置验证器
type ConfigValidator struct {
	validate *validator.Validate
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		validate: validator.New(),
	}
}

// ValidateConfig 验证配置数据
func (v *ConfigValidator) ValidateConfig(configData map[string]interface{}, configSchema map[string]interface{}) error {
	// 将配置数据转换为JSON字符串
	configJSON, err := json.Marshal(configData)
	if err != nil {
		return errors.Wrap(errors.KindDomain, "config_validator.validate", "failed to marshal config data", err)
	}

	// 简化的JSON Schema验证
	// 在生产环境中，建议使用更完整的JSON Schema验证库如 github.com/xeipuuv/gojsonschema
	if err := v.validateJSONSchema(configJSON, configSchema); err != nil {
		return errors.Wrap(errors.KindDomain, "config_validator.validate", "config data validation failed", err)
	}

	return nil
}

// ValidateProviderName 验证供应商名称
func (v *ConfigValidator) ValidateProviderName(providerType ProviderType, providerName string) error {
	if providerName == "" {
		return errors.New(errors.KindDomain, "config_validator.validate_provider_name", "provider name cannot be empty")
	}

	// 检查供应商名称格式
	switch providerType {
	case ProviderTypeOpenAI:
		if providerName != "openai" {
			return errors.New(errors.KindDomain, "config_validator.validate_provider_name", "OpenAI provider name must be 'openai'")
		}
	case ProviderTypeDoubao:
		if providerName != "doubao" {
			return errors.New(errors.KindDomain, "config_validator.validate_provider_name", "Doubao provider name must be 'doubao'")
		}
	case ProviderTypeEdge:
		if providerName != "edge" {
			return errors.New(errors.KindDomain, "config_validator.validate_provider_name", "Edge provider name must be 'edge'")
		}
	case ProviderTypeDeepgram:
		if providerName != "deepgram" {
			return errors.New(errors.KindDomain, "config_validator.validate_provider_name", "Deepgram provider name must be 'deepgram'")
		}
	case ProviderTypeOllama:
		if providerName != "ollama" {
			return errors.New(errors.KindDomain, "config_validator.validate_provider_name", "Ollama provider name must be 'ollama'")
		}
	case ProviderTypeStepfun:
		if providerName != "stepfun" {
			return errors.New(errors.KindDomain, "config_validator.validate_provider_name", "Stepfun provider name must be 'stepfun'")
		}
	case ProviderTypeChatglm:
		if providerName != "chatglm" {
			return errors.New(errors.KindDomain, "config_validator.validate_provider_name", "ChatGLM provider name must be 'chatglm'")
		}
	case ProviderTypeCoze:
		if providerName != "coze" {
			return errors.New(errors.KindDomain, "config_validator.validate_provider_name", "Coze provider name must be 'coze'")
		}
	case ProviderTypeGosherpa:
		if providerName != "gosherpa" {
			return errors.New(errors.KindDomain, "config_validator.validate_provider_name", "Gosherpa provider name must be 'gosherpa'")
		}
	default:
		return errors.New(errors.KindDomain, "config_validator.validate_provider_name", fmt.Sprintf("unknown provider type: %s", providerType))
	}

	return nil
}

// validateJSONSchema 简化的JSON Schema验证
func (v *ConfigValidator) validateJSONSchema(configJSON []byte, schema map[string]interface{}) error {
	// 将schema转换为map[string]interface{}
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return errors.Wrap(errors.KindDomain, "config_validator.validate_json_schema", "failed to marshal schema", err)
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		return errors.Wrap(errors.KindDomain, "config_validator.validate_json_schema", "failed to unmarshal schema", err)
	}

	// 获取schema的properties
	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		return errors.New(errors.KindDomain, "config_validator.validate_json_schema", "invalid schema: missing properties")
	}

	// 将配置数据解析为map
	var configMap map[string]interface{}
	if err := json.Unmarshal(configJSON, &configMap); err != nil {
		return errors.Wrap(errors.KindDomain, "config_validator.validate_json_schema", "failed to unmarshal config data", err)
	}

	// 验证每个必需字段
	requiredFields, _ := schemaMap["required"].([]interface{})
	for _, field := range requiredFields {
		fieldName, ok := field.(string)
		if !ok {
			continue
		}

		if _, exists := configMap[fieldName]; !exists {
			return errors.New(errors.KindDomain, "config_validator.validate_json_schema", fmt.Sprintf("required field missing: %s", fieldName))
		}
	}

	// 验证字段类型和格式
	for fieldName, fieldSchema := range properties {
		fieldSchemaMap, ok := fieldSchema.(map[string]interface{})
		if !ok {
			continue
		}

		fieldValue, exists := configMap[fieldName]
		if !exists {
			continue
		}

		// 验证字段类型
		if fieldType, ok := fieldSchemaMap["type"].(string); ok {
			if err := v.validateFieldType(fieldName, fieldValue, fieldType); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateFieldType 验证字段类型
func (v *ConfigValidator) validateFieldType(fieldName string, value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return errors.New(errors.KindDomain, "config_validator.validate_field_type", fmt.Sprintf("field '%s' must be a string", fieldName))
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int64, int32:
		default:
			return errors.New(errors.KindDomain, "config_validator.validate_field_type", fmt.Sprintf("field '%s' must be a number", fieldName))
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return errors.New(errors.KindDomain, "config_validator.validate_field_type", fmt.Sprintf("field '%s' must be a boolean", fieldName))
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return errors.New(errors.KindDomain, "config_validator.validate_field_type", fmt.Sprintf("field '%s' must be an array", fieldName))
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return errors.New(errors.KindDomain, "config_validator.validate_field_type", fmt.Sprintf("field '%s' must be an object", fieldName))
		}
	}

	return nil
}

// GetConfigSchema 获取供应商的配置模式
func (v *ConfigValidator) GetConfigSchema(providerType ProviderType) map[string]interface{} {
	switch providerType {
	case ProviderTypeOpenAI:
		return map[string]interface{}{
			"type": "object",
			"required": []string{"api_key"},
			"properties": map[string]interface{}{
				"api_key": map[string]interface{}{
					"type": "string",
					"description": "OpenAI API密钥",
					"secret": true,
				},
				"base_url": map[string]interface{}{
					"type": "string",
					"description": "API基础URL，可选",
					"default": "https://api.openai.com/v1",
				},
				"model": map[string]interface{}{
					"type": "string",
					"description": "默认模型名称",
					"default": "gpt-3.5-turbo",
				},
				"max_tokens": map[string]interface{}{
					"type": "integer",
					"description": "最大token数",
					"default": 2048,
				},
				"temperature": map[string]interface{}{
					"type": "number",
					"description": "温度参数",
					"default": 0.7,
				},
			},
		}
	case ProviderTypeDoubao:
		return map[string]interface{}{
			"type": "object",
			"required": []string{"app_key", "app_secret"},
			"properties": map[string]interface{}{
				"app_key": map[string]interface{}{
					"type": "string",
					"description": "豆包应用密钥",
					"secret": true,
				},
				"app_secret": map[string]interface{}{
					"type": "string",
					"description": "豆包应用秘钥",
					"secret": true,
				},
				"endpoint_id": map[string]interface{}{
					"type": "string",
					"description": "端点ID",
				},
			},
		}
	case ProviderTypeEdge:
		return map[string]interface{}{
			"type": "object",
			"required": []string{},
			"properties": map[string]interface{}{
				"voice": map[string]interface{}{
					"type": "string",
					"description": "语音模型",
					"default": "zh-CN-XiaoxiaoNeural",
				},
				"rate": map[string]interface{}{
					"type": "string",
					"description": "语速",
					"default": "+0%",
				},
				"pitch": map[string]interface{}{
					"type": "string",
					"description": "音调",
					"default": "+0Hz",
				},
			},
		}
	default:
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		}
	}
}