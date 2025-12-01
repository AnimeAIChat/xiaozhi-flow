# XiaoZhi Flow 插件开发指南

## 📖 目录

- [概述](#概述)
- [快速开始](#快速开始)
- [插件架构](#插件架构)
- [开发环境搭建](#开发环境搭建)
- [插件开发流程](#插件开发流程)
- [插件类型](#插件类型)
- [API 参考](#api-参考)
- [示例插件](#示例插件)
- [最佳实践](#最佳实践)
- [调试和测试](#调试和测试)
- [部署和分发](#部署和分发)
- [常见问题](#常见问题)

## 📖 概述

XiaoZhi Flow 插件系统基于 HashiCorp go-plugin 框架，支持通过 gRPC 通信的进程级扩展。插件系统设计为：

- **安全隔离**：插件运行在独立的进程中
- **类型安全**：基于 Go 接口的强类型定义
- **易于开发**：简单的 SDK 和丰富的工具
- **高性能**：基于 gRPC 的高效通信
- **热插拔**：支持运行时加载和卸载

### 插件系统能力

- 🎵 **音频处理插件**：ASR（语音识别）、TTS（语音合成）、VAD（语音活动检测）
- 🤖 **大模型插件**：文本生成、对话、嵌入向量生成
- 🔌 **设备控制插件**：IoT设备控制、传感器数据采集
- 🛠️ **通用功能插件**：文件操作、网络请求、数据处理

## 🚀 快速开始

### 1. 创建第一个插件

```bash
# 创建插件目录
mkdir plugins/my-first-plugin
cd plugins/my-first-plugin

# 创建插件主文件
touch main.go
```

### 2. 基础插件代码

```go
package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	v1 "xiaozhi-server-go/api/v1"
	sdk "xiaozhi-server-go/internal/plugin/sdk"
)

// MyPlugin 我的第一个插件
type MyPlugin struct {
	sdk.SimplePluginImpl
	logger hclog.Logger
}

// NewMyPlugin 创建插件实例
func NewMyPlugin(logger hclog.Logger) *MyPlugin {
	info := &v1.PluginInfo{
		ID:          "my-first-plugin",
		Name:        "My First Plugin",
		Version:     "1.0.0",
		Description: "我的第一个XiaoZhi Flow插件",
		Author:      "Your Name",
		Type:        v1.PluginTypeUtility,
		Tags:        []string{"example", "utility"},
		Capabilities: []string{"hello", "math"},
		Metadata: map[string]interface{}{
			"language": "go",
			"created":  "2024-01-01",
		},
	}

	return &MyPlugin{
		SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
		logger:          logger.Named("my-plugin"),
	}
}

// CallTool 实现工具调用
func (p *MyPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
	p.logger.Info("Tool called", "tool", req.ToolName, "args", req.Arguments)

	switch req.ToolName {
	case "hello":
		return p.hello(ctx, req.Arguments)
	case "add":
		return p.add(ctx, req.Arguments)
	default:
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "UNKNOWN_TOOL",
				Message: fmt.Sprintf("未知工具: %s", req.ToolName),
			},
		}
	}
}

// hello 工具实现
func (p *MyPlugin) hello(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
	name, ok := args["name"].(string)
	if !ok {
		name = "World"
	}

	message := fmt.Sprintf("Hello, %s! from My Plugin", name)

	return &v1.CallToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"message":   message,
			"timestamp": fmt.Sprintf("%v", ctx.Value("timestamp")),
		},
		Output: message,
	}
}

// add 数学加法工具
func (p *MyPlugin) add(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
	a, aOk := args["a"].(float64)
	b, bOk := args["b"].(float64)

	if !aOk || !bOk {
		return &v1.CallToolResponse{
			Success: false,
			Error: &v1.ErrorInfo{
				Code:    "INVALID_ARGS",
				Message: "参数 a 和 b 必须是数字",
			},
		}
	}

	result := a + b

	return &v1.CallToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"a":      a,
			"b":      b,
			"result": result,
		},
		Output: fmt.Sprintf("%f + %f = %f", a, b, result),
	}
}

// ListTools 列出可用工具
func (p *MyPlugin) ListTools(ctx context.Context) *v1.ListToolsResponse {
	tools := []*v1.ToolInfo{
		{
			Name:        "hello",
			Description: "向世界问好",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "要问候的名字",
						"default":     "World",
					},
				},
			},
		},
		{
			Name:        "add",
			Description: "数学加法计算",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type":        "number",
						"description": "第一个数字",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "第二个数字",
					},
				},
				"required": []string{"a", "b"},
			},
		},
	}

	return &v1.ListToolsResponse{
		Success: true,
		Tools:   tools,
	}
}

// GetToolSchema 获取工具模式
func (p *MyPlugin) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
	listResp := p.ListTools(ctx)
	if !listResp.Success {
		return &v1.GetToolSchemaResponse{
			Success: false,
			Error:   listResp.Error,
		}
	}

	for _, tool := range listResp.Tools {
		if tool.Name == req.ToolName {
			return &v1.GetToolSchemaResponse{
				Success: true,
				Schema:  tool.InputSchema,
			}
		}
	}

	return &v1.GetToolSchemaResponse{
		Success: false,
		Error: &v1.ErrorInfo{
			Code:    "TOOL_NOT_FOUND",
			Message: fmt.Sprintf("工具 %s 未找到", req.ToolName),
		},
	}
}

func main() {
	// 创建日志记录器
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "my-first-plugin",
		Level:  hclog.Info,
		Output: hclog.DefaultOutput,
	})

	// 创建插件实例
	plugin := NewMyPlugin(logger)

	logger.Info("Starting My First Plugin")

	// 服务插件
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.SimpleHandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"plugin": &sdk.SimplePluginRPC{Impl: plugin},
		},
	})
}
```

### 3. 创建插件配置文件

```yaml
# plugins/my-first-plugin/plugin.yaml
name: My First Plugin
version: 1.0.0
description: 我的第一个XiaoZhi Flow插件
author: Your Name
type: utility
tags:
  - example
  - utility
  - math
capabilities:
  - hello
  - add
metadata:
  language: go
  created_at: "2024-01-01T00:00:00Z"

deployment:
  type: local_binary
  path: ./plugins/my-first-plugin/main.go
  resources:
    max_memory: "64Mi"
    max_cpu: "50m"
  timeout: 10s
  retry_count: 3

config:
  greeting_language: "zh-CN"
  math_precision: 2

environment:
  PLUGIN_LOG_LEVEL: "info"
  PLUGIN_DEBUG: "false"

enabled: true
```

### 4. 编译和测试

```bash
# 编译插件
cd plugins/my-first-plugin
go build -o my-first-plugin main.go

# 测试插件
./my-first-plugin

# 或者在主系统中测试
cd ../../
go run cmd/xiaozhi-server/main.go
```

## 🏗️ 插件架构

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    XiaoZhi Flow 主系统                        │
├─────────────────────────────────────────────────────────────┤
│  统一网关 (Unified Gateway)                                    │
│  ├── 插件路由 (Plugin Router)                                  │
│  ├── MCP 路由 (MCP Router)                                   │
│  └── Provider 路由 (Provider Router)                         │
├─────────────────────────────────────────────────────────────┤
│  插件管理器 (Plugin Manager)                                   │
│  ├── 插件发现 (Plugin Discovery)                              │
│  ├── 插件注册表 (Plugin Registry)                              │
│  ├── 生命周期管理 (Lifecycle Management)                       │
│  └── 健康检查 (Health Check)                                  │
├─────────────────────────────────────────────────────────────┤
│  gRPC 通信层 (gRPC Communication Layer)                        │
│  ├── 插件服务 (Plugin Services)                                │
│  ├── 指标收集 (Metrics Collection)                              │
│  └── 错误处理 (Error Handling)                                 │
├─────────────────────────────────────────────────────────────┤
│  插件运行时 (Plugin Runtime)                                   │
│  ├── 本地二进制 (Local Binary)                                 │
│  ├── 容器化 (Container)                                        │
│  └── 远程服务 (Remote Service)                                 │
└─────────────────────────────────────────────────────────────┘
```

### 插件生命周期

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   发现插件   │ -> │   加载插件   │ -> │   初始化插件  │ -> │   运行插件   │
│ Discovery   │    │   Load      │    │ Initialize  │    │   Running   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                           │
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    │
│   清理资源   │ <- │   卸载插件   │ <- │   健康检查   │ <- │
│  Cleanup    │    │  Unload     │    │ HealthCheck  │    │
└─────────────┘    └─────────────┘    └─────────────┘    │
                                                   └─────────┘
```

## 🛠️ 开发环境搭建

### 系统要求

- Go 1.24+
- Git
- 基础的 Go 开发工具

### 开发工具安装

```bash
# 安装开发工具
go install github.com/air-verse/air@latest
go install github.com/swaggo/swag/cmd/swag@latest

# 验证安装
go version
air -v
swag -v
```

### 项目结构

```
plugins/
├── examples/              # 示例插件
│   ├── hello/            # Hello World 插件
│   ├── calculator/       # 计算器插件
│   └── weather/          # 天气插件
├── audio/                 # 音频处理插件
│   ├── tts/              # 文字转语音
│   └── asr/              # 语音识别
├── llm/                   # 大模型插件
│   ├── openai/           # OpenAI 集成
│   └── ollama/           # 本地模型
├── device/                # 设备控制插件
│   ├── esp32/            # ESP32 控制
│   └── sensors/          # 传感器数据
└── utility/               # 通用功能插件
    ├── file/             # 文件操作
    └── network/          # 网络请求
```

### 插件模板

创建新插件时，可以使用以下模板结构：

```
my-plugin/
├── main.go                # 插件主文件
├── plugin.yaml           # 插件配置
├── go.mod                # Go 模块文件
├── README.md             # 插件说明
├── test/                 # 测试文件
│   └── main_test.go
├── docs/                 # 文档
│   └── api.md
└── assets/               # 静态资源
```

## 🔧 插件开发流程

### 1. 规划插件

- **确定插件类型**：Audio、LLM、Device、Utility
- **定义功能范围**：插件要实现的具体功能
- **设计工具接口**：提供哪些工具给用户使用
- **规划依赖关系**：需要的外部库或服务

### 2. 创建插件项目

```bash
# 使用脚本创建插件模板
./scripts/create-plugin.sh my-plugin --type utility --author "Your Name"

# 或手动创建
mkdir plugins/my-plugin
cd plugins/my-plugin
go mod init my-plugin
```

### 3. 实现插件接口

```go
// 实现 SimplePlugin 接口
type MyPlugin struct {
    sdk.SimplePluginImpl
    // 添加插件特定字段
}

// 必须实现的方法
func (p *MyPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse
func (p *MyPlugin) ListTools(ctx context.Context) *v1.ListToolsResponse
func (p *MyPlugin) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse
```

### 4. 创建插件配置

```yaml
# plugin.yaml
name: My Plugin
version: 1.0.0
type: utility
deployment:
  type: local_binary
  path: ./main.go
enabled: true
```

### 5. 测试插件

```go
// test/main_test.go
func TestMyPlugin(t *testing.T) {
    plugin := NewMyPlugin(hclog.Default())

    // 测试工具调用
    req := &v1.CallToolRequest{
        ToolName: "hello",
        Arguments: map[string]interface{}{
            "name": "Test",
        },
    }

    resp := plugin.CallTool(context.Background(), req)
    assert.True(t, resp.Success)
    assert.Contains(t, resp.Output, "Hello, Test")
}
```

### 6. 构建和部署

```bash
# 构建
go build -o my-plugin main.go

# 测试运行
./my-plugin

# 复制到插件目录
cp my-plugin ../
cp plugin.yaml ../
```

## 📋 插件类型

### 1. Audio Plugin (音频插件)

用于音频处理功能：语音识别(ASR)、语音合成(TTS)、语音活动检测(VAD)等。

```go
type AudioPlugin interface {
    SimplePlugin
    ProcessAudio(ctx context.Context, req *v1.ProcessAudioRequest) *v1.ProcessAudioResponse
    StreamProcessAudio(ctx context.Context, req *v1.StreamProcessAudioRequest) (<-chan *v1.StreamProcessAudioResponse, error)
}
```

**示例场景**：
- 语音识别插件：将语音转换为文本
- 文字转语音插件：将文本转换为语音
- 音频处理插件：降噪、格式转换、音频增强

### 2. LLM Plugin (大模型插件)

用于大语言模型功能：文本生成、对话、嵌入向量生成等。

```go
type LLMPlugin interface {
    SimplePlugin
    GenerateText(ctx context.Context, req *v1.GenerateTextRequest) *v1.GenerateTextResponse
    StreamGenerateText(ctx context.Context, req *v1.StreamGenerateTextRequest) (<-chan *v1.StreamGenerateTextResponse, error)
    GenerateEmbedding(ctx context.Context, req *v1.GenerateEmbeddingRequest) *v1.GenerateEmbeddingResponse
}
```

**示例场景**：
- OpenAI 集成插件：调用 GPT API
- 本地模型插件：集成 Ollama、llama.cpp
- 专用模型插件：如代码生成、翻译等

### 3. Device Plugin (设备控制插件)

用于设备控制和传感器数据采集。

```go
type DevicePlugin interface {
    SimplePlugin
    ControlDevice(ctx context.Context, req *v1.ControlDeviceRequest) *v1.ControlDeviceResponse
    GetDeviceStatus(ctx context.Context, req *v1.GetDeviceStatusRequest) *v1.GetDeviceStatusResponse
    ListDevices(ctx context.Context, req *v1.ListDevicesRequest) *v1.ListDevicesResponse
}
```

**示例场景**：
- ESP32 控制插件：控制 ESP32 设备
- 智能家居插件：控制灯光、空调等
- 传感器插件：读取温度、湿度、光照等

### 4. Utility Plugin (通用功能插件)

用于各种通用功能：文件操作、网络请求、数据处理等。

```go
type UtilityPlugin interface {
    SimplePlugin
    // 可以添加自定义方法
}
```

**示例场景**：
- 文件操作插件：文件读写、格式转换
- 网络请求插件：HTTP 调用、API 集成
- 数据处理插件：数据转换、格式化、验证

## 📚 API 参考

### 核心接口

#### SimplePlugin 接口

```go
type SimplePlugin interface {
    // 生命周期管理
    Initialize(ctx context.Context, config *InitializeConfig) error
    Shutdown(ctx context.Context) error

    // 健康检查
    HealthCheck(ctx context.Context) *v1.HealthStatus

    // 指标收集
    GetMetrics(ctx context.Context) *v1.Metrics

    // 插件信息
    GetInfo() *v1.PluginInfo
    Logger() hclog.Logger

    // 工具调用
    CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse
    ListTools(ctx context.Context) *v1.ListToolsResponse
    GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse
}
```

#### 工具调用请求/响应

```go
type CallToolRequest struct {
    ToolName string                 `json:"tool_name"`
    Arguments map[string]interface{} `json:"arguments"`
    Options   map[string]string      `json:"options"`
}

type CallToolResponse struct {
    Success bool                   `json:"success"`
    Result  map[string]interface{} `json:"result"`
    Output  string                 `json:"output"`
    Error   *ErrorInfo             `json:"error"`
}

type ToolInfo struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"input_schema"`
    Metadata    map[string]string      `json:"metadata"`
}
```

#### 错误处理

```go
type ErrorInfo struct {
    Code    string            `json:"code"`
    Message string            `json:"message"`
    Details string            `json:"details"`
    Context map[string]string `json:"context"`
}
```

#### 指标收集

```go
type Metrics struct {
    Counters   map[string]float64   `json:"counters"`
    Gauges     map[string]float64   `json:"gauges"`
    Histograms map[string]*Histogram `json:"histograms"`
    Timestamp  time.Time            `json:"timestamp"`
}

// 简单插件提供的方法
func (p *SimplePluginImpl) IncrementCounter(name string)
func (p *SimplePluginImpl) SetGauge(name string, value float64)
func (p *SimplePluginImpl) RecordHistogram(name string, value float64)
```

### 配置参数

#### 插件配置结构

```go
type InitializeConfig struct {
    Config      map[string]interface{} `json:"config"`
    Environment map[string]string      `json:"environment"`
}

type PluginConfig struct {
    ID          string                 `yaml:"id"`
    Name        string                 `yaml:"name"`
    Version     string                 `yaml:"version"`
    Description string                 `yaml:"description"`
    Type        string                 `yaml:"type"`
    Deployment  DeploymentConfig       `yaml:"deployment"`
    Config      map[string]interface{} `yaml:"config"`
    Environment map[string]string      `yaml:"environment"`
    Enabled     bool                   `yaml:"enabled"`
}
```

#### 部署配置

```go
type DeploymentConfig struct {
    Type       string            `yaml:"type"`        // local_binary, container, remote_service
    Path       string            `yaml:"path"`        // 二进制路径
    Image      string            `yaml:"image"`      // 容器镜像
    Endpoint   string            `yaml:"endpoint"`   // 远程服务端点
    Resources  ResourceConfig    `yaml:"resources"`
    Timeout    time.Duration     `yaml:"timeout"`
    RetryCount int               `yaml:"retry_count"`
    Options    map[string]string `yaml:"options"`
}
```

## 🎯 示例插件

### 1. 计算器插件

```go
// plugins/utility/calculator/main.go
package main

import (
    "context"
    "fmt"
    "math"

    v1 "xiaozhi-server-go/api/v1"
    sdk "xiaozhi-server-go/internal/plugin/sdk"
)

type CalculatorPlugin struct {
    sdk.SimplePluginImpl
}

func (p *CalculatorPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
    switch req.ToolName {
    case "add":
        return p.calculate(req.Arguments, func(a, b float64) float64 { return a + b })
    case "subtract":
        return p.calculate(req.Arguments, func(a, b float64) float64 { return a - b })
    case "multiply":
        return p.calculate(req.Arguments, func(a, b float64) float64 { return a * b })
    case "divide":
        return p.calculate(req.Arguments, func(a, b float64) float64 { return a / b })
    case "sqrt":
        return p.sqrt(req.Arguments)
    case "pow":
        return p.power(req.Arguments)
    default:
        return unknownTool(req.ToolName)
    }
}

func (p *CalculatorPlugin) calculate(args map[string]interface{}, op func(float64, float64) float64) *v1.CallToolResponse {
    a, aOk := args["a"].(float64)
    b, bOk := args["b"].(float64)

    if !aOk || !bOk {
        return invalidArgs("需要参数 a 和 b")
    }

    result := op(a, b)
    p.IncrementCounter("calculate.total")
    p.RecordHistogram("calculate.result", result)

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "a":      a,
            "b":      b,
            "result": result,
        },
        Output: fmt.Sprintf("%.2f", result),
    }
}

func (p *CalculatorPlugin) sqrt(args map[string]interface{}) *v1.CallToolResponse {
    x, ok := args["x"].(float64)
    if !ok {
        return invalidArgs("需要参数 x")
    }

    if x < 0 {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_INPUT",
                Message: "不能计算负数的平方根",
            },
        }
    }

    result := math.Sqrt(x)
    return &v1.CallToolResponse{
        Success: true,
        Result:  map[string]interface{}{"x": x, "result": result},
        Output:  fmt.Sprintf("%.2f", result),
    }
}

func (p *CalculatorPlugin) power(args map[string]interface{}) *v1.CallToolResponse {
    base, baseOk := args["base"].(float64)
    exp, expOk := args["exp"].(float64)

    if !baseOk || !expOk {
        return invalidArgs("需要参数 base 和 exp")
    }

    result := math.Pow(base, exp)
    return &v1.CallToolResponse{
        Success: true,
        Result:  map[string]interface{}{"base": base, "exp": exp, "result": result},
        Output:  fmt.Sprintf("%.2f", result),
    }
}

func (p *CalculatorPlugin) ListTools(ctx context.Context) *v1.ListToolsResponse {
    tools := []*v1.ToolInfo{
        {
            Name:        "add",
            Description: "加法计算",
            InputSchema: numericBinaryOpSchema("加法", "第一个数字", "第二个数字"),
        },
        {
            Name:        "subtract",
            Description: "减法计算",
            InputSchema: numericBinaryOpSchema("减法", "被减数", "减数"),
        },
        {
            Name:        "multiply",
            Description: "乘法计算",
            InputSchema: numericBinaryOpSchema("乘法", "第一个乘数", "第二个乘数"),
        },
        {
            Name:        "divide",
            Description: "除法计算",
            InputSchema: numericBinaryOpSchema("除法", "被除数", "除数"),
        },
        {
            Name:        "sqrt",
            Description: "计算平方根",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "x": map[string]interface{}{
                        "type":        "number",
                        "description": "要计算平方根的数字",
                        "minimum":     0,
                    },
                },
                "required": []string{"x"},
            },
        },
        {
            Name:        "pow",
            Description: "幂运算",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "base": map[string]interface{}{
                        "type":        "number",
                        "description": "底数",
                    },
                    "exp": map[string]interface{}{
                        "type":        "number",
                        "description": "指数",
                    },
                },
                "required": []string{"base", "exp"},
            },
        },
    }

    return &v1.ListToolsResponse{Success: true, Tools: tools}
}

func numericBinaryOpSchema(description, aDesc, bDesc string) map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "a": map[string]interface{}{
                "type":        "number",
                "description": aDesc,
            },
            "b": map[string]interface{}{
                "type":        "number",
                "description": bDesc,
            },
        },
        "required": []string{"a", "b"},
    }
}

