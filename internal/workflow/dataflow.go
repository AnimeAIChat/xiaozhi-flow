package workflow

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// DataFlowEngine 数据流引擎实现
type DataFlowEngine struct {
	dagEngine DAGEngine
	logger    Logger
}

// NewDataFlowEngine 创建数据流引擎
func NewDataFlowEngine(dagEngine DAGEngine, logger Logger) DataFlow {
	return &DataFlowEngine{
		dagEngine: dagEngine,
		logger:    logger,
	}
}

// PassDataToNode 传递数据到节点
func (e *DataFlowEngine) PassDataToNode(execution *Execution, nodeID string, data map[string]interface{}) error {
	if execution.NodeResults == nil {
		execution.NodeResults = make(map[string]*NodeResult)
	}

	// 获取或创建节点结果
	result, exists := execution.NodeResults[nodeID]
	if !exists {
		result = &NodeResult{
			NodeID: nodeID,
			Status: NodeStatusRunning,
			Inputs: make(map[string]interface{}),
			Outputs: make(map[string]interface{}),
		}
		execution.NodeResults[nodeID] = result
	}

	// 更新输出数据
	if result.Outputs == nil {
		result.Outputs = make(map[string]interface{})
	}

	for key, value := range data {
		result.Outputs[key] = value
	}

	// 更新节点状态
	result.Status = NodeStatusCompleted
	endTime := time.Now()
	result.EndTime = &endTime
	if !result.StartTime.IsZero() {
		result.ElapsedTime = endTime.Sub(result.StartTime)
	}

	e.logger.Debug("Data passed to node", "execution_id", execution.ID, "node_id", nodeID, "data_keys", getKeys(data))

	return nil
}

// GetNodeInputs 获取节点输入数据
func (e *DataFlowEngine) GetNodeInputs(execution *Execution, node *Node, workflow *Workflow) (map[string]interface{}, error) {
	inputs := make(map[string]interface{})

	// 获取节点依赖
	dependencies := e.dagEngine.GetNodeDependencies(node.ID, workflow.Edges)

	// 合并所有依赖节点的输出数据
	mergedData := make(map[string]interface{})
	for _, depID := range dependencies {
		if result, exists := execution.NodeResults[depID]; exists && result.Status == NodeStatusCompleted {
			for key, value := range result.Outputs {
				mergedData[fmt.Sprintf("%s.%s", depID, key)] = value
			}
		} else {
			return nil, fmt.Errorf("dependency node %s is not completed", depID)
		}
	}

	// 合并全局变量和执行上下文
	for key, value := range workflow.Config.Variables {
		mergedData[fmt.Sprintf("global.%s", key)] = value
	}

	for key, value := range execution.Context {
		mergedData[fmt.Sprintf("context.%s", key)] = value
	}

	// 根据节点输入Schema映射数据
	for _, inputSchema := range node.Inputs {
		value, err := e.resolveInputValue(inputSchema, mergedData, execution.Inputs)
		if err != nil {
			if inputSchema.Required {
				return nil, fmt.Errorf("failed to resolve input %s: %w", inputSchema.Name, err)
			}
			// 非必需字段，使用默认值
			if inputSchema.Default != nil {
				inputs[inputSchema.Name] = inputSchema.Default
			}
			continue
		}
		inputs[inputSchema.Name] = value
	}

	e.logger.Debug("Node inputs resolved", "execution_id", execution.ID, "node_id", node.ID, "inputs_count", len(inputs))

	return inputs, nil
}

// resolveInputValue 解析输入值
func (e *DataFlowEngine) resolveInputValue(schema InputSchema, data map[string]interface{}, executionInputs map[string]interface{}) (interface{}, error) {
	// 首先检查执行输入中是否有直接提供该值
	if value, exists := executionInputs[schema.Name]; exists {
		return e.validateAndConvert(schema, value)
	}

	// 查找数据中的匹配项
	if value, exists := data[schema.Name]; exists {
		return e.validateAndConvert(schema, value)
	}

	// 查找依赖节点的输出
	for key, value := range data {
		if strings.HasSuffix(key, "."+schema.Name) {
			return e.validateAndConvert(schema, value)
		}
	}

	// 如果没有找到值且不是必需的，返回默认值
	if !schema.Required && schema.Default != nil {
		return e.validateAndConvert(schema, schema.Default)
	}

	return nil, fmt.Errorf("input value not found for %s", schema.Name)
}

// validateAndConvert 验证并转换值
func (e *DataFlowEngine) validateAndConvert(schema InputSchema, value interface{}) (interface{}, error) {
	// 类型验证和转换
	switch schema.Type {
	case "string":
		strValue, err := e.convertToString(value)
		if err != nil {
			return nil, fmt.Errorf("cannot convert to string: %w", err)
		}
		return e.validateString(strValue, schema)
	case "number":
		numValue, err := e.convertToNumber(value)
		if err != nil {
			return nil, fmt.Errorf("cannot convert to number: %w", err)
		}
		return e.validateNumber(numValue, schema)
	case "boolean":
		boolValue, err := e.convertToBoolean(value)
		if err != nil {
			return nil, fmt.Errorf("cannot convert to boolean: %w", err)
		}
		return boolValue, nil
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("value is not an object")
		}
		return value, nil
	case "array":
		if reflect.TypeOf(value).Kind() != reflect.Slice {
			return nil, fmt.Errorf("value is not an array")
		}
		return value, nil
	default:
		return value, nil
	}
}

// convertToString 转换为字符串
func (e *DataFlowEngine) convertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// convertToNumber 转换为数字
func (e *DataFlowEngine) convertToNumber(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to number", value)
	}
}

