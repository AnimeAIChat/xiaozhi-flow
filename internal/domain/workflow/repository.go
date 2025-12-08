package workflow

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"xiaozhi-server-go/internal/platform/storage"
)

// Repository defines the interface for workflow storage
type Repository interface {
	Create(ctx context.Context, workflow *storage.Workflow) error
	Get(ctx context.Context, id string) (*storage.Workflow, error)
	Update(ctx context.Context, workflow *storage.Workflow) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*storage.Workflow, error)
}

// GormRepository implements Repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GormRepository
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, workflow *storage.Workflow) error {
	return r.db.WithContext(ctx).Create(workflow).Error
}

func (r *GormRepository) Get(ctx context.Context, id string) (*storage.Workflow, error) {
	var workflow storage.Workflow
	if err := r.db.WithContext(ctx).First(&workflow, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &workflow, nil
}

func (r *GormRepository) Update(ctx context.Context, workflow *storage.Workflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

func (r *GormRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&storage.Workflow{}, "id = ?", id).Error
}

func (r *GormRepository) List(ctx context.Context) ([]*storage.Workflow, error) {
	var workflows []*storage.Workflow
	if err := r.db.WithContext(ctx).Find(&workflows).Error; err != nil {
		return nil, err
	}
	return workflows, nil
}
