package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/bradtumy/authorization-service/pkg/tenant"
)

func runStoreTests(t *testing.T, s Store) {
	ctx := context.Background()
	tnt := tenant.Tenant{ID: "t1", Name: "t1", CreatedAt: time.Now().UTC()}
	if err := s.SaveTenant(ctx, tnt); err != nil {
		t.Fatalf("SaveTenant: %v", err)
	}
	got, err := s.LoadTenant(ctx, "t1")
	if err != nil || got.ID != "t1" {
		t.Fatalf("LoadTenant: %v", err)
	}
	list, err := s.ListTenants(ctx)
	if err != nil || len(list) == 0 {
		t.Fatalf("ListTenants: %v", err)
	}
	pol := policy.Policy{ID: "p1", Description: "d"}
	if err := s.SavePolicy(ctx, "t1", pol); err != nil {
		t.Fatalf("SavePolicy: %v", err)
	}
	pList, err := s.LoadPolicies(ctx, "t1")
	if err != nil || len(pList) != 1 {
		t.Fatalf("LoadPolicies: %v", err)
	}
	if err := s.SaveEdge(ctx, "t1", "a", "b"); err != nil {
		t.Fatalf("SaveEdge: %v", err)
	}
	edges, err := s.LoadEdges(ctx, "t1")
	if err != nil || len(edges) != 1 {
		t.Fatalf("LoadEdges: %v", err)
	}
	if err := s.DeleteTenant(ctx, "t1"); err != nil {
		t.Fatalf("DeleteTenant: %v", err)
	}
	if _, err := s.LoadTenant(ctx, "t1"); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestMemoryStore(t *testing.T) {
	runStoreTests(t, NewMemory())
}

func TestSQLiteStore(t *testing.T) {
	os.Remove("test.db")
	s, err := NewSQLite("test.db")
	if err != nil {
		t.Fatalf("new sqlite: %v", err)
	}
	// run migration
	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS tenants(id TEXT PRIMARY KEY, name TEXT, created_at INTEGER);`)
	if err != nil {
		t.Fatalf("migrate tenants: %v", err)
	}
	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS policies(tenant_id TEXT, policy_id TEXT, policy TEXT, PRIMARY KEY(tenant_id, policy_id));`)
	if err != nil {
		t.Fatalf("migrate policies: %v", err)
	}
	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS edges(tenant_id TEXT, src TEXT, dst TEXT, PRIMARY KEY(tenant_id, src, dst));`)
	if err != nil {
		t.Fatalf("migrate edges: %v", err)
	}
	runStoreTests(t, s)
}
