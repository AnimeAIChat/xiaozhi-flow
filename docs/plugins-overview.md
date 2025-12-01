# XiaoZhi Flow 插件系统概览

## 🎯 系统概述

XiaoZhi Flow插件系统是一个强大、灵活的插件架构，支持通过HashiCorp go-plugin框架和gRPC通信协议扩展系统功能。该系统采用Clean Architecture设计，提供完整的生命周期管理、安全隔离和性能监控。

## 🏗️ 架构设计

### 核心组件

```
┌─────────────────────────────────────────────────────────────┐
│                    XiaoZhi Flow 主系统                      │
├─────────────────────────────────────────────────────────────┤
│                      Bootstrap层                            │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Plugin Manager                          │    │
│  │  • Discovery      • Registry                         │    │
│  │  • Runtime        • Lifecycle                        │    │
│  │  • Health Check   • Metrics                          │    │
│  └─────────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│                      SDK层                                  │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Plugin SDK                              │    │
│  │  • Base Classes  • gRPC Client/Server               │    │
│  │  • Metrics       • Logging                          │    │
│  │  • Config        • Utilities                        │    │
│  └─────────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│                      插件层                                  │
│  ┌──────────────┬──────────────┬──────────────┬─────────┐    │
│  │   ASR插件    │   TTS插件    │   LLM插件    │   ...   │    │
│  │  语音识别     │ 文本转语音    │ 大语言模型   │  更多插件 │    │
│  └──────────────┴──────────────┴──────────────┴─────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### 设计原则

- **松耦合**: 插件与主系统通过标准化接口通信
- **可扩展**: 支持动态加载和卸载插件
- **安全隔离**: 每个插件运行在独立的进程中
- **高性能**: 基于gRPC的高效通信
- **可观测**: 完整的监控和日志记录

## 🔌 已实现的插件

### 1. ASR Plugin (语音识别插件)

**功能特性:**
- 语音转文字功能
- 支持多种音频格式 (WAV, MP3, FLAC, AAC等)
- 多语言支持 (中文、英文、日文、韩文等)
- 批量处理能力
- 音频格式检测
- 置信度评分

**工具列表:**
- `speech_to_text`: 语音转文字
- `detect_audio_format`: 检测音频格式
- `batch_transcribe`: 批量转录
- `get_supported_formats`: 获取支持的格式

**使用场景:**
- 语音助手交互
- 语音笔记转录
- 会议记录
- 音频内容分析

**技术特点:**
- 支持实时流式识别
- 智能噪声过滤
- 说话人分离
- 时间戳标记

### 2. TTS Plugin (文本转语音插件)

**功能特性:**
- 文本转语音功能
- 多种语音选择 (不同性别、年龄、口音)
- 多种音频格式输出
- 参数调节 (语速、音调、音量)
- 批量合成
- 文本验证

**工具列表:**
- `text_to_speech`: 文本转语音
- `get_available_voices`: 获取可用语音
- `synthesize_batch`: 批量合成
- `validate_text`: 验证文本
- `get_supported_formats`: 获取支持的格式

**使用场景:**
- 语音助手回复
- 有声读物制作
- 导航语音播报
- 自动化通知

**技术特点:**
- 神经网络语音合成
- SSML标记语言支持
- 情感语音合成
- 实时流式输出

### 3. LLM Plugin (大语言模型插件)

**功能特性:**
- 聊天对话完成
- 文本补全
- 多模型支持 (OpenAI, Anthropic, Azure等)
- Token管理和成本控制
- 提示验证
- 模型信息查询

**工具列表:**
- `chat_completion`: 聊天对话
- `text_completion`: 文本补全
- `get_available_models`: 获取可用模型
- `count_tokens`: 计算token数量
- `validate_prompt`: 验证提示
- `get_model_info`: 获取模型信息

**使用场景:**
- 智能对话系统
- 内容生成
- 文本分析和处理
- 代码生成和解释

**技术特点:**
- 多轮对话支持
- 上下文理解
- 函数调用能力
- 流式输出支持

## 📊 插件性能指标

### 监控维度

每个插件都内置了丰富的性能指标监控：

**系统级指标:**
- 插件启动/关闭次数
- 健康检查状态
- 内存和CPU使用情况
- 错误率统计

**业务级指标:**
- API调用次数和成功率
- 响应时间分布
- 数据处理量统计
- 成本和使用配额

**自定义指标:**
- ASR: 识别准确率、音频处理时长
- TTS: 合成时长、音频文件大小
- LLM: Token使用量、模型调用成本

### 指标查看方式

```bash
# 通过API查看插件指标
curl -X POST http://localhost:8080/api/v1/plugins/{plugin-id}/metrics

# 查看所有插件概览
curl -X GET http://localhost:8080/api/v1/plugins
```

## 🛠️ 开发指南

### 快速创建新插件

1. **使用创建脚本:**
```bash
./scripts/create-plugin.sh
```

2. **手动创建:**
```bash
mkdir -p plugins/my-plugin
cd plugins/my-plugin
# 创建 main.go, plugin.yaml, README.md
```

3. **基础结构:**
```go
package main

import (
    "github.com/hashicorp/go-plugin"
    v1 "xiaozhi-server-go/api/v1"
    sdk "xiaozhi-server-go/internal/plugin/sdk"
)

type MyPlugin struct {
    sdk.SimplePluginImpl
    logger hclog.Logger
}

