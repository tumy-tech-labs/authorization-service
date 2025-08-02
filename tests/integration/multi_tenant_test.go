package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	api "github.com/bradtumy/authorization-service/api"
	"github.com/bradtumy/authorization-service/pkg/graph"
	"github.com/bradtumy/authorization-service/pkg/policy"
	jwt "github.com/dgrijalva/jwt-go"
)

import _ "unsafe"

//go:linkname policyStores github.com/bradtumy/authorization-service/api.policyStores
var policyStores map[string]*policy.PolicyStore

//go:linkname policyEngines github.com/bradtumy/authorization-service/api.policyEngines
var policyEngines map[string]*policy.PolicyEngine

//go:linkname policyGraphs github.com/bradtumy/authorization-service/api.policyGraphs
var policyGraphs map[string]*graph.Graph

//go:linkname policyFiles github.com/bradtumy/authorization-service/api.policyFiles
var policyFiles map[string]string

func token(t *testing.T) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "tester"})
	str, err := tok.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return str
}

func TestMultiTenantIsolation(t *testing.T) {
	router := api.SetupRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	tok := token(t)

	createTenant := func(id string) {
		body := fmt.Sprintf(`{"tenantID":"%s","name":"%s"}`, id, id)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/tenant/create", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("create tenant %s: %v", id, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("create tenant %s status %d", id, resp.StatusCode)
		}
	}
	createTenant("acme")
	createTenant("globex")
	defer func() {
		for _, id := range []string{"acme", "globex"} {
			body := fmt.Sprintf(`{"tenantID":"%s"}`, id)
			req, _ := http.NewRequest(http.MethodPost, srv.URL+"/tenant/delete", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tok)
			http.DefaultClient.Do(req)
		}
	}()

	fileA, err := os.CreateTemp("", "acme*.yaml")
	if err != nil {
		t.Fatalf("tempfile A: %v", err)
	}
	defer os.Remove(fileA.Name())
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
		t.Fatalf("write A: %v", err)
	}

	fileB, err := os.CreateTemp("", "globex*.yaml")
	if err != nil {
		t.Fatalf("tempfile B: %v", err)
	}
	defer os.Remove(fileB.Name())
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
		t.Fatalf("write B: %v", err)
	}

	storeA := policy.NewPolicyStore()
	if err := storeA.LoadPolicies(fileA.Name()); err != nil {
		t.Fatalf("load A: %v", err)
	}
	gA := graph.New()
	policyStores["acme"] = storeA
	policyGraphs["acme"] = gA
	policyEngines["acme"] = policy.NewPolicyEngine(storeA, gA)
	policyFiles["acme"] = fileA.Name()

	storeB := policy.NewPolicyStore()
	if err := storeB.LoadPolicies(fileB.Name()); err != nil {
		t.Fatalf("load B: %v", err)
	}
	gB := graph.New()
	policyStores["globex"] = storeB
	policyGraphs["globex"] = gB
	policyEngines["globex"] = policy.NewPolicyEngine(storeB, gB)
	policyFiles["globex"] = fileB.Name()

	check := func(tenantID string) policy.Decision {
		body := fmt.Sprintf(`{"tenantID":"%s","subject":"alice","resource":"file1","action":"read","conditions":{}}`, tenantID)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/check-access", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("check %s: %v", tenantID, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("check %s status %d", tenantID, resp.StatusCode)
		}
		var dec policy.Decision
		if err := json.NewDecoder(resp.Body).Decode(&dec); err != nil {
			t.Fatalf("decode %s: %v", tenantID, err)
		}
		return dec
	}

	if !check("acme").Allow {
		t.Fatalf("acme expected allow")
	}
	if check("globex").Allow {
		t.Fatalf("globex expected deny")
	}

	policyGraphs["acme"].AddRelation("user:alice", "group:team")
	if !policyGraphs["acme"].HasPath("user:alice", "group:team") {
		t.Fatalf("acme graph missing relation")
	}
	if policyGraphs["globex"].HasPath("user:alice", "group:team") {
		t.Fatalf("graph relation leaked to globex")
	}

	delete(policyStores, "acme")
	delete(policyStores, "globex")
	delete(policyGraphs, "acme")
	delete(policyGraphs, "globex")
	delete(policyEngines, "acme")
	delete(policyEngines, "globex")
	delete(policyFiles, "acme")
	delete(policyFiles, "globex")
}
