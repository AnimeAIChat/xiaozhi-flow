# ✨ 小智 AI Flow 后端服务

[![Release](https://img.shields.io/github/v/release/AnimeAIChat/xiaozhi-server-go?style=flat-square)](https://github.com/AnimeAIChat/xiaozhi-server-go/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/AnimeAIChat/xiaozhi-server-go?style=flat-square)](https://goreportcard.com/report/github.com/AnimeAIChat/xiaozhi-server-go)

小智 AI 是一个语音交互机器人，结合 Qwen、DeepSeek 等强大大模型，通过 MCP 协议连接多端设备（ESP32、Android、Python 等），实现高效自然的人机对话。

本项目是其后端服务，旨在提供一套 **商业级部署方案** —— 高并发、低成本、功能完整、开箱即用。

项目初始基于 [虾哥的 ESP32 开源项目](https://github.com/78/xiaozhi-esp32?tab=readme-ov-file)，目前已形成完整生态，支持多种客户端协议兼容接入。

---

## ✨ 核心优势

| 优势         | 说明                                                   |
| ---------- | ---------------------------------------------------- |
| 🚀 高并发     | 单机支持 3000+ 在线，分布式可扩展至百万用户                            |
| 👥 用户系统    | 完整的用户注册、登录、权限管理能力                                    |
| 💰 支付集成    | 接入支付系统，助力商业闭环                                        |
| 🛠️ 模型接入灵活 | 支持通过 API 调用多种大模型，简化部署，支持定制本地部署                       |
| 📈 商业支持    | 提供 7×24 技术支持与运维保障                                    |
| 🧠 模型兼容    | 支持 ASR（豆包）、TTS（EdgeTTS）、LLM（OpenAI、Ollama）、图文解说（智谱）等 |

---

## ✅ 社区版功能清单

* [x] 支持 websocket 连接
* [x] 支持 PCM / Opus 格式语音对话
* [x] 支持大模型：ASR（豆包流式）、TTS（EdgeTTS/豆包）、LLM（OpenAI API、Ollama）
* [x] 支持语音控制调用摄像头识别图像（智谱 API）
* [x] 支持 auto/manual/realtime 三种对话模式，支持对话实时打断
* [x] 支持 ESP32 小智客户端、Python 客户端、Android 客户端连入，无需校验
* [x] OTA 固件下发
* [x] 支持 MCP 协议（客户端 / 本地 / 服务器），可接入高德地图、天气查询等
* [x] 支持语音控制切换角色声音
* [x] 支持语音控制切换预设角色
* [x] 支持语音控制播放音乐
* [x] 支持单机部署服务
* [x] 支持本地数据库 sqlite
* [x] 支持coze工作流 
* [x] 支持Docker部署
## ✅商务版功能清单
* [x] 社区版所有功能
* [x] 开发团队技术支持
* [x] 后续核心功能免费更新
* [x] 商务版管理后台，更多的功能选项
* [x] 支持多用户管理
* [x] 自定义修改欢迎界面
* [x] 自定义修改版权logo，使用自己公司的商务标识
* [x] 自定义修改Agent角色模板
* [x] 支持更多的模型
* [x] 支持 websocket 和 MQTT+UDP 两种通信协议
* [x] 支持 tts 流式生成及发送
* [x] 支持声音克隆
* [x] 支持知识库
* [x] 支持定制音色（cosyvoice2, indextts）
* [x] 支持通过 OTA 升级固件
* [x] 支持 Coze 工作流
* [x] 支持 Dify 工作流
* [x] 深度优化响应速度
* [x] 支持用户身份验证，激活绑定设备
* [x] 支持设备管理：解绑/禁用
* [x] 支持后台解绑设备
* [x] 支持用户自定义 Agent
* [x] 国际化多语言支持：中文、英语、日语、西班牙语、印尼语等
* [x] 支持MCP接入点
* [x] 支持网络数据库
* [x] 支持分布式部署
* [x] 支持本地部署大模型

商务版测试/体验地址：

https://xiaozhi.xf.bj.cn/login

---

#### OTA 地址配置（必配）

```text
http://your-server-ip:8080/api/ota/
```

---

## 💬 MCP 协议配置

参考：`internal/domain/mcp/README.md`

---

## 🧪 源码安装与运行

### 前置条件

* Go 1.24.2+
* Windows 用户需安装 CGO 和 Opus 库（见下文）

```bash
git clone https://github.com/AnimeAIChat/xiaozhi-server-go.git
cd xiaozhi-server-go
```

---

### Windows 安装 Opus 编译环境

安装 [MSYS2](https://www.msys2.org/)，打开MYSY2 MINGW64控制台，然后输入以下命令：

```bash
pacman -Syu
pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-go mingw-w64-x86_64-opus
pacman -S mingw-w64-x86_64-pkg-config
```

设置环境变量（用于 PowerShell 或系统变量）：

```bash
set PKG_CONFIG_PATH=C:\msys64\mingw64\lib\pkgconfig
set CGO_ENABLED=1
```

尽量在MINGW64环境下运行一次 “go run ./src/main.go” 命令，确保服务正常运行

GO mod如果更新较慢，可以考虑设置go代理，切换国内镜像源。

---

### 测试
* 推荐使用ESP32硬件设备测试，可以最大程度避免兼容问题
* 推荐使用玄凤小智Android客户端，在设置界面增加本地服务的ota地址即可。安卓版本在Release页面发布，可选择最新版本
  <img width="221" height="470" alt="image" src="https://github.com/user-attachments/assets/145a6612-8397-439b-9429-325855a99101" />

  [xiaozhi-0.0.6.apk](https://github.com/AnimeAIChat/xiaozhi-server-go/releases/download/v0.1.0/xiaozhi-0.0.6.apk)
* 可使用其他兼容小智协议的客户端进行测试
---

## 📚 API 文档（Scalar）

* 打开浏览器访问：`http://localhost:8080/docs`，体验由 Scalar 驱动的现代化接口文档界面
* 需要原始 OpenAPI 规范时，可直接访问：`http://localhost:8080/openapi.json`

---

## 💬 社区支持


欢迎提交 Issue、PR 或新功能建议！


<img src="https://github.com/user-attachments/assets/c162b3a1-c299-4cf3-960d-404b21f138cb" width="450" alt="微信群二维码"> 
<img src="https://github.com/user-attachments/assets/074c6aec-cfb5-4a68-8fc2-2d08679e366b" width="450" alt="QQ群二维码">
---

## 🛠️ 定制开发

我们接受各种定制化开发项目，如果您有特定需求，欢迎通过微信联系洽谈。

<img src="https://github.com/user-attachments/assets/e2639bc3-a58a-472f-9e72-b9363f9e79a3" width="450" alt="群主二维码">

## 📄 License

本仓库遵循 `Xiaozhi-server-go Open Source License`（基于 Apache 2.0 增强版）
