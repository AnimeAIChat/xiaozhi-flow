# LLM Plugin - 大语言模型集成插件

## 🎯 概述

LLM插件是一个功能强大的大语言模型集成插件，支持多种LLM提供商和服务。该插件提供了统一的API接口，让开发者可以轻松地在应用中集成各种AI模型。

## ✨ 功能特性

### 核心功能
- **聊天完成**: 支持多轮对话和上下文理解
- **文本完成**: 传统的文本补全功能
- **多模型支持**: 集成OpenAI、Anthropic、Azure等主流提供商
- **参数控制**: 支持温度、top_p、max_tokens等参数调节

### 高级功能
- **Token管理**: 智能计算和预测token使用量
- **成本控制**: 内置成本估算和预算管理
- **提示验证**: 智能验证和优化输入提示
- **缓存机制**: 提高响应速度，减少重复请求
- **流式输出**: 支持实时流式响应
- **函数调用**: 支持工具调用和功能扩展

## 🚀 快速开始

### 1. 构建插件

```bash
# 进入插件目录
cd plugins/examples/llm

# 构建插件
go build -o llm-plugin main.go

# 或者直接运行
go run main.go
```

### 2. 配置环境变量

```bash
# OpenAI
export OPENAI_API_KEY="your-openai-api-key"

# Anthropic
export ANTHROPIC_API_KEY="your-anthropic-api-key"

# Azure OpenAI
export AZURE_OPENAI_API_KEY="your-azure-api-key"
export AZURE_OPENAI_ENDPOINT="your-azure-endpoint"
```

### 3. 运行测试

```bash
# 运行功能测试
go run test_llm.go
```

### 4. 启动插件

```bash
# 启动插件服务
./llm-plugin
```

## 📋 支持的模型

### OpenAI模型
- **GPT-4**: 最强大的语言模型，适合复杂任务
- **GPT-4 Turbo**: 高性能多模态模型，支持视觉理解
- **GPT-3.5 Turbo**: 快速高效的对话模型
- **Text Davinci 003**: 强大的文本生成模型

### Anthropic模型
- **Claude 3 Opus**: 最强大的Claude模型
- **Claude 3 Sonnet**: 平衡性能的Claude模型
- **Claude 3 Haiku**: 快速响应的Claude模型

### Azure OpenAI模型
- **Azure GPT-4**: 企业级GPT-4部署
- **Azure GPT-3.5 Turbo**: 企业级GPT-3.5部署

### 本地模型
- **LLaMA 2 7B**: 本地部署7B参数模型
- **LLaMA 2 13B**: 本地部署13B参数模型

## 🛠️ 使用方法

### 1. 基础聊天对话

```go
request := &v1.CallToolRequest{
    ToolName: "chat_completion",
    Arguments: map[string]interface{}{
        "model": "gpt-3.5-turbo",
        "messages": []map[string]interface{}{
            {
                "role": "system",
                "content": "你是一个友好的AI助手。",
            },
            {
                "role": "user",
                "content": "你好，请介绍一下你自己。",
            },
        },
        "max_tokens":   500,
        "temperature":  0.7,
        "top_p":       1.0,
        "stream":      false,
    },
}

response := plugin.CallTool(ctx, request)
if response.Success {
    result := response.Result.(map[string]interface{})
    choices := result["choices"].([]map[string]interface{})
    message := choices[0]["message"].(map[string]interface{})
    content := message["content"].(string)

    fmt.Printf("AI回复: %s\n", content)
}
```

### 2. 多轮对话

```go
conversation := []map[string]interface{}{
    {"role": "system", "content": "你是一个专业的技术顾问。"},
    {"role": "user", "content": "我想学习Python编程，有什么建议吗？"},
    {"role": "assistant", "content": "学习Python是个很好的选择！建议你从基础语法开始..."},
    {"role": "user", "content": "你能推荐一些适合初学者的项目吗？"},
}

request := &v1.CallToolRequest{
    ToolName: "chat_completion",
    Arguments: map[string]interface{}{
        "model": "gpt-3.5-turbo",
        "messages": conversation,
        "max_tokens": 300,
    },
}
```

### 3. 文本补全

```go
request := &v1.CallToolRequest{
    ToolName: "text_completion",
    Arguments: map[string]interface{}{
        "prompt": "人工智能的发展历程可以追溯到",
        "model": "text-davinci-003",
        "max_tokens": 200,
        "temperature": 0.7,
    },
}
```

### 4. 获取可用模型

```go
request := &v1.CallToolRequest{
    ToolName: "get_available_models",
    Arguments: map[string]interface{}{
        "provider": "openai",  // 可选：openai, anthropic, azure, local
        "type": "chat",        // 可选：chat, completion
    },
}
```

