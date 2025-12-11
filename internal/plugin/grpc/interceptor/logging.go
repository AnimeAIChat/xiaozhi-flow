package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"xiaozhi-server-go/internal/platform/logging"
)

// LoggingInterceptor 创建gRPC日志拦截器
func LoggingInterceptor(logger *logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// 记录请求开始
		logger.InfoTag("gRPC", "开始处理请求",
			"method", info.FullMethod,
			"request_id", getRequestIDFromContext(ctx))

		// 调用处理器
		resp, err := handler(ctx, req)

		// 记录请求完成
		duration := time.Since(start)
		if err != nil {
			logger.ErrorTag("gRPC", "请求处理失败",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"error", err.Error(),
				"request_id", getRequestIDFromContext(ctx))
		} else {
			logger.InfoTag("gRPC", "请求处理成功",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"request_id", getRequestIDFromContext(ctx))
		}

		return resp, err
	}
}

// StreamLoggingInterceptor 创建gRPC流式日志拦截器
func StreamLoggingInterceptor(logger *logging.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// 记录流开始
		logger.InfoTag("gRPC", "开始处理流请求",
			"method", info.FullMethod,
			"request_id", getRequestIDFromContext(ss.Context()))

		// 调用处理器
		err := handler(srv, ss)

		// 记录流完成
		duration := time.Since(start)
		if err != nil {
			logger.ErrorTag("gRPC", "流请求处理失败",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"error", err.Error(),
				"request_id", getRequestIDFromContext(ss.Context()))
		} else {
			logger.InfoTag("gRPC", "流请求处理成功",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"request_id", getRequestIDFromContext(ss.Context()))
		}

		return err
	}
}

// getRequestIDFromContext 从上下文中获取请求ID
func getRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}