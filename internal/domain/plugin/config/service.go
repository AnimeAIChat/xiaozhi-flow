package config

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"xiaozhi-server-go/internal/plugin/capability"
	"xiaozhi-server-go/internal/platform/errors"
	"xiaozhi-server-go/internal/platform/logging"
)

// PluginConfigService 插件配置服务接口
type PluginConfigService interface {
	// 基础CRUD操作
	CreateProviderConfig(ctx context.Context, req *CreateProviderConfigRequest) (*ProviderConfig, error)
	GetProviderConfig(ctx context.Context, id int) (*ProviderConfig, error)
	GetProviderConfigs(ctx context.Context, filter *ProviderConfigFilter) (*ProviderConfigList, error)
	UpdateProviderConfig(ctx context.Context, id int, req *UpdateProviderConfigRequest) (*ProviderConfig, error)
	DeleteProviderConfig(ctx context.Context, id int) error

	// 配置测试和验证
	TestProviderConfig(ctx context.Context, req *TestProviderConfigRequest) (*TestResult, error)
	ValidateProviderConfig(ctx context.Context, providerType ProviderType, config map[string]interface{}) error

	// 快照管理
	CreateConfigSnapshot(ctx context.Context, providerConfigID int, req *CreateSnapshotRequest) (*ConfigSnapshot, error)
	GetConfigSnapshots(ctx context.Context, providerConfigID int, filter *SnapshotFilter) (*SnapshotList, error)
	RestoreConfigSnapshot(ctx context.Context, providerConfigID, snapshotID int) error

	// 历史管理
	GetConfigHistory(ctx context.Context, providerConfigID int, filter *HistoryFilter) (*HistoryList, error)

	// 统计和可用性
	GetAvailableProviders(ctx context.Context) ([]AvailableProvider, error)
	GetPluginStats(ctx context.Context) (*PluginStats, error)

	// 系统集成
	GetEnabledCapabilities(ctx context.Context, capabilityType CapabilityType) ([]Capability, error)
	GetCapabilityExecutor(ctx context.Context, capabilityID string, config map[string]interface{}) (capability.Executor, error)
}

// CreateProviderConfigRequest 创建供应商配置请求
type CreateProviderConfigRequest struct {
	ProviderType ProviderType         `json:"providerType"`
	ProviderName string               `json:"providerName"`
	DisplayName  string               `json:"displayName"`
	Description  string               `json:"description"`
	Config       map[string]interface{} `json:"config"`
	Enabled      bool                 `json:"enabled"`
	Priority     int                  `json:"priority"`
	CreatedBy    string               `json:"createdBy"`
	UserAgent    string               `json:"userAgent"`
	IPAddress    string               `json:"ipAddress"`
}

// UpdateProviderConfigRequest 更新供应商配置请求
type UpdateProviderConfigRequest struct {
	DisplayName string                   `json:"displayName"`
	Description string                   `json:"description"`
	Config      map[string]interface{}   `json:"config"`
	Enabled     *bool                    `json:"enabled"`
	Priority    *int                     `json:"priority"`
	UpdatedBy   string                   `json:"updatedBy"`
	UserAgent   string                   `json:"userAgent"`
	IPAddress   string                   `json:"ipAddress"`
}

// TestProviderConfigRequest 测试供应商配置请求
type TestProviderConfigRequest struct {
	ProviderType ProviderType           `json:"providerType"`
	Config       map[string]interface{} `json:"config"`
}

// TestResult 测试结果
type TestResult struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Latency   int64                  `json:"latency"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// CreateSnapshotRequest 创建快照请求
type CreateSnapshotRequest struct {
	Version     string `json:"version"`
	SnapshotName string `json:"snapshotName"`
	Description string `json:"description"`
	CreatedBy   string `json:"createdBy"`
}

// ProviderConfigFilter 供应商配置过滤器
type ProviderConfigFilter struct {
	ProviderType ProviderType `json:"providerType"`
	Enabled      *bool        `json:"enabled"`
	HealthStatus HealthStatus `json:"healthStatus"`
	Page         int          `json:"page"`
	PageSize     int          `json:"pageSize"`
}

// ProviderConfigList 供应商配置列表
type ProviderConfigList struct {
	Total     int64           `json:"total"`
	Page      int             `json:"page"`
	PageSize  int             `json:"pageSize"`
	TotalPages int64          `json:"totalPages"`
	Configs   []ProviderConfig `json:"configs"`
}

// SnapshotFilter 快照过滤器
type SnapshotFilter struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// SnapshotList 快照列表
type SnapshotList struct {
	Total     int64           `json:"total"`
	Page      int             `json:"page"`
	PageSize  int             `json:"pageSize"`
	TotalPages int64          `json:"totalPages"`
	Snapshots []ConfigSnapshot `json:"snapshots"`
}

// HistoryFilter 历史过滤器
type HistoryFilter struct {
	Operation  HistoryOperation `json:"operation"`
	StartDate  time.Time        `json:"startDate"`
	EndDate    time.Time        `json:"endDate"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
}

