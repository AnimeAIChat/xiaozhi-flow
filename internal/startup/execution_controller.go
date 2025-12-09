package startup

import (
	"context"
	"xiaozhi-server-go/internal/startup/model"
)

// ExecutionController 执行控制器接口
type ExecutionController interface {
	// ExecuteWorkflow 执行工作流
	ExecuteWorkflow(ctx context.Context, workflowID string, inputs map[string]interface{}) (*StartupWorkflowExecution, error)

	// GetExecution 获取执行实例
	GetExecution(executionID string) (*StartupWorkflowExecution, bool)

	// CancelExecution 取消执行
	CancelExecution(executionID string) error

	// PauseExecution 暂停执行
	PauseExecution(executionID string) error

	// ResumeExecution 恢复执行
	ResumeExecution(executionID string) error
}

// ExecutionControllerAdapter 执行控制器适配器
type ExecutionControllerAdapter struct {
	executor *StartupWorkflowExecutor
}

// NewExecutionControllerAdapter 创建执行控制器适配器
func NewExecutionControllerAdapter(executor *StartupWorkflowExecutor) *ExecutionControllerAdapter {
	return &ExecutionControllerAdapter{
		executor: executor,
	}
}

// ExecuteWorkflow 执行工作流
func (a *ExecutionControllerAdapter) ExecuteWorkflow(ctx context.Context, workflowID string, inputs map[string]interface{}) (*StartupWorkflowExecution, error) {
	if a.executor != nil {
		return a.executor.ExecuteWorkflow(ctx, workflowID, inputs)
	}
	return nil, model.NewStartupError("EXECUTOR_NOT_AVAILABLE", "execution controller is not available")
}

// GetExecution 获取执行实例
func (a *ExecutionControllerAdapter) GetExecution(executionID string) (*StartupWorkflowExecution, bool) {
	if a.executor != nil {
		return a.executor.GetExecution(executionID)
	}
	return nil, false
}

// CancelExecution 取消执行
func (a *ExecutionControllerAdapter) CancelExecution(executionID string) error {
	if a.executor != nil {
		return a.executor.CancelExecution(executionID)
	}
	return model.ErrExecutionNotFound
}

// PauseExecution 暂停执行
func (a *ExecutionControllerAdapter) PauseExecution(executionID string) error {
	if a.executor != nil {
		return a.executor.PauseExecution(executionID)
	}
	return model.ErrExecutionNotFound
}

// ResumeExecution 恢复执行
func (a *ExecutionControllerAdapter) ResumeExecution(executionID string) error {
	if a.executor != nil {
		return a.executor.ResumeExecution(executionID)
	}
	return model.ErrExecutionNotFound
}

