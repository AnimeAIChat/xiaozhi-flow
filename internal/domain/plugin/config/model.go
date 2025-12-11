package config

import (
	"time"

	"xiaozhi-server-go/internal/platform/errors"
)

// ProviderType 供应商类型
type ProviderType string

const (
	ProviderTypeOpenAI   ProviderType = "openai"   // OpenAI供应商
	ProviderTypeDoubao   ProviderType = "doubao"   // 豆包供应商
	ProviderTypeEdge     ProviderType = "edge"     // Edge TTS供应商
	ProviderTypeDeepgram ProviderType = "deepgram" // Deepgram供应商
	ProviderTypeOllama   ProviderType = "ollama"   // Ollama本地供应商
	ProviderTypeStepfun  ProviderType = "stepfun"  // Stepfun供应商
	ProviderTypeChatglm  ProviderType = "chatglm"  // ChatGLM供应商
	ProviderTypeCoze     ProviderType = "coze"     // Coze供应商
	ProviderTypeGosherpa ProviderType = "gosherpa" // Gosherpa供应商
)

// CapabilityType 能力类型
type CapabilityType string

const (
	CapabilityTypeLLM  CapabilityType = "llm"  // 大语言模型能力
	CapabilityTypeASR  CapabilityType = "asr"  // 语音识别能力
	CapabilityTypeTTS  CapabilityType = "tts"  // 文字转语音能力
	CapabilityTypeTool CapabilityType = "tool" // 工具能力
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"   // 健康
	HealthStatusUnhealthy HealthStatus = "unhealthy" // 不健康
	HealthStatusUnknown   HealthStatus = "unknown"   // 未知
)

// HistoryOperation 操作类型
type HistoryOperation string

const (
	OperationCreate  HistoryOperation = "create"  // 创建
	OperationUpdate  HistoryOperation = "update"  // 更新
	OperationDelete  HistoryOperation = "delete"  // 删除
	OperationEnable  HistoryOperation = "enable"  // 启用
	OperationDisable HistoryOperation = "disable" // 禁用
	OperationTest    HistoryOperation = "test"    // 测试
)

// ProviderConfig 供应商配置聚合根
type ProviderConfig struct {
	ID              int           `json:"id" gorm:"primaryKey"`
	ProviderType    ProviderType  `json:"providerType" gorm:"type:varchar(100);not null;index"`
	ProviderName    string        `json:"providerName" gorm:"type:varchar(255);not null;uniqueIndex:idx_provider_type_name"`
	DisplayName     string        `json:"displayName" gorm:"type:varchar(255);not null"`
	Description     string        `json:"description" gorm:"type:text"`
	ConfigData      string        `json:"-" gorm:"type:text;not null"`             // 加密的配置数据，不序列化到JSON
	ConfigSchema    string        `json:"configSchema" gorm:"type:text;not null"` // 配置模式定义
	Enabled         bool          `json:"enabled" gorm:"default:true;index"`
	Priority        int           `json:"priority" gorm:"default:100;index"`
	HealthStatus    HealthStatus  `json:"healthStatus" gorm:"type:varchar(50);default:'unknown';index"`
	LastHealthCheck *time.Time    `json:"lastHealthCheck"`
	CreatedAt       time.Time     `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time     `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联数据
	Capabilities []Capability      `json:"capabilities" gorm:"foreignKey:ProviderConfigID;constraint:OnDelete:CASCADE"`
	Snapshots    []ConfigSnapshot  `json:"snapshots,omitempty" gorm:"foreignKey:ProviderConfigID;constraint:OnDelete:CASCADE"`
	History      []ConfigHistory   `json:"-" gorm:"foreignKey:ProviderConfigID;constraint:OnDelete:CASCADE"`
}

