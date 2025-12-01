package test

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pluginv1 "xiaozhi-server-go/api/v1"
	"xiaozhi-server-go/internal/plugin/config"
	"xiaozhi-server-go/internal/plugin/manager"
	"xiaozhi-server-go/internal/plugin/discovery"
)

// TestPluginManager 测试插件管理器
func TestPluginManager(t *testing.T) {
	// 创建测试配置
	pluginConfig := &config.ManagerConfig{
		Enabled: true,
		Discovery: &config.DiscoveryConfig{
			Enabled:      true,
			ScanInterval: 5 * time.Second,
			Paths:        []string{"../../plugins/examples"},
		},
		Registry: &config.RegistryConfig{
			Type: "memory",
			TTL:  5 * time.Minute,
		},
		HealthCheck: &config.HealthCheckConfig{
			Interval:         2 * time.Second,
			Timeout:          1 * time.Second,
			FailureThreshold: 3,
		},
	}

	// 创建日志记录器
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "test-plugin-manager",
		Level:  hclog.Debug,
		Output: hclog.DefaultOutput,
	})

	// 创建插件管理器配置
	managerConfig := &manager.PluginConfig{
		Enabled:      true,
		Discovery:    pluginConfig.Discovery,
		Registry:     pluginConfig.Registry,
		HealthCheck:  pluginConfig.HealthCheck,
	}

	// 创建插件管理器
	pm, err := manager.NewPluginManager(managerConfig, logger)
	require.NoError(t, err)
	require.NotNil(t, pm)

	// 启动插件管理器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = pm.Start(ctx)
	require.NoError(t, err)

	// 等待启动完成
	time.Sleep(1 * time.Second)

	// 测试插件发现
	plugins, err := pm.DiscoverPlugins(ctx)
	require.NoError(t, err)
	t.Logf("Discovered %d plugins", len(plugins))

	// 列出已加载的插件
	loadedPlugins, err := pm.ListPlugins()
	require.NoError(t, err)
	t.Logf("Loaded plugins count: %d", len(loadedPlugins))

	// 健康检查所有插件
	healthStatuses := pm.HealthCheckAll(ctx)
	for pluginID, status := range healthStatuses {
		t.Logf("Plugin %s health: %v", pluginID, status.Healthy)
	}

	// 停止插件管理器
	err = pm.Stop(ctx)
	require.NoError(t, err)
}

// TestHelloPluginConfig 测试Hello插件配置
func TestHelloPluginConfig(t *testing.T) {
	// 加载插件配置
	systemConfig, err := config.LoadConfigFromFile("../../../config/plugins.yaml")
	require.NoError(t, err)

	// 获取Hello插件配置
	helloConfig, exists := systemConfig.GetPluginConfig("hello-plugin")
	require.True(t, exists)
	require.NotNil(t, helloConfig)

	// 验证配置
	assert.Equal(t, "hello-plugin", helloConfig.ID)
	assert.Equal(t, "Hello Plugin", helloConfig.Name)
	assert.Equal(t, "1.0.0", helloConfig.Version)
	assert.Equal(t, "utility", helloConfig.Type)
	assert.Equal(t, "local_binary", helloConfig.Deployment.Type)
	assert.True(t, helloConfig.Enabled)

	// 验证部署配置
	assert.NotEmpty(t, helloConfig.Deployment.Path)
	assert.Equal(t, 30*time.Second, helloConfig.Deployment.Timeout)
	assert.Equal(t, 3, helloConfig.Deployment.RetryCount)

	// 验证配置参数
	assert.Equal(t, "en", helloConfig.Config["greeting_language"])
	assert.Equal(t, "World", helloConfig.Config["default_name"])

	// 验证环境变量
	assert.Equal(t, "info", helloConfig.Environment["PLUGIN_LOG_LEVEL"])
	assert.Equal(t, "false", helloConfig.Environment["PLUGIN_DEBUG"])

	t.Logf("Hello plugin config validated successfully: %+v", helloConfig)
}

// TestPluginInfoValidation 测试插件信息验证
func TestPluginInfoValidation(t *testing.T) {
	info := &pluginv1.PluginInfo{
		Id:          "test-plugin",
		Name:        "Test Plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Author:      "Test Author",
		Type:        pluginv1.PluginType_PLUGIN_TYPE_UTILITY,
		Tags:        []string{"test", "utility"},
		Capabilities: []string{"execute"},
	}

	// 验证必需字段
	assert.NotEmpty(t, info.Id)
	assert.NotEmpty(t, info.Name)
	assert.NotEmpty(t, info.Version)
	assert.NotEqual(t, pluginv1.PluginType_PLUGIN_TYPE_UNSPECIFIED, info.Type)

	// 验证枚举值
	assert.True(t, info.Type >= pluginv1.PluginType_PLUGIN_TYPE_AUDIO && info.Type <= pluginv1.PluginType_PLUGIN_TYPE_CUSTOM)

	t.Logf("Plugin info validation passed: %+v", info)
}

// TestPluginTypes 测试插件类型
func TestPluginTypes(t *testing.T) {
	types := map[pluginv1.PluginType]string{
		pluginv1.PluginType_PLUGIN_TYPE_UNSPECIFIED: "unspecified",
		pluginv1.PluginType_PLUGIN_TYPE_AUDIO:       "audio",
		pluginv1.PluginType_PLUGIN_TYPE_LLM:         "llm",
		pluginv1.PluginType_PLUGIN_TYPE_DEVICE:      "device",
		pluginv1.PluginType_PLUGIN_TYPE_UTILITY:     "utility",
		pluginv1.PluginType_PLUGIN_TYPE_CUSTOM:      "custom",
	}

	for pluginType, expectedName := range types {
		assert.Equal(t, expectedName, pluginType.String())
		t.Logf("Plugin type %d = %s", pluginType, pluginType.String())
	}
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	defaultConfig := config.DefaultConfig()
	require.NotNil(t, defaultConfig)

	// 验证默认配置
	assert.True(t, defaultConfig.Manager.Enabled)
	assert.True(t, defaultConfig.Manager.Discovery.Enabled)
	assert.NotEmpty(t, defaultConfig.Manager.Discovery.Paths)
	assert.Equal(t, "memory", defaultConfig.Manager.Registry.Type)
	assert.True(t, defaultConfig.Gateway.Enabled)
	assert.Len(t, defaultConfig.Gateway.Routing.Rules, 3)

	t.Logf("Default config validated: enabled=%v", defaultConfig.Manager.Enabled)
}

// BenchmarkPluginDiscovery 性能测试：插件发现
func BenchmarkPluginDiscovery(b *testing.B) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "benchmark-discovery",
		Level: hclog.Error, // 减少日志输出
	})

	discoveryConfig := &config.DiscoveryConfig{
		Enabled:      true,
		ScanInterval: 30 * time.Second,
		Paths:        []string{"../../plugins/examples"},
	}

	discovery, err := discovery.NewDiscovery(discoveryConfig, logger)
	require.NoError(b, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = discovery.Start(ctx)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		plugins, err := discovery.Discover(ctx)
		if err != nil {
			b.Fatalf("Discovery failed: %v", err)
		}
		b.Logf("Discovery iteration %d found %d plugins", i, len(plugins))
	}
}