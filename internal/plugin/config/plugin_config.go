package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// PluginConfig 插件配置
type PluginConfig struct {
	ID          string                 `yaml:"id" json:"id"`
	Name        string                 `yaml:"name" json:"name"`
	Version     string                 `yaml:"version" json:"version"`
	Description string                 `yaml:"description" json:"description"`
	Type        string                 `yaml:"type" json:"type"`
	Tags        []string               `yaml:"tags" json:"tags"`
	Deployment  DeploymentConfig       `yaml:"deployment" json:"deployment"`
	Config      map[string]interface{} `yaml:"config" json:"config"`
	Environment map[string]string      `yaml:"environment" json:"environment"`
	Enabled     bool                   `yaml:"enabled" json:"enabled"`
}

// DeploymentConfig 部署配置
type DeploymentConfig struct {
	Type       string            `yaml:"type" json:"type"`        // local_binary, container, remote_service
	Path       string            `yaml:"path" json:"path"`        // 二进制路径
	Image      string            `yaml:"image" json:"image"`      // 容器镜像
	Endpoint   string            `yaml:"endpoint" json:"endpoint"` // 远程服务端点
	Resources  ResourceConfig    `yaml:"resources" json:"resources"`
	Timeout    time.Duration     `yaml:"timeout" json:"timeout"`
	RetryCount int               `yaml:"retry_count" json:"retry_count"`
	Options    map[string]string `yaml:"options" json:"options"`
}

// ResourceConfig 资源配置
type ResourceConfig struct {
	MaxMemory string `yaml:"max_memory" json:"max_memory"`
	MaxCPU    string `yaml:"max_cpu" json:"max_cpu"`
	MaxGPU    string `yaml:"max_gpu" json:"max_gpu"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	Plugins   map[string]*PluginConfig `yaml:"plugins" json:"plugins"`
	Manager   ManagerConfig           `yaml:"manager" json:"manager"`
	Gateway   GatewayConfig           `yaml:"gateway" json:"gateway"`
	Security  SecurityConfig          `yaml:"security" json:"security"`
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	Enabled      bool                `yaml:"enabled" json:"enabled"`
	Discovery    DiscoveryConfig     `yaml:"discovery" json:"discovery"`
	Registry     RegistryConfig      `yaml:"registry" json:"registry"`
	HealthCheck  HealthCheckConfig   `yaml:"health_check" json:"health_check"`
}

// DiscoveryConfig 发现配置
type DiscoveryConfig struct {
	Enabled      bool          `yaml:"enabled" json:"enabled"`
	ScanInterval time.Duration `yaml:"scan_interval" json:"scan_interval"`
	Paths        []string      `yaml:"paths" json:"paths"`
}

// RegistryConfig 注册表配置
type RegistryConfig struct {
	Type string        `yaml:"type" json:"type"` // memory, redis, etcd
	TTL  time.Duration `yaml:"ttl" json:"ttl"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Interval         time.Duration `yaml:"interval" json:"interval"`
	Timeout          time.Duration `yaml:"timeout" json:"timeout"`
	FailureThreshold int           `yaml:"failure_threshold" json:"failure_threshold"`
}

// GatewayConfig 网关配置
type GatewayConfig struct {
	Enabled      bool           `yaml:"enabled" json:"enabled"`
	Routing      RoutingConfig  `yaml:"routing" json:"routing"`
	LoadBalancer LoadBalancerConfig `yaml:"load_balancer" json:"load_balancer"`
}

// RoutingConfig 路由配置
type RoutingConfig struct {
	Rules []RoutingRule `yaml:"rules" json:"rules"`
}

// RoutingRule 路由规则
type RoutingRule struct {
	ID      string `yaml:"id" json:"id"`
	Pattern string `yaml:"pattern" json:"pattern"`
	Target  string `yaml:"target" json:"target"`
	Weight  int    `yaml:"weight" json:"weight"`
}

