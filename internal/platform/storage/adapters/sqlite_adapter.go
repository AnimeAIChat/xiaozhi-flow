package adapters

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"xiaozhi-server-go/internal/platform/storage"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SQLiteAdapter SQLite数据库适配器
type SQLiteAdapter struct {
	db     *gorm.DB
	sqlDB  *sql.DB
	config storage.DatabaseConnection
}

// NewSQLiteAdapter 创建SQLite适配器
func NewSQLiteAdapter() DatabaseAdapter {
	return &SQLiteAdapter{}
}

// Connect 连接SQLite数据库
func (a *SQLiteAdapter) Connect(config storage.DatabaseConnection) (*gorm.DB, error) {
	a.config = config

	// 确保数据目录存在
	if config.Path != "" {
		dir := filepath.Dir(config.Path)
		if err := createDirectoryIfNotExists(dir); err != nil {
			return nil, fmt.Errorf("创建数据库目录失败: %w", err)
		}
	}

	// 创建GORM连接
	db, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("SQLite连接失败: %w", err)
	}

	// 获取原生SQL连接
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取SQL连接失败: %w", err)
	}

	a.db = db
	a.sqlDB = sqlDB

	// 配置连接池
	if err := a.ConfigureConnectionPool(config.ConnectionPool); err != nil {
		return nil, fmt.Errorf("配置连接池失败: %w", err)
	}

	// SQLite特定优化
	if err := a.OptimizeForDatabase(); err != nil {
		return nil, fmt.Errorf("SQLite优化失败: %w", err)
	}

	return db, nil
}

// Disconnect 断开数据库连接
func (a *SQLiteAdapter) Disconnect() error {
	if a.sqlDB != nil {
		return a.sqlDB.Close()
	}
	return nil
}

// ValidateConnection 验证连接是否有效
func (a *SQLiteAdapter) ValidateConnection() bool {
	if a.sqlDB == nil {
		return false
	}
	return a.sqlDB.Ping() == nil
}

// GetConnection 获取GORM数据库连接
func (a *SQLiteAdapter) GetConnection() *gorm.DB {
	return a.db
}

// CreateSchema 创建完整的数据库模式
func (a *SQLiteAdapter) CreateSchema() error {
	if !a.ValidateConnection() {
		return fmt.Errorf("数据库连接无效")
	}

	// 阶段1: 创建所有表
	if err := a.createTablesPhase(); err != nil {
		return fmt.Errorf("表创建阶段失败: %w", err)
	}

	// 阶段2: 验证所有表创建完成
	if err := a.verifyTablesPhase(); err != nil {
		return fmt.Errorf("表验证阶段失败: %w", err)
	}

	// 阶段3: 创建所有索引
	if err := a.createIndexesPhase(); err != nil {
		return fmt.Errorf("索引创建阶段失败: %w", err)
	}

	return nil
}

// createTablesPhase 阶段1: 创建所有表
func (a *SQLiteAdapter) createTablesPhase() error {
	tableDefinitions := a.getTableDefinitions()

	for tableName, model := range tableDefinitions {
		if err := a.CreateTable(tableName, model); err != nil {
			return fmt.Errorf("创建表%s失败: %w", tableName, err)
		}
	}

	return nil
}

// verifyTablesPhase 阶段2: 验证所有表创建完成
func (a *SQLiteAdapter) verifyTablesPhase() error {
	tableDefinitions := a.getTableDefinitions()
	tableNames := make([]string, 0, len(tableDefinitions))

	for tableName := range tableDefinitions {
		tableNames = append(tableNames, tableName)
	}

	return a.ValidateTables(tableNames)
}

// createIndexesPhase 阶段3: 创建所有索引
func (a *SQLiteAdapter) createIndexesPhase() error {
	indexDefinitions := a.getIndexDefinitions()

	for _, indexes := range indexDefinitions {
		for indexName, indexDef := range indexes {
			if err := a.executeDDLWithValidation(indexName, func() error {
				return a.db.Exec(indexDef).Error
			}); err != nil {
				return fmt.Errorf("创建索引%s失败: %w", indexName, err)
			}
		}
	}

	return nil
}

