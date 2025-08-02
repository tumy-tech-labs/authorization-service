package store

import (
	"context"

	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/bradtumy/authorization-service/pkg/tenant"
)

// Edge represents a relation in the authorization graph.
type Edge struct {
	Src string
	Dst string
}

// Store defines operations for persisting tenants, policies and graph edges.
type Store interface {
	SaveTenant(ctx context.Context, t tenant.Tenant) error
	LoadTenant(ctx context.Context, id string) (tenant.Tenant, error)
	ListTenants(ctx context.Context) ([]tenant.Tenant, error)
	DeleteTenant(ctx context.Context, id string) error

	SavePolicy(ctx context.Context, tenantID string, p policy.Policy) error
	LoadPolicies(ctx context.Context, tenantID string) ([]policy.Policy, error)
	ClearPolicies(ctx context.Context, tenantID string) error

	SaveEdge(ctx context.Context, tenantID, src, dst string) error
	LoadEdges(ctx context.Context, tenantID string) ([]Edge, error)
	ClearEdges(ctx context.Context, tenantID string) error
}
