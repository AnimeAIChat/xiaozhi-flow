# TTS Plugin - 文本转语音插件

## 🎯 概述

TTS插件是一个功能强大的文本转语音插件，支持多种语言、语音和音频格式。该插件集成了现代语音合成技术，可以生成高质量的语音输出。

## ✨ 功能特性

### 核心功能
- **文本转语音**: 将文本转换为自然流畅的语音
- **多语音支持**: 提供多种不同性别、年龄、口音的语音选择
- **音频格式**: 支持MP3、WAV、FLAC、AAC、OGG等多种音频格式
- **参数调节**: 支持语速、音调、音量等参数调整

### 高级功能
- **批量合成**: 支持批量处理多个文本
- **SSML支持**: 支持语音合成标记语言进行精细控制
- **文本验证**: 智能验证文本，提供改进建议
- **语音克隆**: 支持自定义语音模型（高级功能）
- **情感语音**: 支持带有情感的语音合成
- **实时缓存**: 提供智能缓存机制提高性能

## 🚀 快速开始

### 1. 构建插件

```bash
# 进入插件目录
cd plugins/examples/tts

# 构建插件
go build -o tts-plugin main.go

# 或者直接运行
go run main.go
```

### 2. 运行测试

```bash
# 运行功能测试
go run test_tts.go
```

### 3. 启动插件

```bash
# 启动插件服务
./tts-plugin
```

## 📋 支持的语言和语音

### 语言支持
- **中文**: zh-CN（简体中文）、zh-TW（繁体中文）
- **英文**: en-US（美式英语）、en-GB（英式英语）
- **日文**: ja-JP
- **韩文**: ko-KR
- **西班牙语**: es-ES
- **法语**: fr-FR
- **德语**: de-DE
- **意大利语**: it-IT
- **葡萄牙语**: pt-BR（巴西）
- **俄语**: ru-RU

### 语音类型
- **性别**: 男声、女声、中性声音
- **年龄**: 儿童、青年、中年、老年
- **口音**: 标准口音、地方口音、外语口音
- **风格**: 新闻播报、日常对话、客服、朗读等

## 🛠️ 使用方法

### 1. 基本文本转语音

```go
request := &v1.CallToolRequest{
    ToolName: "text_to_speech",
    Arguments: map[string]interface{}{
        "text":    "你好，欢迎使用TTS插件！",
        "voice":   "zh-CN-female-1",    // 可选，默认语音
        "format":  "mp3",              // 可选，默认mp3
        "rate":    1.0,                // 可选，语速0.1-2.0
        "pitch":   1.0,                // 可选，音调0.1-2.0
        "volume":  1.0,                // 可选，音量0.1-2.0
    },
}

response := plugin.CallTool(ctx, request)
if response.Success {
    result := response.Result.(map[string]interface{})
    audioData := result["audio_data"].(string) // Base64编码
    duration := result["duration"].(float64)   // 毫秒
    size := result["size"].(int)              // 字节数

    fmt.Printf("合成完成，时长 %.2f 秒，大小 %d 字节\n", duration/1000, size)

    // 解码音频数据
    decodedAudio, _ := base64.StdEncoding.DecodeString(audioData)
    // 保存或播放音频...
}
```

### 2. 获取可用语音

```go
request := &v1.CallToolRequest{
    ToolName: "get_available_voices",
    Arguments: map[string]interface{}{
        "language": "zh-CN",  // 可选，筛选语言
        "gender":   "female", // 可选，筛选性别
    },
}

response := plugin.CallTool(ctx, request)
if response.Success {
    result := response.Result.(map[string]interface{})
    voices := result["voices"].([]map[string]interface{})

    for _, voice := range voices {
        fmt.Printf("ID: %s, 名称: %s, 语言: %s, 性别: %s\n",
            voice["id"], voice["name"], voice["language"], voice["gender"])
    }
}
```

### 3. 批量文本合成

