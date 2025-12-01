# ASR Plugin - 语音识别插件

## 🎯 概述

ASR插件是一个功能强大的语音识别插件，支持多种音频格式转换为文字。该插件集成了现代语音识别技术，可以高效处理各种音频文件。

## ✨ 功能特性

### 核心功能
- **语音转文字**: 支持多种音频格式的语音识别
- **格式检测**: 自动检测音频文件格式和基本信息
- **批量处理**: 支持批量转录多个音频文件
- **多语言支持**: 支持中文、英文、日文、韩文等多种语言

### 高级功能
- **说话人分离**: 支持多说话人场景的语音分离
- **标点符号**: 自动添加标点符号
- **时间戳**: 提供词级别或句子级别的时间戳
- **置信度评分**: 提供识别结果的置信度评分

## 🚀 快速开始

### 1. 构建插件

```bash
# 进入插件目录
cd plugins/examples/asr

# 构建插件
go build -o asr-plugin main.go

# 或者直接运行
go run main.go
```

### 2. 运行测试

```bash
# 运行功能测试
go run test_asr.go
```

### 3. 启动插件

```bash
# 启动插件服务
./asr-plugin
```

## 📋 支持的格式

### 音频格式
- **WAV**: 无损音频格式（推荐）
- **MP3**: 有损压缩格式
- **FLAC**: 无损压缩格式
- **AAC**: 高效有损压缩格式
- **OGG**: 开源有损压缩格式
- **M4A**: Apple音频格式
- **WMA**: Windows Media音频格式

### 语言支持
- **中文**: zh-CN（简体中文）、zh-TW（繁体中文）
- **英文**: en-US（美式英语）、en-GB（英式英语）
- **日文**: ja-JP
- **韩文**: ko-KR
- *更多语言持续添加中...*

## 🛠️ 使用方法

### 1. 语音转文字

```go
request := &v1.CallToolRequest{
    ToolName: "speech_to_text",
    Arguments: map[string]interface{}{
        "audio_data": "base64编码的音频数据",
        "format":     "wav",           // 可选，默认wav
        "language":   "zh-CN",         // 可选，默认zh-CN
    },
}

response := plugin.CallTool(ctx, request)
if response.Success {
    result := response.Result.(map[string]interface{})
    text := result["text"].(string)
    confidence := result["confidence"].(float64)
    duration := result["duration"].(float64)

    fmt.Printf("识别结果: %s\n", text)
    fmt.Printf("置信度: %.2f%%\n", confidence*100)
    fmt.Printf("时长: %.2f秒\n", duration/1000)
}
```

### 2. 检测音频格式

```go
request := &v1.CallToolRequest{
    ToolName: "detect_audio_format",
    Arguments: map[string]interface{}{
        "audio_data": "base64编码的音频数据",
    },
}

response := plugin.CallTool(ctx, request)
if response.Success {
    result := response.Result.(map[string]interface{})
    format := result["format"].(string)
    duration := result["duration"].(float64)

    fmt.Printf("格式: %s\n", format)
    fmt.Printf("预计时长: %.2f秒\n", duration/1000)
}
```

### 3. 批量转录

```go
audioFiles := []map[string]interface{}{
    {
        "filename":   "file1.wav",
        "audio_data": "base64数据1",
    },
    {
        "filename":   "file2.mp3",
        "audio_data": "base64数据2",
    },
}

request := &v1.CallToolRequest{
    ToolName: "batch_transcribe",
    Arguments: map[string]interface{}{
        "audio_files": audioFiles,
        "language":    "zh-CN",
    },
}

response := plugin.CallTool(ctx, request)
if response.Success {
    results := response.Result["results"].([]map[string]interface{})
    for _, result := range results {
        if result["success"].(bool) {
            fmt.Printf("文件 %s: %s\n", result["filename"], result["text"])
        } else {
            fmt.Printf("文件 %s 失败: %s\n", result["filename"], result["error"])
        }
    }
}
```

### 4. 获取支持信息

