package adapters

import (
	"xiaozhi-server-go/internal/platform/storage"
	"gorm.io/gorm"
)

// DatabaseAdapter 定义数据库适配器的统一接口
type DatabaseAdapter interface {
	// 连接管理
	Connect(config storage.DatabaseConnection) (*gorm.DB, error)
	Disconnect() error
	ValidateConnection() bool
	GetConnection() *gorm.DB

	// 模式管理
	CreateSchema() error
	CreateTable(tableName string, model interface{}) error
	CreateIndex(tableName, indexName, indexDef string) error
	ValidateTables(tableNames []string) error

	// 配置管理
	ConfigureConnectionPool(config storage.ConnectionPool) error
	GetDatabaseType() string
	GetCapabilities() []string

	// 性能优化
	OptimizeForDatabase() error
	GetConnectionStats() map[string]interface{}
}

// AdapterCapabilities 定义适配器能力
type AdapterCapabilities struct {
	SupportsTransactions   bool
	SupportsForeignKeys    bool
	SupportsJSON           bool
	SupportsFullText       bool
	MaxConnections        int
	PreferredPoolSize      int
}

// DatabaseConfig 数据库配置（新版本）
type DatabaseConfig struct {
	Type           string                 `json:"type"`           // sqlite, mysql, postgresql
	ConnectionString string                 `json:"connection_string"`
	Options        map[string]interface{} `json:"options"`
	Pool          ConnectionPoolConfig    `json:"pool"`
}

type ConnectionPoolConfig struct {
	MaxOpenConns    int `json:"max_open_conns"`
	MaxIdleConns    int `json:"max_idle_conns"`
	ConnMaxLifetime int `json:"conn_max_lifetime"` // 秒
}

// DDLStatement DDL语句定义
type DDLStatement struct {
	Type        string            `json:"type"`         // table, index, constraint
	Name        string            `json:"name"`         // 对象名称
	Definition  string            `json:"definition"`   // SQL定义
	DependsOn   []string          `json:"depends_on"`   // 依赖关系
	Options     map[string]string `json:"options"`      // 选项参数
}

// ValidationResult 验证结果
type ValidationResult struct {
	Success     bool     `json:"success"`
	ObjectName  string   `json:"object_name"`
	ObjectType  string   `json:"object_type"`
	Issues      []string `json:"issues"`
	MissingDeps []string `json:"missing_deps"`
}