func unknownTool(toolName string) *v1.CallToolResponse {
    return &v1.CallToolResponse{
        Success: false,
        Error: &v1.ErrorInfo{
            Code:    "UNKNOWN_TOOL",
            Message: fmt.Sprintf("未知工具: %s", toolName),
        },
    }
}

func invalidArgs(message string) *v1.CallToolResponse {
    return &v1.CallToolResponse{
        Success: false,
        Error: &v1.ErrorInfo{
            Code:    "INVALID_ARGS",
            Message: message,
        },
    }
}

func main() {
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "calculator-plugin",
        Level:  hclog.Info,
        Output: hclog.DefaultOutput,
    })

    info := &v1.PluginInfo{
        ID:          "calculator",
        Name:        "Calculator Plugin",
        Version:     "1.0.0",
        Description: "数学计算工具插件",
        Author:      "XiaoZhi Team",
        Type:        v1.PluginTypeUtility,
        Tags:        []string{"math", "calculator", "utility"},
        Capabilities: []string{"add", "subtract", "multiply", "divide", "sqrt", "pow"},
    }

    plugin := &CalculatorPlugin{
        SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
    }

    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.SimpleHandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "plugin": &sdk.SimplePluginRPC{Impl: plugin},
        },
    })
}
```

### 2. 天气查询插件

```go
// plugins/utility/weather/main.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"

    v1 "xiaozhi-server-go/api/v1"
    sdk "xiaozhi-server-go/internal/plugin/sdk"
)