// LoadBalancerConfig 负载均衡配置
type LoadBalancerConfig struct {
	Strategy string `yaml:"strategy" json:"strategy"` // round_robin, random, least_connections
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	Sandbox   SandboxConfig   `yaml:"sandbox" json:"sandbox"`
	Signature SignatureConfig `yaml:"signature" json:"signature"`
}

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Type    string `yaml:"type" json:"type"` // docker, gvisor, none
}

// SignatureConfig 签名配置
type SignatureConfig struct {
	Enabled    bool     `yaml:"enabled" json:"enabled"`
	PublicKeys []string `yaml:"public_keys" json:"public_keys"`
}

// DefaultConfig 默认配置
func DefaultConfig() *SystemConfig {
	return &SystemConfig{
		Plugins: make(map[string]*PluginConfig),
		Manager: ManagerConfig{
			Enabled: true,
			Discovery: DiscoveryConfig{
				Enabled:      true,
				ScanInterval: 30 * time.Second,
				Paths:        getDefaultPluginPaths(),
			},
			Registry: RegistryConfig{
				Type: "memory",
				TTL:  5 * time.Minute,
			},
			HealthCheck: HealthCheckConfig{
				Interval:         10 * time.Second,
				Timeout:          5 * time.Second,
				FailureThreshold: 3,
			},
		},
		Gateway: GatewayConfig{
			Enabled: true,
			Routing: RoutingConfig{
				Rules: []RoutingRule{
					{
						ID:      "plugin-rules",
						Pattern: "plugin_*",
						Target:  "plugin",
						Weight:  1,
					},
					{
						ID:      "mcp-rules",
						Pattern: "mcp_*",
						Target:  "mcp",
						Weight:  1,
					},
					{
						ID:      "default-rules",
						Pattern: "*",
						Target:  "provider",
						Weight:  1,
					},
				},
			},
			LoadBalancer: LoadBalancerConfig{
				Strategy: "round_robin",
			},
		},
		Security: SecurityConfig{
			Sandbox: SandboxConfig{
				Enabled: true,
				Type:    "docker",
			},
			Signature: SignatureConfig{
				Enabled: true,
				PublicKeys: []string{
					"./keys/public_key.pem",
				},
			},
		},
	}
}