### 5. Token计算和成本估算

```go
request := &v1.CallToolRequest{
    ToolName: "count_tokens",
    Arguments: map[string]interface{}{
        "messages": []map[string]interface{}{
            {"role": "user", "content": "你好，这是测试文本"},
        },
    },
}
```

### 6. 提示验证

```go
request := &v1.CallToolRequest{
    ToolName: "validate_prompt",
    Arguments: map[string]interface{}{
        "messages": []map[string]interface{}{
            {"role": "system", "content": "你是一个有用的助手"},
            {"role": "user", "content": "请帮我解释量子计算"},
        },
        "model": "gpt-3.5-turbo",
    },
}
```

## ⚙️ 配置选项

### 插件配置 (plugin.yaml)

```yaml
# 默认模型配置
default_model: "gpt-3.5-turbo"
default_max_tokens: 1000
default_temperature: 0.7

# OpenAI配置
openai:
  api_key: ""
  organization: ""
  base_url: "https://api.openai.com/v1"
  timeout: 60s
  max_retries: 3

# Anthropic配置
anthropic:
  api_key: ""
  base_url: "https://api.anthropic.com"
  timeout: 60s

# Azure OpenAI配置
azure:
  api_key: ""
  endpoint: ""
  api_version: "2023-12-01-preview"
  deployment_name: ""

# 限制和配额
limits:
  max_messages_per_request: 50
  max_tokens_per_request: 4000
  max_requests_per_minute: 60
  max_tokens_per_minute: 40000

# 成本控制
cost_control:
  daily_budget: 10.0
  cost_per_1k_tokens:
    openai:
      "gpt-4": 0.03
      "gpt-3.5-turbo": 0.001
```

### 环境变量

```bash
# 基础配置
PLUGIN_LOG_LEVEL=info

# OpenAI
OPENAI_API_KEY=your_openai_api_key
OPENAI_ORGANIZATION=your_organization_id

# Anthropic
ANTHROPIC_API_KEY=your_anthropic_api_key

# Azure OpenAI
AZURE_OPENAI_API_KEY=your_azure_api_key
AZURE_OPENAI_ENDPOINT=your_azure_endpoint
AZURE_OPENAI_DEPLOYMENT_NAME=your_deployment_name

# 本地模型
LOCAL_LLM_BASE_URL=http://localhost:8080
LOCAL_LLM_MODEL_PATH=/path/to/model

# 成本控制
DAILY_BUDGET=10.0
```

## 📊 性能指标

插件内置了丰富的性能指标监控：

- **llm.calls.total**: 总调用次数
- **llm.calls.success**: 成功调用次数
- **llm.calls.unknown**: 未知工具调用次数
- **llm.errors.completion**: 完成生成错误次数
- **llm.completion_duration**: 完成生成时长分布
- **llm.tokens.input**: 输入token总数
- **llm.tokens.output**: 输出token总数
- **llm.models_list.calls**: 模型列表查询次数
- **llm.count_tokens.calls**: token计算调用次数
- **llm.validate_prompt.calls**: 提示验证调用次数
- **llm.model_info.calls**: 模型信息查询次数
- **llm.text_completion.calls**: 文本完成调用次数

## 🔧 错误处理

插件提供了完善的错误处理机制：

```go
// 常见错误代码
- INVALID_ARGUMENT: 参数错误
- COMPLETION_ERROR: 完成生成失败
- MODEL_NOT_FOUND: 模型不存在
- CONTEXT_TOO_LONG: 上下文过长
- TOKEN_LIMIT_EXCEEDED: Token限制超限
- RATE_LIMITED: 调用频率限制
- QUOTA_EXCEEDED: 配额超限
- PROVIDER_ERROR: 提供商服务错误
- TIMEOUT: 请求超时
- AUTHENTICATION_FAILED: 认证失败
```

## 💡 最佳实践

### 1. 提示工程
- 使用清晰、具体的指令
- 添加系统消息定义角色
- 提供示例和上下文
- 避免歧义和模糊表达

### 2. 成本优化
- 选择合适的模型（根据任务复杂度）
- 控制max_tokens参数
- 使用缓存减少重复请求
- 监控token使用情况

### 3. 性能优化
- 合理设置temperature参数
- 使用批处理减少请求次数
- 启用缓存机制
- 选择合适的提供商

### 4. 安全考虑
- 验证和过滤用户输入
- 避免敏感信息泄露
- 设置访问频率限制
- 监控异常使用模式

## 🎛️ 参数说明

