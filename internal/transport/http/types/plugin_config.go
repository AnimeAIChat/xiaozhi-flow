package types

import (
	"time"

	"xiaozhi-server-go/internal/domain/plugin/config"
)

// PluginConfigRequest 创建/更新插件配置请求
type PluginConfigRequest struct {
	ProviderType config.ProviderType        `json:"providerType" binding:"required"`
	ProviderName string                     `json:"providerName" binding:"required"`
	DisplayName  string                     `json:"displayName" binding:"required"`
	Description  string                     `json:"description"`
	Config       map[string]interface{}     `json:"config" binding:"required"`
	Enabled      bool                       `json:"enabled"`
	Priority     int                        `json:"priority"`
}

// PluginConfigResponse 插件配置响应
type PluginConfigResponse struct {
	ID              int                        `json:"id"`
	ProviderType    config.ProviderType        `json:"providerType"`
	ProviderName    string                     `json:"providerName"`
	DisplayName     string                     `json:"displayName"`
	Description     string                     `json:"description"`
	Config          map[string]interface{}     `json:"config"`
	ConfigSchema    map[string]interface{}     `json:"configSchema"`
	Enabled         bool                       `json:"enabled"`
	Priority        int                        `json:"priority"`
	HealthStatus    config.HealthStatus        `json:"healthStatus"`
	LastHealthCheck *time.Time                 `json:"lastHealthCheck"`
	CreatedAt       time.Time                  `json:"createdAt"`
	UpdatedAt       time.Time                  `json:"updatedAt"`
	Capabilities    []CapabilityResponse       `json:"capabilities"`
}

// PluginConfigListResponse 插件配置列表响应
type PluginConfigListResponse struct {
	Total       int64                    `json:"total"`
	Page        int                      `json:"page"`
	PageSize    int                      `json:"pageSize"`
	TotalPages  int64                    `json:"totalPages"`
	Configs     []PluginConfigResponse   `json:"configs"`
}

// CapabilityResponse 能力响应
type CapabilityResponse struct {
	ID                    int                      `json:"id"`
	ProviderConfigID      int                      `json:"providerConfigId"`
	CapabilityID          string                   `json:"capabilityId"`
	CapabilityType        config.CapabilityType    `json:"capabilityType"`
	CapabilityName        string                   `json:"capabilityName"`
	CapabilityDescription string                   `json:"capabilityDescription"`
	InputSchema           map[string]interface{}   `json:"inputSchema"`
	OutputSchema          map[string]interface{}   `json:"outputSchema"`
	Enabled               bool                     `json:"enabled"`
	CreatedAt             time.Time                `json:"createdAt"`
	UpdatedAt             time.Time                `json:"updatedAt"`
}

// AvailableProviderResponse 可用供应商响应
type AvailableProviderResponse struct {
	ProviderType   config.ProviderType    `json:"providerType"`
	ProviderName   string                 `json:"providerName"`
	DisplayName    string                 `json:"displayName"`
	Description    string                 `json:"description"`
	ConfigTemplate map[string]interface{} `json:"configTemplate"`
	ConfigSchema   map[string]interface{} `json:"configSchema"`
	Capabilities   []CapabilityTemplate   `json:"capabilities"`
}

// CapabilityTemplate 能力模板
type CapabilityTemplate struct {
	CapabilityID          string                 `json:"capabilityId"`
	CapabilityType        config.CapabilityType  `json:"capabilityType"`
	CapabilityName        string                 `json:"capabilityName"`
	CapabilityDescription string                 `json:"capabilityDescription"`
	InputSchema           map[string]interface{} `json:"inputSchema"`
	OutputSchema          map[string]interface{} `json:"outputSchema"`
}

// ConfigSnapshotResponse 配置快照响应
type ConfigSnapshotResponse struct {
	ID            int       `json:"id"`
	ProviderConfigID int     `json:"providerConfigId"`
	Version       string    `json:"version"`
	SnapshotName  string    `json:"snapshotName"`
	Description   string    `json:"description"`
	SnapshotData  map[string]interface{} `json:"snapshotData"`
	IsActive      bool      `json:"isActive"`
	CreatedBy     string    `json:"createdBy"`
	CreatedAt     time.Time `json:"createdAt"`
}

