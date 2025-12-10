package manager

import (
	"encoding/json"
	"fmt"

	"xiaozhi-server-go/internal/domain/config/types"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/errors"
	"xiaozhi-server-go/internal/platform/storage"

	"gorm.io/gorm"
)

// DatabaseRepository 基于数据库的配置存储库实现
type DatabaseRepository struct {
	db *gorm.DB
}

// NewDatabaseRepository 创建新的数据库配置存储库
func NewDatabaseRepository(db interface{}) types.Repository {
	if db == nil {
		return &DatabaseRepository{db: storage.GetDB()}
	}
	if gormDB, ok := db.(*gorm.DB); ok {
		return &DatabaseRepository{db: gormDB}
	}
	return &DatabaseRepository{db: storage.GetDB()}
}

// LoadConfig 加载配置
func (r *DatabaseRepository) LoadConfig() (*config.Config, error) {
	// 首先尝试从数据库加载配置
	cfg, err := r.loadConfigFromDB()
	if err != nil {
		// 如果数据库加载失败，返回默认配置
		return config.DefaultConfig(), nil
	}

	if cfg != nil {
		return cfg, nil
	}

	// 如果数据库中没有配置，初始化默认配置
	return r.InitDefaultConfig()
}

// SaveConfig 保存配置
func (r *DatabaseRepository) SaveConfig(cfg *config.Config) error {
	if cfg == nil {
		return errors.Wrap(errors.KindDomain, "config.save", "config cannot be nil", nil)
	}

	// 使用事务确保原子性
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 保存 Providers (LLM, TTS, ASR, VLLLM)
	if err := tx.Exec("DELETE FROM providers").Error; err != nil {
		tx.Rollback()
		return errors.Wrap(errors.KindStorage, "config.save", "failed to delete existing providers", err)
	}

	saveProvider := func(id, pType string, configData interface{}) error {
		p := storage.Provider{
			ID:      id,
			Type:    pType,
			Name:    id, // 默认使用ID作为Name
			Config:  storage.FlexibleJSON{Data: configData},
			Enabled: true,
		}
		
		// 尝试提取更友好的 Name
		if m, ok := configData.(config.LLMConfig); ok {
			p.Name = m.Type
		} else if m, ok := configData.(config.TTSConfig); ok {
			p.Name = m.Type
		} else if m, ok := configData.(config.VLLLMConfig); ok {
			p.Name = m.Type
		} else if m, ok := configData.(map[string]interface{}); ok {
			if name, ok := m["type"].(string); ok {
				p.Name = name
			}
		}

		return tx.Create(&p).Error
	}

	for id, c := range cfg.LLM {
		if err := saveProvider(id, "llm", c); err != nil {
			tx.Rollback()
			return err
		}
	}
	for id, c := range cfg.TTS {
		if err := saveProvider(id, "tts", c); err != nil {
			tx.Rollback()
			return err
		}
	}
	for id, c := range cfg.ASR {
		if err := saveProvider(id, "asr", c); err != nil {
			tx.Rollback()
			return err
		}
	}
	for id, c := range cfg.VLLLM {
		if err := saveProvider(id, "vllm", c); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 2. 保存 Plugins
	if err := tx.Exec("DELETE FROM plugins").Error; err != nil {
		tx.Rollback()
		return errors.Wrap(errors.KindStorage, "config.save", "failed to delete existing plugins", err)
	}

	for id, c := range cfg.Plugins {
		p := storage.Plugin{
			ID:          id,
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
			Config:      storage.FlexibleJSON{Data: c.Config},
			Enabled:     c.Enabled,
		}
		if err := tx.Create(&p).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 3. 保存基础配置 (config_records)
	// 创建副本并清空已保存到其他表的字段，避免重复保存
	cfgCopy := *cfg
	cfgCopy.LLM = nil
	cfgCopy.TTS = nil
	cfgCopy.ASR = nil
	cfgCopy.VLLLM = nil
	cfgCopy.Plugins = nil

	// 将配置转换为键值对并保存到数据库
	configMap, err := r.configToMap(&cfgCopy)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(errors.KindDomain, "config.save", "failed to convert config to map", err)
	}

	// 先将所有现有配置标记为非活跃
	if err := tx.Exec("UPDATE config_records SET is_active = ?", false).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(errors.KindStorage, "config.save", "failed to deactivate existing configs", err)
	}

	// 删除所有非活跃的配置记录
	if err := tx.Exec("DELETE FROM config_records WHERE is_active = ?", false).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(errors.KindStorage, "config.save", "failed to delete inactive configs", err)
	}

	// 保存新的配置记录
	for key, value := range configMap {
		category := r.getCategoryFromKey(key)
		description := r.getDescriptionFromKey(key)

		record := storage.ConfigRecord{
			Key:         key,
			Value:       storage.FlexibleJSON{Data: value},
			Description: description,
			Category:    category,
			Version:     1,
			IsActive:    true,
		}

		if err := tx.Create(&record).Error; err != nil {
			tx.Rollback()
			return errors.Wrap(errors.KindStorage, "config.save", fmt.Sprintf("failed to save config record for key %s", key), err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return errors.Wrap(errors.KindStorage, "config.save", "failed to commit transaction", err)
	}

	return nil
}

// InitDefaultConfig 初始化默认配置
func (r *DatabaseRepository) InitDefaultConfig() (*config.Config, error) {
	defaultCfg := config.DefaultConfig()
	if err := r.SaveConfig(defaultCfg); err != nil {
		return nil, errors.Wrap(errors.KindDomain, "config.init", "failed to save default config", err)
	}

	// 初始化默认的模型选择
	modelSelectionManager := NewModelSelectionManager(r.db)
	if err := modelSelectionManager.InitDefaultSelection(1); err != nil { // 使用管理员用户ID 1
		return nil, errors.Wrap(errors.KindDomain, "config.init", "failed to init default model selection", err)
	}

	return defaultCfg, nil
}

// IsInitialized 检查配置是否已初始化
func (r *DatabaseRepository) IsInitialized() (bool, error) {
	// 检查数据库是否为nil
	if r.db == nil {
		return false, nil // 数据库未初始化，配置也未初始化
	}

	var count int64
	if err := r.db.Model(&storage.ConfigRecord{}).Where("is_active = ?", true).Count(&count).Error; err != nil {
		return false, errors.Wrap(errors.KindStorage, "config.check_init", "failed to check config initialization", err)
	}
	return count > 0, nil
}

// loadConfigFromDB 从数据库加载配置
func (r *DatabaseRepository) loadConfigFromDB() (*config.Config, error) {
	// 检查数据库是否为nil
	if r.db == nil {
		return config.DefaultConfig(), nil
	}

	// 检查数据库连接是否有效
	if _, err := r.db.DB(); err != nil {
		return config.DefaultConfig(), nil
	}

	// 1. 加载基础配置 (config_records)
	var cfg *config.Config
	
	rows, err := r.db.Raw("SELECT key, value FROM config_records WHERE is_active = ?", true).Rows()
	if err == nil {
		defer rows.Close()
		configMap := make(map[string]interface{})
		for rows.Next() {
			var key string
			var valueStr string
			if err := rows.Scan(&key, &valueStr); err != nil {
				continue
			}
			var value interface{}
			if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
				continue
			}
			configMap[key] = value
		}

		if len(configMap) > 0 {
			nested := r.unflattenMap(configMap)
			data, err := json.Marshal(nested)
			if err == nil {
				var c config.Config
				if err := json.Unmarshal(data, &c); err == nil {
					cfg = &c
				}
			}
		}
	}

	// 如果没有基础配置，可能需要初始化默认值，或者返回nil让上层处理
	// 这里我们假设如果config_records为空，则返回nil，触发InitDefaultConfig
	if cfg == nil {
		return nil, nil
	}

	// 2. 加载 Providers (LLM, TTS, ASR, VLLLM)
	var providers []storage.Provider
	if err := r.db.Find(&providers).Error; err == nil {
		if cfg.LLM == nil { cfg.LLM = make(map[string]config.LLMConfig) }
		if cfg.TTS == nil { cfg.TTS = make(map[string]config.TTSConfig) }
		if cfg.ASR == nil { cfg.ASR = make(map[string]interface{}) }
		if cfg.VLLLM == nil { cfg.VLLLM = make(map[string]config.VLLLMConfig) }

		for _, p := range providers {
			if !p.Enabled {
				continue
			}
			
			// 确保 Data 是正确的类型
			var dataBytes []byte
			if p.Config.Data != nil {
				dataBytes, _ = json.Marshal(p.Config.Data)
			}

			switch p.Type {
			case "llm":
				var llmConfig config.LLMConfig
				if len(dataBytes) > 0 {
					json.Unmarshal(dataBytes, &llmConfig)
				}
				cfg.LLM[p.ID] = llmConfig
			case "tts":
				var ttsConfig config.TTSConfig
				if len(dataBytes) > 0 {
					json.Unmarshal(dataBytes, &ttsConfig)
				}
				cfg.TTS[p.ID] = ttsConfig
			case "asr":
				var asrConfig map[string]interface{}
				if len(dataBytes) > 0 {
					json.Unmarshal(dataBytes, &asrConfig)
				}
				cfg.ASR[p.ID] = asrConfig
			case "vllm":
				var vllmConfig config.VLLLMConfig
				if len(dataBytes) > 0 {
					json.Unmarshal(dataBytes, &vllmConfig)
				}
				cfg.VLLLM[p.ID] = vllmConfig
			}
		}
	}

	// 3. 加载 Plugins
	var plugins []storage.Plugin
	if err := r.db.Find(&plugins).Error; err == nil {
		if cfg.Plugins == nil { cfg.Plugins = make(map[string]config.PluginConfig) }
		
		for _, p := range plugins {
			var pluginConfig config.PluginConfig
			pluginConfig.ID = p.ID
			pluginConfig.Name = p.Name
			pluginConfig.Type = p.Type
			pluginConfig.Description = p.Description
			pluginConfig.Enabled = p.Enabled
			
			if p.Config.Data != nil {
				dataBytes, _ := json.Marshal(p.Config.Data)
				json.Unmarshal(dataBytes, &pluginConfig.Config)
			}
			
			cfg.Plugins[p.ID] = pluginConfig
		}
	}

	return cfg, nil
}