type WeatherPlugin struct {
    sdk.SimplePluginImpl
    apikey string
    client *http.Client
}

func (p *WeatherPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
    switch req.ToolName {
    case "get_weather":
        return p.getWeather(ctx, req.Arguments)
    case "get_forecast":
        return p.getForecast(ctx, req.Arguments)
    default:
        return unknownTool(req.ToolName)
    }
}

func (p *WeatherPlugin) getWeather(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    city, ok := args["city"].(string)
    if !ok {
        return invalidArgs("需要参数 city")
    }

    url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
        url.QueryEscape(city), p.apikey)

    resp, err := p.client.Get(url)
    if err != nil {
        p.IncrementCounter("weather.error")
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "API_ERROR",
                Message: fmt.Sprintf("天气API调用失败: %v", err),
            },
        }
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "API_ERROR",
                Message: fmt.Sprintf("天气API返回错误: %d", resp.StatusCode),
            },
        }
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "READ_ERROR",
                Message: fmt.Sprintf("读取响应失败: %v", err),
            },
        }
    }

    var weatherData map[string]interface{}
    if err := json.Unmarshal(body, &weatherData); err != nil {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "PARSE_ERROR",
                Message: fmt.Sprintf("解析JSON失败: %v", err),
            },
        }
    }

    p.IncrementCounter("weather.success")
    p.RecordHistogram("weather.response_time", time.Since(time.Now()).Seconds())

    return &v1.CallToolResponse{
        Success: true,
        Result:  weatherData,
        Output:  formatWeatherOutput(weatherData),
    }
}

