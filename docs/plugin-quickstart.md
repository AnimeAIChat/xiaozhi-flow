# XiaoZhi Flow 插件快速开始

## 🚀 5分钟创建第一个插件

### 1. 创建插件目录和文件

```bash
mkdir -p plugins/hello-world
cd plugins/hello-world
```

### 2. 创建插件主文件 `main.go`

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

// HelloWorldPlugin 简单的Hello World插件
type HelloWorldPlugin struct {
    sdk.SimplePluginImpl
    logger hclog.Logger
}

// NewHelloWorldPlugin 创建插件实例
func NewHelloWorldPlugin(logger hclog.Logger) *HelloWorldPlugin {
    info := &v1.PluginInfo{
        ID:          "hello-world",
        Name:        "Hello World Plugin",
        Version:     "1.0.0",
        Description: "简单的Hello World插件",
        Author:      "You",
        Type:        v1.PluginTypeUtility,
        Tags:        []string{"example", "hello"},
        Capabilities: []string{"greet"},
    }

    return &HelloWorldPlugin{
        SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
        logger:          logger.Named("hello-plugin"),
    }
}

// CallTool 实现工具调用
func (p *HelloWorldPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
    if req.ToolName == "greet" {
        name, ok := req.Arguments["name"].(string)
        if !ok {
            name = "World"
        }

        message := fmt.Sprintf("Hello, %s! from XiaoZhi Flow Plugin", name)

        return &v1.CallToolResponse{
            Success: true,
            Result: map[string]interface{}{
                "message":   message,
                "timestamp": ctx.Value("timestamp"),
            },
            Output: message,
        }
    }

    return &v1.CallToolResponse{
        Success: false,
        Error: &v1.ErrorInfo{
            Code:    "UNKNOWN_TOOL",
            Message: fmt.Sprintf("未知工具: %s", req.ToolName),
        },
    }
}

// ListTools 列出可用工具
func (p *HelloWorldPlugin) ListTools(ctx context.Context) *v1.ListToolsResponse {
    tools := []*v1.ToolInfo{
        {
            Name:        "greet",
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
    }

    return &v1.ListToolsResponse{
        Success: true,
        Tools:   tools,
    }
}

// GetToolSchema 获取工具模式
func (p *HelloWorldPlugin) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
    if req.ToolName == "greet" {
        return &v1.GetToolSchemaResponse{
            Success: true,
            Schema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "name": map[string]interface{}{
                        "type":        "string",
                        "description": "要问候的名字",
                        "default":     "World",
                    },
                },
            },
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
        Name:   "hello-world-plugin",
        Level:  hclog.Info,
        Output: hclog.DefaultOutput,
    })

    // 创建插件实例
    plugin := NewHelloWorldPlugin(logger)

    logger.Info("Starting Hello World Plugin")

    // 服务插件
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.SimpleHandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "plugin": &sdk.SimplePluginRPC{Impl: plugin},
        },
    })
}
```

### 3. 创建插件配置文件 `plugin.yaml`

```yaml
name: Hello World Plugin
version: 1.0.0
description: 简单的Hello World插件
author: You
type: utility
tags:
  - example
  - hello
capabilities:
  - greet
metadata:
  language: go
  created_at: "2024-01-01T00:00:00Z"

deployment:
  type: local_binary
  path: ./main.go
  resources:
    max_memory: "32Mi"
    max_cpu: "50m"
  timeout: 5s
  retry_count: 3

config:
  greeting: "Hello"

environment:
  PLUGIN_LOG_LEVEL: "info"

enabled: true
```

### 4. 运行插件

```bash
# 在插件目录中构建
go build -o hello-world main.go

# 测试运行
./hello-world

# 或直接运行（需要Go环境）
go run main.go
```

### 5. 启动主系统

```bash
# 回到项目根目录
cd ../../

# 启动 XiaoZhi Flow 系统
go run cmd/xiaozhi-server/main.go
```

系统启动后，你应该能在日志中看到：

```
[引导] 初始化插件管理器
[插件] 插件管理器初始化完成
[插件] 发现插件: hello-world
```

## 🎯 验证插件工作

### 1. 查看 API 文档

访问 `http://localhost:8080/docs` 查看插件信息。

### 2. 通过 REST API 测试

```bash
# 列出所有工具
curl -X POST http://localhost:8080/api/v1/plugins/hello-world/tools/list \
  -H "Content-Type: application/json"

# 调用 hello 工具
curl -X POST http://localhost:8080/api/v1/plugins/hello-world/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "greet",
    "arguments": {
      "name": "XiaoZhi"
    }
  }'
```

响应应该类似：
```json
{
  "success": true,
  "result": {
    "message": "Hello, XiaoZhi! from XiaoZhi Flow Plugin"
  },
  "output": "Hello, XiaoZhi! from XiaoZhi Flow Plugin"
}
```

## 🔧 下一步

现在你已经创建了第一个插件！接下来可以：

1. **查看完整文档**：`docs/plugin-development.md`
2. **尝试其他插件类型**：
   - 音频处理插件
   - 大模型集成插件
   - 设备控制插件
3. **添加更多工具**到现有插件
4. **学习高级功能**：
   - 指标收集
   - 错误处理
   - 配置管理
5. **发布插件**到插件市场

## 💡 提示

- 插件目录：`plugins/`
- 配置文件：`config/plugins.yaml`
- 日志文件：`logs/plugin-manager.log`
- API 文档：`http://localhost:8080/docs`

祝你插件开发愉快！🎉