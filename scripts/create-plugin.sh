#!/bin/bash

# XiaoZhi Flow 插件创建工具
# 用法: ./scripts/create-plugin.sh <plugin-name> [options]

set -e

# 默认值
PLUGIN_TYPE="utility"
AUTHOR="Your Name"
DESCRIPTION=""
INTERACTIVE=false

# 显示帮助
show_help() {
    cat << EOF
XiaoZhi Flow 插件创建工具

用法: $0 <plugin-name> [选项]

选项:
  -t, --type TYPE        插件类型 (utility|audio|llm|device) [默认: utility]
  -a, --author AUTHOR    作者姓名 [默认: Your Name]
  -d, --description DESC 插件描述 [默认: 自动生成]
  -i, --interactive    交互式模式
  -h, --help           显示帮助信息

插件类型说明:
  utility - 通用功能插件 (文件操作、网络请求等)
  audio   - 音频处理插件 (ASR、TTS、VAD等)
  llm     - 大模型插件 (文本生成、对话等)
  device  - 设备控制插件 (IoT设备控制、传感器等)

示例:
  $0 my-utility-plugin
  $0 weather-plugin -a "Your Name" -d "天气查询插件"
  $0 tts-plugin -t audio -i

EOF
}

# 解析命令行参数
parse_args() {
    PLUGIN_NAME=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--type)
                PLUGIN_TYPE="$2"
                shift 2
                ;;
            -a|--author)
                AUTHOR="$2"
                shift 2
                ;;
            -d|--description)
                DESCRIPTION="$2"
                shift 2
                ;;
            -i|--interactive)
                INTERACTIVE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            -*)
                echo "未知选项: $1"
                show_help
                exit 1
                ;;
            *)
                if [[ -z "$PLUGIN_NAME" ]]; then
                    PLUGIN_NAME="$1"
                else
                    echo "错误: 多余的参数 '$1'"
                    show_help
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # 检查必需参数
    if [[ -z "$PLUGIN_NAME" ]]; then
        echo "错误: 必须提供插件名称"
        show_help
        exit 1
    fi

    # 验证插件类型
    if [[ ! "$PLUGIN_TYPE" =~ ^(utility|audio|llm|device)$ ]]; then
        echo "错误: 无效的插件类型 '$PLUGIN_TYPE'"
        echo "支持的类型: utility, audio, llm, device"
        exit 1
    fi

    # 规范化插件名称
    PLUGIN_NAME=$(echo "$PLUGIN_NAME" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]//g')
}

# 交互式输入
interactive_input() {
    echo "=== XiaoZhi Flow 插件创建向导 ==="
    echo

    # 输入插件类型
    echo "请选择插件类型:"
    echo "1) utility - 通用功能插件"
    echo "2) audio   - 音频处理插件"
    echo "3) llm     - 大模型插件"
    echo "4) device  - 设备控制插件"
    echo -n "请输入选择 (1-4): "
    read -r choice

    case $choice in
        1) PLUGIN_TYPE="utility" ;;
        2) PLUGIN_TYPE="audio" ;;
        3) PLUGIN_TYPE="llm" ;;
        4) PLUGIN_TYPE="device" ;;
        *) echo "无效选择，使用默认类型: utility"
           PLUGIN_TYPE="utility" ;;
    esac

    # 输入作者
    echo -n "请输入作者姓名 [默认: Your Name]: "
    read -r input
    if [[ -n "$input" ]]; then
        AUTHOR="$input"
    fi

    # 输入描述
    echo -n "请输入插件描述 [可选]: "
    read -r input
    if [[ -n "$input" ]]; then
        DESCRIPTION="$input"
    fi

    echo
    echo "=== 插件信息确认 ==="
    echo "插件名称: $PLUGIN_NAME"
    echo "插件类型: $PLUGIN_TYPE"
    echo "作者: $AUTHOR"
    echo "描述: ${DESCRIPTION:-无}"
    echo
    echo -n "确认创建? (y/n): "
    read -r confirm

    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "已取消创建"
        exit 0
    fi
}

# 生成默认描述
generate_description() {
    if [[ -z "$DESCRIPTION" ]]; then
        case $PLUGIN_TYPE in
            utility)
                DESCRIPTION="XiaoZhi Flow 通用功能插件"
                ;;
            audio)
                DESCRIPTION="XiaoZhi Flow 音频处理插件"
                ;;
            llm)
                DESCRIPTION="XiaoZhi Flow 大模型集成插件"
                ;;
            device)
                DESCRIPTION="XiaoZhi Flow 设备控制插件"
                ;;
        esac
    fi
}

