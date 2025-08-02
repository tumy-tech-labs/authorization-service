package store

import (
	"context"
	"errors"
	"sync"

	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/bradtumy/authorization-service/pkg/tenant"
)

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu       sync.RWMutex
	tenants  map[string]tenant.Tenant
	policies map[string]map[string]policy.Policy       // tenantID -> policyID -> policy
	edges    map[string]map[string]map[string]struct{} // tenantID -> src -> dst set
}

// NewMemory returns a new MemoryStore instance.
func NewMemory() *MemoryStore {
	return &MemoryStore{
		tenants:  make(map[string]tenant.Tenant),
		policies: make(map[string]map[string]policy.Policy),
		edges:    make(map[string]map[string]map[string]struct{}),
	}
}

func (m *MemoryStore) SaveTenant(ctx context.Context, t tenant.Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tenants[t.ID] = t
	return nil
}

func (m *MemoryStore) LoadTenant(ctx context.Context, id string) (tenant.Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tenants[id]
	if !ok {
		return tenant.Tenant{}, errors.New("tenant not found")
	}
	return t, nil
}

func (m *MemoryStore) ListTenants(ctx context.Context) ([]tenant.Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]tenant.Tenant, 0, len(m.tenants))
	for _, t := range m.tenants {
		out = append(out, t)
	}
	return out, nil
}

func (m *MemoryStore) DeleteTenant(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tenants, id)
	delete(m.policies, id)
	delete(m.edges, id)
	return nil
}

func (m *MemoryStore) SavePolicy(ctx context.Context, tenantID string, p policy.Policy) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.policies[tenantID] == nil {
		m.policies[tenantID] = make(map[string]policy.Policy)
	}
	m.policies[tenantID][p.ID] = p
	return nil
}

func (m *MemoryStore) LoadPolicies(ctx context.Context, tenantID string) ([]policy.Policy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mpol := m.policies[tenantID]
	out := make([]policy.Policy, 0, len(mpol))
	for _, p := range mpol {
		out = append(out, p)
	}
	return out, nil
}

func (m *MemoryStore) ClearPolicies(ctx context.Context, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.policies, tenantID)
	return nil
}

func (m *MemoryStore) SaveEdge(ctx context.Context, tenantID, src, dst string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.edges[tenantID] == nil {
		m.edges[tenantID] = make(map[string]map[string]struct{})
	}
	if m.edges[tenantID][src] == nil {
		m.edges[tenantID][src] = make(map[string]struct{})
	}
	m.edges[tenantID][src][dst] = struct{}{}
	return nil
}

func (m *MemoryStore) LoadEdges(ctx context.Context, tenantID string) ([]Edge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tenantEdges := m.edges[tenantID]
	out := []Edge{}
	for src, targets := range tenantEdges {
		for dst := range targets {
			out = append(out, Edge{Src: src, Dst: dst})
		}
	}
	return out, nil
}

func (m *MemoryStore) ClearEdges(ctx context.Context, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.edges, tenantID)
	return nil
}
