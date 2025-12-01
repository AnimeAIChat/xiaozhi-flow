package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"xiaozhi-server-go/internal/plugin/config"
	"xiaozhi-server-go/internal/plugin/sdk"
)

// LocalBinaryRuntime 本地二进制运行时
type LocalBinaryRuntime struct {
	logger  hclog.Logger
	clients map[string]*plugin.Client
	mu      sync.RWMutex
}

// NewLocalBinaryRuntime 创建本地二进制运行时
func NewLocalBinaryRuntime(logger hclog.Logger) *LocalBinaryRuntime {
	return &LocalBinaryRuntime{
		logger:  logger.Named("local-binary-runtime"),
		clients: make(map[string]*plugin.Client),
	}
}

// Start 启动插件
func (r *LocalBinaryRuntime) Start(ctx context.Context, config *config.PluginConfig) (plugin.PluginClient, error) {
	r.logger.Info("Starting local binary plugin", "plugin_id", config.ID, "path", config.Deployment.Path)

	// 验证二进制文件存在
	if _, err := os.Stat(config.Deployment.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin binary not found at %s: %w", config.Deployment.Path, err)
	}

	// 检查文件是否可执行
	if !isExecutable(config.Deployment.Path) {
		return nil, fmt.Errorf("plugin binary is not executable: %s", config.Deployment.Path)
	}

	// 创建客户端配置
	clientConfig := &plugin.ClientConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins:         sdk.PluginMap,
		AllowedProtocols: []plugin.ClientProtocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC,
		},
	}

	// 创建插件客户端命令
	cmd := exec.CommandContext(ctx, config.Deployment.Path)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置环境变量
	if len(config.Environment) > 0 {
		env := os.Environ()
		for k, v := range config.Environment {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	// 创建插件客户端
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  clientConfig.HandshakeConfig,
		Plugins:         clientConfig.Plugins,
		Cmd:             cmd,
		AllowedProtocols: clientConfig.AllowedProtocols,
	})

	// 启动客户端
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to get RPC client: %w", err)
	}

	// 测试连接
	if err := rpcClient.Ping(); err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to ping plugin: %w", err)
	}

	// 存储客户端
	r.mu.Lock()
	r.clients[config.ID] = client
	r.mu.Unlock()

	r.logger.Info("Local binary plugin started successfully", "plugin_id", config.ID)
	return client, nil
}

// Stop 停止插件
func (r *LocalBinaryRuntime) Stop(ctx context.Context, pluginID string) error {
	r.logger.Info("Stopping local binary plugin", "plugin_id", pluginID)

	r.mu.Lock()
	client, exists := r.clients[pluginID]
	if !exists {
		r.mu.Unlock()
		return fmt.Errorf("plugin %s not found", pluginID)
	}
	delete(r.clients, pluginID)
	r.mu.Unlock()

	// 尝试优雅关闭
	if err := client.Kill(); err != nil {
		r.logger.Warn("Failed to kill plugin client gracefully", "plugin_id", pluginID, "error", err)
		// 强制关闭
		if err := client.Kill(); err != nil {
			r.logger.Error("Failed to kill plugin client forcefully", "plugin_id", pluginID, "error", err)
			return err
		}
	}

	r.logger.Info("Local binary plugin stopped successfully", "plugin_id", pluginID)
	return nil
}

// Type 获取运行时类型
func (r *LocalBinaryRuntime) Type() string {
	return "local_binary"
}

// HealthCheck 健康检查
func (r *LocalBinaryRuntime) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	clientCount := len(r.clients)
	r.mu.RUnlock()

	r.logger.Debug("Local binary runtime health check", "active_clients", clientCount)
	return nil
}

// Shutdown 关闭运行时
func (r *LocalBinaryRuntime) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down local binary runtime")

	r.mu.Lock()
	clients := make(map[string]*plugin.Client)
	for id, client := range r.clients {
		clients[id] = client
	}
	r.clients = make(map[string]*plugin.Client)
	r.mu.Unlock()

	// 关闭所有客户端
	for id, client := range clients {
		if err := client.Kill(); err != nil {
			r.logger.Error("Failed to kill client during shutdown", "plugin_id", id, "error", err)
		}
	}

	r.logger.Info("Local binary runtime shutdown complete")
	return nil
}

// isExecutable 检查文件是否可执行
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// 在 Windows 上，检查文件扩展名
	if filepath.Ext(path) == ".exe" {
		return true
	}

	// 在 Unix 系统上检查执行权限
	return info.Mode().Perm()&0111 != 0
}

// ContainerRuntime 容器运行时
type ContainerRuntime struct {
	logger  hclog.Logger
	clients map[string]*plugin.Client
	mu      sync.RWMutex
}

// NewContainerRuntime 创建容器运行时
func NewContainerRuntime(logger hclog.Logger) *ContainerRuntime {
	return &ContainerRuntime{
		logger:  logger.Named("container-runtime"),
		clients: make(map[string]*plugin.Client),
	}
}

// Start 启动容器插件
func (r *ContainerRuntime) Start(ctx context.Context, config *config.PluginConfig) (plugin.PluginClient, error) {
	r.logger.Info("Starting container plugin", "plugin_id", config.ID, "image", config.Deployment.Image)

	// 这里应该实现 Docker 容器启动逻辑
	// 为了简化，这里提供一个基础的实现框架

	// 检查 Docker 是否可用
	if !isDockerAvailable() {
		return nil, fmt.Errorf("docker is not available on this system")
	}

	// 创建容器并启动
	// 这里应该使用 Docker API 启动容器
	// 简化实现，使用本地模拟

	r.logger.Info("Container plugin started (simplified)", "plugin_id", config.ID)
	return nil, fmt.Errorf("container runtime not fully implemented yet")
}

