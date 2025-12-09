package startup

import (
	"context"
	"time"
	"xiaozhi-server-go/internal/startup/model"
	"xiaozhi-server-go/internal/platform/storage"
)

// Stub implementations
func NewFileWorkflowStorage(path string) model.WorkflowStorage { return &StubWorkflowStorage{} }
func NewFileExecutionStorage(path string) model.ExecutionStorage { return &StubExecutionStorage{} }
func NewFileTemplateStorage(path string) model.TemplateStorage { return &StubTemplateStorage{} }

func NewDatabaseWorkflowStorage(config *storage.DatabaseConfig) model.WorkflowStorage { return &StubWorkflowStorage{} }
func NewDatabaseExecutionStorage(config *storage.DatabaseConfig) model.ExecutionStorage { return &StubExecutionStorage{} }
func NewDatabaseTemplateStorage(config *storage.DatabaseConfig) model.TemplateStorage { return &StubTemplateStorage{} }

func NewMemoryWorkflowStorage() model.WorkflowStorage { return &StubWorkflowStorage{} }
func NewMemoryExecutionStorage() model.ExecutionStorage { return &StubExecutionStorage{} }
func NewMemoryTemplateStorage() model.TemplateStorage { return &StubTemplateStorage{} }

type StubWorkflowStorage struct{}
func (s *StubWorkflowStorage) Save(ctx context.Context, workflow *model.StartupWorkflow) error { return nil }
func (s *StubWorkflowStorage) Get(ctx context.Context, id string) (*model.StartupWorkflow, error) { return nil, nil }
func (s *StubWorkflowStorage) Delete(ctx context.Context, id string) error { return nil }
func (s *StubWorkflowStorage) List(ctx context.Context) ([]*model.StartupWorkflow, error) { return nil, nil }
func (s *StubWorkflowStorage) Update(ctx context.Context, workflow *model.StartupWorkflow) error { return nil }

type StubExecutionStorage struct{}
func (s *StubExecutionStorage) Save(ctx context.Context, execution *model.StartupExecution) error { return nil }
func (s *StubExecutionStorage) Get(ctx context.Context, id string) (*model.StartupExecution, error) { return nil, nil }
func (s *StubExecutionStorage) Delete(ctx context.Context, id string) error { return nil }
func (s *StubExecutionStorage) List(ctx context.Context, workflowID string) ([]*model.StartupExecution, error) { return nil, nil }
func (s *StubExecutionStorage) Cleanup(ctx context.Context, olderThan time.Time) error { return nil }

type StubTemplateStorage struct{}
func (s *StubTemplateStorage) Save(ctx context.Context, template *model.StartupWorkflowTemplate) error { return nil }
func (s *StubTemplateStorage) Get(ctx context.Context, id string) (*model.StartupWorkflowTemplate, error) { return nil, nil }
func (s *StubTemplateStorage) Delete(ctx context.Context, id string) error { return nil }
func (s *StubTemplateStorage) List(ctx context.Context) ([]*model.StartupWorkflowTemplate, error) {
	return nil, nil
}
func (s *StubTemplateStorage) Update(ctx context.Context, template *model.StartupWorkflowTemplate) error {
	return nil
}