func (p *WeatherPlugin) getForecast(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    city, ok := args["city"].(string)
    days, daysOk := args["days"].(float64)

    if !ok {
        return invalidArgs("需要参数 city")
    }
    if !daysOk || days < 1 || days > 7 {
        return invalidArgs("参数 days 必须是 1-7 之间的数字")
    }

    // 实现预报逻辑...
    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "city": city,
            "days": days,
            "forecast": "天气预报数据",
        },
        Output: fmt.Sprintf("%s 未来%d天天气预报", city, int(days)),
    }
}

func (p *WeatherPlugin) ListTools(ctx context.Context) *v1.ListToolsResponse {
    tools := []*v1.ToolInfo{
        {
            Name:        "get_weather",
            Description: "获取当前天气信息",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "city": map[string]interface{}{
                        "type":        "string",
                        "description": "城市名称",
                    },
                },
                "required": []string{"city"},
            },
        },
        {
            Name:        "get_forecast",
            Description: "获取天气预报",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "city": map[string]interface{}{
                        "type":        "string",
                        "description": "城市名称",
                    },
                    "days": map[string]interface{}{
                        "type":        "number",
                        "description": "预报天数 (1-7)",
                        "minimum":     1,
                        "maximum":     7,
                        "default":     3,
                    },
                },
                "required": []string{"city"},
            },
        },
    }

    return &v1.ListToolsResponse{Success: true, Tools: tools}
}

func formatWeatherOutput(data map[string]interface{}) string {
    main, ok := data["main"].(map[string]interface{})
    if !ok {
        return "天气数据格式错误"
    }

    temp, _ := main["temp"].(float64)
    humidity, _ := main["humidity"].(float64)

    weather, ok := data["weather"].([]interface{})
    if !ok || len(weather) == 0 {
        return "天气信息不完整"
    }

    weatherInfo, ok := weather[0].(map[string]interface{})
    if !ok {
        return "天气详情格式错误"
    }

    description, _ := weatherInfo["description"].(string)

    return fmt.Sprintf("当前天气：%s，温度：%.1f°C，湿度：%.0f%%", description, temp, humidity)
}

func main() {
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "weather-plugin",
        Level:  hclog.Info,
        Output: hclog.DefaultOutput,
    })

    // 从环境变量读取 API Key
    apikey := os.Getenv("OPENWEATHER_API_KEY")
    if apikey == "" {
        logger.Error("OPENWEATHER_API_KEY 环境变量未设置")
        return
    }

    info := &v1.PluginInfo{
        ID:          "weather",
        Name:        "Weather Plugin",
        Version:     "1.0.0",
        Description: "天气查询插件",
        Author:      "XiaoZhi Team",
        Type:        v1.PluginTypeUtility,
        Tags:        []string{"weather", "api", "utility"},
        Capabilities: []string{"get_weather", "get_forecast"},
    }

    plugin := &WeatherPlugin{
        SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
        apikey:           apikey,
        client: &http.Client{Timeout: 10 * time.Second},
    }

    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.SimpleHandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "plugin": &sdk.SimplePluginRPC{Impl: plugin},
        },
    })
}
```

### 3. 文件操作插件

```go
// plugins/utility/fileops/main.go
package main

import (
    "context"
    "encoding/base64"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"

    v1 "xiaozhi-server-go/api/v1"
    sdk "xiaozhi-server-go/internal/plugin/sdk"
)

type FileOpsPlugin struct {
    sdk.SimplePluginImpl
    baseDir string
}

func (p *FileOpsPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
    switch req.ToolName {
    case "read_file":
        return p.readFile(ctx, req.Arguments)
    case "write_file":
        return p.writeFile(ctx, req.Arguments)
    case "list_files":
        return p.listFiles(ctx, req.Arguments)
    case "create_dir":
        return p.createDir(ctx, req.Arguments)
    case "delete_file":
        return p.deleteFile(ctx, req.Arguments)
    default:
        return unknownTool(req.ToolName)
    }
}

