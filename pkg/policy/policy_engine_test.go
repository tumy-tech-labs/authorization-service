package policy

import (
	"testing"

	"github.com/bradtumy/authorization-service/pkg/graph"
)

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

	engine := NewPolicyEngine(store, graph.New())
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

	engine := NewPolicyEngine(store, graph.New())
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

	engine := NewPolicyEngine(store, graph.New())
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

	engine := NewPolicyEngine(store, graph.New())
	decision := engine.Evaluate("user1", "anyfile", "write", nil)
	if !decision.Allow {
		t.Fatalf("expected wildcard policy to allow access, got %v", decision)
	}
}

func TestEvaluateConditionSatisfied(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["admin"] = Role{Name: "admin", Policies: []string{"policy1"}}
	store.Users["user1"] = User{Username: "user1", Roles: []string{"admin"}}
	store.Policies["policy1"] = Policy{
		ID:         "policy1",
		Subjects:   []Subject{{Role: "admin"}},
		Resource:   []string{"file1"},
		Action:     []string{"read"},
		Effect:     "allow",
		Conditions: map[string]string{"time": "business-hours"},
	}

	engine := NewPolicyEngine(store, graph.New())
	decision := engine.Evaluate("user1", "file1", "read", map[string]string{"time": "10:00"})
	if !decision.Allow {
		t.Fatalf("expected access to be allowed during business hours")
	}
}

func TestEvaluateConditionUnsatisfied(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["admin"] = Role{Name: "admin", Policies: []string{"policy1"}}
	store.Users["user1"] = User{Username: "user1", Roles: []string{"admin"}}
	store.Policies["policy1"] = Policy{
		ID:         "policy1",
		Subjects:   []Subject{{Role: "admin"}},
		Resource:   []string{"file1"},
		Action:     []string{"read"},
		Effect:     "allow",
		Conditions: map[string]string{"time": "business-hours"},
	}

	engine := NewPolicyEngine(store, graph.New())
	decision := engine.Evaluate("user1", "file1", "read", map[string]string{"time": "20:00"})
	if decision.Allow {
		t.Fatalf("expected access to be denied outside business hours")
	}
	if decision.Reason != "conditions not satisfied" {
		t.Fatalf("unexpected reason: %s", decision.Reason)
	}
}

func TestEvaluateWhenSatisfied(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["partner"] = Role{Name: "partner", Policies: []string{"policy1"}}
	store.Users["bob"] = User{Username: "bob", Roles: []string{"partner"}}
	store.Policies["policy1"] = Policy{
		ID:       "policy1",
		Subjects: []Subject{{Role: "partner"}},
		Resource: []string{"dashboard"},
		Action:   []string{"view"},
		Effect:   "allow",
		When:     []string{"context.time == \"business-hours\"", "context.risk < \"medium\""},
	}
	engine := NewPolicyEngine(store, graph.New())
	env := map[string]string{"time": "business-hours", "risk": "low"}
	decision := engine.Evaluate("bob", "dashboard", "view", env)
	if !decision.Allow {
		t.Fatalf("expected access to be allowed when when conditions satisfied")
	}
}

func TestEvaluateWhenUnsatisfied(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["partner"] = Role{Name: "partner", Policies: []string{"policy1"}}
	store.Users["bob"] = User{Username: "bob", Roles: []string{"partner"}}
	store.Policies["policy1"] = Policy{
		ID:       "policy1",
		Subjects: []Subject{{Role: "partner"}},
		Resource: []string{"dashboard"},
		Action:   []string{"view"},
		Effect:   "allow",
		When:     []string{"context.time == \"business-hours\"", "context.risk < \"medium\""},
	}
	engine := NewPolicyEngine(store, graph.New())
	env := map[string]string{"time": "business-hours", "risk": "high"}
	decision := engine.Evaluate("bob", "dashboard", "view", env)
	if decision.Allow {
		t.Fatalf("expected access to be denied when when conditions not satisfied")
	}
	if decision.Reason != "conditions not satisfied" {
		t.Fatalf("unexpected reason: %s", decision.Reason)
	}
}