```go
texts := []interface{}{
    "第一段文本内容",
    "第二段文本内容",
    "第三段文本内容",
}

request := &v1.CallToolRequest{
    ToolName: "synthesize_batch",
    Arguments: map[string]interface{}{
        "texts":  texts,
        "voice":  "zh-CN-female-1",
        "format": "mp3",
    },
}

response := plugin.CallTool(ctx, request)
if response.Success {
    result := response.Result.(map[string]interface{})
    results := result["results"].([]map[string]interface{})

    for i, item := range results {
        if item["success"].(bool) {
            fmt.Printf("文本 %d: 合成成功，大小 %v 字节\n", i, item["size"])
            // 处理音频数据 item["audio_data"]
        } else {
            fmt.Printf("文本 %d: 合成失败 - %v\n", i, item["error"])
        }
    }
}
```

### 4. 文本验证

```go
request := &v1.CallToolRequest{
    ToolName: "validate_text",
    Arguments: map[string]interface{}{
        "text": "这是要验证的文本内容，包含一些特殊符号！@#￥%",
    },
}

response := plugin.CallTool(ctx, request)
if response.Success {
    result := response.Result.(map[string]interface{})
    valid := result["valid"].(bool)
    issues := result["issues"].([]string)
    warnings := result["warnings"].([]string)
    charCount := result["char_count"].(int)
    estimatedDuration := result["estimated_duration"].(float64)

    fmt.Printf("验证结果: %v\n", valid)
    fmt.Printf("字符数: %d\n", charCount)
    fmt.Printf("预计时长: %.2f 秒\n", estimatedDuration/1000)

    if len(issues) > 0 {
        fmt.Printf("问题: %v\n", issues)
    }
    if len(warnings) > 0 {
        fmt.Printf("警告: %v\n", warnings)
    }
}
```

### 5. 获取支持信息

```go
request := &v1.CallToolRequest{
    ToolName: "get_supported_formats",
    Arguments: map[string]interface{}{},
}

response := plugin.CallTool(ctx, request)
if response.Success {
    result := response.Result.(map[string]interface{})
    formats := result["formats"].([]string)
    features := result["voice_features"].([]string)
    languages := result["languages"].([]string)

    fmt.Printf("支持格式: %v\n", formats)
    fmt.Printf("功能特性: %v\n", features)
    fmt.Printf("支持语言: %v\n", languages)
}
```

## ⚙️ 配置选项

### 插件配置 (plugin.yaml)

```yaml
# TTS引擎配置
engine:
  provider: "azure"  # azure, google, aws, baidu, local
  model: "neural"    # neural, standard
  region: "eastasia"
  default_language: "zh-CN"
  default_voice: "zh-CN-XiaoxiaoNeural"

# 音频处理配置
audio:
  sample_rate: 24000
  bit_rate: 128
  channels: 1
  format: "mp3"
  quality: "high"  # low, medium, high

# 语音参数配置
voice:
  default_rate: 1.0
  default_pitch: 1.0
  default_volume: 1.0
  rate_range: [0.1, 2.0]
  pitch_range: [0.1, 2.0]
  volume_range: [0.1, 2.0]

# SSML配置
ssml:
  enabled: true
  supported_tags:
    - "emphasis"
    - "break"
    - "prosody"
    - "say-as"
    - "voice"

# 批处理配置
batch:
  max_texts: 100
  max_total_chars: 50000
  max_concurrent: 5

# 缓存配置
cache:
  enabled: true
  ttl: 3600
  max_size: "1Gi"
```

### 环境变量

```bash
# 基础配置
PLUGIN_LOG_LEVEL=info

# Azure Speech Service
AZURE_SPEECH_KEY=your_azure_speech_key
AZURE_SPEECH_REGION=eastasia

# Google Cloud Text-to-Speech
GOOGLE_CLOUD_KEY=your_google_cloud_key
GOOGLE_CLOUD_PROJECT=your_project_id

# AWS Polly
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
AWS_REGION=us-east-1

# 百度语音合成
BAIDU_APP_ID=your_baidu_app_id
BAIDU_API_KEY=your_baidu_api_key
BAIDU_SECRET_KEY=your_baidu_secret_key
```

## 📊 性能指标

插件内置了丰富的性能指标监控：

