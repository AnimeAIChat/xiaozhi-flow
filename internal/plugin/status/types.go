package status

import (
	"time"

	"xiaozhi-server-go/internal/plugin/capability"
)

// PluginStatusType 插件状态类型
type PluginStatusType string

const (
	StatusUnknown    PluginStatusType = "unknown"
	StatusInstalled  PluginStatusType = "installed"
	StatusEnabled    PluginStatusType = "enabled"
	StatusDisabled   PluginStatusType = "disabled"
	StatusRunning    PluginStatusType = "running"
	StatusStopped    PluginStatusType = "stopped"
	StatusError      PluginStatusType = "error"
)

// HealthStatus 健康状态类型
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// PluginStatus 插件状态信息
type PluginStatus struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	Description     string            `json:"description"`
	Version         string            `json:"version"`
	Status          PluginStatusType  `json:"status"`
	Address         string            `json:"address"`
	Port            int               `json:"port"`
	Capabilities    []CapabilityDef   `json:"capabilities"`
	Config          map[string]interface{} `json:"config,omitempty"`
	HealthStatus    HealthStatus      `json:"health_status"`
	LastHealthCheck time.Time         `json:"last_health_check"`
	Error           string            `json:"error,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// CapabilityDef 插件能力定义
type CapabilityDef struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ConfigSchema CapabilitySchema      `json:"config_schema"`
	InputSchema  CapabilitySchema      `json:"input_schema"`
	OutputSchema CapabilitySchema      `json:"output_schema"`
	Enabled     bool                   `json:"enabled"`
}

// CapabilitySchema 能力配置模式
type CapabilitySchema struct {
	Type       string                    `json:"type"`
	Properties map[string]SchemaProperty `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// SchemaProperty 模式属性
type SchemaProperty struct {
	Type        string      `json:"type"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
	Secret      bool        `json:"secret"`
}

// PluginFilter 插件筛选条件
type PluginFilter struct {
	Type         string            `form:"type" json:"type"`
	Status       PluginStatusType  `form:"status" json:"status"`
	HealthStatus HealthStatus      `form:"health_status" json:"health_status"`
	Page         int               `form:"page" json:"page"`
	PageSize     int               `form:"page_size" json:"page_size"`
	SortBy       string            `form:"sort_by" json:"sort_by"`
	SortOrder    string            `form:"sort_order" json:"sort_order"`
	Search       string            `form:"search" json:"search"`
}

// PluginListResponse 插件列表响应
type PluginListResponse struct {
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
	Plugins    []PluginStatus `json:"plugins"`
}

// PluginControlRequest 插件控制请求
type PluginControlRequest struct {
	Action string                 `json:"action"` // "start", "stop", "restart", "reallocate_port"
	Config map[string]interface{} `json:"config,omitempty"`
}

// PluginControlResponse 插件控制响应
type PluginControlResponse struct {
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	OldStatus    string    `json:"old_status,omitempty"`
	NewStatus    string    `json:"new_status"`
	OldPort      int       `json:"old_port,omitempty"`
	NewPort      int       `json:"new_port,omitempty"`
	ProcessTime  string    `json:"process_time"`
}

// PluginStats 插件统计信息
type PluginStats struct {
	TotalPlugins      int                    `json:"total_plugins"`
	RunningPlugins    int                    `json:"running_plugins"`
	StoppedPlugins    int                    `json:"stopped_plugins"`
	ErrorPlugins      int                    `json:"error_plugins"`
	HealthyPlugins    int                    `json:"healthy_plugins"`
	UnhealthyPlugins  int                    `json:"unhealthy_plugins"`
	ByType            map[string]int         `json:"by_type"`
	ByStatus          map[PluginStatusType]int `json:"by_status"`
	AveragePortUsage  float64                `json:"average_port_usage"`
}

// ConvertFromCapability 从能力定义转换
func ConvertFromCapability(cap capability.Definition) CapabilityDef {
	var properties map[string]SchemaProperty
	if len(cap.ConfigSchema.Properties) > 0 {
		properties = make(map[string]SchemaProperty)
		for k, v := range cap.ConfigSchema.Properties {
			properties[k] = SchemaProperty{
				Type:        v.Type,
				Default:     v.Default,
				Description: v.Description,
				Secret:      v.Secret,
			}
		}
	}

	return CapabilityDef{
		ID:          cap.ID,
		Type:        string(cap.Type),
		Name:        cap.Name,
		Description: cap.Description,
		ConfigSchema: CapabilitySchema{
			Type:       cap.ConfigSchema.Type,
			Properties: properties,
			Required:   cap.ConfigSchema.Required,
		},
		InputSchema: CapabilitySchema{
			Type:       cap.InputSchema.Type,
			Properties: convertProperties(cap.InputSchema.Properties),
			Required:   cap.InputSchema.Required,
		},
		OutputSchema: CapabilitySchema{
			Type:       cap.OutputSchema.Type,
			Properties: convertProperties(cap.OutputSchema.Properties),
			Required:   cap.OutputSchema.Required,
		},
		Enabled: true,
	}
}

// convertProperties 转换属性定义
func convertProperties(props map[string]capability.Property) map[string]SchemaProperty {
	result := make(map[string]SchemaProperty)
	for k, v := range props {
		result[k] = SchemaProperty{
			Type:        v.Type,
			Default:     v.Default,
			Description: v.Description,
			Secret:      v.Secret,
		}
	}
	return result
}

// DefaultPluginFilter 默认筛选条件
func DefaultPluginFilter() PluginFilter {
	return PluginFilter{
		Page:      1,
		PageSize:  20,
		SortBy:    "updated_at",
		SortOrder: "desc",
	}
}

// Validate 验证筛选条件
func (f *PluginFilter) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	if f.SortBy == "" {
		f.SortBy = "updated_at"
	}
	if f.SortOrder == "" {
		f.SortOrder = "desc"
	}
	return nil
}