package storage

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"xiaozhi-server-go/internal/platform/storage/migrations"
	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitConfigStore ensures the underlying configuration store is ready.
// Since we no longer use database-backed configuration, this is a no-op.
func InitConfigStore() error {
	return nil
}

// ConfigStore returns the default configuration store implementation.
// Since we no longer use database-backed configuration, this returns nil.
func ConfigStore() interface{} {
	return nil
}

// Global database instance for backward compatibility
var db *gorm.DB

// InitDatabaseWithConfig initializes database using the provided configuration
func InitDatabaseWithConfig(config DatabaseConnection) error {
	if err := initDatabaseWithConnection(config); err != nil {
		return err
	}

	fmt.Printf("数据库已使用配置文件成功连接\n")
	return nil
}

// ConnectDatabaseWithConfig connects to an existing database using the provided configuration
// This function only connects to an existing database without reinitializing tables
func ConnectDatabaseWithConfig(config DatabaseConnection) error {
	dbPath := config.Path

	// For SQLite, ensure the database file exists
	if config.Type == "sqlite" {
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			return fmt.Errorf("database file does not exist: %s", dbPath)
		}
	}

	var err error
	var gormDB *gorm.DB

	switch config.Type {
	case "sqlite":
		gormDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset)
		gormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	case "postgresql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			config.Host, config.Username, config.Password, config.Database, config.Port, config.SSLMode)
		gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool parameters for long-running connections
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)  // 保持所有连接都处于空闲状态以复用
	sqlDB.SetConnMaxLifetime(0)  // 连接永不自动过期，SQLite专用
	sqlDB.SetConnMaxIdleTime(0)  // 空闲连接永不自动关闭，SQLite专用

	// Verify the database connection is fully operational by running a test query
	var testResult int64
	if err := gormDB.Raw("SELECT 1").Count(&testResult).Error; err != nil {
		return fmt.Errorf("database connection test query failed: %w", err)
	}

	// Set global database instance only after successful validation
	SetDB(gormDB)

	// For existing databases, DO NOT run AutoMigrate
	// AutoMigrate can reset data in existing tables
	// The database should already have the correct schema from previous initialization
	fmt.Printf("数据库已成功连接\n")
	return nil
}

// InitDatabase checks database initialization status without creating it automatically.
func InitDatabase() error {
	// Use environment variable for database path if set (for testing)
	dbPath := os.Getenv("XIAOZHI_DB_PATH")
	if dbPath == "" {
		dataDir := "./data"
		dbPath = filepath.Join(dataDir, "xiaozhi.db")
	}

	// Check if database file exists (only for SQLite)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("数据库文件不存在: %s，请通过配置页面进行初始化\n", dbPath)
		return nil // Don't treat missing database as an error
	} else if err != nil {
		return fmt.Errorf("failed to check database file: %w", err)
	}

	// Database file exists, try to open and initialize it
	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to open existing database: %w", err)
	}

	// Test the database connection before proceeding
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool parameters for SQLite (never expire connections)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(0)  // SQLite专用：连接永不自动过期
	sqlDB.SetConnMaxIdleTime(0)  // SQLite专用：空闲连接永不自动关闭

	// Verify the database connection is fully operational by running a test query
	var testResult int64
	if err := db.Raw("SELECT 1").Count(&testResult).Error; err != nil {
		return fmt.Errorf("database connection test query failed: %w", err)
	}

	// Auto-migrate tables for existing database
	if err := db.AutoMigrate(&AuthClient{}, &DomainEvent{}, &ConfigRecord{}, &ConfigSnapshot{}, &ModelSelection{}, &User{}, &Device{}, &Agent{}, &AgentDialog{}, &VerificationCode{}); err != nil {
		return fmt.Errorf("failed to migrate existing database: %w", err)
	}

	// Run migrations
	migrationManager := NewMigrationManager(db)
	migrationManager.AddMigration(&migrations.Migration001Initial{})
	migrationManager.AddMigration(&migrations.Migration002ConfigTables{})
	migrationManager.AddMigration(&migrations.Migration003ModelSelections{})

	if err := migrationManager.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations on existing database: %w", err)
	}

	// Check admin user status
	if err := initializeAdminUser(db); err != nil {
		return fmt.Errorf("failed to check admin user status: %w", err)
	}

	fmt.Printf("数据库已存在并成功连接: %s\n", dbPath)
	return nil
}

