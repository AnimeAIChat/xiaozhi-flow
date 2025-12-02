# XiaoZhi Flow 工作流引擎使用指南

## 概述

XiaoZhi Flow 工作流引擎是一个功能完整的DAG（有向无环图）工作流执行系统，支持插件化架构、节点Schema定义、数据流传递和并行执行。

## 核心特性

### 🎯 插件进程管理
- **HTTP插件进程模拟器** - 模拟独立的HTTP插件进程
- **启动/停止控制** - 完整的插件生命周期管理
- **健康检查** - 定期健康状态监控
- **统计信息** - 调用次数、成功率、延迟等指标

### 📊 节点Schema系统
- **输入/输出定义** - 完整的Schema定义和验证
- **类型系统** - 支持string、number、boolean、object、array
- **验证规则** - 长度、范围、正则表达式等验证
- **默认值支持** - 可选字段的默认值设置

### 🔄 DAG拓扑排序
- **Kahn算法** - 高效的拓扑排序实现
- **循环检测** - 自动检测和防止循环依赖
- **依赖解析** - 智能节点依赖关系分析
- **执行顺序** - 保证正确的节点执行顺序

### ⚡ 工作流执行器
- **顺序执行** - 按依赖关系顺序执行节点
- **并行执行** - 支持无依赖节点的并行处理
- **错误处理** - 完善的错误处理和重试机制
- **超时控制** - 工作流和节点级别的超时设置

### 🌊 数据流传递
- **智能数据路由** - 自动数据传递到目标节点
- **Schema验证** - 输入数据的类型和验证检查
- **并行数据合并** - 并行节点结果的智能合并
- **表达式支持** - 灵活的数据映射和转换

## 快速开始

### 1. 创建工作流

```go
import "xiaozhi-server-go/internal/workflow"

// 创建简单工作流
workflow := &workflow.Workflow{
    ID:          "my-workflow",
    Name:        "我的工作流",
    Description: "测试工作流",
    Version:     "1.0.0",
    Nodes: []workflow.Node{
        {
            ID:   "start",
            Name: "开始",
            Type: workflow.NodeTypeStart,
            Inputs: []workflow.InputSchema{
                {
                    Name:     "data",
                    Type:     "object",
                    Required: true,
                    Description: "输入数据",
                },
            },
            Position: workflow.Position{X: 100, Y: 100},
        },
        {
            ID:     "process",
            Name:   "处理数据",
            Type:   workflow.NodeTypeTask,
            Plugin: "my-plugin",
            Method: "process_data",
            Inputs: []workflow.InputSchema{
                {
                    Name:     "input",
                    Type:     "object",
                    Required: true,
                },
            },
            Outputs: []workflow.OutputSchema{
                {
                    Name: "result",
                    Type: "object",
                    Description: "处理结果",
                },
            },
            Position: workflow.Position{X: 300, Y: 100},
        },
        {
            ID:   "end",
            Name: "结束",
            Type: workflow.NodeTypeEnd,
            Position: workflow.Position{X: 500, Y: 100},
        },
    },
    Edges: []workflow.Edge{
        {ID: "e1", From: "start", To: "process"},
        {ID: "e2", From: "process", To: "end"},
    },
    Config: workflow.WorkflowConfig{
        Timeout:       5 * time.Minute,
        MaxRetries:    3,
        ParallelLimit: 5,
        EnableLog:     true,
        Variables: map[string]interface{}{
            "env": "production",
        },
    },
}
```

### 2. 初始化组件

```go
// 创建日志器
logger := &workflow.SimpleLogger{}

// 创建插件管理器
pluginManager := workflow.NewHTTPPluginManager(logger)

// 创建DAG引擎
dagEngine := workflow.NewDAGEngine(logger)

// 创建数据流引擎
dataFlow := workflow.NewDataFlowEngine(dagEngine, logger)

// 创建工作流执行器
executor := workflow.NewWorkflowExecutor(pluginManager, dagEngine, dataFlow, logger)
```

### 3. 启动插件

```go
ctx := context.Background()

// 启动插件
plugin, err := pluginManager.StartPlugin(ctx, "my-plugin")
if err != nil {
    log.Fatalf("Failed to start plugin: %v", err)
}

fmt.Printf("Plugin started: %s\n", plugin.Name)
```

### 4. 执行工作流

```go
// 准备输入数据
inputs := map[string]interface{}{
    "data": map[string]interface{}{
        "message": "Hello, Workflow!",
        "timestamp": time.Now().Unix(),
    },
}

// 执行工作流
execution, err := executor.Execute(ctx, workflow, inputs)
if err != nil {
    log.Fatalf("Failed to execute workflow: %v", err)
}

fmt.Printf("Workflow execution started: %s\n", execution.ID)
```

