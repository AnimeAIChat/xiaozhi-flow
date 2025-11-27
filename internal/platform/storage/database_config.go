package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DatabaseConfig 数据库配置文件结构
type DatabaseConfig struct {
	Database     DatabaseConnection `json:"database"`
	Admin        AdminConfig        `json:"admin"`
	Initialized  bool               `json:"initialized"`
	Version      string             `json:"version"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// DatabaseConnection 数据库连接配置
type DatabaseConnection struct {
	Type           string           `json:"type"`             // sqlite, mysql, postgresql
	Path           string           `json:"path,omitempty"`    // SQLite 文件路径
	Host           string           `json:"host,omitempty"`    // 数据库主机
	Port           int              `json:"port,omitempty"`    // 数据库端口
	Database       string           `json:"database,omitempty"` // 数据库名称
	Username       string           `json:"username,omitempty"` // 用户名
	Password       string           `json:"password,omitempty"` // 密码
	SSLMode        string           `json:"ssl_mode,omitempty"` // SSL 模式 (PostgreSQL)
	Charset        string           `json:"charset,omitempty"` // 字符集 (MySQL)
	ConnectionPool ConnectionPool   `json:"connection_pool"`   // 连接池配置
}

// ConnectionPool 连接池配置
type ConnectionPool struct {
	MaxOpenConns    int           `json:"max_open_conns"`    // 最大打开连接数
	MaxIdleConns    int           `json:"max_idle_conns"`    // 最大空闲连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"` // 连接最大生存时间
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

// DatabaseTestStep 数据库测试步骤
type DatabaseTestStep string

const (
	StepNetworkCheck     DatabaseTestStep = "network_check"
	StepDatabaseConnect  DatabaseTestStep = "database_connect"
	StepPermissionCheck  DatabaseTestStep = "permission_check"
	StepTableCreation    DatabaseTestStep = "table_creation"
)

// DatabaseTestResult 数据库测试结果
type DatabaseTestResult struct {
	Step    DatabaseTestStep `json:"step"`
	Status  string           `json:"status"` // success, failed, running, pending
	Message string           `json:"message"`
	Latency int64            `json:"latency,omitempty"` // 毫秒
	Details interface{}       `json:"details,omitempty"`
}

// DatabaseTestProgress 数据库测试进度
type DatabaseTestProgress struct {
	CurrentStep DatabaseTestStep    `json:"current_step"`
	Steps       []DatabaseTestStep `json:"steps"`
	Results     map[string]*DatabaseTestResult `json:"results"`
	IsComplete  bool               `json:"is_complete"`
	OverallStatus string           `json:"overall_status"` // success, failed, in_progress
}

// DatabaseConfigManager 数据库配置管理器
type DatabaseConfigManager struct {
	configPath string
	mutex      sync.RWMutex
}

// NewDatabaseConfigManager 创建数据库配置管理器
func NewDatabaseConfigManager() *DatabaseConfigManager {
	// 确保数据目录存在
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create data directory: %v\n", err)
	}

	return &DatabaseConfigManager{
		configPath: filepath.Join(dataDir, "db.json"),
	}
}

// LoadConfig 加载数据库配置
func (m *DatabaseConfigManager) LoadConfig() (*DatabaseConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 检查文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database config file does not exist: %s", m.configPath)
	}

	// 读取文件
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read database config file: %w", err)
	}

	// 解析JSON
	var config DatabaseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse database config file: %w", err)
	}

	return &config, nil
}

// SaveConfig 保存数据库配置
func (m *DatabaseConfigManager) SaveConfig(config *DatabaseConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 设置时间戳
	config.UpdatedAt = time.Now()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = config.UpdatedAt
	}

	// 验证配置
	if err := m.validateConfigInternal(config); err != nil {
		return fmt.Errorf("invalid database config: %w", err)
	}

	// 序列化JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database config: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write database config file: %w", err)
	}

	return nil
}

// Exists 检查配置文件是否存在
func (m *DatabaseConfigManager) Exists() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, err := os.Stat(m.configPath)
	return !os.IsNotExist(err)
}

// ValidateConfig 验证配置
func (m *DatabaseConfigManager) ValidateConfig(config *DatabaseConfig) error {
	return m.validateConfigInternal(config)
}