// ConfigSnapshotListResponse 配置快照列表响应
type ConfigSnapshotListResponse struct {
	Total     int64                    `json:"total"`
	Page      int                      `json:"page"`
	PageSize  int                      `json:"pageSize"`
	TotalPages int64                    `json:"totalPages"`
	Snapshots []ConfigSnapshotResponse `json:"snapshots"`
}

// ConfigHistoryResponse 配置变更历史响应
type ConfigHistoryResponse struct {
	ID              int                     `json:"id"`
	ProviderConfigID int                    `json:"providerConfigId"`
	Operation       config.HistoryOperation `json:"operation"`
	OldData         map[string]interface{}  `json:"oldData,omitempty"`
	NewData         map[string]interface{}  `json:"newData,omitempty"`
	ChangeSummary   string                  `json:"changeSummary"`
	ChangedFields   []string                `json:"changedFields"`
	CreatedBy       string                  `json:"createdBy"`
	UserAgent       string                  `json:"userAgent"`
	IPAddress       string                  `json:"ipAddress"`
	CreatedAt       time.Time               `json:"createdAt"`
}

// ConfigHistoryListResponse 配置变更历史列表响应
type ConfigHistoryListResponse struct {
	Total   int64                    `json:"total"`
	Page    int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
	TotalPages int64                    `json:"totalPages"`
	History []ConfigHistoryResponse  `json:"history"`
}

// HealthTestRequest 健康测试请求
type HealthTestRequest struct {
	Config map[string]interface{} `json:"config" binding:"required"`
}

// HealthTestResponse 健康测试响应
type HealthTestResponse struct {
	Success    bool                   `json:"success"`
	Message    string                 `json:"message"`
	Latency    int64                  `json:"latency"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// ConfigRestoreRequest 配置恢复请求
type ConfigRestoreRequest struct {
	SnapshotID int `json:"snapshotId" binding:"required"`
}

// ConfigExportRequest 配置导出请求
type ConfigExportRequest struct {
	IncludeSecrets bool `json:"includeSecrets"`
	ProviderIDs    []int `json:"providerIds"`
}

// ConfigImportRequest 配置导入请求
type ConfigImportRequest struct {
	Configs     []map[string]interface{} `json:"configs" binding:"required"`
	Overwrite   bool                     `json:"overwrite"`
	IncludeSecrets bool                   `json:"includeSecrets"`
}

// ConfigValidationResponse 配置验证响应
type ConfigValidationResponse struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ConfigSnapshotRequest 配置快照请求
type ConfigSnapshotRequest struct {
	Version     string `json:"version"`
	SnapshotName string `json:"snapshotName"`
	Description string `json:"description"`
}

// PluginStatsResponse 插件统计响应
type PluginStatsResponse struct {
	TotalProviders    int                          `json:"totalProviders"`
	EnabledProviders  int                          `json:"enabledProviders"`
	HealthyProviders  int                          `json:"healthyProviders"`
	TotalCapabilities int                          `json:"totalCapabilities"`
	EnabledCapabilities int                       `json:"enabledCapabilities"`
	ProviderStats     map[string]ProviderStats     `json:"providerStats"`
	CapabilityStats   map[string]CapabilityStats  `json:"capabilityStats"`
}

// ProviderStats 供应商统计
type ProviderStats struct {
	Type         config.ProviderType `json:"type"`
	Count        int                 `json:"count"`
	EnabledCount int                 `json:"enabledCount"`
	HealthyCount int                 `json:"healthyCount"`
}

// CapabilityStats 能力统计
type CapabilityStats struct {
	Type         config.CapabilityType `json:"type"`
	Count        int                   `json:"count"`
	EnabledCount int                   `json:"enabledCount"`
}