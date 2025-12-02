package workflow

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockLogger 模拟日志器
type MockLogger struct {
	logs []string
}

func (l *MockLogger) Info(msg string, fields ...interface{}) {
	l.logs = append(l.logs, "INFO: "+msg)
}

func (l *MockLogger) Error(msg string, fields ...interface{}) {
	l.logs = append(l.logs, "ERROR: "+msg)
}

func (l *MockLogger) Warn(msg string, fields ...interface{}) {
	l.logs = append(l.logs, "WARN: "+msg)
}

func (l *MockLogger) Debug(msg string, fields ...interface{}) {
	l.logs = append(l.logs, "DEBUG: "+msg)
}

func TestDAGEngine_TopologicalSort(t *testing.T) {
	logger := &MockLogger{}
	engine := NewDAGEngine(logger)

	nodes := []Node{
		{ID: "A", Name: "Node A", Type: NodeTypeTask},
		{ID: "B", Name: "Node B", Type: NodeTypeTask},
		{ID: "C", Name: "Node C", Type: NodeTypeTask},
	}

	edges := []Edge{
		{ID: "e1", From: "A", To: "B"},
		{ID: "e2", From: "B", To: "C"},
	}

	result, err := engine.TopologicalSort(nodes, edges)
	if err != nil {
		t.Fatalf("Topological sort failed: %v", err)
	}

	expected := []string{"A", "B", "C"}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d nodes, got %d", len(expected), len(result))
	}

	for i, nodeID := range expected {
		if result[i] != nodeID {
			t.Errorf("Expected node %s at position %d, got %s", nodeID, i, result[i])
		}
	}
}

func TestDAGEngine_HasCycle(t *testing.T) {
	logger := &MockLogger{}
	engine := NewDAGEngine(logger)

	tests := []struct {
		name     string
		nodes    []Node
		edges    []Edge
		hasCycle bool
	}{
		{
			name: "No cycle",
			nodes: []Node{
				{ID: "A", Type: NodeTypeTask},
				{ID: "B", Type: NodeTypeTask},
				{ID: "C", Type: NodeTypeTask},
			},
			edges: []Edge{
				{From: "A", To: "B"},
				{From: "B", To: "C"},
			},
			hasCycle: false,
		},
		{
			name: "Has cycle",
			nodes: []Node{
				{ID: "A", Type: NodeTypeTask},
				{ID: "B", Type: NodeTypeTask},
				{ID: "C", Type: NodeTypeTask},
			},
			edges: []Edge{
				{From: "A", To: "B"},
				{From: "B", To: "C"},
				{From: "C", To: "A"},
			},
			hasCycle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasCycle := engine.HasCycle(tt.nodes, tt.edges)
			if hasCycle != tt.hasCycle {
				t.Errorf("Expected hasCycle=%v, got %v", tt.hasCycle, hasCycle)
			}
		})
	}
}

func TestDataFlowEngine_GetNodeInputs(t *testing.T) {
	logger := &MockLogger{}
	dagEngine := NewDAGEngine(logger)
	dataFlow := NewDataFlowEngine(dagEngine, logger)

	workflow := &Workflow{
		Nodes: []Node{
			{
				ID:   "node1",
				Type: NodeTypeTask,
				Inputs: []InputSchema{
					{
						Name:     "input1",
						Type:     "string",
						Required: true,
					},
					{
						Name:     "input2",
						Type:     "number",
						Required: false,
						Default:  42,
					},
				},
			},
		},
		Edges: []Edge{},
		Config: WorkflowConfig{
			Variables: map[string]interface{}{
				"global_var": "global_value",
			},
		},
	}

	execution := &Execution{
		ID: "test-exec",
		NodeResults: map[string]*NodeResult{},
		Context: map[string]interface{}{
			"context_var": "context_value",
		},
		Inputs: map[string]interface{}{
			"input1": "test_value",
		},
	}

	// 获取节点输入
	node := &workflow.Nodes[0]
	inputs, err := dataFlow.GetNodeInputs(execution, node, workflow)
	if err != nil {
		t.Fatalf("Failed to get node inputs: %v", err)
	}

	// 验证输入
	if inputs["input1"] != "test_value" {
		t.Errorf("Expected input1=test_value, got %v", inputs["input1"])
	}

	if inputs["input2"] != 42.0 {
		t.Errorf("Expected input2=42, got %v", inputs["input2"])
	}
}

