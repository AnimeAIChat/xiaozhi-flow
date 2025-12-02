# 启动流程集成测试指南

## 修复的问题

### 1. ✅ API路径重复问题
**问题**: 前端请求 `:8080/api/api/startup/workflows/...` (重复的`api`)
**原因**: API客户端baseURL已包含`/api`，但请求路径又添加了`/api`
**修复**: 将前端API请求路径从 `/api/startup/...` 改为 `/startup/...`

### 2. ✅ 后端API端点缺失
**问题**: 404错误 - 后端没有启动流程相关的API端点
**修复**: 创建了完整的启动流程HTTP服务
- 文件: `internal/transport/http/startup/service.go`
- 包含所有必要的API端点

### 3. ✅ 服务注册
**修复**: 在bootstrap中注册启动流程服务
```go
startupService, err := httpstartup.NewService(config, logger)
startupService.Register(groupCtx, apiGroup)
```

## 实现的API端点

### 工作流管理
- `GET /api/startup/workflows` - 获取工作流列表
- `GET /api/startup/workflows/:id` - 获取工作流详情
- `POST /api/startup/workflows/execute` - 执行工作流

### 执行管理
- `GET /api/startup/executions/:id` - 获取执行状态
- `DELETE /api/startup/executions/:id` - 取消执行
- `POST /api/startup/executions/:id/pause` - 暂停执行
- `POST /api/startup/executions/:id/resume` - 恢复执行
- `GET /api/startup/executions` - 获取执行历史

### WebSocket
- `GET /api/startup/ws` - WebSocket连接端点

## 可用工作流

1. **xiaozhi-flow-default-startup** - 默认启动工作流
2. **xiaozhi-flow-parallel-startup** - 并行启动工作流
3. **xiaozhi-flow-minimal-startup** - 最小启动工作流

## 测试步骤

### 1. 启动后端服务
```bash
# 编译
go build -o xiaozhi-server ./cmd/xiaozhi-server

# 运行
./xiaozhi-server
```

### 2. 测试API端点
```bash
# 测试获取工作流列表
curl http://localhost:8080/api/startup/workflows

# 测试获取特定工作流
curl http://localhost:8080/api/startup/workflows/xiaozhi-flow-default-startup

# 测试执行工作流
curl -X POST http://localhost:8080/api/startup/workflows/execute \
  -H "Content-Type: application/json" \
  -d '{"workflow_id": "xiaozhi-flow-default-startup", "inputs": {}}'
```

### 3. 启动前端
```bash
cd web
npm run dev
```

### 4. 访问Dashboard
- URL: http://localhost:3000
- 切换到工作流视图
- 应该看到启动流程控制面板和真实的节点

## 预期结果

### 前端显示
1. **真实启动流程节点**: 11个节点，包括：
   - 初始化配置存储
   - 初始化数据库
   - 加载默认配置
   - 初始化日志系统
   - 等等...

2. **控制面板功能**:
   - 工作流选择下拉框
   - 执行按钮
   - 进度条
   - 状态标签
   - 配置和统计按钮

3. **节点状态**:
   - 不同的颜色表示不同状态
   - 实时状态更新
   - 依赖关系连线

### 后端日志
```
[INFO] 注册启动流程API路由
[INFO] 启动流程API路由注册完成
[INFO] [HTTP] Gin 服务已启动，访问地址 http://localhost:8080
[INFO] 获取启动工作流列表
[INFO] 获取启动工作流详情 workflow_id=xiaozhi-flow-default-startup
```

## 故障排除

### 如果仍然看到404错误
1. 确认后端服务已启动
2. 检查端口是否正确 (默认8080)
3. 查看后端日志确认API路由已注册

### 如果前端显示静态数据
1. 检查网络请求 (F12 -> Network)
2. 确认API请求没有报错
3. 检查WebSocket连接状态

### 如果WebSocket连接失败
1. 确认后端支持WebSocket
2. 检查防火墙设置
3. 确认URL格式正确

## 下一步扩展

1. **真实执行引擎**: 集成实际的启动流程执行引擎
2. **持久化存储**: 将工作流和执行数据存储到数据库
3. **用户界面优化**: 改进节点显示和交互
4. **错误处理**: 完善错误处理和用户反馈
5. **性能监控**: 添加执行时间和性能统计

## 技术架构

```
前端 (React + TypeScript)
├── Dashboard Component
├── useWorkflowState Hook
├── StartupWebSocket Manager
├── API Service
└── Data Converter

后端 (Go + Gin)
├── HTTP Router
├── Startup HTTP Service
├── Startup WebSocket Handler
└── Bootstrap Integration
```

现在启动流程的前后端集成已经完成，可以显示真实的启动流程节点了！