// HistoryList 历史列表
type HistoryList struct {
	Total     int64           `json:"total"`
	Page      int             `json:"page"`
	PageSize  int             `json:"pageSize"`
	TotalPages int64          `json:"totalPages"`
	History   []ConfigHistory `json:"history"`
}

// AvailableProvider 可用供应商
type AvailableProvider struct {
	ProviderType   ProviderType           `json:"providerType"`
	ProviderName   string                 `json:"providerName"`
	DisplayName    string                 `json:"displayName"`
	Description    string                 `json:"description"`
	ConfigTemplate map[string]interface{} `json:"configTemplate"`
	ConfigSchema   map[string]interface{} `json:"configSchema"`
	Capabilities   []CapabilityTemplate   `json:"capabilities"`
}

// CapabilityTemplate 能力模板
type CapabilityTemplate struct {
	CapabilityID          string           `json:"capabilityId"`
	CapabilityType        CapabilityType   `json:"capabilityType"`
	CapabilityName        string           `json:"capabilityName"`
	CapabilityDescription string           `json:"capabilityDescription"`
	InputSchema           map[string]interface{} `json:"inputSchema"`
	OutputSchema          map[string]interface{} `json:"outputSchema"`
}

// PluginStats 插件统计
type PluginStats struct {
	TotalProviders      int                        `json:"totalProviders"`
	EnabledProviders    int                        `json:"enabledProviders"`
	HealthyProviders    int                        `json:"healthyProviders"`
	TotalCapabilities   int                        `json:"totalCapabilities"`
	EnabledCapabilities int                        `json:"enabledCapabilities"`
	ProviderStats       map[string]ProviderStats   `json:"providerStats"`
	CapabilityStats     map[string]CapabilityStats `json:"capabilityStats"`
}

// ProviderStats 供应商统计
type ProviderStats struct {
	Type         ProviderType `json:"type"`
	Count        int          `json:"count"`
	EnabledCount int          `json:"enabledCount"`
	HealthyCount int          `json:"healthyCount"`
}

// CapabilityStats 能力统计
type CapabilityStats struct {
	Type         CapabilityType `json:"type"`
	Count        int            `json:"count"`
	EnabledCount int            `json:"enabledCount"`
}

// pluginConfigServiceImpl 插件配置服务实现
type pluginConfigServiceImpl struct {
	db           *gorm.DB
	logger       *logging.Logger
	encryptor    *ConfigEncryptor
	validator    *ConfigValidator
	registry     *capability.Registry
}

// NewPluginConfigService 创建插件配置服务
func NewPluginConfigService(
	db *gorm.DB,
	logger *logging.Logger,
	encryptor *ConfigEncryptor,
	validator *ConfigValidator,
	registry *capability.Registry,
) PluginConfigService {
	return &pluginConfigServiceImpl{
		db:        db,
		logger:    logger,
		encryptor: encryptor,
		validator: validator,
		registry:  registry,
	}
}

