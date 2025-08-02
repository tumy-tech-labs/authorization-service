package policy

import (
	"os"
	"testing"
)

// Test that policies can be reloaded from file without restarting service.
func TestPolicyReload(t *testing.T) {
	tmp, err := os.CreateTemp("", "policies*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())

	initial := `roles:
  - name: "admin"
    policies: ["policy1"]
users:
  - username: "alice"
    roles: ["admin"]
policies:
  - id: "policy1"
    description: "deny read"
    subjects:
      - role: "admin"
    resource:
      - "file1"
    action:
      - "read"
    effect: "deny"
`
	if _, err := tmp.Write([]byte(initial)); err != nil {
		t.Fatalf("failed to write initial policy: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	store := NewPolicyStore()
	if err := store.LoadPolicies(tmp.Name()); err != nil {
		t.Fatalf("load policies: %v", err)
	}
	engine := NewPolicyEngine(store)

	dec := engine.Evaluate("alice", "file1", "read", nil)
	if dec.Allow {
		t.Fatalf("expected deny decision, got allow")
	}

	updated := `roles:
  - name: "admin"
    policies: ["policy1"]
users:
  - username: "alice"
    roles: ["admin"]
policies:
  - id: "policy1"
    description: "allow read"
    subjects:
      - role: "admin"
    resource:
      - "file1"
    action:
      - "read"
    effect: "allow"
`
	if err := os.WriteFile(tmp.Name(), []byte(updated), 0644); err != nil {
		t.Fatalf("failed to update policy file: %v", err)
	}

	if err := store.LoadPolicies(tmp.Name()); err != nil {
		t.Fatalf("reload policies: %v", err)
	}

	dec = engine.Evaluate("alice", "file1", "read", nil)
	if !dec.Allow {
		t.Fatalf("expected allow decision after reload")
	}
}
