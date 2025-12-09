package workflow

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"xiaozhi-server-go/internal/plugin/capability"
)

// SimpleLogger 简单日志实现
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, fields ...interface{}) {
	log.Printf("[INFO] %s %v", msg, fields)
}

func (l *SimpleLogger) Error(msg string, fields ...interface{}) {
	log.Printf("[ERROR] %s %v", msg, fields)
}

func (l *SimpleLogger) Warn(msg string, fields ...interface{}) {
	log.Printf("[WARN] %s %v", msg, fields)
}

func (l *SimpleLogger) Debug(msg string, fields ...interface{}) {
	log.Printf("[DEBUG] %s %v", msg, fields)
}

// CreateExampleWorkflow 创建示例工作流
func CreateExampleWorkflow() *Workflow {
	return &Workflow{
		ID:          "example-workflow-1",
		Name:        "示例数据处理工作流",
		Description: "演示如何使用工作流引擎处理数据",
		Version:     "1.0.0",
		Nodes: []Node{
			{
				ID:          "start",
				Name:        "开始",
				Type:        NodeTypeStart,
				Description: "工作流开始节点",
				Inputs: []InputSchema{
					{
						Name:        "data",
						Type:        "object",
						Required:    true,
						Description: "输入数据",
					},
				},
				Outputs: []OutputSchema{
					{
						Name:        "data",
						Type:        "object",
						Description: "传递的数据",
					},
				},
				Position: Position{X: 100, Y: 100},
				Status:   NodeStatusPending,
			},
			{
				ID:          "process_data",
				Name:        "数据处理",
				Type:        NodeTypeTask,
				Description: "处理输入数据",
				Plugin:      "http-plugin-1",
				Method:      "process",
				Inputs: []InputSchema{
					{
						Name:        "input",
						Type:        "object",
						Required:    true,
						Description: "待处理的数据",
					},
					{
						Name:        "operation",
						Type:        "string",
						Required:    false,
						Default:     "transform",
						Description: "处理操作类型",
					},
				},
				Outputs: []OutputSchema{
					{
						Name:        "result",
						Type:        "object",
						Description: "处理结果",
					},
					{
						Name:        "status",
						Type:        "string",
						Description: "处理状态",
					},
				},
				Position: Position{X: 300, Y: 100},
				Status:   NodeStatusPending,
				Config: map[string]interface{}{
					"timeout": "30s",
				},
			},
			{
				ID:          "validate_data",
				Name:        "数据验证",
				Type:        NodeTypeTask,
				Description: "验证处理后的数据",
				Plugin:      "http-plugin-2",
				Method:      "validate",
				Inputs: []InputSchema{
					{
						Name:        "data",
						Type:        "object",
						Required:    true,
						Description: "待验证的数据",
					},
				},
				Outputs: []OutputSchema{
					{
						Name:        "valid",
						Type:        "boolean",
						Description: "是否有效",
					},
					{
						Name:        "errors",
						Type:        "array",
						Description: "验证错误列表",
					},
				},
				Position: Position{X: 500, Y: 100},
				Status:   NodeStatusPending,
			},
			{
				ID:          "condition_check",
				Name:        "条件判断",
				Type:        NodeTypeCondition,
				Description: "判断数据是否需要进一步处理",
				Inputs: []InputSchema{
					{
						Name:        "valid",
						Type:        "boolean",
						Required:    true,
						Description: "验证结果",
					},
				},
				Outputs: []OutputSchema{
					{
						Name:        "condition",
						Type:        "string",
						Description: "条件表达式",
					},
					{
						Name:        "result",
						Type:        "boolean",
						Description: "条件结果",
					},
				},
				Position: Position{X: 700, Y: 100},
				Status:   NodeStatusPending,
				Config: map[string]interface{}{
					"condition": "${valid}",
				},
			},
			{
				ID:          "save_data",
				Name:        "保存数据",
				Type:        NodeTypeTask,
				Description: "保存处理完成的数据",
				Plugin:      "http-plugin-3",
				Method:      "save",
				Inputs: []InputSchema{
					{
						Name:        "data",
						Type:        "object",
						Required:    true,
						Description: "待保存的数据",
					},
				},
				Outputs: []OutputSchema{
					{
						Name:        "saved",
						Type:        "boolean",
						Description: "是否保存成功",
					},
					{
						Name:        "id",
						Type:        "string",
						Description: "保存的记录ID",
					},
				},
				Position: Position{X: 900, Y: 50},
				Status:   NodeStatusPending,
			},
			{
				ID:          "notify_error",
				Name:        "错误通知",
				Type:        NodeTypeTask,
				Description: "发送错误通知",
				Plugin:      "http-plugin-4",
				Method:      "notify",
				Inputs: []InputSchema{
					{
						Name:        "message",
						Type:        "string",
						Required:    true,
						Description: "错误消息",
					},
				},
				Outputs: []OutputSchema{
					{
						Name:        "sent",
						Type:        "boolean",
						Description: "是否发送成功",
					},
				},
				Position: Position{X: 900, Y: 150},
				Status:   NodeStatusPending,
			},
			{
				ID:          "end",
				Name:        "结束",
				Type:        NodeTypeEnd,
				Description: "工作流结束节点",
				Inputs: []InputSchema{
					{
						Name:        "final_result",
						Type:        "object",
						Required:    false,
						Description: "最终结果",
					},
				},
				Outputs: []OutputSchema{
					{
						Name:        "workflow_result",
						Type:        "object",
						Description: "工作流执行结果",
					},
				},
				Position: Position{X: 1100, Y: 100},
				Status:   NodeStatusPending,
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "process_data"},
			{ID: "e2", From: "process_data", To: "validate_data"},
			{ID: "e3", From: "validate_data", To: "condition_check"},
			{ID: "e4", From: "condition_check", To: "save_data"},
			{ID: "e5", From: "condition_check", To: "notify_error"},
			{ID: "e6", From: "save_data", To: "end"},
			{ID: "e7", From: "notify_error", To: "end"},
		},
		Config: WorkflowConfig{
			Timeout:       5 * time.Minute,
			MaxRetries:    3,
			ParallelLimit: 3,
			EnableLog:     true,
			Variables: map[string]interface{}{
				"environment": "production",
				"version":     "1.0.0",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateParallelWorkflow 创建并行工作流示例
func CreateParallelWorkflow() *Workflow {
	return &Workflow{
		ID:          "parallel-workflow-1",
		Name:        "并行处理工作流",
		Description: "演示如何使用并行节点处理多个任务",
		Version:     "1.0.0",
		Nodes: []Node{
			{
				ID:          "start",
				Name:        "开始",
				Type:        NodeTypeStart,
				Description: "工作流开始",
				Position:    Position{X: 100, Y: 100},
				Status:      NodeStatusPending,
			},
			{
				ID:          "split",
				Name:        "数据分发",
				Type:        NodeTypeParallel,
				Description: "将数据分发给多个并行任务",
				Position:    Position{X: 300, Y: 100},
				Status:      NodeStatusPending,
			},
			{
				ID:          "task_a",
				Name:        "任务A",
				Type:        NodeTypeTask,
				Description: "并行任务A",
				Plugin:      "http-plugin-1",
				Method:      "process_a",
				Position:    Position{X: 500, Y: 50},
				Status:      NodeStatusPending,
				Inputs: []InputSchema{
					{
						Name:     "data",
						Type:     "object",
						Required: true,
					},
				},
				Outputs: []OutputSchema{
					{
						Name: "result_a",
						Type: "object",
					},
				},
			},
			{
				ID:          "task_b",
				Name:        "任务B",
				Type:        NodeTypeTask,
				Description: "并行任务B",
				Plugin:      "http-plugin-2",
				Method:      "process_b",
				Position:    Position{X: 500, Y: 150},
				Status:      NodeStatusPending,
				Inputs: []InputSchema{
					{
						Name:     "data",
						Type:     "object",
						Required: true,
					},
				},
				Outputs: []OutputSchema{
					{
						Name: "result_b",
						Type: "object",
					},
				},
			},
			{
				ID:          "task_c",
				Name:        "任务C",
				Type:        NodeTypeTask,
				Description: "并行任务C",
				Plugin:      "http-plugin-3",
				Method:      "process_c",
				Position:    Position{X: 500, Y: 250},
				Status:      NodeStatusPending,
				Inputs: []InputSchema{
					{
						Name:     "data",
						Type:     "object",
						Required: true,
					},
				},
				Outputs: []OutputSchema{
					{
						Name: "result_c",
						Type: "object",
					},
				},
			},
			{
				ID:          "merge",
				Name:        "结果合并",
				Type:        NodeTypeMerge,
				Description: "合并并行任务的结果",
				Position:    Position{X: 700, Y: 150},
				Status:      NodeStatusPending,
			},
			{
				ID:          "end",
				Name:        "结束",
				Type:        NodeTypeEnd,
				Description: "工作流结束",
				Position:    Position{X: 900, Y: 150},
				Status:      NodeStatusPending,
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "split"},
			{ID: "e2", From: "split", To: "task_a"},
			{ID: "e3", From: "split", To: "task_b"},
			{ID: "e4", From: "split", To: "task_c"},
			{ID: "e5", From: "task_a", To: "merge"},
			{ID: "e6", From: "task_b", To: "merge"},
			{ID: "e7", From: "task_c", To: "merge"},
			{ID: "e8", From: "merge", To: "end"},
		},
		Config: WorkflowConfig{
			Timeout:       10 * time.Minute,
			MaxRetries:    2,
			ParallelLimit: 5, // 允许更高的并行度
			EnableLog:     true,
			Variables: map[string]interface{}{
				"parallel_mode": true,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// RunExample 运行完整示例
func RunExample() {
	logger := &SimpleLogger{}

	// 创建组件
	registry := capability.NewRegistry()
	dagEngine := NewDAGEngine(logger)
	dataFlow := NewDataFlowEngine(dagEngine, logger)
	executor := NewWorkflowExecutor(nil, registry, dagEngine, dataFlow, logger)

	// 启动插件
	// plugins := []string{"http-plugin-1", "http-plugin-2", "http-plugin-3", "http-plugin-4"}
	// for _, pluginID := range plugins {
	// 	plugin, err := pluginManager.StartPlugin(context.Background(), pluginID)
	// 	if err != nil {
	// 		logger.Error("Failed to start plugin", "plugin_id", pluginID, "error", err)
	// 		return
	// 	}
	// 	logger.Info("Plugin started", "plugin_id", pluginID, "name", plugin.Name)
	// }

	// 创建工作流
	workflow := CreateExampleWorkflow()
	logger.Info("Workflow created", "workflow_id", workflow.ID, "name", workflow.Name)

	// 执行工作流
	inputs := map[string]interface{}{
		"data": map[string]interface{}{
			"id":      12345,
			"name":    "测试数据",
			"content": "这是一个测试数据的内容",
			"type":    "test",
		},
		"operation": "transform",
	}

	execution, err := executor.Execute(context.Background(), workflow, inputs)
	if err != nil {
		logger.Error("Failed to execute workflow", "error", err)
		return
	}

	logger.Info("Workflow execution started", "execution_id", execution.ID)

	// 监控执行状态
	monitorExecution(executor, execution.ID)

	// 运行并行工作流示例
	logger.Info("\n" + strings.Repeat("=", 50))
	logger.Info("Running parallel workflow example")
	logger.Info(strings.Repeat("=", 50))

	parallelWorkflow := CreateParallelWorkflow()
	parallelInputs := map[string]interface{}{
		"data": map[string]interface{}{
			"batch_id": "batch_001",
			"items":    []string{"item1", "item2", "item3"},
		},
	}

	parallelExecution, err := executor.Execute(context.Background(), parallelWorkflow, parallelInputs)
	if err != nil {
		logger.Error("Failed to execute parallel workflow", "error", err)
		return
	}

	logger.Info("Parallel workflow execution started", "execution_id", parallelExecution.ID)
	monitorExecution(executor, parallelExecution.ID)
}

// monitorExecution 监控执行状态
func monitorExecution(executor WorkflowExecutor, executionID string) {
	ctx := context.Background()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-ticker.C:
			execution, exists := executor.GetExecution(executionID)
			if !exists {
				logger.Error("Execution not found", "execution_id", executionID)
				return
			}

			logger.Info("Execution status",
				"execution_id", execution.ID,
				"status", execution.Status,
				"completed_nodes", countCompletedNodes(execution),
				"total_nodes", len(execution.NodeResults))

			// 打印节点状态
			for nodeID, result := range execution.NodeResults {
				logger.Info("Node status",
					"node_id", nodeID,
					"status", result.Status,
					"elapsed_time", result.ElapsedTime)
				if result.Error != "" {
					logger.Error("Node error", "node_id", nodeID, "error", result.Error)
				}
			}

			if execution.Status == ExecutionStatusCompleted {
				logger.Info("Execution completed successfully", "execution_id", executionID)
				printExecutionResult(execution)
				return
			} else if execution.Status == ExecutionStatusFailed {
				logger.Error("Execution failed", "execution_id", executionID, "error", execution.Error)
				return
			} else if execution.Status == ExecutionStatusCancelled {
				logger.Info("Execution cancelled", "execution_id", executionID)
				return
			}

		case <-timeout:
			logger.Error("Execution monitoring timeout", "execution_id", executionID)
			executor.Cancel(executionID)
			return
		}
	}
}

// countCompletedNodes 计算已完成节点数
func countCompletedNodes(execution *Execution) int {
	count := 0
	for _, result := range execution.NodeResults {
		if result.Status == NodeStatusCompleted {
			count++
		}
	}
	return count
}

// printExecutionResult 打印执行结果
func printExecutionResult(execution *Execution) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("EXECUTION RESULT")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("Execution ID: %s\n", execution.ID)
	fmt.Printf("Status: %s\n", execution.Status)
	fmt.Printf("Start Time: %s\n", execution.StartTime.Format(time.RFC3339))
	if execution.EndTime != nil {
		fmt.Printf("End Time: %s\n", execution.EndTime.Format(time.RFC3339))
		fmt.Printf("Duration: %v\n", execution.EndTime.Sub(execution.StartTime))
	}

	fmt.Println("\nInputs:")
	for key, value := range execution.Inputs {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("\nOutputs:")
	for key, value := range execution.Outputs {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("\nNode Results:")
	for nodeID, result := range execution.NodeResults {
		fmt.Printf("  %s:\n", nodeID)
		fmt.Printf("    Status: %s\n", result.Status)
		fmt.Printf("    Elapsed Time: %v\n", result.ElapsedTime)
		if result.Error != "" {
			fmt.Printf("    Error: %s\n", result.Error)
		}
		if len(result.Outputs) > 0 {
			fmt.Printf("    Outputs:\n")
			for key, value := range result.Outputs {
				fmt.Printf("      %s: %v\n", key, value)
			}
		}
	}

	fmt.Println("\nRecent Logs:")
	logCount := len(execution.Logs)
	start := logCount - 10
	if start < 0 {
		start = 0
	}
	for i := start; i < logCount; i++ {
		log := execution.Logs[i]
		fmt.Printf("  [%s] %s: %s\n", log.Timestamp.Format("15:04:05"), log.Level, log.Message)
		if log.NodeID != "" {
			fmt.Printf("    Node: %s\n", log.NodeID)
		}
	}
	fmt.Println(strings.Repeat("=", 50))
}