func TestEvaluateContextIncluded(t *testing.T) {
	store := NewPolicyStore()
	store.Users["user1"] = User{Username: "user1"}
	engine := NewPolicyEngine(store, graph.New())
	env := map[string]string{"ip": "1.2.3.4"}
	dec := engine.Evaluate("user1", "file1", "read", env)
	if dec.Context["ip"] != "1.2.3.4" {
		t.Fatalf("expected context to include env values")
	}
}

func TestEvaluateGroupMembership(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["managers"] = Role{Name: "managers", Policies: []string{"p1"}}
	store.Users["alice"] = User{Username: "alice"}
	store.Policies["p1"] = Policy{
		ID:       "p1",
		Subjects: []Subject{{Role: "managers"}},
		Resource: []string{"file1"},
		Action:   []string{"read"},
		Effect:   "allow",
	}
	g := graph.New()
	g.AddRelation("user:alice", "group:managers")

	engine := NewPolicyEngine(store, g)
	dec := engine.Evaluate("alice", "file1", "read", nil)
	if !dec.Allow {
		t.Fatalf("expected group membership to allow access")
	}
}

func TestEvaluateResourceGroup(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["admin"] = Role{Name: "admin", Policies: []string{"p1"}}
	store.Users["alice"] = User{Username: "alice", Roles: []string{"admin"}}
	store.Policies["p1"] = Policy{
		ID:       "p1",
		Subjects: []Subject{{Role: "admin"}},
		Resource: []string{"teamA"},
		Action:   []string{"read"},
		Effect:   "allow",
	}
	g := graph.New()
	g.AddRelation("group:teamA", "resource:file1")

	engine := NewPolicyEngine(store, g)
	dec := engine.Evaluate("alice", "file1", "read", nil)
	if !dec.Allow {
		t.Fatalf("expected resource group expansion to allow access")
	}
}

func TestEvaluateDelegationChain(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["admin"] = Role{Name: "admin", Policies: []string{"p1"}}
	store.Users["mary"] = User{Username: "mary", Roles: []string{"admin"}}
	store.Users["bob"] = User{Username: "bob"}
	store.Users["alice"] = User{Username: "alice"}
	store.Policies["p1"] = Policy{
		ID:       "p1",
		Subjects: []Subject{{Role: "admin"}},
		Resource: []string{"file1"},
		Action:   []string{"read"},
		Effect:   "allow",
	}
	g := graph.New()
	g.AddRelation("user:alice", "user:bob")
	g.AddRelation("user:bob", "user:mary")

	engine := NewPolicyEngine(store, g)
	dec := engine.Evaluate("alice", "file1", "read", nil)
	if !dec.Allow || dec.Delegator != "mary" {
		t.Fatalf("expected delegation to allow via mary, got %#v", dec)
	}
}

func TestEvaluateDelegationChainInvalid(t *testing.T) {
	store := NewPolicyStore()
	store.Roles["admin"] = Role{Name: "admin", Policies: []string{"p1"}}
	store.Users["mary"] = User{Username: "mary"} // no roles
	store.Users["alice"] = User{Username: "alice"}
	store.Policies["p1"] = Policy{
		ID:       "p1",
		Subjects: []Subject{{Role: "admin"}},
		Resource: []string{"file1"},
		Action:   []string{"read"},
		Effect:   "allow",
	}
	g := graph.New()
	g.AddRelation("user:alice", "user:mary")

	engine := NewPolicyEngine(store, g)
	dec := engine.Evaluate("alice", "file1", "read", nil)
	if dec.Allow {
		t.Fatalf("expected delegation chain to deny access, got %#v", dec)
	}
	if dec.Delegator != "" {
		t.Fatalf("unexpected delegator %q for failed delegation", dec.Delegator)
	}
}
