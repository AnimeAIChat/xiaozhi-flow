package storage

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// FlexibleJSON 是一个灵活的JSON类型，可以处理格式不规范的数据
type FlexibleJSON struct {
	Data interface{}
}

// Value 实现 driver.Valuer 接口
func (j FlexibleJSON) Value() (driver.Value, error) {
	return json.Marshal(j.Data)
}

// Scan 实现 sql.Scanner 接口
func (j *FlexibleJSON) Scan(value interface{}) error {
	if value == nil {
		j.Data = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// 首先尝试解析为JSON
		var result interface{}
		if err := json.Unmarshal(v, &result); err == nil {
			j.Data = result
			return nil
		}

		// 如果JSON解析失败，将字节数组作为字符串处理
		j.Data = string(v)
		return nil

	case string:
		// 首先尝试解析为JSON
		var result interface{}
		if err := json.Unmarshal([]byte(v), &result); err == nil {
			j.Data = result
			return nil
		}

		// 如果JSON解析失败，直接存储字符串
		j.Data = v
		return nil

	default:
		// 对于其他类型，直接存储
		j.Data = v
		return nil
	}
}

// MarshalJSON 实现json.Marshaler接口
func (j FlexibleJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

// UnmarshalJSON 实现json.Unmarshaler接口
func (j *FlexibleJSON) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j.Data)
}

// ConfigRecord 完整的配置记录模型，用于数据库存储
type ConfigRecord struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Key         string         `gorm:"uniqueIndex;not null" json:"key"` // 配置键，如 "server", "web", "llm.openai"
	Value       FlexibleJSON  `gorm:"type:json;not null" json:"value"`    // 配置值，JSON格式（使用灵活类型处理不规范数据）
	Description string         `gorm:"type:text" json:"description"`     // 配置描述
	Category    string         `gorm:"index" json:"category"`           // 配置分类，如 "server", "web", "llm", "tts", "asr"
	Version     int            `gorm:"default:1" json:"version"`         // 配置版本号，用于版本控制
	IsActive    bool           `gorm:"default:true" json:"is_active"`    // 是否为活动配置
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (ConfigRecord) TableName() string {
	return "config_records"
}

// ConfigSnapshot 配置快照，用于备份和版本控制
type ConfigSnapshot struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	Name      string        `gorm:"not null" json:"name"`      // 快照名称
	Version   int           `gorm:"not null" json:"version"`   // 快照版本
	Data      FlexibleJSON  `gorm:"type:json;not null" json:"data"`      // 完整配置数据
	Comment   string        `gorm:"type:text" json:"comment"`  // 快照注释
	CreatedAt time.Time     `json:"created_at"`
}

// TableName 指定表名
func (ConfigSnapshot) TableName() string {
	return "config_snapshots"
}

// ModelSelection 模型选择记录，用于管理用户选择的AI模型提供商
type ModelSelection struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        int       `gorm:"not null;uniqueIndex" json:"user_id"`                   // 用户ID（整数），外键关联users表
	ASRProvider   string    `gorm:"not null" json:"asr_provider"`                        // 选择的ASR提供商
	TTSProvider   string    `gorm:"not null" json:"tts_provider"`                        // 选择的TTS提供商
	LLMProvider   string    `gorm:"not null" json:"llm_provider"`                        // 选择的LLM提供商
	VLLMProvider  string    `gorm:"not null" json:"vllm_provider"`                       // 选择的VLLM提供商
	IsActive      bool      `gorm:"default:true" json:"is_active"`                        // 是否为活动选择
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ModelSelection) TableName() string {
	return "model_selections"
}