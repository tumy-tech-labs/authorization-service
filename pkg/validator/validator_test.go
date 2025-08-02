package validator

import "testing"

func TestValidatePolicyValid(t *testing.T) {
	yaml := []byte(`
roles:
  - name: "admin"
    policies: ["policy1"]
policies:
  - id: "policy1"
    subjects:
      - role: "admin"
    resource: ["*"]
    action: ["read"]
    effect: "allow"
`)
	if err := ValidatePolicyData(yaml); err != nil {
		t.Fatalf("expected valid policy, got error: %v", err)
	}
}

func TestValidatePolicyInvalidRole(t *testing.T) {
	yaml := []byte(`
roles:
  - name: "admin"
    policies: ["policy1"]
policies:
  - id: "policy1"
    subjects:
      - role: "unknown"
    resource: ["*"]
    action: ["read"]
    effect: "allow"
`)
	if err := ValidatePolicyData(yaml); err == nil {
		t.Fatalf("expected error for undefined role")
	}
}

func TestValidatePolicyEmptyAction(t *testing.T) {
	yaml := []byte(`
roles:
  - name: "admin"
    policies: ["policy1"]
policies:
  - id: "policy1"
    subjects:
      - role: "admin"
    resource: ["*"]
    action: []
    effect: "allow"
`)
	if err := ValidatePolicyData(yaml); err == nil {
		t.Fatalf("expected error for empty action")
	}
}
