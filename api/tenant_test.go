package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bradtumy/authorization-service/pkg/graph"
	"github.com/bradtumy/authorization-service/pkg/policy"
)

func TestCheckAccessSingleTenant(t *testing.T) {
	reqBody := `{"tenantID":"default","subject":"user1","resource":"file1","action":"read","conditions":{}}`
	r := httptest.NewRequest(http.MethodPost, "/check-access", strings.NewReader(reqBody))
	w := httptest.NewRecorder()
	CheckAccess(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var dec policy.Decision
	if err := json.NewDecoder(w.Body).Decode(&dec); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !dec.Allow {
		t.Fatalf("expected allow, got %v", dec)
	}
}

func TestCheckAccessMultiTenantIsolation(t *testing.T) {
	fileA, err := os.CreateTemp("", "tenantA*.yaml")
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	defer os.Remove(fileA.Name())
	fileB, err := os.CreateTemp("", "tenantB*.yaml")
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	defer os.Remove(fileB.Name())

	policyA := `roles:
- name: "admin"
  policies: ["p1"]
users:
- username: "alice"
  roles: ["admin"]
policies:
- id: "p1"
  subjects:
    - role: "admin"
  resource:
    - "file1"
  action:
    - "read"
  effect: "allow"
`
	if err := os.WriteFile(fileA.Name(), []byte(policyA), 0644); err != nil {
		t.Fatalf("writeA: %v", err)
	}
	policyB := `roles:
- name: "admin"
  policies: ["p1"]
users:
- username: "alice"
  roles: ["admin"]
policies:
- id: "p1"
  subjects:
    - role: "admin"
  resource:
    - "file1"
  action:
    - "read"
  effect: "deny"
`
	if err := os.WriteFile(fileB.Name(), []byte(policyB), 0644); err != nil {
		t.Fatalf("writeB: %v", err)
	}

	storeA := policy.NewPolicyStore()
	if err := storeA.LoadPolicies(fileA.Name()); err != nil {
		t.Fatalf("loadA: %v", err)
	}
	storeB := policy.NewPolicyStore()
	if err := storeB.LoadPolicies(fileB.Name()); err != nil {
		t.Fatalf("loadB: %v", err)
	}
	gA := graph.New()
	gB := graph.New()
	policyStores["tenantA"] = storeA
	policyGraphs["tenantA"] = gA
	policyEngines["tenantA"] = policy.NewPolicyEngine(storeA, gA)
	policyFiles["tenantA"] = fileA.Name()
	policyStores["tenantB"] = storeB
	policyGraphs["tenantB"] = gB
	policyEngines["tenantB"] = policy.NewPolicyEngine(storeB, gB)
	policyFiles["tenantB"] = fileB.Name()

	reqA := `{"tenantID":"tenantA","subject":"alice","resource":"file1","action":"read","conditions":{}}`
	wA := httptest.NewRecorder()
	rA := httptest.NewRequest(http.MethodPost, "/check-access", strings.NewReader(reqA))
	CheckAccess(wA, rA)
	var decA policy.Decision
	json.NewDecoder(wA.Body).Decode(&decA)
	if !decA.Allow {
		t.Fatalf("tenantA expected allow")
	}

	reqB := `{"tenantID":"tenantB","subject":"alice","resource":"file1","action":"read","conditions":{}}`
	wB := httptest.NewRecorder()
	rB := httptest.NewRequest(http.MethodPost, "/check-access", strings.NewReader(reqB))
	CheckAccess(wB, rB)
	var decB policy.Decision
	json.NewDecoder(wB.Body).Decode(&decB)
	if decB.Allow {
		t.Fatalf("tenantB expected deny")
	}

	reqC := `{"tenantID":"missing","subject":"alice","resource":"file1","action":"read","conditions":{}}`
	wC := httptest.NewRecorder()
	rC := httptest.NewRequest(http.MethodPost, "/check-access", strings.NewReader(reqC))
	CheckAccess(wC, rC)
	if wC.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown tenant, got %d", wC.Code)
	}
}
