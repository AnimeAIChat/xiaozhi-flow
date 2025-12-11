package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"xiaozhi-server-go/internal/platform/logging"
	pluginpb "xiaozhi-server-go/gen/go/api/proto"
)

// GRPCServer gRPC服务器封装
type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	address  string
	logger   *logging.Logger
}

// NewGRPCServer 创建新的gRPC服务器
func NewGRPCServer(address string, logger *logging.Logger) *GRPCServer {
	if logger == nil {
		logger = logging.DefaultLogger
	}

	return &GRPCServer{
		address: address,
		logger:  logger,
	}
}

// RegisterPluginService 注册插件服务
func (s *GRPCServer) RegisterPluginService(service pluginpb.PluginServiceServer) {
	if s.server == nil {
		s.server = grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				// 可以在这里添加拦截器
			),
			grpc.ChainStreamInterceptor(
				// 可以在这里添加流式拦截器
			),
		)
	}

	pluginpb.RegisterPluginServiceServer(s.server, service)
}

// Start 启动gRPC服务器
func (s *GRPCServer) Start() error {
	var err error

	// 创建监听器
	s.listener, err = net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.address, err)
	}

	if s.logger != nil {
		s.logger.InfoTag("gRPC", "gRPC服务器启动",
			"address", s.address)
	}

	// 启动服务器
	return s.server.Serve(s.listener)
}

// Stop 停止gRPC服务器
func (s *GRPCServer) Stop() {
	if s.server != nil {
		if s.logger != nil {
			s.logger.InfoTag("gRPC", "正在停止gRPC服务器",
				"address", s.address)
		}

		// 优雅关闭
		stopped := make(chan struct{})
		go func() {
			s.server.GracefulStop()
			close(stopped)
		}()

		// 等待停止或超时
		select {
		case <-stopped:
			if s.logger != nil {
				s.logger.InfoTag("gRPC", "gRPC服务器已优雅停止")
			}
		case <-time.After(30 * time.Second):
			if s.logger != nil {
				s.logger.WarnTag("gRPC", "gRPC服务器停止超时，强制停止")
			}
			s.server.Stop()
		}
	}

	if s.listener != nil {
		s.listener.Close()
	}
}

// GetAddress 获取服务器地址
func (s *GRPCServer) GetAddress() string {
	return s.address
}

// IsRunning 检查服务器是否正在运行
func (s *GRPCServer) IsRunning() bool {
	return s.server != nil && s.listener != nil
}

// EnableReflection 启用gRPC反射（用于调试）
func (s *GRPCServer) EnableReflection() {
	if s.server != nil {
		reflection.Register(s.server)
		if s.logger != nil {
			s.logger.InfoTag("gRPC", "已启用gRPC反射")
		}
	}
}

// PluginServerBase 插件服务器基础实现
type PluginServerBase struct {
	pluginpb.UnimplementedPluginServiceServer
	logger *logging.Logger
}

// NewPluginServerBase 创建插件服务器基础实现
func NewPluginServerBase(logger *logging.Logger) *PluginServerBase {
	if logger == nil {
		logger = logging.DefaultLogger
	}

	return &PluginServerBase{
		logger: logger,
	}
}

// GetPluginInfo 获取插件信息（基础实现，需要被具体插件重写）
func (s *PluginServerBase) GetPluginInfo(ctx context.Context, req *pluginpb.GetPluginInfoRequest) (*pluginpb.GetPluginInfoResponse, error) {
	s.logger.InfoTag("gRPC", "GetPluginInfo被调用",
		"plugin_id", req.PluginId)

	// 基础实现，返回空信息
	// 具体插件应该重写这个方法
	return &pluginpb.GetPluginInfoResponse{
		PluginInfo: &pluginpb.PluginInfo{
			Id:          req.PluginId,
			Name:        req.PluginId,
			Type:        "unknown",
			Description: "Plugin base implementation",
			Version:     "1.0.0",
			Status:      "unknown",
		},
		Capabilities: []*pluginpb.CapabilityDefinition{},
	}, nil
}

// ExecuteCapability 执行插件能力（基础实现）
func (s *PluginServerBase) ExecuteCapability(ctx context.Context, req *pluginpb.ExecuteCapabilityRequest) (*pluginpb.ExecuteCapabilityResponse, error) {
	s.logger.InfoTag("gRPC", "ExecuteCapability被调用",
		"capability_id", req.CapabilityId)

	// 基础实现，返回错误
	// 具体插件应该重写这个方法
	return &pluginpb.ExecuteCapabilityResponse{
		Success:     false,
		Outputs:     nil,
		ErrorMessage: "ExecuteCapability not implemented in base class",
		StreamFinished: true,
	}, nil
}

// ExecuteCapabilityStream 流式执行插件能力（基础实现）
func (s *PluginServerBase) ExecuteCapabilityStream(req *pluginpb.ExecuteCapabilityRequest, stream pluginpb.PluginService_ExecuteCapabilityStreamServer) error {
	s.logger.InfoTag("gRPC", "ExecuteCapabilityStream被调用",
		"capability_id", req.CapabilityId)

	// 基础实现，返回错误
	// 具体插件应该重写这个方法
	return stream.Send(&pluginpb.ExecuteCapabilityResponse{
		Success:     false,
		Outputs:     nil,
		ErrorMessage: "ExecuteCapabilityStream not implemented in base class",
		StreamFinished: true,
	})
}

// HealthCheck 健康检查
func (s *PluginServerBase) HealthCheck(ctx context.Context, req *pluginpb.HealthCheckRequest) (*pluginpb.HealthCheckResponse, error) {
	s.logger.DebugTag("gRPC", "HealthCheck被调用",
		"plugin_id", req.PluginId)

	return &pluginpb.HealthCheckResponse{
		Status: "healthy",
		Message: "Plugin is running",
		Details: map[string]string{
			"version": "1.0.0",
		},
	}, nil
}

// CreateGRPCClient 创建gRPC客户端连接
func CreateGRPCClient(address string) (pluginpb.PluginServiceClient, *grpc.ClientConn, error) {
	// 创建连接
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", address, err)
	}

	// 创建客户端
	client := pluginpb.NewPluginServiceClient(conn)

	return client, conn, nil
}