```go
request := &v1.CallToolRequest{
    ToolName: "get_supported_formats",
    Arguments: map[string]interface{}{},
}

response := plugin.CallTool(ctx, request)
if response.Success {
    result := response.Result.(map[string]interface{})
    formats := result["formats"].([]string)
    languages := result["languages"].([]string)
    features := result["features"].([]string)

    fmt.Printf("支持格式: %v\n", formats)
    fmt.Printf("支持语言: %v\n", languages)
    fmt.Printf("功能特性: %v\n", features)
}
```

## ⚙️ 配置选项

### 插件配置 (plugin.yaml)

```yaml
# ASR引擎配置
engine:
  provider: "whisper"  # whisper, azure, google, baidu
  model: "base"        # tiny, base, small, medium, large
  language: "zh-CN"

# 音频处理配置
audio:
  sample_rate: 16000
  channels: 1
  bit_depth: 16
  max_duration: 300  # 最大音频时长（秒）

# 批处理配置
batch:
  max_files: 50
  max_total_size: "500Mi"

# 质量控制
quality:
  min_confidence: 0.8
  enable_punctuation: true
  enable_timestamp: true
```

### 环境变量

```bash
# 基础配置
PLUGIN_LOG_LEVEL=info

# Whisper API配置
WHISPER_API_KEY=your_whisper_api_key

# Azure Speech配置
AZURE_SPEECH_KEY=your_azure_key
AZURE_SPEECH_REGION=eastasia

# Google Cloud配置
GOOGLE_CLOUD_KEY=your_google_cloud_key
```

## 📊 性能指标

插件内置了丰富的性能指标监控：

- **asr.calls.total**: 总调用次数
- **asr.calls.success**: 成功调用次数
- **asr.calls.unknown**: 未知工具调用次数
- **asr.errors.decode**: 解码错误次数
- **asr.errors.processing**: 处理错误次数
- **asr.processing_duration**: 处理时长分布
- **asr.confidence**: 识别置信度分布
- **asr.format_detect.calls**: 格式检测调用次数
- **asr.batch.calls**: 批处理调用次数
- **asr.batch.files_count**: 批处理文件数量分布
- **asr.batch.total_duration**: 批处理总时长分布

## 🔧 错误处理

插件提供了完善的错误处理机制：

```go
// 常见错误代码
- INVALID_ARGUMENT: 参数错误
- DECODE_ERROR: 音频数据解码失败
- PROCESSING_ERROR: 语音处理失败
- UNKNOWN_TOOL: 未知工具调用
- TIMEOUT: 处理超时
- RATE_LIMITED: 调用频率限制
```

## 🧪 测试

运行完整的测试套件：

```bash
# 运行所有测试
go test ./...

# 运行基准测试
go test -bench=. ./...

# 运行功能测试
go run test_asr.go
```

## 📈 性能优化建议

1. **音频格式**: 使用WAV格式获得最佳识别效果
2. **音频质量**: 确保音频采样率至少16kHz，建议使用44.1kHz
3. **文件大小**: 单个音频文件建议不超过100MB
4. **批量处理**: 对于大量文件，使用批量转录API
5. **环境选择**: 根据需求选择合适的ASR引擎提供商

## 🛡️ 安全注意事项

1. **数据隐私**: 音频数据在传输过程中使用Base64编码
2. **访问控制**: 建议配置API密钥和访问控制
3. **存储安全**: 处理完成的音频数据不会被持久化存储
4. **网络安全**: 生产环境建议使用HTTPS传输

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

**Q: 识别结果不准确怎么办？**
A: 检查音频质量，调整采样率，尝试不同的ASR引擎提供商。

**Q: 处理速度很慢？**
A: 检查网络连接，考虑使用更小的模型，或者使用批量处理。

**Q: 支持的音频格式有限？**
A: 可以先转换为WAV格式，或者添加新的格式支持。

**Q: 内存占用过高？**
A: 限制批处理的文件数量，或者增加系统内存。

## 📚 相关文档

- [XiaoZhi Flow 插件开发指南](../../docs/plugin-development.md)
- [插件快速开始](../../docs/plugin-quickstart.md)
- [API 文档](../../../docs/api.md)
- [部署指南](../../../docs/deployment.md)