### 5. 监控执行状态

```go
// 监控执行状态
ticker := time.NewTicker(2 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        currentExecution, exists := executor.GetExecution(execution.ID)
        if !exists {
            log.Println("Execution not found")
            break
        }

        fmt.Printf("Status: %s, Completed nodes: %d/%d\n",
            currentExecution.Status,
            countCompletedNodes(currentExecution),
            len(currentExecution.NodeResults))

        if currentExecution.Status == workflow.ExecutionStatusCompleted {
            fmt.Println("Workflow completed successfully!")
            printResults(currentExecution)
            break
        } else if currentExecution.Status == workflow.ExecutionStatusFailed {
            fmt.Printf("Workflow failed: %s\n", currentExecution.Error)
            break
        }

    case <-time.After(30 * time.Second):
        fmt.Println("Execution timeout")
        executor.Cancel(execution.ID)
        break
    }
}
```

## 节点类型详解

### 1. 开始节点 (Start Node)
- **用途**: 工作流的入口点
- **特性**: 接收外部输入，传递给后续节点
- **配置**: 定义输入Schema

```go
{
    ID:   "start",
    Name: "开始",
    Type: workflow.NodeTypeStart,
    Inputs: []workflow.InputSchema{
        {
            Name:     "data",
            Type:     "object",
            Required: true,
        },
    },
}
```

### 2. 任务节点 (Task Node)
- **用途**: 执行具体任务，调用插件
- **特性**: 定义输入输出Schema，关联插件方法
- **配置**: 指定插件ID和方法

```go
{
    ID:     "task",
    Name:   "数据处理",
    Type:   workflow.NodeTypeTask,
    Plugin: "data-processor",
    Method: "transform",
    Inputs: []workflow.InputSchema{
        {
            Name:     "input",
            Type:     "object",
            Required: true,
        },
    },
    Outputs: []workflow.OutputSchema{
        {
            Name: "result",
            Type: "object",
        },
    },
}
```

### 3. 条件节点 (Condition Node)
- **用途**: 根据条件决定执行路径
- **特性**: 评估条件表达式，输出布尔结果
- **配置**: 条件表达式和输入数据

```go
{
    ID:   "condition",
    Name: "条件判断",
    Type: workflow.NodeTypeCondition,
    Inputs: []workflow.InputSchema{
        {
            Name:     "valid",
            Type:     "boolean",
            Required: true,
        },
    },
    Config: map[string]interface{}{
        "condition": "${valid}",
    },
}
```

### 4. 并行节点 (Parallel Node)
- **用途**: 标记并行执行的开始点
- **特性**: 允许多个无依赖任务并行执行
- **配置**: 并行度限制

```go
{
    ID:   "parallel",
    Name: "并行处理",
    Type: workflow.NodeTypeParallel,
}
```

### 5. 合并节点 (Merge Node)
- **用途**: 合并并行节点的执行结果
- **特性**: 收集多个并行任务的输出
- **配置**: 合并策略

```go
{
    ID:   "merge",
    Name: "结果合并",
    Type: workflow.NodeTypeMerge,
}
```

### 6. 结束节点 (End Node)
- **用途**: 工作流的结束点
- **特性**: 收集最终结果，完成执行
- **配置**: 输出Schema定义

```go
{
    ID:   "end",
    Name: "结束",
    Type: workflow.NodeTypeEnd,
    Outputs: []workflow.OutputSchema{
        {
            Name: "workflow_result",
            Type: "object",
        },
    },
}
```

## Schema定义详解

### 输入Schema (InputSchema)

```go
type InputSchema struct {
    Name        string      `json:"name"`         // 字段名称
    Type        string      `json:"type"`         // 数据类型
    Required    bool        `json:"required"`     // 是否必需
    Default     interface{} `json:"default"`      // 默认值
    Description string      `json:"description"`  // 描述
    Validation  *Validation `json:"validation"`   // 验证规则
}
```

#### 支持的数据类型
- **string** - 字符串类型
- **number** - 数字类型（整数和浮点数）
- **boolean** - 布尔类型
- **object** - 对象类型
- **array** - 数组类型

#### 验证规则 (Validation)

```go
type Validation struct {
    MinLength *int     `json:"min_length,omitempty"` // 最小长度（字符串）
    MaxLength *int     `json:"max_length,omitempty"` // 最大长度（字符串）
    Min       *float64 `json:"min,omitempty"`        // 最小值（数字）
    Max       *float64 `json:"max,omitempty"`        // 最大值（数字）
    Pattern   string   `json:"pattern,omitempty"`    // 正则表达式
    Enum      []string `json:"enum,omitempty"`       // 枚举值
}
```

