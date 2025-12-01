package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/protobuf/types/known/structpb"

	pluginv1 "github.com/kalicyh/xiaozhi-flow/api/v1"
	sdk "github.com/kalicyh/xiaozhi-flow/internal/plugin/sdk"
)

// HelloPlugin 示例Hello插件
type HelloPlugin struct {
	sdk.BasePluginImpl
	logger hclog.Logger
}

// NewHelloPlugin 创建Hello插件
func NewHelloPlugin(logger hclog.Logger) *HelloPlugin {
	info := &pluginv1.PluginInfo{
		Id:          "hello-plugin",
		Name:        "Hello Plugin",
		Version:     "1.0.0",
		Description: "A simple hello world plugin for demonstration",
		Author:      "XiaoZhi Flow Team",
		Type:        pluginv1.PluginType_PLUGIN_TYPE_UTILITY,
		Tags:        []string{"example", "hello", "utility"},
		Capabilities: []string{"greet", "echo", "time"},
		Metadata: map[string]interface{}{
			"language":    "go",
			"created_at":  time.Now().Format(time.RFC3339),
			"maintainer":  "xiaozhi-flow",
		},
	}

	return &HelloPlugin{
		BasePluginImpl: *sdk.NewBasePlugin(info, logger),
		logger:         logger.Named("hello-plugin"),
	}
}

// Execute 执行插件功能
func (p *HelloPlugin) Execute(ctx context.Context, method string, params map[string]interface{}) (map[string]interface{}, error) {
	p.logger.Info("Executing plugin method", "method", method, "params", params)

	switch method {
	case "greet":
		return p.greet(ctx, params)
	case "echo":
		return p.echo(ctx, params)
	case "time":
		return p.getCurrentTime(ctx, params)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// greet 问候方法
func (p *HelloPlugin) greet(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		name = "World"
	}

	message := fmt.Sprintf("Hello, %s! From Hello Plugin v1.0.0", name)

	p.IncrementCounter("greet.total")
	p.SetGauge("last_greet_time", float64(time.Now().Unix()))

	return map[string]interface{}{
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
		"source":    "hello-plugin",
	}, nil
}

// echo 回显方法
func (p *HelloPlugin) echo(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	text, ok := params["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text parameter is required")
	}

	p.IncrementCounter("echo.total")
	p.RecordHistogram("echo_length", float64(len(text)))

	return map[string]interface{}{
		"echo":      text,
		"length":    len(text),
		"timestamp": time.Now().Format(time.RFC3339),
		"source":    "hello-plugin",
	}, nil
}

// getCurrentTime 获取当前时间
func (p *HelloPlugin) getCurrentTime(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	format, ok := params["format"].(string)
	if !ok {
		format = time.RFC3339
	}

	now := time.Now()
	var timeStr string
	switch format {
	case "unix":
		timeStr = fmt.Sprintf("%d", now.Unix())
	case "rfc3339":
		timeStr = now.Format(time.RFC3339)
	default:
		timeStr = now.Format(format)
	}

	p.IncrementCounter("time.total")

	return map[string]interface{}{
		"time":      timeStr,
		"format":    format,
		"timestamp": now.Format(time.RFC3339),
		"source":    "hello-plugin",
	}, nil
}


// CallTool 实现通用工具调用接口
func (p *HelloPlugin) CallTool(ctx context.Context, req *pluginv1.CallToolRequest) (*pluginv1.CallToolResponse, error) {
	p.logger.Info("CallTool called", "tool", req.ToolName, "args", req.Arguments)

	// 将Arguments转换为map
	var args map[string]interface{}
	if req.Arguments != nil && req.Arguments.Fields != nil {
		args = req.Arguments.AsMap()
	}

	switch req.ToolName {
	case "greet":
		result, err := p.greet(ctx, args)
		if err != nil {
			return &pluginv1.CallToolResponse{
				Success: false,
				Error: &pluginv1.ErrorInfo{
					Code:    "GREET_ERROR",
					Message: err.Error(),
				},
			}, nil
		}
		return &pluginv1.CallToolResponse{
			Success: true,
			Result:  structToProtoStruct(result),
			Output:  result["message"].(string),
		}, nil

	case "echo":
		result, err := p.echo(ctx, args)
		if err != nil {
			return &pluginv1.CallToolResponse{
				Success: false,
				Error: &pluginv1.ErrorInfo{
					Code:    "ECHO_ERROR",
					Message: err.Error(),
				},
			}, nil
		}
		return &pluginv1.CallToolResponse{
			Success: true,
			Result:  structToProtoStruct(result),
			Output:  result["echo"].(string),
		}, nil

	case "get_time":
		result, err := p.getCurrentTime(ctx, args)
		if err != nil {
			return &pluginv1.CallToolResponse{
				Success: false,
				Error: &pluginv1.ErrorInfo{
					Code:    "TIME_ERROR",
					Message: err.Error(),
				},
			}, nil
		}
		return &pluginv1.CallToolResponse{
			Success: true,
			Result:  structToProtoStruct(result),
			Output:  result["time"].(string),
		}, nil

	default:
		return &pluginv1.CallToolResponse{
			Success: false,
			Error: &pluginv1.ErrorInfo{
				Code:    "UNKNOWN_TOOL",
				Message: fmt.Sprintf("Unknown tool: %s", req.ToolName),
			},
		}, nil
	}
}