func TestPluginManager_StartAndStopPlugin(t *testing.T) {
	logger := &MockLogger{}
	manager := NewHTTPPluginManager(logger)

	ctx := context.Background()
	pluginID := "test-plugin"

	// 启动插件
	plugin, err := manager.StartPlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("Failed to start plugin: %v", err)
	}

	if plugin.ID != pluginID {
		t.Errorf("Expected plugin ID %s, got %s", pluginID, plugin.ID)
	}

	if plugin.Status != PluginStatusRunning {
		t.Errorf("Expected plugin status %s, got %s", PluginStatusRunning, plugin.Status)
	}

	// 停止插件
	err = manager.StopPlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("Failed to stop plugin: %v", err)
	}

	// 验证插件状态
	plugin, exists := manager.GetPlugin(pluginID)
	if !exists {
		t.Error("Plugin should still exist after stopping")
	}

	if plugin.Status != PluginStatusStopped {
		t.Errorf("Expected plugin status %s, got %s", PluginStatusStopped, plugin.Status)
	}
}

func TestWorkflowExecutor_ExecuteSimpleWorkflow(t *testing.T) {
	logger := &MockLogger{}
	pluginManager := NewHTTPPluginManager(logger)
	dagEngine := NewDAGEngine(logger)
	dataFlow := NewDataFlowEngine(dagEngine, logger)
	executor := NewWorkflowExecutor(pluginManager, dagEngine, dataFlow, logger)

	// 启动插件
	ctx := context.Background()
	plugin, err := pluginManager.StartPlugin(ctx, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to start plugin: %v", err)
	}

	// 创建简单工作流
	workflow := &Workflow{
		ID:   "simple-workflow",
		Name: "Simple Test Workflow",
		Nodes: []Node{
			{
				ID:   "start",
				Name: "Start",
				Type: NodeTypeStart,
				Inputs: []InputSchema{
					{Name: "data", Type: "object", Required: true},
				},
				Position: Position{X: 0, Y: 0},
			},
			{
				ID:     "task",
				Name:   "Task",
				Type:   NodeTypeTask,
				Plugin: plugin.ID,
				Method: "echo",
				Inputs: []InputSchema{
					{Name: "input", Type: "object", Required: true},
				},
				Position: Position{X: 100, Y: 0},
			},
			{
				ID:   "end",
				Name: "End",
				Type: NodeTypeEnd,
				Position: Position{X: 200, Y: 0},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "task"},
			{ID: "e2", From: "task", To: "end"},
		},
		Config: WorkflowConfig{
			Timeout:       30 * time.Second,
			MaxRetries:    1,
			ParallelLimit: 3,
			EnableLog:     true,
		},
	}

	// 执行工作流
	inputs := map[string]interface{}{
		"data": map[string]interface{}{
			"message": "Hello, World!",
		},
	}

	execution, err := executor.Execute(ctx, workflow, inputs)
	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	// 等待执行完成
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			currentExecution, exists := executor.GetExecution(execution.ID)
			if !exists {
				t.Error("Execution not found")
				return
			}

			if currentExecution.Status == ExecutionStatusCompleted {
				t.Log("Workflow execution completed successfully")
				return
			} else if currentExecution.Status == ExecutionStatusFailed {
				t.Errorf("Workflow execution failed: %s", currentExecution.Error)
				return
			}

		case <-timeout:
			t.Error("Workflow execution timeout")
			executor.Cancel(execution.ID)
			return
		}
	}
}

