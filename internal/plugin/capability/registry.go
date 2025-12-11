package capability

import (
	"fmt"
	"sync"
)

type Registry struct {
	providers       map[string]Provider
	capabilities    map[string]Definition
	capToProvider   map[string]string // capabilityID -> providerID
	mu              sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		providers:     make(map[string]Provider),
		capabilities:  make(map[string]Definition),
		capToProvider: make(map[string]string),
	}
}

func (r *Registry) Register(providerID string, p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[providerID] = p
	for _, cap := range p.GetCapabilities() {
		r.capabilities[cap.ID] = cap
		r.capToProvider[cap.ID] = providerID
	}
}

func (r *Registry) GetExecutor(capabilityID string) (Executor, error) {
	r.mu.RLock()
	providerID, ok := r.capToProvider[capabilityID]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("capability not found: %s", capabilityID)
	}

	r.mu.RLock()
	provider, ok := r.providers[providerID]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider not found for capability: %s", capabilityID)
	}

	return provider.CreateExecutor(capabilityID)
}

// GetProvider 获取指定ID的提供者
func (r *Registry) GetProvider(providerID string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[providerID]
	return provider, ok
}

// GetAllProviders 获取所有提供者
func (r *Registry) GetAllProviders() map[string][]Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]Provider)
	for providerID, provider := range r.providers {
		result[providerID] = []Provider{provider}
	}
	return result
}

func (r *Registry) ListCapabilities() []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	caps := make([]Definition, 0, len(r.capabilities))
	for _, c := range r.capabilities {
		caps = append(caps, c)
	}
	return caps
}