// ListTools 列出可用工具
func (p *HelloPlugin) ListTools(ctx context.Context, req *pluginv1.ListToolsRequest) (*pluginv1.ListToolsResponse, error) {
	tools := []*pluginv1.ToolInfo{
		{
			Name:        "greet",
			Description: "Greet someone with a personalized message",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "The name to greet",
						"default":     "World",
					},
				},
			},
		},
		{
			Name:        "echo",
			Description: "Echo back the provided text",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "The text to echo back",
					},
				},
				"required": []string{"text"},
			},
		},
		{
			Name:        "get_time",
			Description: "Get the current time",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Time format (unix, rfc3339, or custom)",
						"default":     "rfc3339",
					},
				},
			},
		},
	}

	return &pluginv1.ListToolsResponse{
		Success: true,
		Tools:   tools,
	}, nil
}

// GetToolSchema 获取工具模式
func (p *HelloPlugin) GetToolSchema(ctx context.Context, req *pluginv1.GetToolSchemaRequest) (*pluginv1.GetToolSchemaResponse, error) {
	listResp, err := p.ListTools(ctx, &pluginv1.ListToolsRequest{})
	if err != nil {
		return &pluginv1.GetToolSchemaResponse{
			Success: false,
			Error: &pluginv1.ErrorInfo{
				Code:    "LIST_TOOLS_ERROR",
				Message: err.Error(),
			},
		}, nil
	}

	for _, tool := range listResp.Tools {
		if tool.Name == req.ToolName {
			return &pluginv1.GetToolSchemaResponse{
				Success: true,
				Schema:  mapToProtoStruct(tool.InputSchema),
			}, nil
		}
	}

	return &pluginv1.GetToolSchemaResponse{
		Success: false,
		Error: &pluginv1.ErrorInfo{
			Code:    "TOOL_NOT_FOUND",
			Message: fmt.Sprintf("tool not found: %s", req.ToolName),
		},
	}, nil
}

func main() {
	// 创建日志记录器
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "hello-plugin",
		Level:  hclog.Info,
		Output: os.Stderr,
	})

	// 创建插件实例
	plugin := NewHelloPlugin(logger)

	logger.Info("Starting Hello Plugin")

	// 服务插件
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"plugin":  &sdk.PluginPlugin{Impl: plugin},
			"utility": &sdk.UtilityPluginPlugin{Impl: plugin},
		},
	})
}

// 辅助函数

// structToProtoStruct 将Go map转换为protobuf Struct
func structToProtoStruct(data map[string]interface{}) *structpb.Struct {
	if data == nil {
		return nil
	}

	result := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}

	for key, value := range data {
		result.Fields[key] = interfaceToProtoValue(value)
	}

	return result
}

// mapToProtoStruct 将Go map转换为protobuf Struct（别名）
func mapToProtoStruct(data map[string]interface{}) *structpb.Struct {
	return structToProtoStruct(data)
}

// interfaceToProtoStruct 将interface{}转换为protobuf Value
func interfaceToProtoValue(v interface{}) *structpb.Value {
	if v == nil {
		return &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: structpb.NullValue_NULL_VALUE}}
	}

	switch val := v.(type) {
	case string:
		return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: val}}
	case int:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(val)}}
	case int32:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(val)}}
	case int64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(val)}}
	case float32:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(val)}}
	case float64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: val}}
	case bool:
		return &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: val}}
	case map[string]interface{}:
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: structToProtoStruct(val)}}
	case []interface{}:
		list := make([]*structpb.Value, len(val))
		for i, item := range val {
			list[i] = interfaceToProtoValue(item)
		}
		return &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: list}}}
	default:
		// 默认转换为字符串
		return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: fmt.Sprintf("%v", v)}}
	}
}