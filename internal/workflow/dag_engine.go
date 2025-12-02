package workflow

import (
	"fmt"
	"sort"
)

// DAGEngineImpl DAG引擎实现
type DAGEngineImpl struct {
	logger Logger
}

// NewDAGEngine 创建DAG引擎
func NewDAGEngine(logger Logger) DAGEngine {
	return &DAGEngineImpl{
		logger: logger,
	}
}

// TopologicalSort 拓扑排序
func (e *DAGEngineImpl) TopologicalSort(nodes []Node, edges []Edge) ([]string, error) {
	// 构建邻接表和入度表
	adjacency := make(map[string][]string)
	inDegree := make(map[string]int)
	nodeSet := make(map[string]bool)

	// 初始化节点集合
	for _, node := range nodes {
		nodeSet[node.ID] = true
		inDegree[node.ID] = 0
	}

	// 构建邻接表和入度
	for _, edge := range edges {
		if !nodeSet[edge.From] || !nodeSet[edge.To] {
			return nil, fmt.Errorf("edge references non-existent node: %s -> %s", edge.From, edge.To)
		}

		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
		inDegree[edge.To]++

		// 确保from节点在入度表中
		if _, exists := inDegree[edge.From]; !exists {
			inDegree[edge.From] = 0
		}
	}

	// 检查是否有循环依赖
	if e.HasCycle(nodes, edges) {
		return nil, fmt.Errorf("workflow contains cycles")
	}

	// Kahn算法进行拓扑排序
	queue := make([]string, 0)
	result := make([]string, 0)

	// 找到所有入度为0的节点
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	// 按字母顺序排序入度为0的节点，确保确定性输出
	sort.Strings(queue)

	for len(queue) > 0 {
		// 取出第一个节点
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// 遍历当前节点的所有邻居
		for _, neighbor := range adjacency[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				// 插入到正确位置保持排序
				queue = insertSorted(queue, neighbor)
			}
		}
	}

	// 检查是否所有节点都被处理
	if len(result) != len(nodes) {
		return nil, fmt.Errorf("topological sort failed: cycle detected")
	}

	e.logger.Debug("Topological sort completed", "nodes_count", len(result))

	return result, nil
}

// HasCycle 检查循环依赖
func (e *DAGEngineImpl) HasCycle(nodes []Node, edges []Edge) bool {
	// 构建邻接表
	adjacency := make(map[string][]string)
	nodeSet := make(map[string]bool)

	for _, node := range nodes {
		nodeSet[node.ID] = true
	}

	for _, edge := range edges {
		if !nodeSet[edge.From] || !nodeSet[edge.To] {
			continue // 忽略无效边
		}
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}

	// DFS检测循环
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for nodeID := range nodeSet {
		if !visited[nodeID] {
			if e.hasCycleDFS(nodeID, adjacency, visited, recursionStack) {
				return true
			}
		}
	}

	return false
}

// hasCycleDFS DFS检测循环的辅助函数
func (e *DAGEngineImpl) hasCycleDFS(nodeID string, adjacency map[string][]string, visited, recursionStack map[string]bool) bool {
	visited[nodeID] = true
	recursionStack[nodeID] = true

	// 遍历所有邻居
	for _, neighbor := range adjacency[nodeID] {
		if !visited[neighbor] {
			if e.hasCycleDFS(neighbor, adjacency, visited, recursionStack) {
				return true
			}
		} else if recursionStack[neighbor] {
			// 如果在当前递归栈中找到节点，说明存在循环
			return true
		}
	}

	// 从递归栈中移除当前节点
	recursionStack[nodeID] = false
	return false
}

// GetExecutableNodes 获取可执行节点
func (e *DAGEngineImpl) GetExecutableNodes(execution *Execution, workflow *Workflow) ([]string, error) {
	// 构建节点状态映射
	nodeStatus := make(map[string]NodeStatus)
	for _, node := range workflow.Nodes {
		// 检查节点是否已完成
		if result, exists := execution.NodeResults[node.ID]; exists {
			nodeStatus[node.ID] = result.Status
		} else {
			nodeStatus[node.ID] = NodeStatusPending
		}
	}

	// 构建邻接表
	adjacency := make(map[string][]string)
	for _, edge := range workflow.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}

	var executableNodes []string

	for _, node := range workflow.Nodes {
		// 跳过已完成的节点
		if nodeStatus[node.ID] == NodeStatusCompleted ||
		   nodeStatus[node.ID] == NodeStatusRunning ||
		   nodeStatus[node.ID] == NodeStatusFailed {
			continue
		}

		// 检查所有依赖节点是否已完成
		if e.canExecuteNode(node.ID, adjacency, nodeStatus, workflow) {
			executableNodes = append(executableNodes, node.ID)
		}
	}

	e.logger.Debug("Found executable nodes", "count", len(executableNodes))

	return executableNodes, nil
}