// configToMap 将配置对象转换为键值对映射
func (r *DatabaseRepository) configToMap(cfg *config.Config) (map[string]interface{}, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal(data, &configMap); err != nil {
		return nil, err
	}

	// 展平嵌套结构为键值对
	flattened := make(map[string]interface{})
	r.flattenMap("", configMap, flattened)
	return flattened, nil
}

// flattenMap 将嵌套映射展平为键值对
func (r *DatabaseRepository) flattenMap(prefix string, src map[string]interface{}, dst map[string]interface{}) {
	for key, value := range src {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if nested, ok := value.(map[string]interface{}); ok {
			r.flattenMap(fullKey, nested, dst)
		} else {
			dst[fullKey] = value
		}
	}
}

// unflattenMap 将展平的键值对重新构建为嵌套映射
func (r *DatabaseRepository) unflattenMap(src map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range src {
		r.setNestedValue(result, key, value)
	}

	return result
}

// setNestedValue 在嵌套映射中设置值
func (r *DatabaseRepository) setNestedValue(m map[string]interface{}, key string, value interface{}) {
	parts := r.splitKey(key)
	current := m

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
		} else {
			if current[part] == nil {
				current[part] = make(map[string]interface{})
			}
			if nested, ok := current[part].(map[string]interface{}); ok {
				current = nested
			} else {
				newMap := make(map[string]interface{})
				current[part] = newMap
				current = newMap
			}
		}
	}
}

