package runtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/kalicyh/xiaozhi-flow/internal/plugin/config"
	"github.com/kalicyh/xiaozhi-flow/internal/plugin/sdk"
)

// Runtime 运行时接口
type Runtime interface {
	// 启动插件
	Start(ctx context.Context, config *config.PluginConfig) (plugin.PluginClient, error)
	// 停止插件
	Stop(ctx context.Context, pluginID string) error
	// 获取运行时类型
	Type() string
	// 健康检查
	HealthCheck(ctx context.Context) error
	// 停止运行时
	Shutdown(ctx context.Context) error
}

// Manager 运行时管理器接口
type Manager interface {
	// 获取运行时
	GetRuntime(runtimeType string) (Runtime, error)
	// 注册运行时
	RegisterRuntime(runtimeType string, runtime Runtime) error
	// 列出运行时
	ListRuntimes() []string
	// 停止所有运行时
	Shutdown(ctx context.Context) error
}

// managerImpl 运行时管理器实现
type managerImpl struct {
	logger   hclog.Logger
	runtimes map[string]Runtime
	mu       sync.RWMutex
}

// NewManager 创建运行时管理器
func NewManager(logger hclog.Logger) (Manager, error) {
	m := &managerImpl{
		logger:   logger.Named("runtime-manager"),
		runtimes: make(map[string]Runtime),
	}

	// 注册内置运行时
	if err := m.registerBuiltinRuntimes(); err != nil {
		return nil, fmt.Errorf("failed to register builtin runtimes: %w", err)
	}

	return m, nil
}

// GetRuntime 获取运行时
func (m *managerImpl) GetRuntime(runtimeType string) (Runtime, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	runtime, exists := m.runtimes[runtimeType]
	if !exists {
		return nil, fmt.Errorf("runtime type %s not found", runtimeType)
	}

	return runtime, nil
}

// RegisterRuntime 注册运行时
func (m *managerImpl) RegisterRuntime(runtimeType string, runtime Runtime) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.runtimes[runtimeType]; exists {
		return fmt.Errorf("runtime type %s already registered", runtimeType)
	}

	m.runtimes[runtimeType] = runtime
	m.logger.Info("Runtime registered", "type", runtimeType)

	return nil
}

// ListRuntimes 列出运行时
func (m *managerImpl) ListRuntimes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]string, 0, len(m.runtimes))
	for runtimeType := range m.runtimes {
		types = append(types, runtimeType)
	}

	return types
}

// Shutdown 停止所有运行时
func (m *managerImpl) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for runtimeType, runtime := range m.runtimes {
		if err := runtime.Shutdown(ctx); err != nil {
			m.logger.Error("Failed to shutdown runtime", "type", runtimeType, "error", err)
		}
	}

	m.runtimes = make(map[string]Runtime)
	m.logger.Info("All runtimes shutdown")

	return nil
}

// registerBuiltinRuntimes 注册内置运行时
func (m *managerImpl) registerBuiltinRuntimes() error {
	// 注册本地二进制运行时
	localRuntime := NewLocalBinaryRuntime(m.logger)
	if err := m.RegisterRuntime("local_binary", localRuntime); err != nil {
		return fmt.Errorf("failed to register local_binary runtime: %w", err)
	}

	// 注册容器运行时
	containerRuntime := NewContainerRuntime(m.logger)
	if err := m.RegisterRuntime("container", containerRuntime); err != nil {
		return fmt.Errorf("failed to register container runtime: %w", err)
	}

	// 注册远程服务运行时
	remoteRuntime := NewRemoteServiceRuntime(m.logger)
	if err := m.RegisterRuntime("remote_service", remoteRuntime); err != nil {
		return fmt.Errorf("failed to register remote_service runtime: %w", err)
	}

	m.logger.Info("Builtin runtimes registered", "count", 3)
	return nil
}

// createClientConfig 创建客户端配置
func createClientConfig(config *config.PluginConfig) *plugin.ClientConfig {
	return &plugin.ClientConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins:         sdk.PluginMap,
		Cmd:             nil, // 将由具体的运行时设置
		AllowedProtocols: []plugin.ClientProtocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC,
		},
	}
}