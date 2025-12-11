package interceptor

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"xiaozhi-server-go/internal/platform/logging"
)

// AuthInterceptor 创建gRPC认证拦截器
func AuthInterceptor(logger *logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 检查是否需要认证的方法
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// 从metadata中获取认证信息
		md, ok := grpc.ServerContextFromTransportStream(ctx)
		if !ok {
			logger.WarnTag("gRPC", "认证失败：无法获取transport context",
				"method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "认证失败")
		}

		// 获取认证token
		token := md.Value()[string("authorization")]
		if len(token) == 0 || token[0] == "" {
			logger.WarnTag("gRPC", "认证失败：缺少认证token",
				"method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "缺少认证token")
		}

		// 验证token（这里简化处理，实际应该验证token的有效性）
		if !validateToken(token[0]) {
			logger.WarnTag("gRPC", "认证失败：无效的token",
				"method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "无效的token")
		}

		logger.DebugTag("gRPC", "认证成功",
			"method", info.FullMethod)

		return handler(ctx, req)
	}
}

// StreamAuthInterceptor 创建gRPC流式认证拦截器
func StreamAuthInterceptor(logger *logging.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 检查是否需要认证的方法
		if isPublicMethod(info.FullMethod) {
			return handler(srv, ss)
		}

		// 从metadata中获取认证信息
		md, ok := ss.Context().Value("metadata").(map[string][]string)
		if !ok {
			logger.WarnTag("gRPC", "流认证失败：无法获取metadata",
				"method", info.FullMethod)
			return status.Error(codes.Unauthenticated, "认证失败")
		}

		// 获取认证token
		token := md["authorization"]
		if len(token) == 0 || token[0] == "" {
			logger.WarnTag("gRPC", "流认证失败：缺少认证token",
				"method", info.FullMethod)
			return status.Error(codes.Unauthenticated, "缺少认证token")
		}

		// 验证token
		if !validateToken(token[0]) {
			logger.WarnTag("gRPC", "流认证失败：无效的token",
				"method", info.FullMethod)
			return status.Error(codes.Unauthenticated, "无效的token")
		}

		logger.DebugTag("gRPC", "流认证成功",
			"method", info.FullMethod)

		return handler(srv, ss)
	}
}

// isPublicMethod 检查是否为公共方法（不需要认证）
func isPublicMethod(method string) bool {
	publicMethods := []string{
		"/plugin.PluginService/GetPluginInfo",
		"/plugin.PluginService/HealthCheck",
	}

	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}
	return false
}

// validateToken 验证token（简化实现）
func validateToken(token string) bool {
	// 这里应该实现真正的token验证逻辑
	// 目前简化处理，非空token即认为有效
	return token != ""
}

// CorsInterceptor 创建CORS拦截器（用于gRPC-Web）
func CorsInterceptor(logger *logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 设置CORS headers
		if headers, ok := ctx.Value("headers").(map[string][]string); ok {
			headers["Access-Control-Allow-Origin"] = []string{"*"}
			headers["Access-Control-Allow-Methods"] = []string{"POST", "GET", "OPTIONS"}
			headers["Access-Control-Allow-Headers"] = []string{"Content-Type", "Authorization"}
		}

		return handler(ctx, req)
	}
}

// RecoveryInterceptor 创建恢复拦截器
func RecoveryInterceptor(logger *logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.ErrorTag("gRPC", "gRPC服务发生panic",
					"method", info.FullMethod,
					"panic", r)

				err = status.Error(codes.Internal, "服务器内部错误")
			}
		}()

		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor 创建流式恢复拦截器
func StreamRecoveryInterceptor(logger *logging.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.ErrorTag("gRPC", "gRPC流服务发生panic",
					"method", info.FullMethod,
					"panic", r)

				err = status.Error(codes.Internal, "服务器内部错误")
			}
		}()

		return handler(srv, ss)
	}
}