### Temperature (温度)
- **0.0-0.3**: 更确定、更一致的输出
- **0.7-1.0**: 平衡的创造性和一致性
- **1.0-2.0**: 更随机、更有创造性的输出

### Top_p
- **0.1**: 选择最可能的token
- **0.5**: 中等多样性
- **1.0**: 全部可能的token

### Max Tokens
- 根据需求设置合适的长度
- 考虑成本和响应时间
- 为上下文留出空间

## 🧪 测试

运行完整的测试套件：

```bash
# 运行所有测试
go test ./...

# 运行基准测试
go test -bench=. ./...

# 运行功能测试
go run test_llm.go
```

测试覆盖：
- 聊天完成功能
- 文本完成功能
- 模型管理
- Token计算
- 提示验证
- 错误处理
- 性能指标

## 🔌 集成示例

### 与语音助手集成

```go
// ASR -> LLM -> TTS 流程
func processVoiceToText(audioData []byte) string {
    // 1. ASR识别语音
    asrResult := callASRPlugin(audioData)
    userText := asrResult["text"].(string)

    // 2. LLM处理文本
    llmRequest := &v1.CallToolRequest{
        ToolName: "chat_completion",
        Arguments: map[string]interface{}{
            "messages": []map[string]interface{}{
                {"role": "user", "content": userText},
            },
        },
    }
    llmResult := callLLMPlugin(llmRequest)
    response := llmResult["choices"].([]map[string]interface{})[0]["message"].(map[string]interface{})["content"].(string)

    // 3. TTS生成语音
    ttsRequest := &v1.CallToolRequest{
        ToolName: "text_to_speech",
        Arguments: map[string]interface{}{
            "text": response,
        },
    }
    ttsResult := callTTSPlugin(ttsRequest)

    return ttsResult["audio_data"].(string)
}
```

## 🚀 高级功能

### 1. 流式输出

```go
request := &v1.CallToolRequest{
    ToolName: "chat_completion",
    Arguments: map[string]interface{}{
        "messages": []map[string]interface{}{
            {"role": "user", "content": "请写一个长故事"},
        },
        "stream": true,
    },
}
```

### 2. 函数调用

```go
request := &v1.CallToolRequest{
    ToolName: "chat_completion",
    Arguments: map[string]interface{}{
        "messages": []map[string]interface{}{
            {"role": "user", "content": "现在北京的天气怎么样？"},
        },
        "functions": []map[string]interface{}{
            {
                "name": "get_weather",
                "description": "获取指定城市的天气信息",
                "parameters": map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "city": map[string]interface{}{
                            "type": "string",
                            "description": "城市名称",
                        },
                    },
                    "required": []string{"city"},
                },
            },
        },
    },
}
```

### 3. 嵌入向量

```go
// 生成文本嵌入向量
embeddings, _ := plugin.CallTool(ctx, &v1.CallToolRequest{
    ToolName: "create_embeddings",
    Arguments: map[string]interface{}{
        "input": []string{"这是一个测试文本"},
        "model": "text-embedding-ada-002",
    },
})
```

## 🛡️ 安全和合规

### 数据隐私
- 支持数据加密传输
- 可配置数据保留策略
- 支持私有化部署

### 访问控制
- API密钥认证
- 请求频率限制
- IP白名单支持

### 内容安全
- 内置内容过滤
- 敏感信息检测
- 输出内容审核

## 🔮 未来规划

- 支持更多LLM提供商
- 添加更多本地模型支持
- 实现更智能的缓存策略
- 添加模型微调功能
- 支持多模态输入（图像、音频）
- 增强的函数调用能力

## 🤝 贡献指南

欢迎提交Issue和Pull Request来改进这个插件：

1. Fork项目
2. 创建特性分支
3. 提交更改
4. 创建Pull Request

## 📄 许可证

本插件遵循项目整体许可证。

## 🆘 故障排除

### 常见问题

**Q: API调用失败？**
A: 检查API密钥配置、网络连接和配额限制。

**Q: 响应速度慢？**
A: 尝试更小的模型，启用缓存，或使用本地部署。

**Q: Token限制错误？**
A: 减少输入长度，或使用支持更大上下文的模型。

**Q: 成本过高？**
A: 选择经济模型，控制token使用，启用缓存。

**Q: 模型响应不符合预期？**
A: 优化提示词，调整temperature参数，添加系统消息。

## 📚 相关文档

- [XiaoZhi Flow 插件开发指南](../../docs/plugin-development.md)
- [插件快速开始](../../docs/plugin-quickstart.md)
- [LLM API参考](../../../docs/llm-api.md)
- [提示工程指南](../../../docs/prompt-engineering.md)
- [成本优化指南](../../../docs/cost-optimization.md)