- **tts.calls.total**: 总调用次数
- **tts.calls.success**: 成功调用次数
- **tts.calls.unknown**: 未知工具调用次数
- **tts.errors.synthesis**: 合成错误次数
- **tts.synthesis_duration**: 合成时长分布
- **tts.audio_size**: 生成音频大小分布
- **tts.voices_list.calls**: 语音列表查询次数
- **tts.validate_text.calls**: 文本验证调用次数
- **tts.batch.calls**: 批量合成调用次数
- **tts.batch.texts_count**: 批处理文本数量分布
- **tts.batch.total_duration**: 批处理总时长分布
- **tts.batch.total_size**: 批处理总大小分布

## 🔧 错误处理

插件提供了完善的错误处理机制：

```go
// 常见错误代码
- INVALID_ARGUMENT: 参数错误
- TEXT_TOO_LONG: 文本过长
- SYNTHESIS_ERROR: 语音合成失败
- UNKNOWN_TOOL: 未知工具调用
- VOICE_NOT_FOUND: 语音不存在
- FORMAT_NOT_SUPPORTED: 格式不支持
- TIMEOUT: 处理超时
- RATE_LIMITED: 调用频率限制
- QUOTA_EXCEEDED: 配额超限
- ENGINE_ERROR: TTS引擎错误
```

## 🎵 SSML支持

插件支持SSML（语音合成标记语言）进行精细控制：

```xml
<speak>
    <prosody rate="0.9" pitch="10%">欢迎使用</prosody>
    <emphasis level="strong">TTS插件</emphasis>
    <break time="500ms"/>
    这是<say-as interpret-as="characters">SSML</say-as>示例
</speak>
```

支持的SSML标签：
- `<prosody>`: 控制语速、音调、音量
- `<emphasis>`: 强调特定词语
- `<break>`: 插入停顿
- `<say-as>`: 指定文本解释方式
- `<voice>`: 切换语音

## 🧪 测试

运行完整的测试套件：

```bash
# 运行所有测试
go test ./...

# 运行基准测试
go test -bench=. ./...

# 运行功能测试
go run test_tts.go
```

测试覆盖：
- 基础文本转语音
- 不同语音和格式
- 参数调节
- 批量处理
- 错误处理
- 性能指标

## 📈 性能优化建议

1. **文本预处理**: 移除不必要的标点和空白字符
2. **批量处理**: 对于多个文本使用批量API
3. **音频格式选择**: 根据需求选择合适的音频格式
4. **缓存利用**: 启用缓存避免重复合成
5. **参数调优**: 根据场景调整语音参数
6. **并发控制**: 合理控制并发数量避免过载

## 🛡️ 安全注意事项

1. **内容过滤**: 建议添加内容安全过滤
2. **访问控制**: 配置API密钥和访问限制
3. **频率限制**: 设置调用频率限制防止滥用
4. **配额管理**: 监控使用量避免超限
5. **隐私保护**: 敏感内容建议本地处理

## 🌐 多云支持

插件支持多个TTS服务提供商：

### Microsoft Azure Speech
- 高质量神经网络语音
- 丰富的语音选择
- 实时流式合成

### Google Cloud Text-to-Speech
- WaveNet高质量语音
- 多语言支持
- 自定义语音训练

### Amazon Polly
- SSML完全支持
- 神经语音和标准语音
- 语音标记功能

### 百度语音合成
- 中文语音优化
- 多种中文发音人
- 离线语音合成支持

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

**Q: 语音质量不够好怎么办？**
A: 尝试神经网络模型，调整采样率，选择合适的语音。

**Q: 合成速度很慢？**
A: 检查网络连接，启用缓存，使用批量处理。

**Q: 某些字符发音不准确？**
A: 使用SSML标签，或者替换为同义词。

**Q: 如何添加新的语音？**
A: 修改语音配置文件，或集成新的TTS提供商。

**Q: 批量处理失败？**
A: 检查文本内容，减少并发数，增加超时时间。

## 📚 相关文档

- [XiaoZhi Flow 插件开发指南](../../docs/plugin-development.md)
- [插件快速开始](../../docs/plugin-quickstart.md)
- [SSML参考文档](../../../docs/ssml.md)
- [API 文档](../../../docs/api.md)
- [部署指南](../../../docs/deployment.md)