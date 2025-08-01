package policy

import "testing"

func TestEvaluateSubjectMismatch(t *testing.T) {
	store := NewPolicyStore()
	// setup role with policy
	store.Roles["admin"] = Role{Name: "admin", Policies: []string{"policy1"}}
	// user with role admin
	store.Users["user1"] = User{Username: "user1", Roles: []string{"admin"}}
	// policy with subject that does not match role
	store.Policies["policy1"] = Policy{
		ID:       "policy1",
		Subjects: []Subject{{Role: "editor"}},
		Resource: []string{"file1"},
		Action:   []string{"read"},
		Effect:   "allow",
	}

	engine := NewPolicyEngine(store)
	decision := engine.Evaluate("user1", "file1", "read", nil)
	if decision.Allow {
		t.Fatalf("expected evaluation to fail due to subject mismatch")
	}
}

func TestEvaluateAllow(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["admin"] = Role{Name: "admin", Policies: []string{"policy1"}}
	store.Users["user1"] = User{Username: "user1", Roles: []string{"admin"}}
	store.Policies["policy1"] = Policy{
		ID:       "policy1",
		Subjects: []Subject{{Role: "admin"}},
		Resource: []string{"file1"},
		Action:   []string{"read"},
		Effect:   "allow",
	}

	engine := NewPolicyEngine(store)
	decision := engine.Evaluate("user1", "file1", "read", nil)
	if !decision.Allow {
		t.Fatalf("expected evaluation to allow access, got %v", decision)
	}
	if decision.PolicyID != "policy1" {
		t.Fatalf("expected policy1, got %s", decision.PolicyID)
	}
}

func TestEvaluateDeny(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["admin"] = Role{Name: "admin", Policies: []string{"policy1"}}
	store.Users["user1"] = User{Username: "user1", Roles: []string{"admin"}}
	store.Policies["policy1"] = Policy{
		ID:       "policy1",
		Subjects: []Subject{{Role: "admin"}},
		Resource: []string{"file1"},
		Action:   []string{"read"},
		Effect:   "deny",
	}

	engine := NewPolicyEngine(store)
	decision := engine.Evaluate("user1", "file1", "read", nil)
	if decision.Allow {
		t.Fatalf("expected evaluation to deny access, got %v", decision)
	}
	if decision.Reason != "denied by policy" {
		t.Fatalf("expected deny reason, got %s", decision.Reason)
	}
}

func TestEvaluateWildcard(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["admin"] = Role{Name: "admin", Policies: []string{"policy1"}}
	store.Users["user1"] = User{Username: "user1", Roles: []string{"admin"}}
	store.Policies["policy1"] = Policy{
		ID:       "policy1",
		Subjects: []Subject{{Role: "admin"}},
		Resource: []string{"*"},
		Action:   []string{"*"},
		Effect:   "allow",
	}

	engine := NewPolicyEngine(store)
	decision := engine.Evaluate("user1", "anyfile", "write", nil)
	if !decision.Allow {
		t.Fatalf("expected wildcard policy to allow access, got %v", decision)
	}
}