func (p *FileOpsPlugin) readFile(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    path, ok := args["path"].(string)
    if !ok {
        return invalidArgs("需要参数 path")
    }

    // 安全检查：确保路径在允许的目录内
    fullPath := filepath.Join(p.baseDir, path)
    if !isPathSafe(fullPath, p.baseDir) {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "PATH_ERROR",
                Message: "路径不安全",
            },
        }
    }

    data, err := os.ReadFile(fullPath)
    if err != nil {
        p.IncrementCounter("file_ops.read_error")
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "READ_ERROR",
                Message: fmt.Sprintf("读取文件失败: %v", err),
            },
        }
    }

    content := string(data)
    p.IncrementCounter("file_ops.read_success")
    p.RecordHistogram("file_ops.read_size", float64(len(content)))

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "path":    path,
            "content": content,
            "size":    len(content),
        },
        Output: fmt.Sprintf("文件读取成功，大小：%d 字节", len(content)),
    }
}

func (p *FileOpsPlugin) writeFile(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    path, ok := args["path"].(string)
    content, ok := args["content"].(string)
    if !ok {
        return invalidArgs("需要参数 path 和 content")
    }

    fullPath := filepath.Join(p.baseDir, path)
    if !isPathSafe(fullPath, p.baseDir) {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "PATH_ERROR",
                Message: "路径不安全",
            },
        }
    }

    // 创建目录
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "MKDIR_ERROR",
                Message: fmt.Sprintf("创建目录失败: %v", err),
            },
        }
    }

    err := os.WriteFile(fullPath, []byte(content), 0644)
    if err != nil {
        p.IncrementCounter("file_ops.write_error")
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "WRITE_ERROR",
                Message: fmt.Sprintf("写入文件失败: %v", err),
            },
        }
    }

    p.IncrementCounter("file_ops.write_success")
    p.RecordHistogram("file_ops.write_size", float64(len(content)))

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "path":    path,
            "size":    len(content),
        },
        Output: fmt.Sprintf("文件写入成功，大小：%d 字节", len(content)),
    }
}

func (p *FileOpsPlugin) listFiles(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    path, ok := args["path"].(string)
    if !ok {
        path = "."
    }

    fullPath := filepath.Join(p.baseDir, path)
    if !isPathSafe(fullPath, p.baseDir) {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "PATH_ERROR",
                Message: "路径不安全",
            },
        }
    }

    entries, err := os.ReadDir(fullPath)
    if err != nil {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "READDIR_ERROR",
                Message: fmt.Sprintf("读取目录失败: %v", err),
            },
        }
    }

    var files []map[string]interface{}
    for _, entry := range entries {
        info, err := entry.Info()
        if err != nil {
            continue
        }

        file := map[string]interface{}{
            "name":    entry.Name(),
            "size":    info.Size(),
            "is_dir":  entry.IsDir(),
            "mod_time": info.ModTime(),
        }

        files = append(files, file)
    }

    p.IncrementCounter("file_ops.list_success")

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "path":  path,
            "files": files,
            "count": len(files),
        },
        Output: fmt.Sprintf("找到 %d 个文件/目录", len(files)),
    }
}

func isPathSafe(path, baseDir string) bool {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return false
    }

    absBase, err := filepath.Abs(baseDir)
    if err != nil {
        return false
    }

    rel, err := filepath.Rel(absBase, absPath)
    if err != nil {
        return false
    }

    return !filepath.IsAbs(rel) && !strings.Contains(rel, "..")
}

func (p *FileOpsPlugin) createDir(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    path, ok := args["path"].(string)
    if !ok {
        return invalidArgs("需要参数 path")
    }

    fullPath := filepath.Join(p.baseDir, path)
    if !isPathSafe(fullPath, p.baseDir) {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "PATH_ERROR",
                Message: "路径不安全",
            },
        }
    }

    err := os.MkdirAll(fullPath, 0755)
    if err != nil {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "MKDIR_ERROR",
                Message: fmt.Sprintf("创建目录失败: %v", err),
            },
        }
    }

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "path": path,
        },
        Output: fmt.Sprintf("目录创建成功: %s", path),
    }
}

func (p *FileOpsPlugin) deleteFile(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    path, ok := args["path"].(string)
    if !ok {
        return invalidArgs("需要参数 path")
    }

    fullPath := filepath.Join(p.baseDir, path)
    if !isPathSafe(fullPath, p.baseDir) {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "PATH_ERROR",
                Message: "路径不安全",
            },
        }
    }

    err := os.Remove(fullPath)
    if err != nil {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "DELETE_ERROR",
                Message: fmt.Sprintf("删除文件失败: %v", err),
            },
        }
    }

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "path": path,
        },
        Output: fmt.Sprintf("文件删除成功: %s", path),
    }
}

func (p *FileOpsPlugin) ListTools(ctx context.Context) *v1.ListToolsResponse {
    tools := []*v1.ToolInfo{
        {
            Name:        "read_file",
            Description: "读取文件内容",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "path": map[string]interface{}{
                        "type":        "string",
                        "description": "文件路径",
                    },
                },
                "required": []string{"path"},
            },
        },
        {
            Name:        "write_file",
            Description: "写入文件内容",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "path": map[string]interface{}{
                        "type":        "string",
                        "description": "文件路径",
                    },
                    "content": map[string]interface{}{
                        "type":        "string",
                        "description": "文件内容",
                    },
                },
                "required": []string{"path", "content"},
            },
        },
        {
            Name:        "list_files",
            Description: "列出目录中的文件",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "path": map[string]interface{}{
                        "type":        "string",
                        "description": "目录路径（默认：当前目录）",
                        "default":     ".",
                    },
                },
            },
        },
        {
            Name:        "create_dir",
            Description: "创建目录",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "path": map[string]interface{}{
                        "type":        "string",
                        "description": "目录路径",
                    },
                },
                "required": []string{"path"},
            },
        },
        {
            Name:        "delete_file",
            Description: "删除文件",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "path": map[string]interface{}{
                        "type":        "string",
                        "description": "文件路径",
                    },
                },
                "required": []string{"path"},
            },
        },
    }

    return &v1.ListToolsResponse{Success: true, Tools: tools}
}