// canExecuteNode 检查节点是否可以执行
func (e *DAGEngineImpl) canExecuteNode(nodeID string, adjacency map[string][]string, nodeStatus map[string]NodeStatus, workflow *Workflow) bool {
	// 开始节点总是可以执行
	for _, node := range workflow.Nodes {
		if node.ID == nodeID && node.Type == NodeTypeStart {
			return true
		}
	}

	// 检查是否有前置节点
	var dependencies []string
	for from, neighbors := range adjacency {
		for _, to := range neighbors {
			if to == nodeID {
				dependencies = append(dependencies, from)
			}
		}
	}

	// 如果没有前置节点，可以执行
	if len(dependencies) == 0 {
		return true
	}

	// 检查所有前置节点是否已完成
	for _, depID := range dependencies {
		status, exists := nodeStatus[depID]
		if !exists || status != NodeStatusCompleted {
			return false
		}
	}

	return true
}

// GetNodeDependencies 获取节点依赖
func (e *DAGEngineImpl) GetNodeDependencies(nodeID string, edges []Edge) []string {
	var dependencies []string
	for _, edge := range edges {
		if edge.To == nodeID {
			dependencies = append(dependencies, edge.From)
		}
	}
	return dependencies
}

// GetNodeDependents 获取节点的后续节点
func (e *DAGEngineImpl) GetNodeDependents(nodeID string, edges []Edge) []string {
	var dependents []string
	for _, edge := range edges {
		if edge.From == nodeID {
			dependents = append(dependents, edge.To)
		}
	}
	return dependents
}

// GetCriticalPath 获取关键路径
func (e *DAGEngineImpl) GetCriticalPath(workflow *Workflow, execution *Execution) ([]string, time.Duration) {
	// 构建邻接表
	adjacency := make(map[string][]string)
	inDegree := make(map[string]int)
	nodeSet := make(map[string]*Node)

	for i := range workflow.Nodes {
		node := &workflow.Nodes[i]
		nodeSet[node.ID] = node
		inDegree[node.ID] = 0
	}

	for _, edge := range workflow.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
		inDegree[edge.To]++
	}

	// 计算最长路径
 longestPath := make(map[string]time.Duration)
 longestPathNodes := make(map[string][]string)

	// 找到起始节点
	var startNodes []string
	for nodeID, degree := range inDegree {
		if degree == 0 {
			startNodes = append(startNodes, nodeID)
			longestPath[nodeID] = 0
			longestPathNodes[nodeID] = []string{nodeID}
		}
	}

	// 拓扑排序并计算最长路径
	queue := startNodes
	sorted := make([]string, 0)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		for _, next := range adjacency[current] {
			currentDuration := longestPath[current]
			if result, exists := execution.NodeResults[current]; exists {
				currentDuration = result.ElapsedTime
			}

			proposedDuration := longestPath[current] + currentDuration
			if proposedDuration > longestPath[next] {
				longestPath[next] = proposedDuration
				longestPathNodes[next] = append(longestPathNodes[current], next)
			}

			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	// 找到关键路径
	var criticalPath []string
	var maxDuration time.Duration

	for _, path := range longestPathNodes {
		var totalDuration time.Duration
		for _, nodeID := range path {
			if result, exists := execution.NodeResults[nodeID]; exists {
				totalDuration += result.ElapsedTime
			}
		}

		if totalDuration > maxDuration {
			maxDuration = totalDuration
			criticalPath = path
		}
	}

	return criticalPath, maxDuration
}

// insertSorted 将字符串插入到已排序的切片中
func insertSorted(slice []string, value string) []string {
	index := sort.SearchStrings(slice, value)
	if index == len(slice) {
		return append(slice, value)
	}
	if slice[index] != value {
		slice = append(slice, "")
		copy(slice[index+1:], slice[index:])
		slice[index] = value
	}
	return slice
}

// ValidateWorkflow 验证工作流的正确性
func (e *DAGEngineImpl) ValidateWorkflow(workflow *Workflow) error {
	// 检查基本字段
	if workflow.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	if len(workflow.Nodes) == 0 {
		return fmt.Errorf("workflow must have at least one node")
	}

	// 检查节点ID唯一性
	nodeIDs := make(map[string]bool)
	for _, node := range workflow.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node ID is required")
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true
	}

	// 检查边的有效性
	for _, edge := range workflow.Edges {
		if !nodeIDs[edge.From] {
			return fmt.Errorf("edge references non-existent node: %s", edge.From)
		}
		if !nodeIDs[edge.To] {
			return fmt.Errorf("edge references non-existent node: %s", edge.To)
		}
		if edge.From == edge.To {
			return fmt.Errorf("self-loop detected for node: %s", edge.From)
		}
	}

	// 检查循环依赖
	if e.HasCycle(workflow.Nodes, workflow.Edges) {
		return fmt.Errorf("workflow contains cycles")
	}

	// 检查是否有开始和结束节点
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
		return fmt.Errorf("workflow must have at least one start node")
	}

	if !hasEnd {
		return fmt.Errorf("workflow must have at least one end node")
	}

	return nil
}