# 创建插件目录结构
create_plugin_structure() {
    PLUGIN_DIR="plugins/$PLUGIN_NAME"
    echo "创建插件目录: $PLUGIN_DIR"

    mkdir -p "$PLUGIN_DIR"/{test,docs,assets}

    # 创建模块文件
    cat > "$PLUGIN_DIR/go.mod" << EOF
module $PLUGIN_NAME

go 1.24

require (
    xiaozhi-server-go v1.0.0
)

replace xiaozhi-server-go => ../../
EOF

    # 创建 README
    cat > "$PLUGIN_DIR/README.md" << EOF
# $PLUGIN_NAME

$DESCRIPTION

## 功能特性

- 支持的工具：
  - 工具1: 描述
  - 工具2: 描述

## 安装

1. 将插件复制到 XiaoZhi Flow 的 plugins 目录
2. 重启 XiaoZhi Flow 服务

## 使用

### 工具调用示例

\`\`\`json
{
  "tool_name": "tool_name",
  "arguments": {
    "param1": "value1"
  }
}
\`\`\`

## 开发

\`\`\`bash
# 构建插件
go build -o $PLUGIN_NAME main.go

# 运行插件
./$PLUGIN_NAME

# 测试插件
go test ./test/
\`\`\`

## 作者

- **作者**: $AUTHOR
- **版本**: 1.0.0
- **许可证**: MIT

## 许可证

MIT License
EOF

    # 创建测试文件
    cat > "$PLUGIN_DIR/test/main_test.go" << EOF
package main

import (
    "context"
    "testing"

    "github.com/hashicorp/go-hclog"
    "github.com/stretchr/testify/assert"

    v1 "xiaozhi-server-go/api/v1"
    sdk "xiaozhi-server-go/internal/plugin/sdk"
)

func Test${PLUGIN_NAME^}_CallTool(t *testing.T) {
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "test",
        Level:  hclog.Debug,
    })

    plugin := New${PLUGIN_NAME^}Plugin(logger)

    // 测试调用
    req := &v1.CallToolRequest{
        ToolName: "your_tool",
        Arguments: map[string]interface{}{
            "param": "value",
        },
    }

    resp := plugin.CallTool(context.Background(), req)
    assert.NotNil(t, resp)
    // 添加具体的断言
}

func Test${PLUGIN_NAME^}_ListTools(t *testing.T) {
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "test",
        Level:  hclog.Debug,
    })

    plugin := New${PLUGIN_NAME^}Plugin(logger)

    resp := plugin.ListTools(context.Background())
    assert.True(t, resp.Success)
    assert.NotEmpty(t, resp.Tools)
}

func Test${PLUGIN_NAME^}_GetToolSchema(t *testing.T) {
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "test",
        Level:  hclog.Debug,
    })

    plugin := New${PLUGIN_NAME^}Plugin(logger)

    resp := plugin.GetToolSchema(context.Background(), &v1.GetToolSchemaRequest{
        ToolName: "your_tool",
    })

    // 根据实际情况调整断言
    assert.NotNil(t, resp)
}
EOF

    # 创建 API 文档
    mkdir -p "$PLUGIN_DIR/docs"
    cat > "$PLUGIN_DIR/docs/api.md" << EOF
# $PLUGIN_NAME API 文档

## 工具列表

### tool_name

**描述**: 工具的详细描述

**参数**:
- \`param1\` (string): 参数描述
- \`param2\` (number): 参数描述

**示例**:
\`\`\`json
{
  "tool_name": "tool_name",
  "arguments": {
    "param1": "value1",
    "param2": 123
  }
}
\`\`\`

**响应**:
\`\`\`json
{
  "success": true,
  "result": {
    "field1": "value1",
    "field2": "value2"
  },
  "output": "处理结果"
}
\`\`\`
EOF

    echo "插件目录结构创建完成"
}

# 生成插件主文件
generate_main_file() {
    local plugin_name_pascal
    plugin_name_pascal=$(echo "$PLUGIN_NAME" | sed 's/\(^.\|-\([a-z]\)/\u\1/g')
    local plugin_class_name="${plugin_name_pascal}Plugin"

    echo "生成插件主文件..."

    case $PLUGIN_TYPE in
        utility)
            generate_utility_plugin "$plugin_class_name"
            ;;
        audio)
            generate_audio_plugin "$plugin_class_name"
            ;;
        llm)
            generate_llm_plugin "$plugin_class_name"
            ;;
        device)
            generate_device_plugin "$plugin_class_name"
            ;;
    esac
}

generate_utility_plugin() {
    local class_name=$1
    cat > "$PLUGIN_DIR/main.go" << EOF
package main

import (
    "context"
    "fmt"
    "github.com/hashicorp/go-hclog"
    "github.com/hashicorp/go-plugin"

    v1 "xiaozhi-server-go/api/v1"
    sdk "xiaozhi-server-go/internal/plugin/sdk"
)

// $class_name 通用功能插件
type $class_name struct {
    sdk.SimplePluginImpl
    logger hclog.Logger
}

// New$class_name 创建插件实例
func New$class_name(logger hclog.Logger) *$class_name {
    info := &v1.PluginInfo{
        ID:          "$PLUGIN_NAME",
        Name:        "$PLUGIN_NAME Plugin",
        Version:     "1.0.0",
        Description: "$DESCRIPTION",
        Author:      "$AUTHOR",
        Type:        v1.PluginTypeUtility,
        Tags:        []string{"utility"},
        Capabilities: []string{"process_data", "format_output"},
    }

    return &$class_name{
        SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
        logger:          logger.Named("$PLUGIN_NAME"),
    }
}

// CallTool 实现工具调用
func (p *$class_name) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
    p.logger.Info("Tool called", "tool", req.ToolName, "args", req.Arguments)

    switch req.ToolName {
    case "process_data":
        return p.processData(ctx, req.Arguments)
    case "format_output":
        return p.formatOutput(ctx, req.Arguments)
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

// processData 数据处理工具
func (p *$class_name) processData(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    data, ok := args["data"].(string)
    if !ok {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: "需要参数 data",
            },
        }
    }

    processedData := fmt.Sprintf("Processed: %s", data)

    p.IncrementCounter("process_data.total")

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "original":  data,
            "processed": processedData,
            "length":    len(data),
        },
        Output: processedData,
    }
}

// format_output 格式化输出工具
func (p *$class_name) formatOutput(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    format, ok := args["format"].(string)
    data, ok := args["data"].(string)
    if !ok {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: "需要参数 format 和 data",
            },
        }
    }

    var output string
    switch format {
    case "json":
        output = fmt.Sprintf(\`{"data": "%s", "processed": true}\`, data)
    case "xml":
        output = fmt.Sprintf(\`<data>%s</data>\`, data)
    case "yaml":
        output = fmt.Sprintf(\`data: "%s"\`, data)
    default:
        output = data
    }

    p.IncrementCounter("format_output.total")

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "format": format,
            "output": output,
        },
        Output: output,
    }
}

// ListTools 列出可用工具
func (p *$class_name) ListTools(ctx context.Context) *v1.ListToolsResponse {
    tools := []*v1.ToolInfo{
        {
            Name:        "process_data",
            Description: "处理数据",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "data": map[string]interface{}{
                        "type":        "string",
                        "description": "要处理的数据",
                    },
                },
                "required": []string{"data"},
            },
        },
        {
            Name:        "format_output",
            Description: "格式化输出",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "format": map[string]interface{}{
                        "type":        "string",
                        "description": "输出格式 (json|xml|yaml)",
                        "enum":        []interface{}{"json", "xml", "yaml"},
                        "default":     "json",
                    },
                    "data": map[string]interface{}{
                        "type":        "string",
                        "description": "要格式化的数据",
                    },
                },
                "required": []string{"format", "data"},
            },
        },
    }

    return &v1.ListToolsResponse{
        Success: true,
        Tools:   tools,
    }
}

// GetToolSchema 获取工具模式
func (p *$class_name) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
    tools := p.ListTools(ctx)
    if !tools.Success {
        return &v1.GetToolSchemaResponse{
            Success: false,
            Error:   tools.Error,
        }
    }

    for _, tool := range tools.Tools {
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
        Name:   "$PLUGIN_NAME",
        Level:  hclog.Info,
        Output: hclog.DefaultOutput,
    })

    // 创建插件实例
    plugin := New$class_name(logger)

    logger.Info("Starting $PLUGIN_NAME Plugin")

    // 服务插件
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.SimpleHandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "plugin": &sdk.SimplePluginRPC{Impl: plugin},
        },
    })
}
EOF
}

generate_audio_plugin() {
    local class_name=$1
    cat > "$PLUGIN_DIR/main.go" << EOF
package main

import (
    "context"
    "fmt"
    "github.com/hashicorp/go-hclog"
    "github.com/hashicorp/go-plugin"

    v1 "xiaozhi-server-go/api/v1"
    sdk "xiaozhi-server-go/internal/plugin/sdk"
)

// $class_name 音频处理插件
type $class_name struct {
    sdk.SimplePluginImpl
    logger hclog.Logger
}

// New$class_name 创建插件实例
func New$class_name(logger hclog.Logger) *$class_name {
    info := &v1.PluginInfo{
        ID:          "$PLUGIN_NAME",
        Name:        "$PLUGIN_NAME Plugin",
        Version:     "1.0.0",
        Description: "$DESCRIPTION",
        Author:      "$AUTHOR",
        Type:        v1.PluginTypeAudio,
        Tags:        []string{"audio"},
        Capabilities: []string{"process_audio", "get_format"},
    }

    return &$class_name{
        SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
        logger:          logger.Named("$PLUGIN_NAME"),
    }
}

// CallTool 实现工具调用
func (p *$class_name) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
    p.logger.Info("Audio tool called", "tool", req.ToolName)

    switch req.ToolName {
    case "process_audio":
        return p.processAudio(ctx, req.Arguments)
    case "get_format":
        return p.getFormat(ctx, req.Arguments)
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

// processAudio 音频处理
func (p *$class_name) processAudio(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    audioData, ok := args["audio_data"].(string)
    if !ok {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: "需要参数 audio_data",
            },
        }
    }

    format, _ := args["format"].(string)
    if format == "" {
        format = "auto"
    }

    p.IncrementCounter("process_audio.total")

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "audio_data": audioData,
            "format":     format,
            "length":     len(audioData),
            "processed":  true,
        },
        Output: fmt.Sprintf("音频处理完成，格式: %s，大小: %d 字节", format, len(audioData)),
    }
}

// getFormat 获取音频格式
func (p *$class_name) getFormat(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    audioData, ok := args["audio_data"].(string)
    if !ok {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: "需要参数 audio_data",
            },
        }
    }

    // 简单的格式检测
    var format string
    if len(audioData) > 4 {
        header := audioData[:4]
        switch header {
        case "RIFF":
            format = "wav"
        case "ID3":
            format = "mp3"
        case "OggS":
            format = "ogg"
        default:
            format = "unknown"
        }
    } else {
        format = "unknown"
    }

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "audio_data": audioData,
            "format":     format,
            "length":     len(audioData),
        },
        Output: fmt.Sprintf("音频格式: %s", format),
    }
}

// ListTools 列出可用工具
func (p *$class_name) ListTools(ctx context.Context) *v1.ListToolsResponse {
    tools := []*v1.ToolInfo{
        {
            Name:        "process_audio",
            Description: "处理音频数据",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "audio_data": map[string]interface{}{
                        "type":        "string",
                        "description": "Base64编码的音频数据",
                    },
                    "format": map[string]interface{}{
                        "type":        "string",
                        "description": "音频格式 (wav|mp3|ogg|auto)",
                        "default":     "auto",
                    },
                },
                "required": []string{"audio_data"},
            },
        },
        {
            Name:        "get_format",
            Description: "获取音频格式",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "audio_data": map[string]interface{}{
                        "type":        "string",
                        "description": "Base64编码的音频数据",
                    },
                },
                "required": []string{"audio_data"},
            },
        },
    }

    return &v1.ListToolsResponse{
        Success: true,
        Tools:   tools,
    }
}

// GetToolSchema 获取工具模式
func (p *$class_name) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
    tools := p.ListTools(ctx)
    if !tools.Success {
        return &v1.GetToolSchemaResponse{
            Success: false,
            Error:   tools.Error,
        }
    }

    for _, tool := range tools.Tools {
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
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "$PLUGIN_NAME",
        Level:  hclog.Info,
        Output: hclog.DefaultOutput,
    })

    plugin := New$class_name(logger)
    logger.Info("Starting $PLUGIN_NAME Audio Plugin")

    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.SimpleHandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "plugin": &sdk.SimplePluginRPC{Impl: plugin},
        },
    })
}
EOF
}

generate_llm_plugin() {
    local class_name=$1
    cat > "$PLUGIN_DIR/main.go" << EOF
package main

import (
    "context"
    "fmt"
    "strings"
    "github.com/hashicorp/go-hclog"
    "github.com/hashicorp/go-plugin"

    v1 "xiaozhi-server-go/api/v1"
    sdk "xiaozhi-server-go/internal/plugin/sdk"
)

// $class_name 大模型插件
type $class_name struct {
    sdk.SimplePluginImpl
    logger hclog.Logger
}

// New$class_name 创建插件实例
func New$class_name(logger hclog.Logger) *$class_name {
    info := &v1.PluginInfo{
        ID:          "$PLUGIN_NAME",
        Name:        "$PLUGIN_NAME Plugin",
        Version:     "1.0.0",
        Description: "$DESCRIPTION",
        Author:      "$AUTHOR",
        Type:        v1.PluginTypeLLM,
        Tags:        []string{"llm", "ai"},
        Capabilities: []string{"generate_text", "complete_text"},
    }

    return &$class_name{
        SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
        logger:          logger.Named("$PLUGIN_NAME"),
    }
}

// CallTool 实现工具调用
func (p *$class_name) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
    p.logger.Info("LLM tool called", "tool", req.ToolName)

    switch req.ToolName {
    case "generate_text":
        return p.generateText(ctx, req.Arguments)
    case "complete_text":
        return p.completeText(ctx, req.Arguments)
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

// generateText 文本生成
func (p *$class_name) generateText(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    prompt, ok := args["prompt"].(string)
    if !ok {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: "需要参数 prompt",
            },
        }
    }

    maxTokens, _ := args["max_tokens"].(float64)
    if maxTokens == 0 {
        maxTokens = 100
    }

    // 简单的文本生成逻辑
    words := strings.Fields(prompt)
    if len(words) > 50 {
        maxTokens = 50
    }

    generatedText := fmt.Sprintf("%s [Generated Text - Length: %d, Max Tokens: %.0f]",
        strings.Join(words[:min(len(words), int(maxTokens))], " "),
        len(words), maxTokens)

    p.IncrementCounter("generate_text.total")
    p.RecordHistogram("generate_text.prompt_length", float64(len(prompt)))

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "prompt":      prompt,
            "text":        generatedText,
            "max_tokens":  maxTokens,
            "word_count":  len(words),
        },
        Output: generatedText,
    }
}

// completeText 文本补全
func (p *$class_name) completeText(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    prefix, ok := args["prefix"].(string)
    if !ok {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: "需要参数 prefix",
            },
        }
    }

    suffix, _ := args["suffix"].(string)

    // 简单的文本补全逻辑
    suggestions := []string{
        prefix + " [Suggestion 1]",
        prefix + " [Suggestion 2]",
        prefix + " [Suggestion 3]",
    }

    if suffix != "" {
        for i := range suggestions {
            suggestions[i] += " " + suffix
        }
    }

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "prefix":     prefix,
            "suffix":     suffix,
            "suggestions": suggestions,
        },
        Output: strings.Join(suggestions, "\n"),
    }
}

// min 返回两个数中的较小值
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// ListTools 列出可用工具
func (p *$class_name) ListTools(ctx context.Context) *v1.ListToolsResponse {
    tools := []*v1.ToolInfo{
        {
            Name:        "generate_text",
            Description: "生成文本",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "prompt": map[string]interface{}{
                        "type":        "string",
                        "description": "输入提示",
                    },
                    "max_tokens": map[string]interface{}{
                        "type":        "number",
                        "description": "最大生成长度",
                        "default":     100,
                    },
                },
                "required": []string{"prompt"},
            },
        },
        {
            Name:        "complete_text",
            Description: "文本补全",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "prefix": map[string]interface{}{
                        "type":        "string",
                        "description": "文本前缀",
                    },
                    "suffix": map[string]interface{}{
                        "type":        "string",
                        "description": "文本后缀",
                    },
                },
                "required": []string{"prefix"},
            },
        },
    }

    return &v1.ListToolsResponse{
        Success: true,
        Tools:   tools,
    }
}

// GetToolSchema 获取工具模式
func (p *$class_name) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
    tools := p.ListTools(ctx)
    if !tools.Success {
        return &v1.GetToolSchemaResponse{
            Success: false,
            Error:   tools.Error,
        }
    }

    for _, tool := range tools.Tools {
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
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "$PLUGIN_NAME",
        Level:  hclog.Info,
        Output: hclog.DefaultOutput,
    })

    plugin := New$class_name(logger)
    logger.Info("Starting $PLUGIN_NAME LLM Plugin")

    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.SimpleHandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "plugin": &sdk.SimplePluginRPC{Impl: plugin},
        },
    })
}
EOF
}

generate_device_plugin() {
    local class_name=$1
    cat > "$PLUGIN_DIR/main.go" << EOF
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/hashicorp/go-hclog"
    "github.com/hashicorp/go-plugin"

    v1 "xiaozhi-server-go/api/v1"
    sdk "xiaozhi-server-go/internal/plugin/sdk"
)

// $class_name 设备控制插件
type $class_name struct {
    sdk.SimplePluginImpl
    logger hclog.Logger
}

// New$class_name 创建插件实例
func New$class_name(logger hclog.Logger) *$class_name {
    info := &v1.PluginInfo{
        ID:          "$PLUGIN_NAME",
        Name:        "$PLUGIN_NAME Plugin",
        Version:     "1.0.0",
        Description: "$DESCRIPTION",
        Author:      "$AUTHOR",
        Type:        v1.PluginTypeDevice,
        Tags:        []string{"device", "iot"},
        Capabilities: []string{"control_device", "get_status", "list_devices"},
    }

    return &$class_name{
        SimplePluginImpl: *sdk.NewSimplePlugin(info, logger),
        logger:          logger.Named("$PLUGIN_NAME"),
    }
}

// CallTool 实现工具调用
func (p *$class_name) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
    p.logger.Info("Device tool called", "tool", req.ToolName)

    switch req.ToolName {
    case "control_device":
        return p.controlDevice(ctx, req.Arguments)
    case "get_status":
        return p.getDeviceStatus(ctx, req.Arguments)
    case "list_devices":
        return p.listDevices(ctx, req.Arguments)
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

// controlDevice 控制设备
func (p *$class_name) controlDevice(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    deviceID, ok := args["device_id"].(string)
    if !ok {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: "需要参数 device_id",
            },
        }
    }

    action, ok := args["action"].(string)
    if !ok {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: "需要参数 action",
            },
        }
    }

    // 模拟设备控制
    p.IncrementCounter("control_device.total")
    p.IncrementCounter(fmt.Sprintf("control_device.%s", action))

    result := map[string]interface{}{
        "device_id": deviceID,
        "action":    action,
        "status":    "success",
        "timestamp": time.Now().Unix(),
    }

    return &v1.CallToolResponse{
        Success: true,
        Result:  result,
        Output: fmt.Sprintf("设备 %s 执行操作 %s 成功", deviceID, action),
    }
}

// getDeviceStatus 获取设备状态
func (p *$class_name) getDeviceStatus(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    deviceID, ok := args["device_id"].(string)
    if !ok {
        return &v1.CallToolResponse{
            Success: false,
            Error: &v1.ErrorInfo{
                Code:    "INVALID_ARGS",
                Message: "需要参数 device_id",
            },
        }
    }

    // 模拟设备状态
    status := map[string]interface{}{
        "device_id": deviceID,
        "online":   true,
        "battery":  85,
        "signal":  -45,
        "last_seen": time.Now(),
        "sensors": map[string]interface{}{
            "temperature": 22.5,
            "humidity":    65.2,
        },
    }

    p.IncrementCounter("get_status.total")

    return &v1.CallToolResponse{
        Success: true,
        Result:  status,
        Output: fmt.Sprintf("设备 %s 状态: 在线，电量: %d%%", deviceID, status["battery"]),
    }
}

// listDevices 列出设备
func (p *$class_name) listDevices(ctx context.Context, args map[string]interface{}) *v1.CallToolResponse {
    // 模拟设备列表
    devices := []map[string]interface{}{
        {
            "device_id":   "sensor_001",
            "name":       "温度传感器",
            "type":       "sensor",
            "online":     true,
            "location":   "客厅",
        },
        {
            "device_id":   "switch_001",
            "name":       "智能开关",
            "type":       "actuator",
            "online":     true,
            "location":   "卧室",
        },
        {
            "device_id":   "camera_001",
            "name":       "网络摄像头",
            "type":       "camera",
            "online":     false,
            "location":   "门口",
        },
    }

    p.IncrementCounter("list_devices.total")

    return &v1.CallToolResponse{
        Success: true,
        Result: map[string]interface{}{
            "devices": devices,
            "count":   len(devices),
            "timestamp": time.Now().Unix(),
        },
        Output: fmt.Sprintf("找到 %d 个设备", len(devices)),
    }
}

// ListTools 列出可用工具
func (p *class_name) ListTools(ctx context.Context) *v1.ListToolsResponse {
    tools := []*v1.ToolInfo{
        {
            Name:        "control_device",
            Description: "控制设备",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "device_id": map[string]interface{}{
                        "type":        "string",
                        "description": "设备ID",
                    },
                    "action": map[string]interface{}{
                        "type":        "string",
                        "description": "控制动作 (on|off|toggle)",
                    },
                },
                "required": []string{"device_id", "action"},
            },
        },
        {
            Name:        "get_status",
            Description: "获取设备状态",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "device_id": map[string]interface{}{
                        "type":        "string",
                        "description": "设备ID",
                    },
                },
                "required": []string{"device_id"},
            },
        },
        {
            Name:        "list_devices",
            Description: "列出所有设备",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "type": map[string]interface{}{
                        "type":        "string",
                        "description": "设备类型过滤 (sensor|actuator|camera)",
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
func (p *class_name) GetToolSchema(ctx context.Context, req *v1.GetToolSchemaRequest) *v1.GetToolSchemaResponse {
    tools := p.ListTools(ctx)
    if !tools.Success {
        return &v1.GetToolSchemaResponse{
            Success: false,
            Error:   tools.Error,
        }
    }

    for _, tool := range tools.Tools {
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
    logger := hclog.New(&hclog.LoggerOptions{
        Name:   "$PLUGIN_NAME",
        Level:  hclog.Info,
        Output: hclog.DefaultOutput,
    })

    plugin := New$class_name(logger)
    logger.Info("Starting $PLUGIN_NAME Device Plugin")

    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.SimpleHandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "plugin": &sdk.SimplePluginRPC{Impl: plugin},
        },
    })
}
EOF
}

# 生成配置文件
generate_config() {
    cat > "$PLUGIN_DIR/plugin.yaml" << EOF
name: $PLUGIN_NAME Plugin
version: 1.0.0
description: $DESCRIPTION
author: $AUTHOR
type: $PLUGIN_TYPE
tags:
EOF

    # 根据插件类型添加特定标签
    case $PLUGIN_TYPE in
        utility)
            cat >> "$PLUGIN_DIR/plugin.yaml" << EOF
  - utility
  - tools
EOF
            ;;
        audio)
            cat >> "$PLUGIN_DIR/plugin.yaml" << EOF
  - audio
  - processing
EOF
            ;;
        llm)
            cat >> "$PLUGIN_DIR/plugin.yaml" << EOF
  - llm
  - ai
  - generation
EOF
            ;;
        device)
            cat >> "$PLUGIN_DIR/plugin.yaml" << EOF
  - device
  - iot
  - control
EOF
            ;;
    esac

    cat >> "$PLUGIN_DIR/plugin.yaml" << EOF
capabilities:
EOF

    # 根据插件类型添加特定能力
    case $PLUGIN_TYPE in
        utility)
            cat >> "$PLUGIN_DIR/plugin.yaml" << EOF
  - process_data
  - format_output
EOF
            ;;
        audio)
            cat >> "$PLUGIN_DIR/plugin.yaml" << EOF
  - process_audio
  - get_format
EOF
            ;;
        llm)
            cat >> "$PLUGIN_DIR/plugin.yaml" << EOF
  - generate_text
  - complete_text
EOF
            ;;
        device)
            cat >> "$PLUGIN_DIR/plugin.yaml" << EOF
  - control_device
  - get_status
  - list_devices
EOF
            ;;
    esac

    cat >> "$PLUGIN_DIR/plugin.yaml" << EOF
metadata:
  language: go
  created_at: "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  framework: "xiaozhi-flow-plugin-sdk"

deployment:
  type: local_binary
  path: ./main.go
  resources:
    max_memory: "64Mi"
    max_cpu: "100m"
  timeout: 10s
  retry_count: 3

config:
  log_level: "info"

environment:
  PLUGIN_LOG_LEVEL: "info"

enabled: true
EOF
}

# 创建构建脚本
generate_build_script() {
    cat > "$PLUGIN_DIR/build.sh" << 'EOF'
#!/bin/bash

set -e

# 构建配置
PLUGIN_NAME="$PLUGIN_NAME"
VERSION="1.0.0"
BUILD_DIR="build"
DIST_DIR="dist"

# 清理旧的构建
rm -rf $BUILD_DIR $DIST_DIR
mkdir -p $BUILD_DIR $DIST_DIR

echo "构建 $PLUGIN_NAME v$VERSION..."

# 构建插件
go build -ldflags "-X main.version=$VERSION" -o $BUILD_DIR/$PLUGIN_NAME main.go

# 复制文件
cp plugin.yaml $BUILD_DIR/
cp README.md $BUILD_DIR/ 2>/dev/null || true
cp -r docs $BUILD_DIR/ 2>/dev/null || true
cp -r assets $BUILD_DIR/ 2>/dev/null || true

# 创建分发包
cd $BUILD_DIR
tar -czf ../$DIST_DIR/${PLUGIN_NAME}-${VERSION}.tar.gz *
cd ..

echo "构建完成: $DIST_DIR/${PLUGIN_NAME}-${VERSION}.tar.gz"
EOF
    chmod +x "$PLUGIN_DIR/build.sh"

    # 创建测试脚本
    cat > "$PLUGIN_DIR/test.sh" << 'EOF'
#!/bin/bash

set -e

echo "运行 $PLUGIN_NAME 测试..."

# 运行单元测试
go test ./test/ -v

# 集成测试（如果有）
if [ -f "integration_test.go" ]; then
    echo "运行集成测试..."
    go test -run Integration .
fi

echo "测试完成"
EOF
    chmod +x "$PLUGIN_DIR/test.sh"

    # 创建安装脚本
    cat > "$PLUGIN_DIR/install.sh" << 'EOF
#!/bin/bash

set -e

PLUGIN_NAME="$PLUGIN_NAME"
INSTALL_DIR="${1:-../../plugins}"

echo "安装 $PLUGIN_NAME 到 $INSTALL_DIR"

# 创建安装目录
mkdir -p "$INSTALL_DIR"

# 复制文件
cp main.go "$INSTALL_DIR/"
cp plugin.yaml "$INSTALL_DIR/"

# 复制其他文件（如果存在）
if [ -d "docs" ]; then
    cp -r docs "$INSTALL_DIR/"
fi

if [ -d "assets" ]; then
    cp -r assets "$INSTALL_DIR/"
fi

# 设置权限
chmod +x "$INSTALL_DIR/main.go" 2>/dev/null || true

echo "$PLUGIN_NAME 安装完成"
echo "插件位置: $INSTALL_DIR/"
EOF
    chmod +x "$PLUGIN_DIR/install.sh"

    # 创建开发脚本
    cat > "$PLUGIN_DIR/dev.sh" << 'EOF
#!/bin/bash

set -e

echo "开发模式运行 $PLUGIN_NAME"

# 设置环境变量
export PLUGIN_LOG_LEVEL=debug
export PLUGIN_RELOAD=true

# 运行插件
go run main.go
EOF
    chmod +x "$PLUGIN_DIR/dev.sh"
}

# 创建完成消息
create_completion_message() {
    echo
    echo "🎉 插件创建成功！"
    echo
    echo "插件信息:"
    echo "  名称: $PLUGIN_NAME"
    echo "  类型: $PLUGIN_TYPE"
    echo "  作者: $AUTHOR"
    echo "  描述: $DESCRIPTION"
    echo
    echo "插件目录: plugins/$PLUGIN_NAME/"
    echo
    echo "下一步操作:"
    echo "1. 进入插件目录: cd plugins/$PLUGIN_NAME"
    echo "2. 构建插件: ./build.sh"
    echo "3. 测试插件: ./test.sh"
    echo "4. 安装插件: ./install.sh"
    echo "5. 开发模式: ./dev.sh"
    echo
    echo "或者直接运行: go run main.go"
    echo
    echo "详细文档请参考: docs/plugin-development.md"
    echo
}

# 主函数
main() {
    parse_args "$@"

    # 交互式模式
    if [[ "$INTERACTIVE" == true ]]; then
        interactive_input
    fi

    # 生成描述
    generate_description

    # 创建目录结构
    create_plugin_structure

    # 生成配置文件
    generate_config

    # 生成主文件
    generate_main_file

    # 生成脚本
    generate_build_script

    # 显示完成消息
    create_completion_message
}

# 执行主函数
main "$@"
EOF

chmod +x "$0"