func (p *MyPlugin) CallTool(ctx context.Context, req *v1.CallToolRequest) *v1.CallToolResponse {
    // 实现工具调用逻辑
}
```

### 插件开发最佳实践

1. **错误处理:**
   - 使用标准的错误代码和消息
   - 提供详细的错误信息
   - 记录错误日志用于调试

2. **性能优化:**
   - 使用连接池管理外部资源
   - 实现适当的缓存策略
   - 控制并发数量避免过载

3. **安全考虑:**
   - 验证输入参数
   - 实现访问控制
   - 避免敏感信息泄露

4. **可观测性:**
   - 记录关键业务指标
   - 使用结构化日志
   - 实现健康检查接口

## 🚀 部署和运行

### 环境要求

- Go 1.19+
- HashiCorp go-plugin
- gRPC相关依赖

### 启动顺序

1. **启动主系统:**
```bash
go run cmd/xiaozhi-server/main.go
```

2. **插件自动发现:**
   - 系统会自动扫描 `plugins/` 目录
   - 加载所有可用的插件配置
   - 启动已启用的插件

3. **验证部署:**
```bash
# 检查插件状态
curl -X GET http://localhost:8080/api/v1/plugins

# 测试插件功能
curl -X POST http://localhost:8080/api/v1/plugins/asr-plugin/tools/call \
  -H "Content-Type: application/json" \
  -d '{"tool_name": "get_supported_formats", "arguments": {}}'
```

### 配置管理

**全局配置 (`config/plugins.yaml`):**
```yaml
plugins:
  discovery:
    paths: ["plugins/"]
    auto_reload: true

  runtime:
    default_timeout: 30s
    max_memory: "512Mi"
    max_cpu: "200m"

  security:
    enable_sandbox: true
    allowed_domains: ["localhost"]
```

**插件配置 (`plugins/*/plugin.yaml`):**
```yaml
name: My Plugin
version: 1.0.0
type: utility
enabled: true
deployment:
  type: local_binary
  resources:
    max_memory: "256Mi"
```

## 📚 文档资源

### 开发文档
- [插件开发指南](plugin-development.md) - 详细的开发指南
- [插件快速开始](plugin-quickstart.md) - 5分钟创建第一个插件
- [API参考文档](api-reference.md) - 完整的API文档

### 插件示例
- [ASR插件文档](../plugins/examples/asr/README.md) - 语音识别插件
- [TTS插件文档](../plugins/examples/tts/README.md) - 文本转语音插件
- [LLM插件文档](../plugins/examples/llm/README.md) - 大语言模型插件
- [Hello World插件](../plugins/examples/hello/README.md) - 基础示例

### 工具和脚本
- [插件创建脚本](../scripts/create-plugin.sh) - 自动化插件创建工具
- [测试脚本](../scripts/test-plugins.sh) - 批量测试插件

## 🔮 未来规划

### 短期目标
- [ ] 实现统一网关，集成插件、MCP和Provider
- [ ] 添加插件系统的Web管理界面
- [ ] 完善插件市场和分发机制
- [ ] 增强安全沙箱功能

### 中期目标
- [ ] 支持更多插件类型 (图像处理、设备控制等)
- [ ] 实现插件热更新
- [ ] 添加插件依赖管理
- [ ] 实现分布式插件部署

### 长期目标
- [ ] 构建插件生态系统
- [ ] 支持多语言插件开发
- [ ] 实现插件版本管理
- [ ] 添加插件性能分析和优化工具

## 🤝 贡献指南

### 如何贡献

1. **报告问题:**
   - 在GitHub上提交Issue
   - 提供详细的错误信息和复现步骤

2. **开发新插件:**
   - Fork项目
   - 创建插件目录和代码
   - 编写测试和文档
   - 提交Pull Request

3. **改进现有功能:**
   - 修复Bug
   - 添加新功能
   - 优化性能
   - 完善文档

### 代码规范

- 遵循Go语言最佳实践
- 使用有意义的变量和函数名
- 添加适当的注释和文档
- 确保所有代码都有测试覆盖

## 🆘 支持和帮助

### 获取帮助
- 查看文档: `docs/` 目录下的详细文档
- 运行测试: 使用提供的测试脚本
- 查看示例: 参考现有插件的实现

### 常见问题

**Q: 插件启动失败怎么办？**
A: 检查插件配置文件，确保依赖已安装，查看错误日志。

**Q: 如何调试插件？**
A: 启用详细日志模式，使用测试脚本，检查gRPC连接。

**Q: 插件性能如何优化？**
A: 监控性能指标，优化算法，使用缓存，调整资源配置。

**Q: 如何确保插件安全？**
A: 验证输入，使用沙箱，限制资源访问，定期更新依赖。

## 📈 总结

XiaoZhi Flow插件系统提供了一个强大、灵活、可扩展的架构，让开发者可以轻松地为系统添加新功能。通过标准化的接口、完善的工具链和丰富的文档，插件开发变得简单而高效。

无论是简单的工具扩展还是复杂的AI模型集成，插件系统都能提供必要的支持和保障。我们期待社区贡献更多的插件，共同构建丰富的生态系统。

---

**更多信息:**
- 项目主页: https://github.com/kalicyh/xiaozhi-flow
- 文档站点: https://xiaozhi-flow.dev
- 社区讨论: https://github.com/kalicyh/xiaozhi-flow/discussions