### 输出Schema (OutputSchema)

```go
type OutputSchema struct {
    Name        string `json:"name"`        // 字段名称
    Type        string `json:"type"`        // 数据类型
    Description string `json:"description"` // 描述
}
```

## 插件开发

### HTTP插件接口

HTTP插件需要实现以下端点：

#### 1. 健康检查端点
```
GET /health
```

响应：
```json
{
    "status": "healthy",
    "plugin_id": "my-plugin",
    "timestamp": 1640995200
}
```

#### 2. 方法调用端点
```
POST /call
```

请求体：
```json
{
    "method": "process_data",
    "payload": {
        "input": "data",
        "options": {}
    }
}
```

响应：
```json
{
    "success": true,
    "data": {
        "result": "processed_data"
    },
    "message": "Operation completed successfully"
}
```

#### 3. 插件信息端点
```
GET /info
```

响应：
```json
{
    "id": "my-plugin",
    "name": "Data Processor Plugin",
    "version": "1.0.0",
    "type": "http",
    "status": "running"
}
```

### 插件实现示例

```go
package main

import (
    "encoding/json"
    "net/http"
)

func main() {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status":    "healthy",
            "plugin_id": "my-plugin",
            "timestamp": time.Now().Unix(),
        })
    })

    http.HandleFunc("/call", func(w http.ResponseWriter, r *http.Request) {
        var request map[string]interface{}
        if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        method := request["method"].(string)
        payload := request["payload"].(map[string]interface{})

        result := handleMethod(method, payload)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
    })

    log.Println("Plugin server started on :8080")
    http.ListenAndServe(":8080", nil)
}

func handleMethod(method string, payload map[string]interface{}) map[string]interface{} {
    switch method {
    case "process_data":
        data := payload["input"].(string)
        return map[string]interface{}{
            "success": true,
            "data": map[string]interface{}{
                "result": strings.ToUpper(data),
                "processed_at": time.Now().Format(time.RFC3339),
            },
        }
    default:
        return map[string]interface{}{
            "success": false,
            "error":   "Unknown method: " + method,
        }
    }
}
```

## 并行执行

### 并行工作流示例

```go
parallelWorkflow := &workflow.Workflow{
    ID:   "parallel-workflow",
    Name: "并行处理工作流",
    Nodes: []workflow.Node{
        {ID: "start", Type: workflow.NodeTypeStart},
        {ID: "split", Type: workflow.NodeTypeParallel},
        {ID: "task_a", Type: workflow.NodeTypeTask, Plugin: "plugin-a"},
        {ID: "task_b", Type: workflow.NodeTypeTask, Plugin: "plugin-b"},
        {ID: "task_c", Type: workflow.NodeTypeTask, Plugin: "plugin-c"},
        {ID: "merge", Type: workflow.NodeTypeMerge},
        {ID: "end", Type: workflow.NodeTypeEnd},
    },
    Edges: []workflow.Edge{
        {From: "start", To: "split"},
        {From: "split", To: "task_a"},
        {From: "split", To: "task_b"},
        {From: "split", To: "task_c"},
        {From: "task_a", To: "merge"},
        {From: "task_b", To: "merge"},
        {From: "task_c", To: "merge"},
        {From: "merge", To: "end"},
    },
    Config: workflow.WorkflowConfig{
        ParallelLimit: 5, // 允许最多5个并行任务
    },
}
```

### 并行配置

- **ParallelLimit** - 最大并行执行节点数
- **自动并行** - 系统自动识别可并行节点
- **资源管理** - 合理控制并行度避免资源耗尽

## 数据流传递

### 数据传递规则

1. **依赖传递** - 数据从已完成节点传递到依赖它的节点
2. **Schema映射** - 根据节点Schema自动映射数据
3. **类型转换** - 自动进行基本类型转换
4. **默认值** - 可选字段使用默认值

### 数据访问模式

```go
// 直接访问输入数据
inputs["field_name"]

// 访问依赖节点输出
inputs["previous_node.output_field"]

// 访问全局变量
inputs["global.variable_name"]

// 访问执行上下文
inputs["context.context_field"]
```

### 数据合并策略

并行节点的数据合并：
- **前缀命名** - 使用节点ID作为前缀避免冲突
- **展平合并** - 将所有并行节点输出合并到一个对象
- **结构化保留** - 保持原始数据结构

## 错误处理

### 错误处理策略

1. **节点级错误** - 单个节点失败不影响其他节点
2. **重试机制** - 可配置的重试次数和策略
3. **错误传播** - 错误信息在依赖链中传播
4. **优雅降级** - 部分失败时的处理策略

### 错误配置

