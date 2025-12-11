package v1

import "time"

// SystemStatus 系统状态
type SystemStatus struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	StartTime   time.Time         `json:"start_time"`
	Uptime      int64             `json:"uptime"` // 运行时间（秒）
	Environment string            `json:"environment"`
	GoVersion   string            `json:"go_version"`
	Services    []ServiceStatus   `json:"services"`
	Database    *DatabaseStatus    `json:"database,omitempty"`
	Statistics  SystemStatistics   `json:"statistics"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"`     // running, stopped, error
	Health    string            `json:"health"`    // healthy, unhealthy
	Message   string            `json:"message,omitempty"`
	Uptime    int64             `json:"uptime,omitempty"`
	StartTime time.Time         `json:"start_time"`
	Checks    []HealthCheck      `json:"checks,omitempty"`
}

// HealthCheck 健康检查
type HealthCheck struct {
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	Message  string    `json:"message,omitempty"`
	Duration int64     `json:"duration"` // 毫秒
	Details  interface{} `json:"details,omitempty"`
}

// DatabaseStatus 数据库状态
type DatabaseStatus struct {
	Status        string    `json:"status"`
	Connection    string    `json:"connection"` // connected, disconnected
	Type          string    `json:"type"`       // mysql, postgresql, sqlite
	Host          string    `json:"host"`
	Port          int       `json:"port"`
	Database      string    `json:"database"`
	MaxConnections int       `json:"max_connections"`
	OpenConnections int       `json:"open_connections"`
	SlowQueries    int       `json:"slow_queries"`
}

// SystemStatistics 系统统计
type SystemStatistics struct {
	TotalRequests    int64 `json:"total_requests"`
	SuccessRequests  int64 `json:"success_requests"`
	ErrorRequests    int64 `json:"error_requests"`
	ActiveSessions   int64 `json:"active_sessions"`
	TotalUsers       int64 `json:"total_users"`
	RegisteredUsers  int64 `json:"registered_users"`
	StorageUsed      int64 `json:"storage_used"`
	MemoryUsage      int64 `json:"memory_usage"`
	CPUUsage         float64 `json:"cpu_usage"`
	DiskUsage        float64 `json:"disk_usage"`
}

// HealthCheckRequest 健康检查请求
type HealthCheckRequest struct {
	Checks   []string `json:"checks,omitempty"`
	Timeout  int      `form:"timeout,default=30"` // 超时时间（秒）
}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status    string         `json:"status"`    // healthy, unhealthy, degraded
	Timestamp time.Time       `json:"timestamp"`
	Duration  int64           `json:"duration"`  // 检查耗时（毫秒）
	Checks    []HealthCheckResult `json:"checks"`
	Overall   string         `json:"overall"`
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Duration  int64     `json:"duration"`
	Details   interface{} `json:"details,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// SystemConfigRequest 系统配置更新请求
type SystemConfigRequest struct {
	Setting string      `json:"setting"`
	Value   interface{} `json:"value"`
}

// SystemConfigResponse 系统配置响应
type SystemConfigResponse struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Type      string      `json:"type"`      // string, number, boolean, object
	UpdatedBy  string      `json:"updated_by"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `json:"type" binding:"required,oneof=sqlite mysql postgresql"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database" binding:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
	Options  string `json:"options,omitempty"`
}

// DatabaseTestRequest 数据库测试请求
type DatabaseTestRequest struct {
	Config DatabaseConfig `json:"config" binding:"required"`
}

// DatabaseTestResponse 数据库测试响应
type DatabaseTestResponse struct {
	Success   bool              `json:"success"`
	Connected bool              `json:"connected"`
	Message   string            `json:"message"`
	Latency   int64             `json:"latency"` // 连接延迟（毫秒）
	Version   string            `json:"version,omitempty"`
	Details   interface{}       `json:"details,omitempty"`
}

// DatabaseSchema 数据库模式
type DatabaseSchema struct {
	Name    string        `json:"name"`
	Tables  []TableInfo   `json:"tables"`
	Indexes []IndexInfo  `json:"indexes"`
	ForeignKeys []ForeignKeyInfo `json:"foreign_keys"`
}

// TableInfo 表信息
type TableInfo struct {
	Name            string      `json:"name"`
	Engine          string      `json:"engine"`
	Charset         string      `json:"charset"`
	Collation       string      `json:"collation"`
	Rows            int64       `json:"rows"`
	DataLength       int64       `json:"data_length"`
	IndexLength      int64       `json:"index_length"`
	AutoIncrement    int64       `json:"auto_increment"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Columns         []ColumnInfo `json:"columns"`
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Nullable     bool        `json:"nullable"`
	Default      interface{} `json:"default,omitempty"`
	MaxLength    int         `json:"max_length,omitempty"`
	PrimaryKey   bool        `json:"primary_key"`
	AutoIncrement bool        `json:"auto_increment"`
}

// IndexInfo 索引信息
type IndexInfo struct {
	Name      string      `json:"name"`
	Type       string      `json:"type"` // btree, hash, fulltext
	Unique    bool        `json:"unique"`
	Columns   []string    `json:"columns"`
	TableName string      `json:"table_name"`
}