// splitKey 按点分割键
func (r *DatabaseRepository) splitKey(key string) []string {
	// 简单的点分割，实际实现中可能需要处理转义
	var parts []string
	var current string

	for _, char := range key {
		if char == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// getCategoryFromKey 从键获取分类
func (r *DatabaseRepository) getCategoryFromKey(key string) string {
	parts := r.splitKey(key)
	if len(parts) > 0 {
		return parts[0]
	}
	return "general"
}

// GetConfigValue 获取单个配置项的值
func (r *DatabaseRepository) GetConfigValue(key string) (interface{}, error) {
	var valueStr string
	row := r.db.Raw("SELECT value FROM config_records WHERE key = ? AND is_active = ?", key, true).Row()
	if err := row.Scan(&valueStr); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.Wrap(errors.KindDomain, "config.get_value", fmt.Sprintf("config key %s not found", key), err)
		}
		return nil, errors.Wrap(errors.KindStorage, "config.get_value", fmt.Sprintf("failed to get config value for key %s", key), err)
	}

	var value interface{}
	if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
		return nil, errors.Wrap(errors.KindStorage, "config.get_value", fmt.Sprintf("failed to unmarshal config value for key %s", key), err)
	}

	return value, nil
}

// GetBoolConfigValue 获取布尔类型的配置值
func (r *DatabaseRepository) GetBoolConfigValue(key string) (bool, error) {
	value, err := r.GetConfigValue(key)
	if err != nil {
		return false, err
	}

	if boolValue, ok := value.(bool); ok {
		return boolValue, nil
	}

	return false, errors.Wrap(errors.KindDomain, "config.get_bool", fmt.Sprintf("config key %s is not a boolean value", key), nil)
}

// GetStringArrayConfigValue 获取字符串数组类型的配置值
func (r *DatabaseRepository) GetStringArrayConfigValue(key string) ([]string, error) {
	value, err := r.GetConfigValue(key)
	if err != nil {
		return nil, err
	}

	if arrayValue, ok := value.([]interface{}); ok {
		result := make([]string, len(arrayValue))
		for i, v := range arrayValue {
			if str, ok := v.(string); ok {
				result[i] = str
			} else {
				return nil, errors.Wrap(errors.KindDomain, "config.get_string_array", fmt.Sprintf("config key %s contains non-string value at index %d", key, i), nil)
			}
		}
		return result, nil
	}

	return nil, errors.Wrap(errors.KindDomain, "config.get_string_array", fmt.Sprintf("config key %s is not a string array", key), nil)
}

// getDescriptionFromKey 从键获取描述
func (r *DatabaseRepository) getDescriptionFromKey(key string) string {
	descriptions := map[string]string{
		"server":        "服务器配置",
		"log":           "日志配置",
		"web":           "Web服务配置",
		"transport":     "传输层配置",
		"system":        "系统配置",
		"audio":         "音频配置",
		"pool":          "连接池配置",
		"mcp_pool":      "MCP连接池配置",
		"quick_reply":   "快速回复配置",
		"local_mcp_fun": "本地MCP函数配置",
		"asr":           "语音识别配置",
		"tts":           "语音合成配置",
		"llm":           "大语言模型配置",
		"vllm":          "视觉语言模型配置",
		"mcp":           "模型上下文协议配置",
		"selected":      "选中的服务配置",
	}

	category := r.getCategoryFromKey(key)
	if desc, ok := descriptions[category]; ok {
		return desc
	}
	return fmt.Sprintf("%s 配置", category)
}