// ValidateDBConnection validates that the database connection is fully operational
func ValidateDBConnection(database *gorm.DB) bool {
	if database == nil {
		return false
	}

	// Get underlying SQL connection
	sqlDB, err := database.DB()
	if err != nil {
		fmt.Printf("[DEBUG] ValidateDBConnection: Failed to get underlying DB: %v\n", err)
		return false
	}

	// Test the connection with a ping
	if err := sqlDB.Ping(); err != nil {
		fmt.Printf("[DEBUG] ValidateDBConnection: Ping failed: %v\n", err)
		return false
	}

	// Verify with a simple test query
	var testResult int64
	if err := database.Raw("SELECT 1").Count(&testResult).Error; err != nil {
		fmt.Printf("[DEBUG] ValidateDBConnection: Test query failed: %v\n", err)
		return false
	}

	return true
}

// GetDB returns the global database instance.
func GetDB() *gorm.DB {
	if db == nil {
		// Database not initialized yet, return nil instead of panic
		return nil
	}

	// 对于SQLite，连接配置为永不超时，所以不需要频繁Ping
	// 只在获取底层连接失败时才返回nil
	_, err := db.DB()
	if err != nil {
		fmt.Printf("[DEBUG] GetDB: Failed to get underlying DB: %v\n", err)
		return nil
	}

	// 直接返回，信任SQLite连接池配置
	return db
}

// SetDB sets the global database instance.
func SetDB(database *gorm.DB) {
	db = database
}

// AuthClient represents the authentication client model for GORM
type AuthClient struct {
	ID        uint           `gorm:"primaryKey"`
	ClientID  string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"client_id"`
	Username  string         `gorm:"not null"                               json:"username"`
	Password  string         `gorm:"not null"                               json:"password"`
	IP        string         `                                              json:"ip"`
	DeviceID  string         `                                              json:"device_id"`
	CreatedAt time.Time      `                                              json:"created_at"`
	ExpiresAt *time.Time     `                                              json:"expires_at,omitempty"`
	Metadata  datatypes.JSON `                                              json:"metadata,omitempty"`
}

// DomainEvent 领域事件存储模型
type DomainEvent struct {
	ID        uint           `gorm:"primaryKey"`
	EventType string         `gorm:"index;not null"` // 事件类型
	SessionID string         `gorm:"index"`          // 会话ID
	UserID    string         `gorm:"index"`          // 用户ID
	Data      datatypes.JSON `gorm:"not null"`       // 事件数据
	CreatedAt time.Time      `gorm:"index"`          // 创建时间
}

// Agent 智能体模型
type Agent struct {
	ID                 uint           `gorm:"primaryKey"`
	Name               string         `gorm:"not null"`
	LLM                string         `gorm:"default:'ChatGLMLLM'"`
	Language           string         `gorm:"default:'普通话'"`
	Voice              string         `gorm:"default:'zh_female_wanwanxiaohe_moon_bigtts'"`
	VoiceName          string         `gorm:"default:'湾湾小何'"`
	Prompt             string         `gorm:"type:text"`
	ASRSpeed           int            `gorm:"default:2"`
	SpeakSpeed         int            `gorm:"default:2"`
	Tone               int            `gorm:"default:50"`
	UserID             uint           `gorm:"not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	LastConversationAt time.Time
	EnabledTools       string         `gorm:"type:text"`
	Conversationid     string
	HeadImg            string         `gorm:"type:varchar(255)"`
	Description        string         `gorm:"type:text"`
	CatalogyID         uint
	Extra              string         `gorm:"type:text"`
}

