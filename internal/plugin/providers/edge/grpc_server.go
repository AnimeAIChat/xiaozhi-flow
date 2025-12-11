package edge

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"xiaozhi-server-go/internal/plugin/grpc/server"
	pluginpb "xiaozhi-server-go/gen/go/api/proto"
	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/plugin/capability"
)

// GRPCServer Edge插件的gRPC服务实现
type GRPCServer struct {
	*server.PluginServerBase
	provider *Provider
	logger   *logging.Logger
}

// NewGRPCServer 创建Edge gRPC服务器
func NewGRPCServer(provider *Provider, logger *logging.Logger) *GRPCServer {
	return &GRPCServer{
		PluginServerBase: server.NewPluginServerBase(logger),
		provider:        provider,
		logger:          logger,
	}
}

// GetPluginInfo 获取Edge插件信息
func (s *GRPCServer) GetPluginInfo(ctx context.Context, req *pluginpb.GetPluginInfoRequest) (*pluginpb.GetPluginInfoResponse, error) {
	if s.logger != nil {
		s.logger.InfoTag("gRPC", "获取Edge插件信息",
			"plugin_id", req.PluginId)
	}

	capabilities := s.provider.GetCapabilities()
	pbCapabilities := make([]*pluginpb.CapabilityDefinition, len(capabilities))

	for i, cap := range capabilities {
		pbCapabilities[i] = &pluginpb.CapabilityDefinition{
			Id:          cap.ID,
			Type:        string(cap.Type),
			Name:        cap.Name,
			Description: cap.Description,
			ConfigSchema: convertSchemaToPB(cap.ConfigSchema),
			InputSchema:  convertSchemaToPB(cap.InputSchema),
			OutputSchema: convertSchemaToPB(cap.OutputSchema),
			Enabled:     true,
		}
	}

	pluginInfo := &pluginpb.PluginInfo{
		Id:          "edge",
		Name:        "Microsoft Edge TTS",
		Type:        "TTS",
		Description: "Microsoft Edge Text to Speech服务提供商",
		Version:     "1.0.0",
		Status:      "active",
		UpdatedAt:   timestamppb.Now(),
	}

	return &pluginpb.GetPluginInfoResponse{
		PluginInfo:  pluginInfo,
		Capabilities: pbCapabilities,
	}, nil
}

// ExecuteCapability 执行Edge插件能力
func (s *GRPCServer) ExecuteCapability(ctx context.Context, req *pluginpb.ExecuteCapabilityRequest) (*pluginpb.ExecuteCapabilityResponse, error) {
	if s.logger != nil {
		s.logger.InfoTag("gRPC", "执行Edge插件能力",
			"capability_id", req.CapabilityId)
	}

	// 创建执行器
	executor, err := s.provider.CreateExecutor(req.CapabilityId)
	if err != nil {
		return &pluginpb.ExecuteCapabilityResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("创建执行器失败: %v", err),
			StreamFinished: true,
		}, nil
	}

	// 转换配置和输入
	config := convertPBToMap(req.Config)
	inputs := convertPBToMap(req.Inputs)

	// Edge TTS不支持流式执行，使用非流式执行
	outputs, err := executor.Execute(ctx, config, inputs)
	if err != nil {
		return &pluginpb.ExecuteCapabilityResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("执行失败: %v", err),
			StreamFinished: true,
		}, nil
	}

	return &pluginpb.ExecuteCapabilityResponse{
		Success:        true,
		Outputs:        convertMapToPB(outputs),
		StreamFinished: true,
	}, nil
}

// ExecuteCapabilityStream 流式执行Edge插件能力
func (s *GRPCServer) ExecuteCapabilityStream(req *pluginpb.ExecuteCapabilityRequest, stream pluginpb.PluginService_ExecuteCapabilityStreamServer) error {
	if s.logger != nil {
		s.logger.InfoTag("gRPC", "流式执行Edge插件能力",
			"capability_id", req.CapabilityId)
	}

	// Edge TTS不支持流式执行，返回错误
	return stream.Send(&pluginpb.ExecuteCapabilityResponse{
		Success:        false,
		ErrorMessage:   "Edge TTS不支持流式执行",
		StreamFinished: true,
	})
}