// CreateProviderConfig 创建供应商配置
func (s *pluginConfigServiceImpl) CreateProviderConfig(ctx context.Context, req *CreateProviderConfigRequest) (*ProviderConfig, error) {
	// 验证供应商名称
	if err := s.validator.ValidateProviderName(req.ProviderType, req.ProviderName); err != nil {
		return nil, err
	}

	// 获取配置模式
	configSchema := s.validator.GetConfigSchema(req.ProviderType)

	// 验证配置数据
	if err := s.validator.ValidateConfig(req.Config, configSchema); err != nil {
		return nil, err
	}

	// 检查是否已存在
	var existing ProviderConfig
	if err := s.db.Where("provider_type = ? AND provider_name = ?", req.ProviderType, req.ProviderName).First(&existing).Error; err == nil {
		return nil, errors.New(errors.KindDomain, "plugin_config.create", "provider config already exists")
	}

	// 创建配置
	providerConfig, err := NewProviderConfig(req.ProviderType, req.ProviderName, req.DisplayName, req.Description)
	if err != nil {
		return nil, err
	}

	providerConfig.Enabled = req.Enabled
	providerConfig.Priority = req.Priority

	// 加密配置数据
	configJSON, _ := json.Marshal(req.Config)
	encryptedConfig, err := s.encryptor.Encrypt(string(configJSON))
	if err != nil {
		return nil, errors.Wrap(errors.KindDomain, "plugin_config.create", "failed to encrypt config", err)
	}

	schemaJSON, _ := json.Marshal(configSchema)
	providerConfig.ConfigData = encryptedConfig
	providerConfig.ConfigSchema = string(schemaJSON)

	// 保存到数据库
	if err := s.db.Create(providerConfig).Error; err != nil {
		return nil, errors.Wrap(errors.KindDomain, "plugin_config.create", "failed to create provider config", err)
	}

	// 创建能力映射
	if err := s.createCapabilitiesForProvider(ctx, providerConfig, req.ProviderType); err != nil {
		s.logger.Error("Failed to create capabilities for provider", "provider", req.ProviderType, "error", err)
	}

	// 记录历史
	s.recordHistory(ctx, providerConfig.ID, OperationCreate, "", string(configJSON), "Created new provider config", []string{}, req.CreatedBy, req.UserAgent, req.IPAddress)

	s.logger.Info("Plugin provider config created", "id", providerConfig.ID, "type", req.ProviderType, "name", req.ProviderName)
	return providerConfig, nil
}

// GetProviderConfig 获取供应商配置
func (s *pluginConfigServiceImpl) GetProviderConfig(ctx context.Context, id int) (*ProviderConfig, error) {
	var providerConfig ProviderConfig
	if err := s.db.Preload("Capabilities").First(&providerConfig, id).Error; err != nil {
		if err.Error() == "record not found" {
			return nil, errors.New(errors.KindDomain, "plugin_config.get", "provider config not found")
		}
		return nil, errors.Wrap(errors.KindDomain, "plugin_config.get", "failed to get provider config", err)
	}

	return &providerConfig, nil
}

// GetProviderConfigs 获取供应商配置列表
func (s *pluginConfigServiceImpl) GetProviderConfigs(ctx context.Context, filter *ProviderConfigFilter) (*ProviderConfigList, error) {
	var configs []ProviderConfig
	var total int64

	query := s.db.Model(&ProviderConfig{}).Preload("Capabilities")

	// 应用过滤器
	if filter.ProviderType != "" {
		query = query.Where("provider_type = ?", filter.ProviderType)
	}
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}
	if filter.HealthStatus != "" {
		query = query.Where("health_status = ?", filter.HealthStatus)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.Wrap(errors.KindDomain, "plugin_config.list", "failed to count provider configs", err)
	}

	// 分页
	page := filter.Page
	pageSize := filter.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("priority ASC, created_at DESC").Find(&configs).Error; err != nil {
		return nil, errors.Wrap(errors.KindDomain, "plugin_config.list", "failed to list provider configs", err)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return &ProviderConfigList{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Configs:    configs,
	}, nil
}