func main() {
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "fileops-plugin",
        Level:  hclog.Info,
        Output: hclog.DefaultOutput,
    })

    // 从环境变量读取基础目录，默认为 ./data/files
    baseDir := os.Getenv("PLUGIN_BASE_DIR")
    if baseDir == "" {
        baseDir = "./data/files"
    }

    // 确保基础目录存在
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        logger.Error("创建基础目录失败", "error", err)
        return
    }

    info := &v1.PluginInfo{
        ID:          "fileops",
        Name:        "File Operations Plugin",
        Version:     "1.0.0",
        Description: "文件操作插件",
        Author:      "XiaoZhi Team",
        Type:        v1.PluginTypeUtility,
        Tags:        []string{"file", "storage", "utility"},
        Capabilities: []string{"read_file", "write_file", "list_files", "create_dir", "delete_file"},
    }

    plugin := &FileOpsPlugin{
        SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
        baseDir:           baseDir,
    }

    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.SimpleHandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "plugin": &sdk.SimplePluginRPC{Impl: plugin},
        },
    })
}
```

## 💡 最佳实践

### 1. 错误处理

```go
// 好的错误处理
func (p *MyPlugin) doSomething(args map[string]interface{}) *v1.CallToolResponse {
    // 参数验证
    if param, ok := args["required_param"].(string); !ok {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: "缺少必需参数 'required_param'",
                Context: map[string]string{
                    "received_args": fmt.Sprintf("%v", args),
                },
            },
        }
    }

    // 业务逻辑错误
    if err := someOperation(param); err != nil {
        p.IncrementCounter("operation_error")
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "OPERATION_FAILED",
                Message: fmt.Sprintf("操作失败: %v", err),
                Context: map[string]string{
                    "param": param,
                },
            },
        }
    }

    // 成功
    p.IncrementCounter("operation_success")
    return &v1.CallToolResponse{
        Success: true,
        Result:  result,
        Output:  "操作成功",
    }
}
```

### 2. 指标收集

```go
// 收集有意义的指标
func (p *MyPlugin) processRequest(req *v1.CallToolRequest) *v1.CallToolResponse {
    start := time.Now()
    defer func() {
        p.RecordHistogram("request_duration", time.Since(start).Seconds())
    }()

    p.IncrementCounter("request_total")

    // 处理请求...
    success := true

    if success {
        p.IncrementCounter("request_success")
    } else {
        p.IncrementCounter("request_error")
    }

    p.SetGauge("active_requests", 0)

    return response
}
```

### 3. 配置管理

```go
type MyPluginConfig struct {
    APIKey     string `yaml:"api_key"`
    Timeout    int    `yaml:"timeout"`
    MaxRetries int    `yaml:"max_retries"`
    Debug      bool   `yaml:"debug"`
}

func (p *MyPlugin) Initialize(ctx context.Context, config *sdk.InitializeConfig) error {
    // 解析插件配置
    pluginConfig := &MyPluginConfig{}
    if err := mapstructure.Decode(config.Config, pluginConfig); err != nil {
        return fmt.Errorf("解析配置失败: %w", err)
    }

    // 验证配置
    if pluginConfig.APIKey == "" {
        return fmt.Errorf("API Key 不能为空")
    }

    // 存储配置
    p.config = pluginConfig

    p.logger.Info("插件配置加载成功",
        "timeout", pluginConfig.Timeout,
        "max_retries", pluginConfig.MaxRetries,
    )

    return nil
}
```

### 4. 资源管理

```go
func (p *MyPlugin) Initialize(ctx context.Context, config *sdk.InitializeConfig) error {
    // 创建资源
    p.httpClient = &http.Client{
        Timeout: time.Duration(p.config.Timeout) * time.Second,
    }

    p.dbConnection = createDatabaseConnection()

    // 设置清理函数
    go func() {
        <-ctx.Done()
        p.cleanup()
    }()

    return nil
}

func (p *MyPlugin) cleanup() {
    if p.httpClient != nil {
        p.httpClient.CloseIdleConnections()
    }

    if p.dbConnection != nil {
        p.dbConnection.Close()
    }
}

func (p *MyPlugin) Shutdown(ctx context.Context) error {
    p.cleanup()
    return nil
}
```

### 5. 安全考虑

```go
// 输入验证
func validateInput(input string) error {
    if len(input) > 1000 {
        return fmt.Errorf("输入过长")
    }

    if strings.Contains(input, "..") {
        return fmt.Errorf("包含不安全路径")
    }

    return nil
}

// 路径安全检查
func isPathSafe(path, baseDir string) bool {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return false
    }

    absBase, err := filepath.Abs(baseDir)
    if err != nil {
        return false
    }

    rel, err := filepath.Rel(absBase, absPath)
    if err != nil {
        return false
    }

    return !filepath.IsAbs(rel) && !strings.Contains(rel, "..")
}

// 资源限制
func (p *MyPlugin) checkResourceLimits() error {
    if runtime.NumGoroutine() > 100 {
        return fmt.Errorf("goroutine 数量过多")
    }

    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    if m.Alloc > 100*1024*1024 { // 100MB
        return fmt.Errorf("内存使用过多")
    }

    return nil
}
```

### 6. 日志记录

```go
// 结构化日志
func (p *MyPlugin) processRequest(req *v1.CallToolRequest) {
    p.logger.Info("处理请求",
        "tool", req.ToolName,
        "args_len", len(req.Arguments),
        "request_id", ctx.Value("request_id"),
    )

    result, err := doWork(req)

    if err != nil {
        p.logger.Error("请求处理失败",
            "tool", req.ToolName,
            "error", err,
            "request_id", ctx.Value("request_id"),
        )
    } else {
        p.logger.Info("请求处理成功",
            "tool", req.ToolName,
            "result_size", len(fmt.Sprintf("%v", result)),
            "request_id", ctx.Value("request_id"),
        )
    }
}
```

## 🧪 调试和测试

### 单元测试

```go
package main

import (
    "context"
    "testing"
    "time"

    "github.com/hashicorp/go-hclog"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    v1 "xiaozhi-server-go/api/v1"
    sdk "xiaozhi-server-go/internal/plugin/sdk"
)

func TestMyPlugin_CallTool(t *testing.T) {
    // 创建插件实例
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "test",
        Level:  hclog.Debug,
    })

    plugin := NewMyPlugin(logger)

    // 初始化插件
    config := &sdk.InitializeConfig{
        Config: map[string]interface{}{
            "api_key": "test_key",
        },
    }

    err := plugin.Initialize(context.Background(), config)
    require.NoError(t, err)

    // 测试工具调用
    t.Run("hello 工具", func(t *testing.T) {
        req := &v1.CallToolRequest{
            ToolName: "hello",
            Arguments: map[string]interface{}{
                "name": "测试用户",
            },
        }

        resp := plugin.CallTool(context.Background(), req)

        assert.True(t, resp.Success)
        assert.Contains(t, resp.Output, "Hello, 测试用户")
        assert.Equal(t, "Hello, 测试用户! from My Plugin", resp.Result["message"])
    })

    t.Run("未知工具", func(t *testing.T) {
        req := &v1.CallToolRequest{
            ToolName: "unknown_tool",
            Arguments: map[string]interface{}{},
        }

        resp := plugin.CallTool(context.Background(), req)

        assert.False(t, resp.Success)
        assert.Equal(t, "UNKNOWN_TOOL", resp.Error.Code)
    })

    // 清理
    err = plugin.Shutdown(context.Background())
    require.NoError(t, err)
}

func TestMyPlugin_ListTools(t *testing.T) {
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "test",
        Level:  hclog.Debug,
    })

    plugin := NewMyPlugin(logger)

    resp := plugin.ListTools(context.Background())

    assert.True(t, resp.Success)
    assert.NotEmpty(t, resp.Tools)

    // 验证工具信息
    toolNames := make([]string, len(resp.Tools))
    for i, tool := range resp.Tools {
        toolNames[i] = tool.Name
        assert.NotEmpty(t, tool.Description)
        assert.NotNil(t, tool.InputSchema)
    }

    assert.Contains(t, toolNames, "hello")
}