// AgentDialog 智能体对话模型
type AgentDialog struct {
	ID             uint      `gorm:"primaryKey"`
	Conversationid string
	AgentID        uint      `gorm:"index"`
	UserID         uint      `gorm:"index"`
	Dialog         string    `gorm:"type:text"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Device 设备模型
type Device struct {
	ID               uint           `gorm:"primaryKey"`
	AgentID          *uint          `gorm:"index"`
	UserID           *uint          `gorm:"index"`
	Name             string         `gorm:"not null"`
	DeviceID         string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	ClientID         string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Version          string
	OTA              bool           `gorm:"default:true"`
	RegisterTime     int64
	LastActiveTime   int64
	RegisterTimeV2   time.Time
	LastActiveTimeV2 time.Time
	Online           bool
	AuthCode         string
	AuthStatus       string
	BoardType        string
	ChipModelName    string
	Channel          int
	SSID             string
	Application      string
	Language         string         `gorm:"default:'zh-CN'"`
	DeviceCode       string
	DeletedAt        gorm.DeletedAt `gorm:"index"`
	Extra            string         `gorm:"type:text"`
	Conversationid   string
	Mode             string
	LastIP           string
	Stats            string         `gorm:"type:text"`
	TotalTokens      int64          `gorm:"default:0"`
	UsedTokens       int64          `gorm:"default:0"`
	LastSessionEndAt *time.Time
}

// User 用户模型
type User struct {
	ID          uint      `gorm:"primaryKey"`
	Username    string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password    string    `json:"-"`
	Nickname    string    `gorm:"type:varchar(255)"`
	HeadImg     string    `gorm:"type:varchar(255)"`
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Email       string    `gorm:"type:varchar(255);uniqueIndex;"`
	Status      uint      `gorm:"default:1"`
	PhoneNumber string    `gorm:"type:varchar(20);"`
	Extra       string    `gorm:"type:text"`
}

// ServerConfig 服务器配置模型
type ServerConfig struct {
	ID     uint   `gorm:"primaryKey"`
	CfgStr string `gorm:"type:text"`
}

// VerificationCode 验证码模型
type VerificationCode struct {
	ID        uint           `gorm:"primarykey"`
	Code      string         `gorm:"unique;not null;size:6"`
	Purpose   string         `gorm:"not null;size:50"`
	UserID    *string        `gorm:"size:100"`
	DeviceID  *string        `gorm:"size:100"`
	ExpiresAt time.Time      `gorm:"not null"`
	UsedAt    *time.Time
	IsUsed    bool           `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// initializeAdminUser 检查管理员用户状态（不再自动创建）
func initializeAdminUser(db *gorm.DB) error {
	// 检查是否已存在管理员用户
	var count int64
	if err := db.Model(&User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check admin user count: %w", err)
	}

	if count > 0 {
		// 管理员用户已存在，系统已初始化
		fmt.Printf("系统已初始化，找到 %d 个管理员用户\n", count)
	} else {
		// 管理员用户不存在，需要通过配置页面进行初始化
		fmt.Printf("系统尚未初始化，请访问配置页面进行初始化\n")
	}

	return nil
}

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[n.Int64()]
	}
	return string(password)
}

// initDatabaseWithConnection 使用指定连接配置初始化数据库
func initDatabaseWithConnection(config DatabaseConnection) error {
	var err error

	// 根据数据库类型创建连接
	switch strings.ToLower(config.Type) {
	case "sqlite":
		// 确保目录存在
		if config.Path != "" {
			dir := filepath.Dir(config.Path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory: %w", err)
			}
		}

		db, err = gorm.Open(sqlite.Open(config.Path), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			return fmt.Errorf("failed to connect to sqlite database: %w", err)
		}

	case "mysql":
		// 需要导入 MySQL 驱动
		// 注意：需要在 import 中添加 _ "gorm.io/driver/mysql"
		return fmt.Errorf("MySQL support not yet implemented")

	case "postgresql", "postgres":
		// 需要导入 PostgreSQL 驱动
		// 注意：需要在 import 中添加 _ "gorm.io/driver/postgres"
		return fmt.Errorf("PostgreSQL support not yet implemented")

	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Use enhanced connection pool settings for better stability
	maxOpenConns := config.ConnectionPool.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 25
	}
	maxIdleConns := config.ConnectionPool.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 25
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(0)  // SQLite专用：连接永不自动过期
	sqlDB.SetConnMaxIdleTime(0)  // SQLite专用：空闲连接永不自动关闭

	// Test the database connection before proceeding with migrations
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Verify the database connection is fully operational by running a test query
	var testResult int64
	if err := db.Raw("SELECT 1").Count(&testResult).Error; err != nil {
		return fmt.Errorf("database connection test query failed: %w", err)
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(&AuthClient{}, &DomainEvent{}, &ConfigRecord{}, &ConfigSnapshot{}, &ModelSelection{}, &User{}, &Device{}, &Agent{}, &AgentDialog{}, &VerificationCode{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Run migrations
	migrationManager := NewMigrationManager(db)
	migrationManager.AddMigration(&migrations.Migration001Initial{})
	migrationManager.AddMigration(&migrations.Migration002ConfigTables{})
	migrationManager.AddMigration(&migrations.Migration003ModelSelections{})

	if err := migrationManager.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
