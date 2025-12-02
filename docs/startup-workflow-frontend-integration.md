# 启动流程前端集成说明

## 概述

本文档描述了如何将后端启动流程工作流集成到前端Dashboard中，实现真实启动流程节点的显示和管理。

## 实现的功能

### 1. API服务扩展
- **文件**: `web/src/services/api.ts`
- **新增方法**:
  - `getStartupWorkflows()` - 获取启动工作流列表
  - `getStartupWorkflow(id)` - 获取特定工作流详情
  - `executeStartupWorkflow(id, inputs)` - 执行工作流
  - `getStartupExecutionStatus(id)` - 获取执行状态
  - `cancelStartupExecution(id)` - 取消执行
  - `pauseStartupExecution(id)` - 暂停执行
  - `resumeStartupExecution(id)` - 恢复执行

### 2. WebSocket连接管理
- **文件**: `web/src/services/startupWebSocket.ts`
- **功能**:
  - 实时连接启动流程WebSocket服务
  - 自动重连机制
  - 消息订阅和事件处理
  - 执行控制（开始、暂停、取消等）

### 3. 数据转换工具
- **文件**: `web/src/utils/startupDataConverter.ts`
- **功能**:
  - 后端启动节点 → ReactFlow节点格式转换
  - 状态映射和样式更新
  - 自动布局算法
  - 统计信息生成

### 4. 工作流状态管理
- **文件**: `web/src/components/Dashboard/hooks/useWorkflowState.ts`
- **功能**:
  - 集成API和WebSocket数据
  - 实时状态更新
  - 执行控制方法
  - 错误处理和fallback机制

### 5. 启动流程控制界面
- **文件**: `web/src/components/Dashboard/components/StartupControls/`
- **功能**:
  - 工作流选择和执行
  - 执行状态监控
  - 进度显示
  - 控制按钮（开始、暂停、取消等）
  - 执行统计和配置

## 使用方法

### 1. 基本使用
```typescript
import { useWorkflowState } from './hooks/useWorkflowState';

const {
  nodes,          // ReactFlow节点
  edges,          // ReactFlow边
  workflow,       // 工作流定义
  execution,      // 当前执行
  isLoading,      // 加载状态
  isConnected,    // WebSocket连接状态
  executeWorkflow,
  cancelExecution,
  pauseExecution,
  resumeExecution
} = useWorkflowState({
  autoConnect: true,
  workflowId: 'xiaozhi-flow-default-startup'
});
```

### 2. 执行工作流
```typescript
try {
  const executionId = await executeWorkflow({
    env: 'development',
    debug: true
  });
  console.log('执行开始:', executionId);
} catch (error) {
  console.error('执行失败:', error);
}
```

### 3. 监听执行状态
通过WebSocket自动监听，无需手动处理。执行状态会自动更新节点样式和进度。

## 节点状态说明

| 后端状态 | 前端显示 | 颜色 | 图标 |
|---------|---------|------|------|
| pending | 等待中 | 灰色 | ⏰ |
| running | 运行中 | 绿色 | ⚡ |
| completed | 已完成 | 蓝色 | ✅ |
| failed | 失败 | 红色 | ❌ |
| paused | 已暂停 | 橙色 | ⏸️ |
| cancelled | 已取消 | 灰色 | ❌ |

## 节点类型映射

| 后端类型 | 前端ReactFlow类型 | 颜色 |
|---------|-----------------|------|
| storage | database | #1890ff |
| config | config | #52c41a |
| service | api | #722ed1 |
| auth | api | #fa8c16 |
| plugin | cloud | #13c2c2 |

## WebSocket消息类型

### 客户端发送
- `ping` - 心跳
- `execute_workflow` - 执行工作流
- `cancel_execution` - 取消执行
- `pause_execution` - 暂停执行
- `resume_execution` - 恢复执行
- `get_execution_status` - 获取执行状态
- `subscribe` - 订阅执行事件
- `unsubscribe` - 取消订阅

### 服务器发送
- `connection_established` - 连接建立确认
- `execution_start` - 执行开始
- `execution_progress` - 执行进度更新
- `execution_end` - 执行结束
- `node_start` - 节点开始
- `node_progress` - 节点进度
- `node_complete` - 节点完成
- `node_error` - 节点错误
- `pong` - 心跳响应

## 配置选项

### useWorkflowState Hook选项
```typescript
interface UseWorkflowStateOptions {
  autoConnect?: boolean;    // 是否自动连接WebSocket (默认: true)
  workflowId?: string;      // 工作流ID (默认: 'xiaozhi-flow-default-startup')
  executionId?: string;     // 指定执行ID
}
```

### 可用的工作流
- `xiaozhi-flow-default-startup` - 默认启动工作流
- `xiaozhi-flow-parallel-startup` - 并行启动工作流
- `xiaozhi-flow-minimal-startup` - 最小启动工作流

## 错误处理

### 1. 连接失败
- WebSocket连接失败时，会自动尝试重连（最多5次）
- 显示连接状态指示器
- 提供手动重连选项

### 2. API失败
- API请求失败时，自动fallback到静态演示数据
- 显示错误信息
- 提供刷新功能

### 3. 执行错误
- 节点执行失败时，显示错误状态
- 提供错误详情查看
- 支持重新执行

## 性能优化

### 1. 数据缓存
- 工作流定义缓存
- 执行历史缓存
- 节点状态缓存

### 2. 实时更新
- 仅更新变化的节点
- 节点样式按需更新
- 边动画智能控制

### 3. 内存管理
- 自动清理WebSocket处理器
- 组件卸载时断开连接
- 避免内存泄漏

## 调试

### 1. 日志
所有操作都会记录详细的日志，可通过以下方式查看：
```typescript
import { log } from '../../../utils/logger';

// 查看工作流相关日志
log.debug('workflow', null, 'debug');
```

### 2. 开发工具
- WebSocket连接统计
- API调用历史
- 执行状态监控

## 后续扩展

### 1. 支持的功能
- 工作流编辑器
- 执行历史查看
- 性能监控
- 报警和通知

### 2. UI改进
- 响应式设计
- 主题支持
- 国际化
- 无障碍支持

## 注意事项

1. **CORS配置**: 确保后端WebSocket支持CORS
2. **认证**: 需要在WebSocket连接时传递认证信息
3. **性能**: 大型工作流可能影响渲染性能，考虑虚拟化
4. **兼容性**: 确保浏览器支持WebSocket
5. **错误处理**: 实现优雅的错误处理和用户反馈

## 故障排除

### 问题1: 无法连接WebSocket
- 检查后端服务是否启动
- 确认端口和路径正确
- 检查防火墙和网络配置

### 问题2: 节点不显示
- 检查工作流定义是否正确
- 确认数据转换无错误
- 查看控制台错误信息

### 问题3: 执行状态不更新
- 确认WebSocket连接正常
- 检查消息订阅状态
- 验证事件处理逻辑