// ForeignKeyInfo 外键信息
type ForeignKeyInfo struct {
	Name           string      `json:"name"`
	TableName      string      `json:"table_name"`
	ColumnName      string      `json:"column_name"`
	ReferTable     string      `json:"refer_table"`
	ReferColumn    string      `json:"refer_column"`
	OnDelete        string      `json:"on_delete"`
	OnUpdate        string      `json:"on_update"`
}

// InitRequest 系统初始化请求
type InitRequest struct {
	DatabaseConfig  DatabaseConfig `json:"database_config" binding:"required"`
	AdminConfig     AdminConfig    `json:"admin_config" binding:"required"`
	Providers       map[string]interface{} `json:"providers"`
	SystemConfig     interface{}   `json:"system_config,omitempty"`
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

// ConnectionTestResult 连接测试结果
type ConnectionTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ProviderList 提供商列表响应
type ProviderList struct {
	Providers []Provider `json:"providers"`
}

// Provider 提供商信息
type Provider struct {
	Type     string                 `json:"type"`     // llm, tts, asr, vllm
	Name     string                 `json:"name"`
	Status   string                 `json:"status"`   // active, inactive, error
	Config   map[string]interface{} `json:"config"`
	Metadata map[string]interface{} `json:"metadata"`
	Enabled  bool                   `json:"enabled"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ProviderConfigRequest 提供商配置请求
type ProviderConfigRequest struct {
	Type     string                 `json:"type" binding:"required"`
	Name     string                 `json:"name" binding:"required"`
	Config   map[string]interface{} `json:"config" binding:"required"`
	Enabled  bool                   `json:"enabled"`
}

// ===== 统一系统API相关类型 =====

// UnifiedSystemInfo 统一系统信息响应
type UnifiedSystemInfo struct {
	Basic  SystemBasicInfo  `json:"basic"`           // 普通用户可见
	Admin  SystemAdminInfo  `json:"admin,omitempty"` // 管理员专用
	Health SystemHealthInfo `json:"health"`            // 所有用户可见
	Time   SystemTimeInfo   `json:"time"`              // 所有用户可见
}

// SystemBasicInfo 基础系统信息（所有用户可见）
type SystemBasicInfo struct {
	Status    string    `json:"status"`           // 系统状态
	Uptime    int64     `json:"uptime"`           // 运行时间（秒）
	Version   string    `json:"version"`          // 版本号
	Timestamp time.Time `json:"timestamp"`        // 时间戳
}

// SystemAdminInfo 管理员专用信息
type SystemAdminInfo struct {
	Memory     string                 `json:"memory"`      // 内存使用率
	CPU        float64                `json:"cpu"`         // CPU使用率
	Disk       string                 `json:"disk"`        // 磁盘使用
	Services   []string                `json:"services"`    // 服务列表
	Load       SystemLoadInfo         `json:"load"`        // 系统负载
	Logs       SystemLogsInfo          `json:"logs"`        // 日志信息
	Config     SystemConfigInfo        `json:"config"`      // 配置信息
}

// SystemHealthInfo 健康检查信息
type SystemHealthInfo struct {
	Overall    string                `json:"overall"`     // 整体状态
	Components []HealthComponent     `json:"components"` // 组件状态
	Timestamp  time.Time             `json:"timestamp"`  // 检查时间
}

// HealthComponent 单个健康检查组件
type HealthComponent struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Latency   int64  `json:"latency"`
	Error     string `json:"error,omitempty"`
}

// SystemTimeInfo 系统时间信息
type SystemTimeInfo struct {
	CurrentTime time.Time `json:"current_time"` // 当前时间
	Timezone    string    `json:"timezone"`     // 时区
	Uptime      int64     `json:"uptime"`       // 运行时间（秒）
}

// SystemLoadInfo 系统负载信息
type SystemLoadInfo struct {
	CPU        float64 `json:"cpu"`        // CPU负载
	Memory     float64 `json:"memory"`     // 内存负载
	Disk       float64 `json:"disk"`       // 磁盘负载
	Network    float64 `json:"network"`    // 网络负载
}

// SystemLogsInfo 系统日志信息
type SystemLogsInfo struct {
	Level      string `json:"level"`       // 日志级别
	Count      int    `json:"count"`       // 日志数量
	LastLog    string `json:"last_log"`    // 最后日志内容
	TotalLines int64  `json:"total_lines"` // 总日志行数
}

// SystemConfigInfo 系统配置信息
type SystemConfigInfo struct {
	Initialized bool                   `json:"initialized"` // 是否已初始化
	NeedsSetup  bool                   `json:"needs_setup"`  // 是否需要设置
	ConfigPath  string                 `json:"config_path"`  // 配置文件路径
	Database    DatabaseConfigInfo      `json:"database"`    // 数据库配置
}

// DatabaseConfigInfo 数据库配置信息
type DatabaseConfigInfo struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Status   string `json:"status"`
}

// SystemOperationRequest 系统操作请求
type SystemOperationRequest struct {
	Operation string                 `json:"operation" binding:"required"` // 操作类型
	Options   map[string]interface{} `json:"options,omitempty"`     // 操作选项
}

// 系统操作类型
const (
	OperationHealthCheck = "health_check"
	OperationRestart     = "restart"
	OperationRefresh     = "refresh"
	OperationCleanup     = "cleanup"
)