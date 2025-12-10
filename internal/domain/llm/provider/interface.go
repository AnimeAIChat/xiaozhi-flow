package provider

import (
	"context"
	"xiaozhi-server-go/internal/domain/llm/repository"
)

// Provider defines the interface for LLM providers
type Provider interface {
	Generate(ctx context.Context, req repository.GenerateRequest) (*repository.GenerateResult, error)
	Stream(ctx context.Context, req repository.GenerateRequest) (<-chan repository.ResponseChunk, error)
	Type() string
}