// Stop 停止容器插件
func (r *ContainerRuntime) Stop(ctx context.Context, pluginID string) error {
	r.logger.Info("Stopping container plugin", "plugin_id", pluginID)

	r.mu.Lock()
	delete(r.clients, pluginID)
	r.mu.Unlock()

	// 停止并删除容器
	// 这里应该使用 Docker API 停止容器

	r.logger.Info("Container plugin stopped", "plugin_id", pluginID)
	return nil
}

// Type 获取运行时类型
func (r *ContainerRuntime) Type() string {
	return "container"
}

// HealthCheck 健康检查
func (r *ContainerRuntime) HealthCheck(ctx context.Context) error {
	// 检查 Docker 守护进程是否运行
	if !isDockerAvailable() {
		return fmt.Errorf("docker daemon is not running")
	}

	r.mu.RLock()
	clientCount := len(r.clients)
	r.mu.RUnlock()

	r.logger.Debug("Container runtime health check", "active_containers", clientCount)
	return nil
}

// Shutdown 关闭运行时
func (r *ContainerRuntime) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down container runtime")

	r.mu.Lock()
	clients := make(map[string]*plugin.Client)
	for id, client := range r.clients {
		clients[id] = client
	}
	r.clients = make(map[string]*plugin.Client)
	r.mu.Unlock()

	// 停止所有容器
	for id := range clients {
		if err := r.Stop(ctx, id); err != nil {
			r.logger.Error("Failed to stop container during shutdown", "container_id", id, "error", err)
		}
	}

	r.logger.Info("Container runtime shutdown complete")
	return nil
}

// isDockerAvailable 检查 Docker 是否可用
func isDockerAvailable() bool {
	// 简单的 Docker 可用性检查
	// 在实际实现中，应该检查 Docker 守护进程是否运行
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "version")
	err := cmd.Run()
	return err == nil
}

// RemoteServiceRuntime 远程服务运行时
type RemoteServiceRuntime struct {
	logger  hclog.Logger
	clients map[string]*plugin.Client
	mu      sync.RWMutex
}

// NewRemoteServiceRuntime 创建远程服务运行时
func NewRemoteServiceRuntime(logger hclog.Logger) *RemoteServiceRuntime {
	return &RemoteServiceRuntime{
		logger:  logger.Named("remote-service-runtime"),
		clients: make(map[string]*plugin.Client),
	}
}

// Start 连接到远程服务插件
func (r *RemoteServiceRuntime) Start(ctx context.Context, config *config.PluginConfig) (plugin.PluginClient, error) {
	r.logger.Info("Connecting to remote service plugin", "plugin_id", config.ID, "endpoint", config.Deployment.Endpoint)

	// 创建客户端配置
	clientConfig := &plugin.ClientConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins:         sdk.PluginMap,
		AllowedProtocols: []plugin.ClientProtocol{
			plugin.ProtocolGRPC,
		},
	}

	// 创建 gRPC 客户端连接
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  clientConfig.HandshakeConfig,
		Plugins:         clientConfig.Plugins,
		AllowedProtocols: clientConfig.AllowedProtocols,
		// 这里需要设置 GRPC 客户端连接配置
		// 简化实现
	})

	// 测试连接
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to connect to remote service: %w", err)
	}

	if err := rpcClient.Ping(); err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to ping remote service: %w", err)
	}

	// 存储客户端
	r.mu.Lock()
	r.clients[config.ID] = client
	r.mu.Unlock()

	r.logger.Info("Remote service plugin connected successfully", "plugin_id", config.ID)
	return client, nil
}

// Stop 断开远程服务连接
func (r *RemoteServiceRuntime) Stop(ctx context.Context, pluginID string) error {
	r.logger.Info("Disconnecting remote service plugin", "plugin_id", pluginID)

	r.mu.Lock()
	client, exists := r.clients[pluginID]
	if !exists {
		r.mu.Unlock()
		return fmt.Errorf("plugin %s not found", pluginID)
	}
	delete(r.clients, pluginID)
	r.mu.Unlock()

	if err := client.Kill(); err != nil {
		r.logger.Error("Failed to disconnect remote service", "plugin_id", pluginID, "error", err)
		return err
	}

	r.logger.Info("Remote service plugin disconnected successfully", "plugin_id", pluginID)
	return nil
}

// Type 获取运行时类型
func (r *RemoteServiceRuntime) Type() string {
	return "remote_service"
}

// HealthCheck 健康检查
func (r *RemoteServiceRuntime) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	clientCount := len(r.clients)
	r.mu.RUnlock()

	r.logger.Debug("Remote service runtime health check", "active_connections", clientCount)
	return nil
}

// Shutdown 关闭运行时
func (r *RemoteServiceRuntime) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down remote service runtime")

	r.mu.Lock()
	clients := make(map[string]*plugin.Client)
	for id, client := range r.clients {
		clients[id] = client
	}
	r.clients = make(map[string]*plugin.Client)
	r.mu.Unlock()

	// 断开所有连接
	for id, client := range clients {
		if err := client.Kill(); err != nil {
			r.logger.Error("Failed to disconnect during shutdown", "plugin_id", id, "error", err)
		}
	}

	r.logger.Info("Remote service runtime shutdown complete")
	return nil
}