// LoadConfigFromFile 从文件加载配置
func LoadConfigFromFile(configPath string) (*SystemConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，返回默认配置
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// SaveConfigToFile 保存配置到文件
func SaveConfigToFile(config *SystemConfig, configPath string) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig 验证配置
func ValidateConfig(config *SystemConfig) error {
	// 验证插件配置
	for id, plugin := range config.Plugins {
		if plugin.ID == "" {
			plugin.ID = id
		}

		if plugin.Name == "" {
			return fmt.Errorf("plugin %s: name is required", id)
		}

		if plugin.Deployment.Type == "" {
			return fmt.Errorf("plugin %s: deployment type is required", id)
		}

		switch plugin.Deployment.Type {
		case "local_binary":
			if plugin.Deployment.Path == "" {
				return fmt.Errorf("plugin %s: path is required for local_binary deployment", id)
			}
		case "container":
			if plugin.Deployment.Image == "" {
				return fmt.Errorf("plugin %s: image is required for container deployment", id)
			}
		case "remote_service":
			if plugin.Deployment.Endpoint == "" {
				return fmt.Errorf("plugin %s: endpoint is required for remote_service deployment", id)
			}
		default:
			return fmt.Errorf("plugin %s: unsupported deployment type %s", id, plugin.Deployment.Type)
		}
	}

	// 验证管理器配置
	if config.Manager.HealthCheck.Interval <= 0 {
		return fmt.Errorf("health check interval must be positive")
	}

	if config.Manager.HealthCheck.Timeout <= 0 {
		return fmt.Errorf("health check timeout must be positive")
	}

	if config.Manager.HealthCheck.FailureThreshold <= 0 {
		return fmt.Errorf("health check failure threshold must be positive")
	}

	// 验证网关配置
	if config.Gateway.Enabled {
		if len(config.Gateway.Routing.Rules) == 0 {
			return fmt.Errorf("gateway is enabled but no routing rules defined")
		}

		for i, rule := range config.Gateway.Routing.Rules {
			if rule.ID == "" {
				return fmt.Errorf("routing rule %d: ID is required", i)
			}
			if rule.Pattern == "" {
				return fmt.Errorf("routing rule %s: pattern is required", rule.ID)
			}
			if rule.Target == "" {
				return fmt.Errorf("routing rule %s: target is required", rule.ID)
			}
		}
	}

	return nil
}

// MergeConfigs 合并配置
func MergeConfigs(base, override *SystemConfig) *SystemConfig {
	result := *base

	// 合并插件配置
	if override.Plugins != nil {
		if result.Plugins == nil {
			result.Plugins = make(map[string]*PluginConfig)
		}
		for id, plugin := range override.Plugins {
			result.Plugins[id] = plugin
		}
	}

	// 合并管理器配置
	if override.Manager.Enabled != result.Manager.Enabled {
		result.Manager.Enabled = override.Manager.Enabled
	}
	if override.Manager.Discovery.Enabled {
		result.Manager.Discovery = override.Manager.Discovery
	}
	if override.Manager.Registry.Type != "" {
		result.Manager.Registry = override.Manager.Registry
	}
	if override.Manager.HealthCheck.Interval > 0 {
		result.Manager.HealthCheck = override.Manager.HealthCheck
	}

	// 合并网关配置
	if override.Gateway.Enabled {
		result.Gateway = override.Gateway
	}

	// 合并安全配置
	if override.Security.Sandbox.Enabled {
		result.Security.Sandbox = override.Security.Sandbox
	}
	if override.Security.Signature.Enabled {
		result.Security.Signature = override.Security.Signature
	}

	return &result
}

// GetPluginConfig 获取插件配置
func (config *SystemConfig) GetPluginConfig(pluginID string) (*PluginConfig, bool) {
	plugin, exists := config.Plugins[pluginID]
	return plugin, exists
}

// SetPluginConfig 设置插件配置
func (config *SystemConfig) SetPluginConfig(pluginConfig *PluginConfig) {
	if config.Plugins == nil {
		config.Plugins = make(map[string]*PluginConfig)
	}
	config.Plugins[pluginConfig.ID] = pluginConfig
}

// RemovePluginConfig 移除插件配置
func (config *SystemConfig) RemovePluginConfig(pluginID string) {
	if config.Plugins != nil {
		delete(config.Plugins, pluginID)
	}
}

// ListPluginConfigs 列出所有插件配置
func (config *SystemConfig) ListPluginConfigs() []*PluginConfig {
	plugins := make([]*PluginConfig, 0, len(config.Plugins))
	for _, plugin := range config.Plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// GetEnabledPluginConfigs 获取启用的插件配置
func (config *SystemConfig) GetEnabledPluginConfigs() []*PluginConfig {
	var plugins []*PluginConfig
	for _, plugin := range config.Plugins {
		if plugin.Enabled {
			plugins = append(plugins, plugin)
		}
	}
	return plugins
}

// getDefaultPluginPaths 获取默认插件路径
func getDefaultPluginPaths() []string {
	paths := []string{"./plugins"}

	// 添加系统插件路径
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".xiaozhi-flow", "plugins"))
	}

	paths = append(paths, "/opt/xiaozhi-flow/plugins")

	// 过滤存在的路径
	var existingPaths []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existingPaths = append(existingPaths, path)
		}
	}

	if len(existingPaths) == 0 {
		// 如果没有现有路径，至少返回当前目录下的plugins
		return []string{"./plugins"}
	}

	return existingPaths
}