// CreateTable 创建单个表
func (a *SQLiteAdapter) CreateTable(tableName string, model interface{}) error {
	if !a.ValidateConnection() {
		return fmt.Errorf("数据库连接无效")
	}

	// 使用改进的DDL执行策略
	return a.executeDDLWithValidation(tableName, func() error {
		// 使用GORM AutoMigrate创建表
		return a.db.AutoMigrate(model)
	})
}

// CreateIndex 创建索引
func (a *SQLiteAdapter) CreateIndex(tableName, indexName, indexDef string) error {
	if !a.ValidateConnection() {
		return fmt.Errorf("数据库连接无效")
	}

	// 执行索引创建SQL
	if err := a.db.Exec(indexDef).Error; err != nil {
		return fmt.Errorf("创建索引%s失败: %w", indexName, err)
	}

	// 验证索引是否创建成功
	return a.verifyIndexExists(indexName)
}

// ValidateTables 验证表结构
func (a *SQLiteAdapter) ValidateTables(tableNames []string) error {
	if !a.ValidateConnection() {
		return fmt.Errorf("数据库连接无效")
	}

	for _, tableName := range tableNames {
		if err := a.verifyTableExists(tableName); err != nil {
			return fmt.Errorf("验证表%s失败: %w", tableName, err)
		}
	}

	return nil
}

// ConfigureConnectionPool 配置连接池
func (a *SQLiteAdapter) ConfigureConnectionPool(config storage.ConnectionPool) error {
	if a.sqlDB == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 设置默认值
	maxOpenConns := config.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 10 // SQLite专用：限制并发连接数
	}

	maxIdleConns := config.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 3 // SQLite专用：减少空闲连接数
	}

	connMaxLifetime := time.Duration(config.ConnMaxLifetime) * time.Second
	if connMaxLifetime == 0 {
		connMaxLifetime = 5 * time.Minute // SQLite专用：设置合理的生命周期
	}

	// 配置连接池
	a.sqlDB.SetMaxOpenConns(maxOpenConns)
	a.sqlDB.SetMaxIdleConns(maxIdleConns)
	a.sqlDB.SetConnMaxLifetime(connMaxLifetime)
	a.sqlDB.SetConnMaxIdleTime(1 * time.Minute) // SQLite专用：定期清理空闲连接

	return nil
}

// GetDatabaseType 获取数据库类型
func (a *SQLiteAdapter) GetDatabaseType() string {
	return "sqlite"
}

// GetCapabilities 获取数据库能力
func (a *SQLiteAdapter) GetCapabilities() []string {
	return []string{
		"transactions",
		"foreign_keys",
		"json",
		"full_text",
		"wal_mode",
		"memory_temp_store",
		"mmap_io",
		"vacuum",
		"analyze",
	}
}

// OptimizeForDatabase SQLite特定优化
func (a *SQLiteAdapter) OptimizeForDatabase() error {
	if a.sqlDB == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// SQLite性能优化设置
	settings := []string{
		"PRAGMA journal_mode=WAL",                // WAL模式提高并发性能
		"PRAGMA synchronous=NORMAL",             // 平衡性能和数据安全
		"PRAGMA cache_size=-10000",               // 10MB缓存
		"PRAGMA temp_store=MEMORY",               // 临时表存储在内存中
		"PRAGMA mmap_size=268435456",            // 256MB内存映射
		"PRAGMA busy_timeout=5000",               // 5秒忙超时
		"PRAGMA foreign_keys=OFF",                // 禁用外键约束（初始化阶段）
		"PRAGMA defer_foreign_keys=OFF",          // 禁用延迟外键约束
	}

	for _, setting := range settings {
		if _, err := a.sqlDB.Exec(setting); err != nil {
			return fmt.Errorf("SQLite设置失败 %s: %w", setting, err)
		}
	}

	return nil
}