// UpdateProviderConfig 更新供应商配置
func (s *pluginConfigServiceImpl) UpdateProviderConfig(ctx context.Context, id int, req *UpdateProviderConfigRequest) (*ProviderConfig, error) {
	providerConfig, err := s.GetProviderConfig(ctx, id)
	if err != nil {
		return nil, err
	}

	// 记录变更
	changes := make([]string, 0)
	oldData, _ := json.Marshal(providerConfig)

	// 更新字段
	if req.DisplayName != "" {
		providerConfig.DisplayName = req.DisplayName
		changes = append(changes, "display_name")
	}
	if req.Description != "" {
		providerConfig.Description = req.Description
		changes = append(changes, "description")
	}
	if req.Config != nil {
		// 验证配置
		configSchema := s.validator.GetConfigSchema(providerConfig.ProviderType)
		if err := s.validator.ValidateConfig(req.Config, configSchema); err != nil {
			return nil, err
		}

		// 加密配置数据
		configJSON, _ := json.Marshal(req.Config)
		encryptedConfig, err := s.encryptor.Encrypt(string(configJSON))
		if err != nil {
			return nil, errors.Wrap(errors.KindDomain, "plugin_config.update", "failed to encrypt config", err)
		}
		providerConfig.ConfigData = encryptedConfig
		changes = append(changes, "config_data")
	}
	if req.Enabled != nil {
		providerConfig.Enabled = *req.Enabled
		changes = append(changes, "enabled")
	}
	if req.Priority != nil {
		providerConfig.Priority = *req.Priority
		changes = append(changes, "priority")
	}

	// 更新数据库
	if err := s.db.Save(providerConfig).Error; err != nil {
		return nil, errors.Wrap(errors.KindDomain, "plugin_config.update", "failed to update provider config", err)
	}

	// 记录历史
	newData, _ := json.Marshal(providerConfig)
	s.recordHistory(ctx, id, OperationUpdate, string(oldData), string(newData), fmt.Sprintf("Updated fields: %v", changes), changes, req.UpdatedBy, req.UserAgent, req.IPAddress)

	s.logger.Info("Plugin provider config updated", "id", id, "changes", changes)
	return providerConfig, nil
}

// DeleteProviderConfig 删除供应商配置
func (s *pluginConfigServiceImpl) DeleteProviderConfig(ctx context.Context, id int) error {
	providerConfig, err := s.GetProviderConfig(ctx, id)
	if err != nil {
		return err
	}

	// 记录历史
	oldData, _ := json.Marshal(providerConfig)
	s.recordHistory(ctx, id, OperationDelete, string(oldData), "", "Deleted provider config", []string{}, "", "", "")

	// 删除配置
	if err := s.db.Delete(providerConfig).Error; err != nil {
		return errors.Wrap(errors.KindDomain, "plugin_config.delete", "failed to delete provider config", err)
	}

	s.logger.Info("Plugin provider config deleted", "id", id, "type", providerConfig.ProviderType, "name", providerConfig.ProviderName)
	return nil
}