// Capability 能力实体
type Capability struct {
	ID                    int             `json:"id" gorm:"primaryKey"`
	ProviderConfigID      int             `json:"providerConfigId" gorm:"not null;index"`
	CapabilityID          string          `json:"capabilityId" gorm:"type:varchar(255);not null;uniqueIndex:idx_config_capability"`
	CapabilityType        CapabilityType  `json:"capabilityType" gorm:"type:varchar(50);not null;index"`
	CapabilityName        string          `json:"capabilityName" gorm:"type:varchar(255);not null"`
	CapabilityDescription string          `json:"capabilityDescription" gorm:"type:text"`
	InputSchema           string          `json:"inputSchema" gorm:"type:text"`
	OutputSchema          string          `json:"outputSchema" gorm:"type:text"`
	Enabled               bool            `json:"enabled" gorm:"default:true;index"`
	CreatedAt             time.Time       `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt             time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ConfigSnapshot 配置快照实体
type ConfigSnapshot struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	ProviderConfigID int     `json:"providerConfigId" gorm:"not null;index"`
	Version       string    `json:"version" gorm:"type:varchar(50);not null;index"`
	SnapshotName  string    `json:"snapshotName" gorm:"type:varchar(255);not null"`
	Description   string    `json:"description" gorm:"type:text"`
	SnapshotData  string    `json:"snapshotData" gorm:"type:text;not null"`
	IsActive      bool      `json:"isActive" gorm:"default:false;index"`
	CreatedBy     string    `json:"createdBy" gorm:"type:varchar(255)"`
	CreatedAt     time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// ConfigHistory 配置变更历史实体
type ConfigHistory struct {
	ID              int              `json:"id" gorm:"primaryKey"`
	ProviderConfigID int              `json:"providerConfigId" gorm:"not null;index"`
	Operation       HistoryOperation  `json:"operation" gorm:"type:varchar(50);not null;index"`
	OldData         string           `json:"-" gorm:"type:text"`                 // 变更前数据，不序列化到JSON
	NewData         string           `json:"-" gorm:"type:text"`                 // 变更后数据，不序列化到JSON
	ChangeSummary   string           `json:"changeSummary" gorm:"type:varchar(1000)"`
	ChangedFields   string           `json:"-" gorm:"type:text"`                 // 变更字段列表，不序列化到JSON
	CreatedBy       string           `json:"createdBy" gorm:"type:varchar(255)"`
	UserAgent       string           `json:"userAgent" gorm:"type:text"`
	IPAddress       string           `json:"ipAddress" gorm:"type:varchar(45)"`
	CreatedAt       time.Time        `json:"createdAt" gorm:"autoCreateTime;index"`
}

// TableName 指定表名
func (ProviderConfig) TableName() string {
	return "plugin_provider_configs"
}

func (Capability) TableName() string {
	return "plugin_capabilities"
}

func (ConfigSnapshot) TableName() string {
	return "plugin_config_snapshots"
}

func (ConfigHistory) TableName() string {
	return "plugin_config_history"
}

// NewProviderConfig 创建新的供应商配置
func NewProviderConfig(providerType ProviderType, providerName, displayName, description string) (*ProviderConfig, error) {
	if providerType == "" {
		return nil, errors.New(errors.KindDomain, "provider_config.new", "provider type cannot be empty")
	}
	if providerName == "" {
		return nil, errors.New(errors.KindDomain, "provider_config.new", "provider name cannot be empty")
	}
	if displayName == "" {
		return nil, errors.New(errors.KindDomain, "provider_config.new", "display name cannot be empty")
	}

	now := time.Now()
	return &ProviderConfig{
		ProviderType: providerType,
		ProviderName: providerName,
		DisplayName:  displayName,
		Description:  description,
		Enabled:      true,
		Priority:     100,
		HealthStatus: HealthStatusUnknown,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// UpdateConfig 更新配置数据
func (pc *ProviderConfig) UpdateConfig(configData, configSchema string) error {
	pc.ConfigData = configData
	pc.ConfigSchema = configSchema
	pc.UpdatedAt = time.Now()
	return nil
}

// SetEnabled 设置启用状态
func (pc *ProviderConfig) SetEnabled(enabled bool) {
	pc.Enabled = enabled
	pc.UpdatedAt = time.Now()
}

// UpdateHealthStatus 更新健康状态
func (pc *ProviderConfig) UpdateHealthStatus(status HealthStatus) {
	pc.HealthStatus = status
	now := time.Now()
	pc.LastHealthCheck = &now
	pc.UpdatedAt = now
}

// SetPriority 设置优先级
func (pc *ProviderConfig) SetPriority(priority int) {
	pc.Priority = priority
	pc.UpdatedAt = time.Now()
}

// IsHealthy 检查是否健康
func (pc *ProviderConfig) IsHealthy() bool {
	return pc.HealthStatus == HealthStatusHealthy
}

// GetEnabledCapabilities 获取启用的能力列表
func (pc *ProviderConfig) GetEnabledCapabilities() []Capability {
	var enabled []Capability
	for _, cap := range pc.Capabilities {
		if cap.Enabled {
			enabled = append(enabled, cap)
		}
	}
	return enabled
}

// NewCapability 创建新的能力
func NewCapability(providerConfigID int, capabilityID, capabilityName, capabilityType, description string) (*Capability, error) {
	if providerConfigID <= 0 {
		return nil, errors.New(errors.KindDomain, "capability.new", "invalid provider config id")
	}
	if capabilityID == "" {
		return nil, errors.New(errors.KindDomain, "capability.new", "capability ID cannot be empty")
	}
	if capabilityName == "" {
		return nil, errors.New(errors.KindDomain, "capability.new", "capability name cannot be empty")
	}

	now := time.Now()
	return &Capability{
		ProviderConfigID:      providerConfigID,
		CapabilityID:          capabilityID,
		CapabilityType:        CapabilityType(capabilityType),
		CapabilityName:        capabilityName,
		CapabilityDescription: description,
		Enabled:               true,
		CreatedAt:             now,
		UpdatedAt:             now,
	}, nil
}

// SetEnabled 设置能力启用状态
func (c *Capability) SetEnabled(enabled bool) {
	c.Enabled = enabled
	c.UpdatedAt = time.Now()
}

// UpdateSchemas 更新模式定义
func (c *Capability) UpdateSchemas(inputSchema, outputSchema string) {
	c.InputSchema = inputSchema
	c.OutputSchema = outputSchema
	c.UpdatedAt = time.Now()
}

// NewConfigSnapshot 创建新的配置快照
func NewConfigSnapshot(providerConfigID int, version, snapshotName, description, snapshotData, createdBy string) (*ConfigSnapshot, error) {
	if providerConfigID <= 0 {
		return nil, errors.New(errors.KindDomain, "config_snapshot.new", "invalid provider config id")
	}
	if version == "" {
		return nil, errors.New(errors.KindDomain, "config_snapshot.new", "version cannot be empty")
	}
	if snapshotName == "" {
		return nil, errors.New(errors.KindDomain, "config_snapshot.new", "snapshot name cannot be empty")
	}
	if snapshotData == "" {
		return nil, errors.New(errors.KindDomain, "config_snapshot.new", "snapshot data cannot be empty")
	}

	return &ConfigSnapshot{
		ProviderConfigID: providerConfigID,
		Version:          version,
		SnapshotName:     snapshotName,
		Description:      description,
		SnapshotData:     snapshotData,
		CreatedBy:        createdBy,
		CreatedAt:        time.Now(),
	}, nil
}

// SetActive 设置为激活状态
func (cs *ConfigSnapshot) SetActive(active bool) {
	cs.IsActive = active
}

// NewConfigHistory 创建新的配置变更历史
func NewConfigHistory(providerConfigID int, operation HistoryOperation, oldData, newData, changeSummary, changedFields, createdBy, userAgent, ipAddress string) (*ConfigHistory, error) {
	if providerConfigID <= 0 {
		return nil, errors.New(errors.KindDomain, "config_history.new", "invalid provider config id")
	}
	if operation == "" {
		return nil, errors.New(errors.KindDomain, "config_history.new", "operation cannot be empty")
	}

	return &ConfigHistory{
		ProviderConfigID: providerConfigID,
		Operation:        operation,
		OldData:          oldData,
		NewData:          newData,
		ChangeSummary:    changeSummary,
		ChangedFields:    changedFields,
		CreatedBy:        createdBy,
		UserAgent:        userAgent,
		IPAddress:        ipAddress,
		CreatedAt:        time.Now(),
	}, nil
}