// executeDDLWithValidation 执行DDL并验证结果（SQLite专用解决方案）
func (a *SQLiteAdapter) executeDDLWithValidation(objectName string, ddlFunc func() error) error {
	if !a.ValidateConnection() {
		return fmt.Errorf("数据库连接无效")
	}

	// 1. 执行DDL操作
	if err := ddlFunc(); err != nil {
		return fmt.Errorf("DDL执行失败: %w", err)
	}

	// 2. 强制提交并刷新连接状态
	if err := a.refreshConnection(); err != nil {
		return fmt.Errorf("连接刷新失败: %w", err)
	}

	// 3. 验证对象是否创建成功
	if err := a.verifyObjectExists(objectName); err != nil {
		return fmt.Errorf("验证对象%s失败: %w", objectName, err)
	}

	// 4. 短暂延迟确保DDL完全生效
	time.Sleep(10 * time.Millisecond)

	return nil
}

// refreshConnection 刷新连接状态
func (a *SQLiteAdapter) refreshConnection() error {
	if a.sqlDB == nil {
		return fmt.Errorf("SQL连接未初始化")
	}

	// 执行Ping操作强制刷新连接状态
	if err := a.sqlDB.Ping(); err != nil {
		return fmt.Errorf("连接Ping失败: %w", err)
	}

	// 执行简单的查询确保连接活跃
	var result int
	if err := a.sqlDB.QueryRow("SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("连接活跃性检查失败: %w", err)
	}

	return nil
}

// verifyObjectExists 验证数据库对象是否存在（表或索引）
func (a *SQLiteAdapter) verifyObjectExists(objectName string) error {
	var count int64

	// 首先检查是否为表
	if err := a.db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?", objectName).Scan(&count).Error; err != nil {
		return fmt.Errorf("查询表%s失败: %w", objectName, err)
	}

	if count > 0 {
		return nil // 表存在
	}

	// 如果不是表，检查是否为索引
	if err := a.db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name = ?", objectName).Scan(&count).Error; err != nil {
		return fmt.Errorf("查询索引%s失败: %w", objectName, err)
	}

	if count > 0 {
		return nil // 索引存在
	}

	// 既不是表也不是索引
	return fmt.Errorf("数据库对象%s不存在", objectName)
}

// GetConnectionStats 获取连接统计信息
func (a *SQLiteAdapter) GetConnectionStats() map[string]interface{} {
	if a.sqlDB == nil {
		return map[string]interface{}{
			"connected": false,
		}
	}

	stats := a.sqlDB.Stats()
	return map[string]interface{}{
		"connected":          a.ValidateConnection(),
		"open_connections":   stats.OpenConnections,
		"in_use":            stats.InUse,
		"idle":              stats.Idle,
		"wait_count":        stats.WaitCount,
		"wait_duration":     stats.WaitDuration.String(),
		"max_idle_closed":    stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// verifyTableExists 验证表是否存在
func (a *SQLiteAdapter) verifyTableExists(tableName string) error {
	var count int64
	if err := a.db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?", tableName).Scan(&count).Error; err != nil {
		return fmt.Errorf("查询表%s失败: %w", tableName, err)
	}

	if count == 0 {
		return fmt.Errorf("表%s不存在", tableName)
	}

	return nil
}

// verifyIndexExists 验证索引是否存在
func (a *SQLiteAdapter) verifyIndexExists(indexName string) error {
	var count int64
	if err := a.db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name = ?", indexName).Scan(&count).Error; err != nil {
		return fmt.Errorf("查询索引%s失败: %w", indexName, err)
	}

	if count == 0 {
		return fmt.Errorf("索引%s不存在", indexName)
	}

	return nil
}

// createDirectoryIfNotExists 创建目录（如果不存在）
func createDirectoryIfNotExists(dir string) error {
	// 检查目录是否已存在
	if _, err := os.Stat(dir); err == nil {
		return nil // 目录已存在
	}

	// 创建目录（包括所有必要的父目录）
	return os.MkdirAll(dir, 0755)
}