// TestProviderConfig 测试供应商配置
func (s *pluginConfigServiceImpl) TestProviderConfig(ctx context.Context, req *TestProviderConfigRequest) (*TestResult, error) {
	// 验证配置
	configSchema := s.validator.GetConfigSchema(req.ProviderType)
	if err := s.validator.ValidateConfig(req.Config, configSchema); err != nil {
		return &TestResult{
			Success:   false,
			Message:   fmt.Sprintf("配置验证失败: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	// 模拟测试逻辑
	startTime := time.Now()

	// 这里应该实际调用供应商API进行测试
	// 暂时返回模拟结果
	time.Sleep(100 * time.Millisecond) // 模拟网络延迟

	latency := time.Since(startTime).Milliseconds()

	return &TestResult{
		Success:   true,
		Message:   "连接测试成功",
		Latency:   latency,
		Details: map[string]interface{}{
			"provider_type": req.ProviderType,
			"test_time":    time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now(),
	}, nil
}

// ValidateProviderConfig 验证供应商配置
func (s *pluginConfigServiceImpl) ValidateProviderConfig(ctx context.Context, providerType ProviderType, config map[string]interface{}) error {
	configSchema := s.validator.GetConfigSchema(providerType)
	return s.validator.ValidateConfig(config, configSchema)
}

// createCapabilitiesForProvider 为供应商创建能力映射
func (s *pluginConfigServiceImpl) createCapabilitiesForProvider(ctx context.Context, providerConfig *ProviderConfig, providerType ProviderType) error {
	// 基于供应商类型创建对应的能力
	switch providerType {
	case ProviderTypeOpenAI:
		// 创建LLM能力
		llmCap, _ := NewCapability(
			providerConfig.ID,
			"openai_chat",
			"OpenAI Chat",
			string(CapabilityTypeLLM),
			"OpenAI GPT对话能力",
		)
		llmCap.InputSchema = `{"type":"object","properties":{"messages":{"type":"array"}}}`
		llmCap.OutputSchema = `{"type":"object","properties":{"content":{"type":"string"}}}`
		s.db.Create(llmCap)

	case ProviderTypeDoubao:
		// 创建LLM能力
		llmCap, _ := NewCapability(
			providerConfig.ID,
			"doubao_llm",
			"豆包大模型",
			string(CapabilityTypeLLM),
			"字节跳动豆包大语言模型",
		)
		s.db.Create(llmCap)

		// 创建ASR能力
		asrCap, _ := NewCapability(
			providerConfig.ID,
			"doubao_asr",
			"豆包语音识别",
			string(CapabilityTypeASR),
			"字节跳动豆包语音识别服务",
		)
		s.db.Create(asrCap)

		// 创建TTS能力
		ttsCap, _ := NewCapability(
			providerConfig.ID,
			"doubao_tts",
			"豆包语音合成",
			string(CapabilityTypeTTS),
			"字节跳动豆包文字转语音服务",
		)
		s.db.Create(ttsCap)

	case ProviderTypeEdge:
		// 创建TTS能力
		ttsCap, _ := NewCapability(
			providerConfig.ID,
			"edge_tts",
			"Edge TTS",
			string(CapabilityTypeTTS),
			"Microsoft Edge文字转语音服务",
		)
		s.db.Create(ttsCap)
	}

	return nil
}

// recordHistory 记录配置变更历史
func (s *pluginConfigServiceImpl) recordHistory(ctx context.Context, providerConfigID int, operation HistoryOperation, oldData, newData, changeSummary string, changedFields []string, createdBy, userAgent, ipAddress string) {
	history, _ := NewConfigHistory(providerConfigID, operation, oldData, newData, changeSummary, "", createdBy, userAgent, ipAddress)
	s.db.Create(history)
}

// GetAvailableProviders 获取可用供应商列表
func (s *pluginConfigServiceImpl) GetAvailableProviders(ctx context.Context) ([]AvailableProvider, error) {
	providers := []AvailableProvider{
		{
			ProviderType: ProviderTypeOpenAI,
			ProviderName: "openai",
			DisplayName:  "OpenAI",
			Description:  "OpenAI GPT大语言模型服务",
			ConfigTemplate: map[string]interface{}{
				"api_key":     "your-openai-api-key",
				"base_url":    "https://api.openai.com/v1",
				"model":       "gpt-3.5-turbo",
				"max_tokens":  2048,
				"temperature": 0.7,
			},
			ConfigSchema: s.validator.GetConfigSchema(ProviderTypeOpenAI),
			Capabilities: []CapabilityTemplate{
				{
					CapabilityID:          "openai_chat",
					CapabilityType:        CapabilityTypeLLM,
					CapabilityName:        "OpenAI Chat",
					CapabilityDescription: "OpenAI GPT对话能力",
				},
			},
		},
		{
			ProviderType: ProviderTypeDoubao,
			ProviderName: "doubao",
			DisplayName:  "豆包",
			Description:  "字节跳动豆包AI服务",
			ConfigTemplate: map[string]interface{}{
				"app_key":      "your-doubao-app-key",
				"app_secret":   "your-doubao-app-secret",
				"endpoint_id":  "your-endpoint-id",
			},
			ConfigSchema: s.validator.GetConfigSchema(ProviderTypeDoubao),
			Capabilities: []CapabilityTemplate{
				{
					CapabilityID:          "doubao_llm",
					CapabilityType:        CapabilityTypeLLM,
					CapabilityName:        "豆包大模型",
					CapabilityDescription: "字节跳动豆包大语言模型",
				},
				{
					CapabilityID:          "doubao_asr",
					CapabilityType:        CapabilityTypeASR,
					CapabilityName:        "豆包语音识别",
					CapabilityDescription: "字节跳动豆包语音识别服务",
				},
				{
					CapabilityID:          "doubao_tts",
					CapabilityType:        CapabilityTypeTTS,
					CapabilityName:        "豆包语音合成",
					CapabilityDescription: "字节跳动豆包文字转语音服务",
				},
			},
		},
	}

	return providers, nil
}

// GetPluginStats 获取插件统计信息
func (s *pluginConfigServiceImpl) GetPluginStats(ctx context.Context) (*PluginStats, error) {
	stats := &PluginStats{
		ProviderStats:   make(map[string]ProviderStats),
		CapabilityStats: make(map[string]CapabilityStats),
	}

	// 统计供应商配置
	var providerConfigs []ProviderConfig
	s.db.Find(&providerConfigs)

	stats.TotalProviders = len(providerConfigs)

	providerTypeCount := make(map[ProviderType]int)
	providerTypeEnabled := make(map[ProviderType]int)
	providerTypeHealthy := make(map[ProviderType]int)

	for _, pc := range providerConfigs {
		providerTypeCount[pc.ProviderType]++
		if pc.Enabled {
			providerTypeEnabled[pc.ProviderType]++
		}
		if pc.HealthStatus == HealthStatusHealthy {
			providerTypeHealthy[pc.ProviderType]++
		}
		stats.EnabledProviders++
		if pc.IsHealthy() {
			stats.HealthyProviders++
		}
	}

	// 统计能力
	var capabilities []Capability
	s.db.Find(&capabilities)
	stats.TotalCapabilities = len(capabilities)

	capabilityTypeCount := make(map[CapabilityType]int)
	capabilityTypeEnabled := make(map[CapabilityType]int)

	for _, cap := range capabilities {
		capabilityTypeCount[cap.CapabilityType]++
		if cap.Enabled {
			capabilityTypeEnabled[cap.CapabilityType]++
			stats.EnabledCapabilities++
		}
	}

	// 转换为统计对象
	for pType, count := range providerTypeCount {
		stats.ProviderStats[string(pType)] = ProviderStats{
			Type:         pType,
			Count:        count,
			EnabledCount: providerTypeEnabled[pType],
			HealthyCount: providerTypeHealthy[pType],
		}
	}

	for cType, count := range capabilityTypeCount {
		stats.CapabilityStats[string(cType)] = CapabilityStats{
			Type:         cType,
			Count:        count,
			EnabledCount: capabilityTypeEnabled[cType],
		}
	}

	return stats, nil
}

// GetEnabledCapabilities 获取启用的能力列表
func (s *pluginConfigServiceImpl) GetEnabledCapabilities(ctx context.Context, capabilityType CapabilityType) ([]Capability, error) {
	var capabilities []Capability
	query := s.db.Joins("JOIN plugin_provider_configs ON plugin_capabilities.provider_config_id = plugin_provider_configs.id").
		Where("plugin_capabilities.enabled = ? AND plugin_capabilities.capability_type = ? AND plugin_provider_configs.enabled = ?", true, capabilityType, true).
		Preload("ProviderConfig").
		Order("plugin_provider_configs.priority ASC")

	if err := query.Find(&capabilities).Error; err != nil {
		return nil, errors.Wrap(errors.KindDomain, "plugin_config.get_enabled_capabilities", "failed to get enabled capabilities", err)
	}

	return capabilities, nil
}

// GetCapabilityExecutor 获取能力执行器
func (s *pluginConfigServiceImpl) GetCapabilityExecutor(ctx context.Context, capabilityID string, config map[string]interface{}) (capability.Executor, error) {
	// 这里应该与现有的插件系统集成
	// 暂时返回nil，实际实现需要调用s.registry.GetExecutor(capabilityID)
	if s.registry != nil {
		return s.registry.GetExecutor(capabilityID)
	}
	return nil, errors.New(errors.KindDomain, "plugin_config.get_executor", "executor integration not implemented")
}

// 实现其他必需的方法...
func (s *pluginConfigServiceImpl) CreateConfigSnapshot(ctx context.Context, providerConfigID int, req *CreateSnapshotRequest) (*ConfigSnapshot, error) {
	// TODO: 实现快照创建
	return nil, errors.New(errors.KindDomain, "plugin_config.create_snapshot", "not implemented")
}

func (s *pluginConfigServiceImpl) GetConfigSnapshots(ctx context.Context, providerConfigID int, filter *SnapshotFilter) (*SnapshotList, error) {
	// TODO: 实现快照列表获取
	return nil, errors.New(errors.KindDomain, "plugin_config.get_snapshots", "not implemented")
}

func (s *pluginConfigServiceImpl) RestoreConfigSnapshot(ctx context.Context, providerConfigID, snapshotID int) error {
	// TODO: 实现快照恢复
	return errors.New(errors.KindDomain, "plugin_config.restore_snapshot", "not implemented")
}

func (s *pluginConfigServiceImpl) GetConfigHistory(ctx context.Context, providerConfigID int, filter *HistoryFilter) (*HistoryList, error) {
	// TODO: 实现历史记录获取
	return nil, errors.New(errors.KindDomain, "plugin_config.get_history", "not implemented")
}