func TestWorkflowValidation(t *testing.T) {
	logger := &MockLogger{}
	engine := NewDAGEngine(logger)

	tests := []struct {
		name      string
		workflow  *Workflow
		shouldErr bool
	}{
		{
			name: "Valid workflow",
			workflow: &Workflow{
				ID:   "valid",
				Nodes: []Node{
					{ID: "start", Type: NodeTypeStart},
					{ID: "end", Type: NodeTypeEnd},
				},
				Edges: []Edge{
					{From: "start", To: "end"},
				},
			},
			shouldErr: false,
		},
		{
			name: "Empty workflow",
			workflow: &Workflow{
				ID:    "empty",
				Nodes: []Node{},
			},
			shouldErr: true,
		},
		{
			name: "Workflow with cycle",
			workflow: &Workflow{
				ID:   "cycle",
				Nodes: []Node{
					{ID: "A", Type: NodeTypeTask},
					{ID: "B", Type: NodeTypeTask},
				},
				Edges: []Edge{
					{From: "A", To: "B"},
					{From: "B", To: "A"},
				},
			},
			shouldErr: true,
		},
		{
			name: "Workflow without start node",
			workflow: &Workflow{
				ID:   "no-start",
				Nodes: []Node{
					{ID: "A", Type: NodeTypeTask},
					{ID: "B", Type: NodeTypeEnd},
				},
				Edges: []Edge{
					{From: "A", To: "B"},
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateWorkflow(tt.workflow)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Expected error=%v, got %v", tt.shouldErr, err)
			}
		})
	}
}

func TestDataValidation(t *testing.T) {
	logger := &MockLogger{}
	dagEngine := NewDAGEngine(logger)
	dataFlow := NewDataFlowEngine(dagEngine, logger)

	tests := []struct {
		name      string
		data      map[string]interface{}
		schemas   []InputSchema
		shouldErr bool
	}{
		{
			name: "Valid data",
			data: map[string]interface{}{
				"name":  "John",
				"age":   30,
				"email": "john@example.com",
			},
			schemas: []InputSchema{
				{Name: "name", Type: "string", Required: true},
				{Name: "age", Type: "number", Required: true},
				{Name: "email", Type: "string", Required: false},
			},
			shouldErr: false,
		},
		{
			name: "Missing required field",
			data: map[string]interface{}{
				"age": 30,
			},
			schemas: []InputSchema{
				{Name: "name", Type: "string", Required: true},
				{Name: "age", Type: "number", Required: true},
			},
			shouldErr: true,
		},
		{
			name: "Invalid type",
			data: map[string]interface{}{
				"name": "John",
				"age":  "thirty", // Should be number
			},
			schemas: []InputSchema{
				{Name: "name", Type: "string", Required: true},
				{Name: "age", Type: "number", Required: true},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dataFlow.ValidateData(tt.data, tt.schemas)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Expected error=%v, got %v", tt.shouldErr, err)
			}
		})
	}
}

// BenchmarkDAGTopologicalSort 性能测试
func BenchmarkDAGTopologicalSort(b *testing.B) {
	logger := &MockLogger{}
	engine := NewDAGEngine(logger)

	// 创建较大的DAG用于性能测试
	nodes := make([]Node, 100)
	for i := 0; i < 100; i++ {
		nodes[i] = Node{
			ID:   fmt.Sprintf("node-%d", i),
			Name: fmt.Sprintf("Node %d", i),
			Type: NodeTypeTask,
		}
	}

	edges := make([]Edge, 99)
	for i := 0; i < 99; i++ {
		edges[i] = Edge{
			ID:   fmt.Sprintf("edge-%d", i),
			From: fmt.Sprintf("node-%d", i),
			To:   fmt.Sprintf("node-%d", i+1),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.TopologicalSort(nodes, edges)
		if err != nil {
			b.Fatalf("Topological sort failed: %v", err)
		}
	}
}

func TestExampleWorkflowCreation(t *testing.T) {
	workflow := CreateExampleWorkflow()

	if workflow.ID == "" {
		t.Error("Workflow ID should not be empty")
	}

	if len(workflow.Nodes) == 0 {
		t.Error("Workflow should have nodes")
	}

	if len(workflow.Edges) == 0 {
		t.Error("Workflow should have edges")
	}

	// 验证节点类型
	hasStart := false
	hasEnd := false
	for _, node := range workflow.Nodes {
		if node.Type == NodeTypeStart {
			hasStart = true
		}
		if node.Type == NodeTypeEnd {
			hasEnd = true
		}
	}

	if !hasStart {
		t.Error("Workflow should have a start node")
	}

	if !hasEnd {
		t.Error("Workflow should have an end node")
	}
}