```go
Config: workflow.WorkflowConfig{
    MaxRetries: 3,           // 最大重试次数
    Timeout:    30 * time.Second, // 超时时间
    EnableLog:  true,        // 启用错误日志
}
```

### 错误监控

```go
// 检查节点错误
for nodeID, result := range execution.NodeResults {
    if result.Status == workflow.NodeStatusFailed {
        fmt.Printf("Node %s failed: %s\n", nodeID, result.Error)
    }
}

// 检查工作流错误
if execution.Status == workflow.ExecutionStatusFailed {
    fmt.Printf("Workflow failed: %s\n", execution.Error)
}
```

## 监控和日志

### 执行状态监控

```go
// 获取执行状态
execution, exists := executor.GetExecution(executionID)
if !exists {
    return fmt.Errorf("execution not found")
}

// 检查状态
switch execution.Status {
case workflow.ExecutionStatusRunning:
    fmt.Println("Workflow is running")
case workflow.ExecutionStatusCompleted:
    fmt.Println("Workflow completed successfully")
case workflow.ExecutionStatusFailed:
    fmt.Printf("Workflow failed: %s\n", execution.Error)
}
```

### 日志查看

```go
// 获取执行日志
logs, err := executor.GetExecutionLogs(executionID)
if err != nil {
    return err
}

// 打印日志
for _, log := range logs {
    fmt.Printf("[%s] %s: %s\n",
        log.Timestamp.Format("15:04:05"),
        log.Level,
        log.Message)
}
```

### 性能指标

```go
// 插件统计
plugins := pluginManager.ListPlugins()
for id, plugin := range plugins {
    fmt.Printf("Plugin %s:\n", id)
    fmt.Printf("  Call count: %d\n", plugin.Stats.CallCount)
    fmt.Printf("  Success rate: %.2f%%\n",
        float64(plugin.Stats.SuccessCount)/float64(plugin.Stats.CallCount)*100)
    fmt.Printf("  Avg latency: %v\n", plugin.Stats.AvgLatency)
}

// 节点执行时间
for nodeID, result := range execution.NodeResults {
    fmt.Printf("Node %s executed in %v\n", nodeID, result.ElapsedTime)
}
```

## 最佳实践

### 1. 工作流设计

- **单一职责** - 每个节点只做一件事
- **合理粒度** - 节点粒度适中，避免过于复杂
- **明确依赖** - 清晰定义节点间的依赖关系
- **错误处理** - 为每个节点配置适当的错误处理

### 2. Schema设计

- **完整定义** - 为所有输入输出定义Schema
- **类型安全** - 使用正确的数据类型
- **验证规则** - 添加适当的验证规则
- **文档化** - 提供清晰的字段描述

### 3. 插件开发

- **幂等性** - 插件方法应该是幂等的
- **错误处理** - 返回明确的错误信息
- **性能优化** - 避免长时间阻塞操作
- **资源管理** - 合理管理资源使用

### 4. 监控运维

- **日志记录** - 记录关键操作和错误
- **性能监控** - 监控执行时间和资源使用
- **告警设置** - 设置适当的告警阈值
- **定期维护** - 定期清理和优化

## 示例项目

完整的使用示例请参考：
- `internal/workflow/example.go` - 完整示例代码
- `internal/workflow/workflow_test.go` - 单元测试

运行示例：
```go
import "xiaozhi-server-go/internal/workflow"

func main() {
    // 运行完整示例
    workflow.RunExample()
}
```

## 故障排除

### 常见问题

1. **循环依赖**
   - 错误：`workflow contains cycles`
   - 解决：检查工作流边的定义，确保没有循环

2. **插件启动失败**
   - 错误：`plugin start failed`
   - 解决：检查插件端口占用和配置

3. **数据验证失败**
   - 错误：`validation failed`
   - 解决：检查输入数据是否符合Schema定义

4. **执行超时**
   - 错误：`execution timeout`
   - 解决：增加超时时间或优化节点性能

### 调试技巧

1. **启用详细日志**
   ```go
   config := workflow.WorkflowConfig{
       EnableLog: true,
   }
   ```

2. **查看执行日志**
   ```go
   logs, _ := executor.GetExecutionLogs(executionID)
   ```

3. **检查节点状态**
   ```go
   for nodeID, result := range execution.NodeResults {
       fmt.Printf("%s: %s\n", nodeID, result.Status)
   }
   ```

## API参考

详细的API文档请参考类型定义文件：
- `internal/workflow/types.go` - 核心类型定义
- `internal/workflow/plugin_manager.go` - 插件管理器
- `internal/workflow/executor.go` - 工作流执行器
- `internal/workflow/dag_engine.go` - DAG引擎
- `internal/workflow/dataflow.go` - 数据流引擎

---

*如有问题或建议，欢迎提交Issue或PR。*