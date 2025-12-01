package adapters

import (
	"fmt"
	"sync"

	"xiaozhi-server-go/internal/platform/storage"
)

// AdapterFactory 适配器工厂
type AdapterFactory struct {
	mu        sync.RWMutex
	adapters  map[string]func() DatabaseAdapter
	supported map[string]AdapterCapabilities
}

// NewAdapterFactory 创建新的适配器工厂
func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{
		adapters:  make(map[string]func() DatabaseAdapter),
		supported: make(map[string]AdapterCapabilities),
	}
}

// RegisterAdapter 注册适配器
func (f *AdapterFactory) RegisterAdapter(dbType string, creator func() DatabaseAdapter, capabilities AdapterCapabilities) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.adapters[dbType] = creator
	f.supported[dbType] = capabilities
}

// CreateAdapter 创建适配器
func (f *AdapterFactory) CreateAdapter(dbType string) (DatabaseAdapter, error) {
	f.mu.RLock()
	creator, exists := f.adapters[dbType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	return creator(), nil
}

// GetSupportedTypes 获取支持的数据库类型
func (f *AdapterFactory) GetSupportedTypes() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	types := make([]string, 0, len(f.adapters))
	for dbType := range f.adapters {
		types = append(types, dbType)
	}
	return types
}

// GetCapabilities 获取数据库类型的能力
func (f *AdapterFactory) GetCapabilities(dbType string) (AdapterCapabilities, error) {
	f.mu.RLock()
	capabilities, exists := f.supported[dbType]
	f.mu.RUnlock()

	if !exists {
		return AdapterCapabilities{}, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	return capabilities, nil
}

// IsSupported 检查是否支持指定的数据库类型
func (f *AdapterFactory) IsSupported(dbType string) bool {
	f.mu.RLock()
	_, exists := f.adapters[dbType]
	f.mu.RUnlock()
	return exists
}

// ValidateConfig 验证数据库配置
func (f *AdapterFactory) ValidateConfig(config storage.DatabaseConnection) error {
	if !f.IsSupported(config.Type) {
		return fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	// 根据数据库类型验证特定配置
	switch config.Type {
	case "sqlite":
		if config.Path == "" {
			return fmt.Errorf("SQLite需要指定数据库路径")
		}
	case "mysql":
		if config.Host == "" || config.Database == "" {
			return fmt.Errorf("MySQL需要指定主机和数据库名")
		}
	case "postgresql":
		if config.Host == "" || config.Database == "" {
			return fmt.Errorf("PostgreSQL需要指定主机和数据库名")
		}
	}

	return nil
}

// 全局工厂实例
var GlobalFactory = NewAdapterFactory()

// 初始化函数，注册所有适配器
func init() {
	// 注册SQLite适配器
	GlobalFactory.RegisterAdapter("sqlite", func() DatabaseAdapter {
		return NewSQLiteAdapter()
	}, AdapterCapabilities{
		SupportsTransactions:   true,
		SupportsForeignKeys:    true,
		SupportsJSON:           true,
		SupportsFullText:       true,
		MaxConnections:        10,
		PreferredPoolSize:      3,
	})

	// TODO: 后续添加MySQL和PostgreSQL
	// GlobalFactory.RegisterAdapter("mysql", func() DatabaseAdapter {
	//     return NewMySQLAdapter()
	// }, AdapterCapabilities{...})

	// GlobalFactory.RegisterAdapter("postgresql", func() DatabaseAdapter {
	//     return NewPostgreSQLAdapter()
	// }, AdapterCapabilities{...})
}

// CreateDatabaseAdapter 便捷函数：根据配置创建数据库适配器
func CreateDatabaseAdapter(config storage.DatabaseConnection) (DatabaseAdapter, error) {
	if err := GlobalFactory.ValidateConfig(config); err != nil {
		return nil, err
	}

	return GlobalFactory.CreateAdapter(config.Type)
}

// GetSupportedDatabaseTypes 获取所有支持的数据库类型
func GetSupportedDatabaseTypes() []string {
	return GlobalFactory.GetSupportedTypes()
}