func TestMyPlugin_GetToolSchema(t *testing.T) {
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "test",
        Level:  hclog.Debug,
    })

    plugin := NewMyPlugin(logger)

    t.Run("存在的工具", func(t *testing.T) {
        req := &v1.GetToolSchemaRequest{
            ToolName: "hello",
        }

        resp := plugin.GetToolSchema(context.Background(), req)

        assert.True(t, resp.Success)
        assert.NotNil(t, resp.Schema)
        assert.Equal(t, "object", resp.Schema["type"])
    })

    t.Run("不存在的工具", func(t *testing.T) {
        req := &v1.GetToolSchemaRequest{
            ToolName: "unknown_tool",
        }

        resp := plugin.GetToolSchema(context.Background(), req)

        assert.False(t, resp.Success)
        assert.Equal(t, "TOOL_NOT_FOUND", resp.Error.Code)
    })
}

// 基准测试
func BenchmarkMyPlugin_CallTool(b *testing.B) {
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "benchmark",
        Level:  hclog.Error, // 减少日志输出
    })

    plugin := NewMyPlugin(logger)
    plugin.Initialize(context.Background(), &sdk.InitializeConfig{})

    req := &v1.CallToolRequest{
        ToolName: "hello",
        Arguments: map[string]interface{}{
            "name": "Benchmark User",
        },
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        resp := plugin.CallTool(context.Background(), req)
        if !resp.Success {
            b.Fatalf("Tool call failed: %v", resp.Error)
        }
    }
}

// 集成测试
func TestMyPlugin_Integration(t *testing.T) {
    // 测试完整的使用流程
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "integration",
        Level:  hclog.Debug,
    })

    plugin := NewMyPlugin(logger)

    // 1. 初始化
    config := &sdk.InitializeConfig{
        Config: map[string]interface{}{
            "api_key": "test_key",
        },
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    err := plugin.Initialize(ctx, config)
    require.NoError(t, err)

    // 2. 获取插件信息
    info := plugin.GetInfo()
    assert.Equal(t, "my-first-plugin", info.ID)
    assert.Equal(t, "My First Plugin", info.Name)

    // 3. 健康检查
    health := plugin.HealthCheck(ctx)
    assert.True(t, health.Healthy)

    // 4. 获取工具列表
    tools := plugin.ListTools(ctx)
    assert.True(t, tools.Success)
    assert.NotEmpty(t, tools.Tools)

    // 5. 调用工具
    for _, tool := range tools.Tools {
        req := &v1.CallToolRequest{
            ToolName: tool.Name,
        }

        // 为有参数的工具添加参数
        if tool.Name == "hello" {
            req.Arguments = map[string]interface{}{
                "name": "Integration Test",
            }
        }

        resp := plugin.CallTool(ctx, req)
        if !resp.Success {
            t.Logf("Tool %s failed: %v", tool.Name, resp.Error)
        }
    }

    // 6. 获取指标
    metrics := plugin.GetMetrics(ctx)
    assert.NotNil(t, metrics)
    assert.Greater(t, len(metrics.Counters), 0)

    // 7. 关闭
    err = plugin.Shutdown(ctx)
    require.NoError(t, err)
}
```

### 集成测试

```go
package main

import (
    "context"
    "testing"
    "time"

    pluginmanager "xiaozhi-server-go/internal/plugin/manager"
    "github.com/hashicorp/go-hclog"
    "github.com/stretchr/testify/require"
)

func TestPluginManager_Integration(t *testing.T) {
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "integration-test",
        Level:  hclog.Debug,
    })

    // 创建插件管理器配置
    config := &pluginmanager.PluginConfig{
        Enabled: true,
        Discovery: &pluginmanager.DiscoveryConfig{
            Enabled:      true,
            ScanInterval: 5 * time.Second,
            Paths:        []string{"../../plugins/examples"},
        },
        Registry: &pluginmanager.RegistryConfig{
            Type: "memory",
            TTL:  5 * time.Minute,
        },
        HealthCheck: &pluginmanager.HealthCheckConfig{
            Interval:         2 * time.Second,
            Timeout:          1 * time.Second,
            FailureThreshold: 3,
        },
    }

    // 创建插件管理器
    pm, err := pluginmanager.NewPluginManager(config, logger)
    require.NoError(t, err)

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 启动插件管理器
    err = pm.Start(ctx)
    require.NoError(t, err)

    // 等待插件发现
    time.Sleep(2 * time.Second)

    // 列出插件
    plugins, err := pm.ListPlugins()
    require.NoError(t, err)

    t.Logf("发现 %d 个插件", len(plugins))

    // 健康检查
    healthStatuses := pm.HealthCheckAll(ctx)
    for pluginID, status := range healthStatuses {
        t.Logf("插件 %s 健康状态: %v", pluginID, status.Healthy)
    }

    // 停止插件管理器
    err = pm.Stop(ctx)
    require.NoError(t, err)
}
```

### 调试技巧

```go
// 1. 使用调试日志
func (p *MyPlugin) debugMethod() {
    p.logger.Debug("调试信息",
        "goroutines", runtime.NumGoroutine(),
        "memory", getMemoryUsage(),
        "config", p.config,
    )
}

// 2. 性能分析
func (p *MyPlugin) profileMethod() {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        p.logger.Info("方法执行时间", "duration", duration)
        p.RecordHistogram("method_duration", duration.Seconds())
    }()

    // 方法实现
}

// 3. 条件断点
func (p *MyPlugin) conditionalBreakpoint() {
    if os.Getenv("DEBUG_PLUGIN") == "true" {
        runtime.Breakpoint()
    }
}

// 4. 状态检查
func (p *MyPlugin) healthCheckDetailed() *v1.HealthStatus {
    status := &v1.HealthStatus{
        Healthy:   true,
        Status:    "healthy",
        Timestamp: time.Now(),
    }

    checks := []string{}
    details := make(map[string]string)

    // 检查内存使用
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    if m.Alloc > 100*1024*1024 { // 100MB
        status.Healthy = false
        status.Status = "high_memory"
        details["memory_usage"] = fmt.Sprintf("%d MB", m.Alloc/1024/1024)
    }
    checks = append(checks, "memory")

    // 检查goroutine数量
    if runtime.NumGoroutine() > 50 {
        status.Healthy = false
        status.Status = "too_many_goroutines"
        details["goroutines"] = fmt.Sprintf("%d", runtime.NumGoroutine())
    }
    checks = append(checks, "goroutines")

    status.Checks = checks
    status.Details = details

    return status
}
```

## 🚀 部署和分发

### 1. 插件打包

```bash
# 创建构建脚本
#!/bin/bash
# build.sh

set -e

PLUGIN_NAME="my-plugin"
VERSION=$(cat plugin.yaml | grep version | awk '{print $2}')
BUILD_DIR="build"
DIST_DIR="dist"

# 清理旧的构建
rm -rf $BUILD_DIR $DIST_DIR
mkdir -p $BUILD_DIR $DIST_DIR

# 构建插件
echo "构建插件 $PLUGIN_NAME v$VERSION..."
go build -ldflags "-X main.version=$VERSION" -o $BUILD_DIR/$PLUGIN_NAME main.go

# 复制配置文件
cp plugin.yaml $BUILD_DIR/
cp README.md $BUILD_DIR/ 2>/dev/null || true
cp -r docs $BUILD_DIR/ 2>/dev/null || true

# 创建分发包
cd $BUILD_DIR
tar -czf ../$DIST_DIR/${PLUGIN_NAME}-${VERSION}.tar.gz *
cd ..