// HealthCheck Edge插件健康检查
func (s *GRPCServer) HealthCheck(ctx context.Context, req *pluginpb.HealthCheckRequest) (*pluginpb.HealthCheckResponse, error) {
	if s.logger != nil {
		s.logger.DebugTag("gRPC", "Edge插件健康检查")
	}

	// 检查插件是否能正常创建执行器
	_, err := s.provider.CreateExecutor("edge_tts")
	if err != nil {
		return &pluginpb.HealthCheckResponse{
			Status:  "unhealthy",
			Message: fmt.Sprintf("无法创建TTS执行器: %v", err),
			Details: map[string]string{
				"error": err.Error(),
			},
		}, nil
	}

	return &pluginpb.HealthCheckResponse{
		Status:  "healthy",
		Message: "Edge插件运行正常",
		Details: map[string]string{
			"version":     "1.0.0",
			"capabilities": "edge_tts",
		},
	}, nil
}

// 辅助函数

// convertSchemaToPB 转换Schema到protobuf格式
func convertSchemaToPB(schema capability.Schema) *structpb.Struct {
	if len(schema.Properties) == 0 && len(schema.Required) == 0 {
		return nil
	}

	result := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}

	if schema.Type != "" {
		result.Fields["type"] = structpb.NewStringValue(schema.Type)
	}

	if len(schema.Properties) > 0 {
		properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
		for key, prop := range schema.Properties {
			propStruct := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
			propStruct.Fields["type"] = structpb.NewStringValue(prop.Type)
			if prop.Description != "" {
				propStruct.Fields["description"] = structpb.NewStringValue(prop.Description)
			}
			if prop.Default != nil {
				propStruct.Fields["default"] = convertInterfaceToPB(prop.Default)
			}
			if prop.Secret {
				propStruct.Fields["secret"] = structpb.NewBoolValue(prop.Secret)
			}
			properties.Fields[key] = structpb.NewStructValue(propStruct)
		}
		result.Fields["properties"] = structpb.NewStructValue(properties)
	}

	if len(schema.Required) > 0 {
		required := &structpb.ListValue{Values: make([]*structpb.Value, len(schema.Required))}
		for i, req := range schema.Required {
			required.Values[i] = structpb.NewStringValue(req)
		}
		result.Fields["required"] = structpb.NewListValue(required)
	}

	return result
}

// convertPBToMap 转换protobuf结构到map
func convertPBToMap(pb *structpb.Struct) map[string]interface{} {
	if pb == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})
	for key, value := range pb.Fields {
		result[key] = convertPBToInterface(value)
	}

	return result
}

// convertMapToPB 转换map到protobuf结构
func convertMapToPB(m map[string]interface{}) *structpb.Struct {
	if m == nil {
		return nil
	}

	fields := make(map[string]*structpb.Value)
	for key, value := range m {
		fields[key] = convertInterfaceToPB(value)
	}

	return &structpb.Struct{Fields: fields}
}

// convertPBToInterface 转换protobuf值到Go接口
func convertPBToInterface(v *structpb.Value) interface{} {
	switch v.Kind.(type) {
	case *structpb.Value_NullValue:
		return nil
	case *structpb.Value_NumberValue:
		return v.GetNumberValue()
	case *structpb.Value_StringValue:
		return v.GetStringValue()
	case *structpb.Value_BoolValue:
		return v.GetBoolValue()
	case *structpb.Value_StructValue:
		return convertPBToMap(v.GetStructValue())
	case *structpb.Value_ListValue:
		list := v.GetListValue()
		result := make([]interface{}, len(list.Values))
		for i, item := range list.Values {
			result[i] = convertPBToInterface(item)
		}
		return result
	default:
		return nil
	}
}

// convertInterfaceToPB 转换Go接口到protobuf值
func convertInterfaceToPB(v interface{}) *structpb.Value {
	switch val := v.(type) {
	case nil:
		return structpb.NewNullValue()
	case bool:
		return structpb.NewBoolValue(val)
	case int32:
		return structpb.NewNumberValue(float64(val))
	case int64:
		return structpb.NewNumberValue(float64(val))
	case float32:
		return structpb.NewNumberValue(float64(val))
	case float64:
		return structpb.NewNumberValue(val)
	case string:
		return structpb.NewStringValue(val)
	case map[string]interface{}:
		return structpb.NewStructValue(convertMapToPB(val))
	case []interface{}:
		list := &structpb.ListValue{}
		for _, item := range val {
			list.Values = append(list.Values, convertInterfaceToPB(item))
		}
		return structpb.NewListValue(list)
	default:
		return structpb.NewStringValue(fmt.Sprintf("%v", val))
	}
}