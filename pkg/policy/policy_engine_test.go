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
	if engine.Evaluate("user1", "file1", "read", nil) {
		t.Fatalf("expected evaluation to fail due to subject mismatch")
	}
}