echo "插件构建完成: $DIST_DIR/${PLUGIN_NAME}-${VERSION}.tar.gz"
```

### 2. 插件安装

```bash
# 创建安装脚本
#!/bin/bash
# install.sh

set -e

PLUGIN_FILE=$1
INSTALL_DIR=${2:-"./plugins"}

if [ -z "$PLUGIN_FILE" ]; then
    echo "用法: $0 <plugin.tar.gz> [install_dir]"
    exit 1
fi

if [ ! -f "$PLUGIN_FILE" ]; then
    echo "插件文件不存在: $PLUGIN_FILE"
    exit 1
fi

# 创建安装目录
mkdir -p $INSTALL_DIR

# 解压插件
echo "安装插件到 $INSTALL_DIR..."
tar -xzf $PLUGIN_FILE -C $INSTALL_DIR

# 设置权限
chmod +x $INSTALL_DIR/*/main.go 2>/dev/null || true

echo "插件安装完成"
```

### 3. 插件配置

```yaml
# config/plugins.yaml
plugins:
  my-plugin:
    enabled: true
    deployment:
      type: local_binary
      path: ./plugins/my-plugin/main.go
    config:
      api_key: "${MY_PLUGIN_API_KEY}"
      timeout: 30
      debug: false
    environment:
      PLUGIN_LOG_LEVEL: "info"
```

### 4. Docker 部署

```dockerfile
# plugins/my-plugin/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o my-plugin main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/my-plugin .
COPY --from=builder /app/plugin.yaml .

EXPOSE 8080
CMD ["./my-plugin"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  my-plugin:
    build: ./plugins/my-plugin
    environment:
      - MY_PLUGIN_API_KEY=${API_KEY}
      - PLUGIN_LOG_LEVEL=info
    volumes:
      - ./plugins:/app/plugins
    restart: unless-stopped
```

## ❓ 常见问题

### 1. 插件无法被发现

**问题**: 插件在 plugins 目录中但没有被系统发现

**解决方案**:
```bash
# 检查插件配置
cat plugins/my-plugin/plugin.yaml

# 检查文件权限
ls -la plugins/my-plugin/

# 检查日志
tail -f logs/plugin-manager.log

# 验证插件格式
go build -o /tmp/test plugins/my-plugin/main.go
/tmp/test --help
```

### 2. 工具调用失败

**问题**: 工具调用返回错误

**解决方案**:
```go
// 在插件中添加详细日志
func (p *MyPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
    p.logger.Debug("工具调用开始",
        "tool", req.ToolName,
        "args", req.Arguments,
    )

    // 参数验证
    if err := validateArgs(req.Arguments); err != nil {
        p.logger.Error("参数验证失败", "error", err)
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: err.Error(),
                Context: map[string]string{
                    "received_args": fmt.Sprintf("%v", req.Arguments),
                },
            },
        }
    }

    // 业务逻辑...
}
```

### 3. 插件内存泄漏

**问题**: 插件运行一段时间后内存使用过高

**解决方案**:
```go
// 定期清理资源
func (p *MyPlugin) cleanupRoutine() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-p.ctx.Done():
            return
        case <-ticker.C:
            p.cleanup()
            p.checkResourceUsage()
        }
    }
}

func (p *MyPlugin) checkResourceUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    if m.Alloc > 100*1024*1024 { // 100MB
        p.logger.Warn("内存使用过高",
            "alloc", m.Alloc/1024/1024,
            "goroutines", runtime.NumGoroutine(),
        )

        // 触发垃圾回收
        runtime.GC()
    }
}
```

### 4. 插件性能问题

**问题**: 插件响应时间过长

**解决方案**:
```go
// 添加超时控制
func (p *MyPlugin) CallTool(ctx context.Context, req *v1.CallToolResponse) *v1.CallToolResponse {
    // 设置超时
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // 使用带超时的操作
    result, err := p.doOperationWithTimeout(ctx, req)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return &v1.CallToolResponse{
                Success: false,
                Error: &v1.ErrorInfo{
                    Code:    "TIMEOUT",
                    Message: "操作超时",
                },
            }
        }
        return handleError(err)
    }

    return result
}

// 并发控制
func (p *MyPlugin) doOperationWithTimeout(ctx context.Context, req *v1.CallToolResponse) (*v1.CallToolResponse, error) {
    // 限制并发数
    semaphore := make(chan struct{}, 10)

    semaphore <- struct{}{}
    defer func() { <-semaphore }()

    return p.doOperation(req)
}
```

### 5. 插件配置错误

**问题**: 插件配置解析失败

**解决方案**:
```go
// 配置验证
func (p *MyPlugin) validateConfig(config map[string]interface{}) error {
    if apiKey, ok := config["api_key"].(string); !ok || apiKey == "" {
        return fmt.Errorf("api_key 不能为空")
    }

    if timeout, ok := config["timeout"].(int); !ok || timeout <= 0 {
        return fmt.Errorf("timeout 必须是正整数")
    }

    if timeout > 300 {
        return fmt.Errorf("timeout 不能超过 300 秒")
    }

    return nil
}

// 默认配置
func (p *MyPlugin) setDefaultConfig(config map[string]interface{}) {
    if _, ok := config["timeout"]; !ok {
        config["timeout"] = 30
    }

    if _, ok := config["retry_count"]; !ok {
        config["retry_count"] = 3
    }

    if _, ok := config["debug"]; !ok {
        config["debug"] = false
    }
}
```

### 6. 插件通信问题

**问题**: 插件与主系统通信失败

**解决方案**:
```go
// 通信重试机制
func (p *MyPlugin) callWithRetry(fn func() error, maxRetries int) error {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        if i > 0 {
            p.logger.Debug("重试通信", "attempt", i+1, "error", lastErr)
            time.Sleep(time.Duration(i) * time.Second)
        }

        if err := fn(); err != nil {
            lastErr = err
            continue
        }

        return nil
    }

    return fmt.Errorf("通信失败，已重试 %d 次: %w", maxRetries, lastErr)
}

// 连接健康检查
func (p *MyPlugin) checkConnection() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return p.httpClient.Get("https://api.example.com/health")
}
```

---

## 📚 更多资源

### 官方文档
- [XiaoZhi Flow 官方文档](https://docs.xiaozhi-flow.dev)
- [插件 API 参考](https://docs.xiaozhi-flow.dev/plugins/api)
- [开发者指南](https://docs.xiaozhi-flow.dev/developers)

### 社区资源
- [GitHub 仓库](https://github.com/xiaozhi-flow/plugins)
- [插件市场](https://market.xiaozhi-flow.dev/plugins)
- [开发者论坛](https://forum.xiaozhi-flow.dev)

### 示例仓库
- [示例插件集合](https://github.com/xiaozhi-flow/plugin-examples)
- [插件模板](https://github.com/xiaozhi-flow/plugin-template)
- [开发工具](https://github.com/xiaozhi-flow/plugin-tools)

---

如果你在开发过程中遇到问题，可以：

1. 查看日志文件了解详细错误信息
2. 查看插件配置是否正确
3. 参考示例插件的实现
4. 在社区论坛寻求帮助
5. 提交 Issue 到 GitHub 仓库

祝你开发愉快！🎉