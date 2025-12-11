package migrations

import (
	"gorm.io/gorm"
)

// Migration004PluginConfigTables 插件配置表迁移 - 创建插件供应商配置管理系统
type Migration004PluginConfigTables struct{}

func (m *Migration004PluginConfigTables) Version() string {
	return "004_plugin_config_tables"
}

func (m *Migration004PluginConfigTables) Description() string {
	return "Create plugin provider config management system with capabilities, snapshots and history"
}

func (m *Migration004PluginConfigTables) Up(db *gorm.DB) error {
	// 创建供应商配置主表
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS plugin_provider_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			provider_type VARCHAR(100) NOT NULL,           -- 供应商类型：openai, doubao, edge等
			provider_name VARCHAR(255) NOT NULL,           -- 供应商名称标识符
			display_name VARCHAR(255) NOT NULL,            -- 显示名称
			description TEXT,                              -- 供应商描述
			config_data TEXT NOT NULL,                     -- 加密的配置数据（JSON格式）
			config_schema TEXT NOT NULL,                   -- 配置模式定义（JSON Schema）
			enabled BOOLEAN DEFAULT TRUE,                  -- 是否启用
			priority INTEGER DEFAULT 100,                  -- 优先级（数字越小优先级越高）
			health_status VARCHAR(50) DEFAULT 'unknown',   -- 健康状态：healthy, unhealthy, unknown
			last_health_check DATETIME,                    -- 最后健康检查时间
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(provider_type, provider_name)           -- 每种类型的供应商名称唯一
		)
	`).Error; err != nil {
		return err
	}

	// 创建能力映射表
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS plugin_capabilities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			provider_config_id INTEGER NOT NULL,           -- 外键关联供应商配置
			capability_id VARCHAR(255) NOT NULL,           -- 能力ID：openai_chat, doubao_asr等
			capability_type VARCHAR(50) NOT NULL,          -- 能力类型：llm, asr, tts
			capability_name VARCHAR(255) NOT NULL,         -- 能力显示名称
			capability_description TEXT,                   -- 能力描述
			input_schema TEXT,                             -- 输入模式定义
			output_schema TEXT,                            -- 输出模式定义
			enabled BOOLEAN DEFAULT TRUE,                  -- 是否启用该能力
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(provider_config_id) REFERENCES plugin_provider_configs(id) ON DELETE CASCADE,
			UNIQUE(provider_config_id, capability_id)      -- 每个供应商的能力ID唯一
		)
	`).Error; err != nil {
		return err
	}

	// 创建配置快照表
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS plugin_config_snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			provider_config_id INTEGER NOT NULL,           -- 外键关联供应商配置
			version VARCHAR(50) NOT NULL,                  -- 版本号（如：v1.0.0, v2.1.0等）
			snapshot_name VARCHAR(255) NOT NULL,           -- 快照名称
			description TEXT,                              -- 快照描述
			snapshot_data TEXT NOT NULL,                   -- 快照配置数据（JSON格式）
			is_active BOOLEAN DEFAULT FALSE,               -- 是否为当前激活的快照
			created_by VARCHAR(255),                       -- 创建者
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(provider_config_id) REFERENCES plugin_provider_configs(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		return err
	}

	// 创建配置变更历史表
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS plugin_config_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			provider_config_id INTEGER NOT NULL,           -- 外键关联供应商配置
			operation VARCHAR(50) NOT NULL,                -- 操作类型：create, update, delete, enable, disable
			old_data TEXT,                                 -- 变更前的数据（JSON格式）
			new_data TEXT,                                 -- 变更后的数据（JSON格式）
			change_summary VARCHAR(1000),                  -- 变更摘要
			changed_fields TEXT,                           -- 变更字段列表（JSON数组）
			created_by VARCHAR(255),                       -- 操作用户
			user_agent TEXT,                               -- 用户代理
			ip_address VARCHAR(45),                        -- IP地址
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(provider_config_id) REFERENCES plugin_provider_configs(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		return err
	}

	// 创建索引
	// 供应商配置表索引
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_configs_type ON plugin_provider_configs(provider_type)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_configs_enabled ON plugin_provider_configs(enabled)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_configs_priority ON plugin_provider_configs(priority)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_configs_health ON plugin_provider_configs(health_status)`).Error; err != nil {
		return err
	}

	// 能力表索引
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_capabilities_config_id ON plugin_capabilities(provider_config_id)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_capabilities_type ON plugin_capabilities(capability_type)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_capabilities_enabled ON plugin_capabilities(enabled)`).Error; err != nil {
		return err
	}

	// 快照表索引
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_snapshots_config_id ON plugin_config_snapshots(provider_config_id)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_snapshots_version ON plugin_config_snapshots(version)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_snapshots_active ON plugin_config_snapshots(is_active)`).Error; err != nil {
		return err
	}

	// 历史表索引
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_history_config_id ON plugin_config_history(provider_config_id)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_history_operation ON plugin_config_history(operation)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_plugin_history_created_at ON plugin_config_history(created_at)`).Error; err != nil {
		return err
	}

	return nil
}

func (m *Migration004PluginConfigTables) Down(db *gorm.DB) error {
	// 删除表（按依赖关系逆序删除）
	if err := db.Exec(`DROP TABLE IF EXISTS plugin_config_history`).Error; err != nil {
		return err
	}
	if err := db.Exec(`DROP TABLE IF EXISTS plugin_config_snapshots`).Error; err != nil {
		return err
	}
	if err := db.Exec(`DROP TABLE IF EXISTS plugin_capabilities`).Error; err != nil {
		return err
	}
	if err := db.Exec(`DROP TABLE IF EXISTS plugin_provider_configs`).Error; err != nil {
		return err
	}

	return nil
}