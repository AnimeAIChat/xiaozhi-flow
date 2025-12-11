package server

import (
	"fmt"
	"sync"

	"xiaozhi-server-go/internal/platform/logging"
	pluginpb "xiaozhi-server-go/gen/go/api/proto"
)

// BaseGRPCProvider gRPC插件提供者基础实现
type BaseGRPCProvider struct {
	Logger         *logging.Logger
	GRPCServer     *GRPCServer
	ServiceAddress string
	Mu             sync.RWMutex
	PluginID       string
	ServiceFactory func() pluginpb.PluginServiceServer
}

// NewBaseGRPCProvider 创建新的BaseGRPCProvider
func NewBaseGRPCProvider(pluginID string, logger *logging.Logger, factory func() pluginpb.PluginServiceServer) *BaseGRPCProvider {
	if logger == nil {
		logger = logging.DefaultLogger
	}
	return &BaseGRPCProvider{
		PluginID:       pluginID,
		Logger:         logger,
		ServiceFactory: factory,
	}
}

// StartGRPCServer 启动gRPC服务
func (p *BaseGRPCProvider) StartGRPCServer(address string) error {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if p.GRPCServer != nil {
		return fmt.Errorf("%s gRPC server already started at %s", p.PluginID, p.ServiceAddress)
	}

	if p.Logger != nil {
		p.Logger.InfoTag("gRPC", fmt.Sprintf("启动%s插件gRPC服务器", p.PluginID),
			"address", address)
	}

	// 创建gRPC服务器
	p.GRPCServer = NewGRPCServer(address, p.Logger)

	// 创建gRPC服务实例
	service := p.ServiceFactory()

	// 注册服务
	p.GRPCServer.RegisterPluginService(service)

	// 启用反射（用于调试）
	p.GRPCServer.EnableReflection()

	// 启动服务器
	go func() {
		if err := p.GRPCServer.Start(); err != nil {
			if p.Logger != nil {
				p.Logger.ErrorTag("gRPC", fmt.Sprintf("%s gRPC服务器启动失败", p.PluginID),
					"address", address,
					"error", err.Error())
			}
		} else {
			p.Mu.Lock()
			p.ServiceAddress = address
			p.Mu.Unlock()
			if p.Logger != nil {
				p.Logger.InfoTag("gRPC", fmt.Sprintf("%s插件gRPC服务器启动成功", p.PluginID),
					"address", address)
			}
		}
	}()

	return nil
}

// StopGRPCServer 停止gRPC服务器
func (p *BaseGRPCProvider) StopGRPCServer() error {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if p.GRPCServer == nil {
		return fmt.Errorf("%s gRPC server not started", p.PluginID)
	}

	if p.Logger != nil {
		p.Logger.InfoTag("gRPC", fmt.Sprintf("停止%s插件gRPC服务器", p.PluginID),
			"address", p.ServiceAddress)
	}

	p.GRPCServer.Stop()

	p.GRPCServer = nil
	p.ServiceAddress = ""

	return nil
}

// GetServiceAddress 获取gRPC服务地址
func (p *BaseGRPCProvider) GetServiceAddress() string {
	p.Mu.RLock()
	defer p.Mu.RUnlock()
	return p.ServiceAddress
}

// GetPluginID 返回插件ID
func (p *BaseGRPCProvider) GetPluginID() string {
	return p.PluginID
}
