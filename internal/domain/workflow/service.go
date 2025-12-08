package workflow

import (
	"context"
	"time"

	"github.com/google/uuid"
	"xiaozhi-server-go/internal/platform/storage"
)

// Service defines the workflow service
type Service struct {
	repo Repository
}

// NewService creates a new workflow service
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateWorkflow creates a new workflow
func (s *Service) CreateWorkflow(ctx context.Context, name, description string, graphData interface{}) (*storage.Workflow, error) {
	workflow := &storage.Workflow{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		GraphData:   storage.FlexibleJSON{Data: graphData},
		IsActive:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, workflow); err != nil {
		return nil, err
	}

	return workflow, nil
}

// GetWorkflow gets a workflow by ID
func (s *Service) GetWorkflow(ctx context.Context, id string) (*storage.Workflow, error) {
	return s.repo.Get(ctx, id)
}

// UpdateWorkflow updates a workflow
func (s *Service) UpdateWorkflow(ctx context.Context, id string, name, description string, graphData interface{}, isActive bool) (*storage.Workflow, error) {
	workflow, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, nil
	}

	workflow.Name = name
	workflow.Description = description
	workflow.GraphData = storage.FlexibleJSON{Data: graphData}
	workflow.IsActive = isActive
	workflow.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, workflow); err != nil {
		return nil, err
	}

	return workflow, nil
}

// DeleteWorkflow deletes a workflow
func (s *Service) DeleteWorkflow(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ListWorkflows lists all workflows
func (s *Service) ListWorkflows(ctx context.Context) ([]*storage.Workflow, error) {
	return s.repo.List(ctx)
}
