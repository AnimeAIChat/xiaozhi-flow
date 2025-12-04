# Dashboardv2 - 基于 Rete.js 的重构版本

这是原始 Dashboard 组件的现代化重构版本，使用 Rete.js 2.0 作为节点编辑器框架。

## 特性

- ✅ **现代化架构**: 基于 Rete.js 理念构建的可视化工作流编辑器
- ✅ **后端集成**: 完整的 xiaozhi-flow-default-startup 启动流程支持
- ✅ **HTTP API 通信**: 通过 REST API 与后端启动流程服务交互
- ✅ **状态轮询**: 智能轮询机制实时更新执行状态
- ✅ **模块化设计**: 完全基于插件的可扩展架构
- ✅ **类型安全**: 完整的 TypeScript 支持
- ✅ **节点系统**: 支持 5 种不同类型的节点
  - 数据库节点 (Database)
  - API 节点 (API)
  - AI 节点 (AI)
  - 云服务节点 (Cloud)
  - 配置节点 (Config)
- ✅ **向后兼容**: 保持与原版 Dashboard 相同的 API 接口

### 🚀 新增后端启动流程功能

- **HTTP API 交互**: 通过标准的 REST API 与后端通信
- **智能轮询**: 高效的轮询机制实时获取执行状态
- **执行控制**: 支持 start、pause、resume、cancel 操作
- **进度可视化**: 实时显示执行进度和状态
- **状态监控**: 执行状态实时更新和统计信息
- **错误处理**: 完善的错误处理和降级机制
- **模拟模式**: 无后端时提供完整的模拟执行体验

## 文件结构

```
Dashboardv2/
├── Dashboardv2.tsx           # 主组件 - 完整的仪表板
├── SimpleReteEditor.tsx      # 核心编辑器 - 启动流程编辑器
├── SimpleWorkflowCanvas.tsx  # 展示组件 - 介绍页面
├── HttpStartupAdapter.tsx    # HTTP 启动流程适配器 - 后端集成
├── nodes/                    # 节点类型定义
│   ├── BaseNode.tsx         # 基础节点类和渲染
│   ├── DatabaseNode.tsx     # 数据库节点
│   ├── ApiNode.tsx          # API 节点
│   ├── AiNode.tsx           # AI 节点
│   ├── CloudNode.tsx        # 云服务节点
│   ├── ConfigNode.tsx       # 配置节点
│   └── index.ts             # 导出文件
├── index.ts                  # 组件导出
└── README.md                 # 详细文档
```

## 使用方法

### 基本使用

```tsx
import Dashboardv2 from '@/components/Dashboardv2';

function App() {
  return <Dashboardv2 />;
}
```

### 启动流程集成

Dashboardv2 自动集成了 `xiaozhi-flow-default-startup` 启动流程：

```tsx
import Dashboardv2 from '@/components/Dashboardv2';

function App() {
  return (
    <Dashboardv2
      // 可选配置
      workflowId="xiaozhi-flow-default-startup"
      autoConnect={true}
    />
  );
}
```

#### 启动流程特性

1. **HTTP API 通信**: 通过标准 REST API 与后端启动流程服务交互
2. **智能轮询**: 自动轮询执行状态，实时更新界面
3. **进度监控**: 实时显示执行进度、完成节点数、执行时间等
4. **执行控制**:
   - 🎬 执行启动流程
   - ⏸️ 暂停执行
   - ▶️ 恢复执行
   - ⏹️ 取消执行
5. **降级机制**: 后端不可用时自动切换到模拟模式
6. **错误处理**: 完善的 HTTP 错误处理和用户反馈

#### 默认工作流节点

`xiaozhi-flow-default-startup` 包含以下节点：

1. **存储初始化** - 初始化数据库连接和存储系统
2. **配置加载** - 加载系统配置和环境变量
3. **服务启动** - 启动核心服务组件
4. **认证设置** - 配置认证和授权系统
5. **插件加载** - 加载和初始化插件系统

#### API 端点

组件使用以下 HTTP API 端点：
- `GET /api/startup/workflows/{id}` - 获取工作流定义
- `POST /api/startup/workflows/{id}/execute` - 执行工作流
- `GET /api/startup/executions/{id}/status` - 获取执行状态
- `POST /api/startup/executions/{id}/pause` - 暂停执行
- `POST /api/startup/executions/{id}/resume` - 恢复执行
- `POST /api/startup/executions/{id}/cancel` - 取消执行

### 节点类型

#### 数据库节点 (DatabaseNode)
- **功能**: 连接和操作数据库
- **输入**: config (配置)
- **输出**: data (数据)

#### API 节点 (ApiNode)
- **功能**: 调用 REST API 和 GraphQL
- **输入**: input, config
- **输出**: response

#### AI 节点 (AiNode)
- **功能**: 集成 AI 服务
- **输入**: prompt, context
- **输出**: result

#### 云服务节点 (CloudNode)
- **功能**: 连接云平台服务
- **输入**: data, credentials
- **输出**: output

#### 配置节点 (ConfigNode)
- **功能**: 工作流配置管理
- **输入**: 无
- **输出**: settings

## 状态管理

每个节点都包含以下状态属性：
- `status`: 'running' | 'stopped' | 'warning'
- `metrics`: 性能指标对象
- `description`: 节点描述
- `label`: 显示标签

## 开发计划

### Phase 1 - 基础功能 ✅
- [x] 安装和配置 Rete.js 2.0
- [x] 创建基础节点架构
- [x] 实现可视化展示界面
- [x] 保持与原版 Dashboard 的兼容性

### Phase 2 - 核心功能 ✅
- [x] 完整的节点编辑器
- [x] 节点拖拽和连接
- [x] 实时数据流可视化
- [x] 工作流执行引擎
- [x] 节点编辑对话框
- [x] 工具栏和控制面板
- [x] 状态管理和监控

### Phase 3 - 高级功能 (待开发)
- [ ] 自定义节点编辑器
- [ ] 插件系统
- [ ] 工作流模板
- [ ] 性能监控
- [ ] 数据持久化
- [ ] 协作编辑

## 技术栈

- **核心框架**: Rete.js 2.0
- **UI 库**: Ant Design
- **样式**: Tailwind CSS
- **构建工具**: Rsbuild
- **类型检查**: TypeScript

## 与原版的区别

| 特性 | 原版 (ReactFlow) | 新版 (Rete.js) |
|------|-----------------|----------------|
| 框架 | ReactFlow | Rete.js 2.0 |
| 节点系统 | 自定义组件 | ClassicPreset |
| 插件架构 | 无 | 基于插件系统 |
| 可扩展性 | 有限 | 高度可扩展 |
| 性能 | 良好 | 优秀 |

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 发起 Pull Request

## 许可证

与主项目保持一致的许可证。