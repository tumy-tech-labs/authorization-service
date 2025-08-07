package user

import "testing"

func TestCRUD(t *testing.T) {
	Reset()
	EnablePersistence(false)
	if _, err := Create("acme", "alice", []string{"TenantAdmin"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if !HasRole("acme", "alice", "TenantAdmin") {
		t.Fatalf("alice missing role")
	}
	if err := AssignRoles("acme", "alice", []string{"PolicyAdmin"}); err != nil {
		t.Fatalf("assign: %v", err)
	}
	u, err := Get("acme", "alice")
	if err != nil || len(u.Roles) != 1 || u.Roles[0] != "PolicyAdmin" {
		t.Fatalf("get after assign")
	}
	list := List("acme")
	if len(list) != 1 || list[0].Username != "alice" {
		t.Fatalf("list incorrect")
	}
	if err := Delete("acme", "alice"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := Get("acme", "alice"); err == nil {
		t.Fatalf("expected not found")
	}
}