// convertToBoolean 转换为布尔值
func (e *DataFlowEngine) convertToBoolean(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int, int32, int64:
		return reflect.ValueOf(value).Int() != 0, nil
	case float32, float64:
		return reflect.ValueOf(value).Float() != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to boolean", value)
	}
}

// validateString 验证字符串
func (e *DataFlowEngine) validateString(value string, schema InputSchema) (string, error) {
	// 长度验证
	if schema.Validation != nil {
		if schema.Validation.MinLength != nil && len(value) < *schema.Validation.MinLength {
			return "", fmt.Errorf("string length %d is less than minimum %d", len(value), *schema.Validation.MinLength)
		}
		if schema.Validation.MaxLength != nil && len(value) > *schema.Validation.MaxLength {
			return "", fmt.Errorf("string length %d is greater than maximum %d", len(value), *schema.Validation.MaxLength)
		}
		// 模式验证
		if schema.Validation.Pattern != "" {
			// 这里应该使用正则表达式验证，简化处理
			e.logger.Debug("Pattern validation skipped for simplicity", "pattern", schema.Validation.Pattern)
		}
		// 枚举验证
		if len(schema.Validation.Enum) > 0 {
			found := false
			for _, enum := range schema.Validation.Enum {
				if value == enum {
					found = true
					break
				}
			}
			if !found {
				return "", fmt.Errorf("value '%s' is not in allowed enum values", value)
			}
		}
	}

	return value, nil
}

// validateNumber 验证数字
func (e *DataFlowEngine) validateNumber(value float64, schema InputSchema) (float64, error) {
	if schema.Validation != nil {
		if schema.Validation.Min != nil && value < *schema.Validation.Min {
			return 0, fmt.Errorf("number %f is less than minimum %f", value, *schema.Validation.Min)
		}
		if schema.Validation.Max != nil && value > *schema.Validation.Max {
			return 0, fmt.Errorf("number %f is greater than maximum %f", value, *schema.Validation.Max)
		}
	}

	return value, nil
}

// MergeParallelData 合并并行节点数据
func (e *DataFlowEngine) MergeParallelData(execution *Execution, nodeIDs []string) (map[string]interface{}, error) {
	mergedData := make(map[string]interface{})

	for _, nodeID := range nodeIDs {
		result, exists := execution.NodeResults[nodeID]
		if !exists {
			return nil, fmt.Errorf("node result not found: %s", nodeID)
		}

		if result.Status != NodeStatusCompleted {
			return nil, fmt.Errorf("node %s is not completed", nodeID)
		}

		// 合并输出数据，使用节点ID作为前缀避免冲突
		for key, value := range result.Outputs {
			mergedData[fmt.Sprintf("%s.%s", nodeID, key)] = value
		}
	}

	e.logger.Debug("Parallel data merged", "execution_id", execution.ID, "node_count", len(nodeIDs), "data_keys", getKeys(mergedData))

	return mergedData, nil
}

// TransformData 数据转换
func (e *DataFlowEngine) TransformData(data map[string]interface{}, mappings map[string]string) map[string]interface{} {
	transformed := make(map[string]interface{})

	for targetKey, sourcePath := range mappings {
		if value, exists := e.getValueByPath(data, sourcePath); exists {
			transformed[targetKey] = value
		}
	}

	// 保留未映射的数据
	for key, value := range data {
		if _, mapped := transformed[key]; !mapped {
			transformed[key] = value
		}
	}

	return transformed
}

// getValueByPath 通过路径获取值
func (e *DataFlowEngine) getValueByPath(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if value, exists := current[part]; exists {
			if i == len(parts)-1 {
				return value, true
			}

			if nextMap, ok := value.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return nil, false
}

// ValidateData 验证数据完整性
func (e *DataFlowEngine) ValidateData(data map[string]interface{}, schemas []InputSchema) error {
	for _, schema := range schemas {
		if schema.Required {
			if _, exists := data[schema.Name]; !exists {
				return fmt.Errorf("required input %s is missing", schema.Name)
			}
		}

		if value, exists := data[schema.Name]; exists {
			_, err := e.validateAndConvert(schema, value)
			if err != nil {
				return fmt.Errorf("validation failed for input %s: %w", schema.Name, err)
			}
		}
	}

	return nil
}

// CloneData 克隆数据
func (e *DataFlowEngine) CloneData(data map[string]interface{}) map[string]interface{} {
	clone := make(map[string]interface{})
	for key, value := range data {
		clone[key] = value
	}
	return clone
}

// getKeys 获取map的所有键
func getKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}

// PathExpression 路径表达式
type PathExpression struct {
	Source string `json:"source"`
	Path   string `json:"path"`
	Target string `json:"target"`
}

// ExpressionEvaluator 表达式求值器
type ExpressionEvaluator struct {
	logger Logger
}

// NewExpressionEvaluator 创建表达式求值器
func NewExpressionEvaluator(logger Logger) *ExpressionEvaluator {
	return &ExpressionEvaluator{
		logger: logger,
	}
}

// EvaluateExpression 求值表达式
func (eval *ExpressionEvaluator) EvaluateExpression(expression string, data map[string]interface{}) (interface{}, error) {
	// 这里应该实现完整的表达式求值逻辑
	// 目前实现简单的变量替换
	result := expression
	for key, value := range data {
		placeholder := fmt.Sprintf("${%s}", key)
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}

	// 尝试解析为JSON或基本类型
	if result != expression {
		// 有替换发生，尝试解析结果
		var parsed interface{}
		if err := json.Unmarshal([]byte(result), &parsed); err == nil {
			return parsed, nil
		}
		return result, nil
	}

	return data[expression], nil
}