// getTableDefinitions 获取表定义
func (a *SQLiteAdapter) getTableDefinitions() map[string]interface{} {
	return map[string]interface{}{
		"auth_clients":      &storage.AuthClient{},
		"domain_events":     &storage.DomainEvent{},
		"config_records":    &storage.ConfigRecord{},
		"config_snapshots":  &storage.ConfigSnapshot{},
		"model_selections":  &storage.ModelSelection{},
		"users":            &storage.User{},
		"devices":          &storage.Device{},
		"agents":           &storage.Agent{},
		"agent_dialogs":    &storage.AgentDialog{},
		"verification_codes": &storage.VerificationCode{},
	}
}

// getIndexDefinitions 获取索引定义
func (a *SQLiteAdapter) getIndexDefinitions() map[string]map[string]string {
	return map[string]map[string]string{
		"auth_clients": {
			"idx_auth_clients_client_id": "CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_clients_client_id ON auth_clients(client_id)",
		},
		"domain_events": {
			"idx_domain_events_event_type":  "CREATE INDEX IF NOT EXISTS idx_domain_events_event_type ON domain_events(event_type)",
			"idx_domain_events_session_id": "CREATE INDEX IF NOT EXISTS idx_domain_events_session_id ON domain_events(session_id)",
			"idx_domain_events_user_id":    "CREATE INDEX IF NOT EXISTS idx_domain_events_user_id ON domain_events(user_id)",
			"idx_domain_events_created_at": "CREATE INDEX IF NOT EXISTS idx_domain_events_created_at ON domain_events(created_at)",
		},
		"config_records": {
			"idx_config_records_key":     "CREATE UNIQUE INDEX IF NOT EXISTS idx_config_records_key ON config_records(key)",
			"idx_config_records_category": "CREATE INDEX IF NOT EXISTS idx_config_records_category ON config_records(category)",
		},
		"config_snapshots": {
			"idx_config_snapshots_name":    "CREATE INDEX IF NOT EXISTS idx_config_snapshots_name ON config_snapshots(name)",
			"idx_config_snapshots_version": "CREATE INDEX IF NOT EXISTS idx_config_snapshots_version ON config_snapshots(version)",
		},
		"model_selections": {
			"idx_model_selections_user_id": "CREATE UNIQUE INDEX IF NOT EXISTS idx_model_selections_user_id ON model_selections(user_id)",
		},
		"users": {
			"idx_users_username": "CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username)",
			"idx_users_email":    "CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		},
		"devices": {
			"idx_devices_agent_id":    "CREATE INDEX IF NOT EXISTS idx_devices_agent_id ON devices(agent_id)",
			"idx_devices_user_id":     "CREATE INDEX IF NOT EXISTS idx_devices_user_id ON devices(user_id)",
			"idx_devices_device_id":   "CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_device_id ON devices(device_id)",
			"idx_devices_client_id":   "CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_client_id ON devices(client_id)",
			"idx_devices_deleted_at":  "CREATE INDEX IF NOT EXISTS idx_devices_deleted_at ON devices(deleted_at)",
		},
		"agents": {
			"idx_agents_user_id":       "CREATE INDEX IF NOT EXISTS idx_agents_user_id ON agents(user_id)",
			"idx_agents_created_at":    "CREATE INDEX IF NOT EXISTS idx_agents_created_at ON agents(created_at)",
		},
		"agent_dialogs": {
			"idx_agent_dialogs_agent_id": "CREATE INDEX IF NOT EXISTS idx_agent_dialogs_agent_id ON agent_dialogs(agent_id)",
			"idx_agent_dialogs_user_id":  "CREATE INDEX IF NOT EXISTS idx_agent_dialogs_user_id ON agent_dialogs(user_id)",
		},
		"verification_codes": {
			"uni_verification_codes_code":       "CREATE UNIQUE INDEX IF NOT EXISTS uni_verification_codes_code ON verification_codes(code)",
			"idx_verification_codes_deleted_at": "CREATE INDEX IF NOT EXISTS idx_verification_codes_deleted_at ON verification_codes(deleted_at)",
		},
	}
}