// validateConfigInternal 内部配置验证
func (m *DatabaseConfigManager) validateConfigInternal(config *DatabaseConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证数据库类型
	dbType := strings.ToLower(config.Database.Type)
	if dbType != "sqlite" && dbType != "mysql" && dbType != "postgresql" && dbType != "postgres" {
		return fmt.Errorf("unsupported database type: %s (supported: sqlite, mysql, postgresql)", config.Database.Type)
	}

	// 标准化数据库名称
	if dbType == "postgres" {
		config.Database.Type = "postgresql"
	}

	// 根据数据库类型验证必需字段
	switch config.Database.Type {
	case "sqlite":
		if config.Database.Path == "" {
			return fmt.Errorf("sqlite database path is required")
		}
		// 确保路径是绝对路径或相对路径
		if !filepath.IsAbs(config.Database.Path) && !strings.HasPrefix(config.Database.Path, "./") {
			config.Database.Path = "./" + config.Database.Path
		}

	case "mysql", "postgresql":
		if config.Database.Host == "" {
			return fmt.Errorf("%s host is required", config.Database.Type)
		}
		if config.Database.Database == "" {
			return fmt.Errorf("%s database name is required", config.Database.Type)
		}
		if config.Database.Username == "" {
			return fmt.Errorf("%s username is required", config.Database.Type)
		}
		if config.Database.Port <= 0 {
			// 设置默认端口
			if config.Database.Type == "mysql" {
				config.Database.Port = 3306
			} else if config.Database.Type == "postgresql" {
				config.Database.Port = 5432
			}
		}
	}

	// 验证连接池配置
	if config.Database.ConnectionPool.MaxOpenConns <= 0 {
		config.Database.ConnectionPool.MaxOpenConns = 25
	}
	if config.Database.ConnectionPool.MaxIdleConns <= 0 {
		config.Database.ConnectionPool.MaxIdleConns = 5
	}
	if config.Database.ConnectionPool.MaxIdleConns > config.Database.ConnectionPool.MaxOpenConns {
		config.Database.ConnectionPool.MaxIdleConns = config.Database.ConnectionPool.MaxOpenConns
	}
	if config.Database.ConnectionPool.ConnMaxLifetime <= 0 {
		config.Database.ConnectionPool.ConnMaxLifetime = 5 * time.Minute
	}

	// 验证管理员配置
	if config.Admin.Username == "" {
		return fmt.Errorf("admin username is required")
	}
	if len(config.Admin.Username) < 3 {
		return fmt.Errorf("admin username must be at least 3 characters")
	}
	if config.Admin.Password == "" {
		return fmt.Errorf("admin password is required")
	}
	if len(config.Admin.Password) < 6 {
		return fmt.Errorf("admin password must be at least 6 characters")
	}

	// 设置默认版本
	if config.Version == "" {
		config.Version = "1.0.0"
	}

	return nil
}

// GetDefaultConfig 获取默认配置
func (m *DatabaseConfigManager) GetDefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Database: DatabaseConnection{
			Type: "sqlite",
			Path: "./data/xiaozhi.db",
			ConnectionPool: ConnectionPool{
				MaxOpenConns:    25,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
			},
		},
		Admin: AdminConfig{
			Username: "admin",
			Password: generateRandomPassword(12),
			Email:    "admin@xiaozhi.local",
		},
		Initialized: false,
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// GetConnectionString 获取数据库连接字符串
func (m *DatabaseConfigManager) GetConnectionString(config *DatabaseConfig) (string, error) {
	if err := m.ValidateConfig(config); err != nil {
		return "", err
	}

	switch config.Database.Type {
	case "sqlite":
		return config.Database.Path, nil

	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Database,
			func() string {
				if config.Database.Charset == "" {
					return "utf8mb4"
				}
				return config.Database.Charset
			}())
		return dsn, nil

	case "postgresql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			config.Database.Host,
			config.Database.Username,
			config.Database.Password,
			config.Database.Database,
			config.Database.Port,
			func() string {
				if config.Database.SSLMode == "" {
					return "prefer"
				}
				return config.Database.SSLMode
			}())
		return dsn, nil

	default:
		return "", fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}
}

// CreateInitialConfig 创建初始配置
func (m *DatabaseConfigManager) CreateInitialConfig() error {
	if m.Exists() {
		return fmt.Errorf("database config already exists")
	}

	config := m.GetDefaultConfig()
	return m.SaveConfig(config)
}

// RemoveConfig 移除配置文件
func (m *DatabaseConfigManager) RemoveConfig() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.Exists() {
		return nil // 文件不存在，视为成功
	}

	return os.Remove(m.configPath)
}

// GetConfigPath 获取配置文件路径
func (m *DatabaseConfigManager) GetConfigPath() string {
	return m.configPath
}