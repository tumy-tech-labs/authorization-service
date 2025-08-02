package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bradtumy/authorization-service/internal/middleware"
	"github.com/bradtumy/authorization-service/pkg/graph"
	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/bradtumy/authorization-service/pkg/policycompiler"
	"github.com/bradtumy/authorization-service/pkg/store"
	"github.com/bradtumy/authorization-service/pkg/tenant"
	"github.com/bradtumy/authorization-service/pkg/validator"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var (
	policyStores  map[string]*policy.PolicyStore
	policyEngines map[string]*policy.PolicyEngine
	policyGraphs  map[string]*graph.Graph
	policyFiles   map[string]string
	backend       store.Store
	compiler      policycompiler.Compiler
)

func init() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: could not load .env file: %v", err)
	}

	policyStores = make(map[string]*policy.PolicyStore)
	policyEngines = make(map[string]*policy.PolicyEngine)
	policyGraphs = make(map[string]*graph.Graph)
	policyFiles = make(map[string]string)

	var err error
	backend, err = store.New()
	if err != nil {
		panic("failed to init store: " + err.Error())
	}

	defaultTenant := "default"
	defaultFile := os.Getenv("POLICY_FILE")
	if defaultFile == "" {
		defaultFile = "configs/policies.yaml"
	}

	store := policy.NewPolicyStore()
	if err := store.LoadPolicies(defaultFile); err != nil {
		panic("Failed to load policies: " + err.Error())
	}
	g := graph.New()
	policyStores[defaultTenant] = store
	policyGraphs[defaultTenant] = g
	policyEngines[defaultTenant] = policy.NewPolicyEngine(store, g)
	policyFiles[defaultTenant] = defaultFile
	def := Tenant{ID: defaultTenant, Name: "default", CreatedAt: time.Now()}
	if err := backend.SaveTenant(context.Background(), def); err != nil {
		panic("failed to save default tenant: " + err.Error())
	}

	compiler = policycompiler.NewOpenAICompiler(os.Getenv("OPENAI_API_KEY"))
}

type AccessRequest struct {
	TenantID   string            `json:"tenantID"`
	Subject    string            `json:"subject"`
	Resource   string            `json:"resource"`
	Action     string            `json:"action"`
	Conditions map[string]string `json:"conditions"`
}

type CompileRequest struct {
	TenantID string `json:"tenantID"`
	Rule     string `json:"rule"`
}

type TenantRequest struct {
	TenantID string `json:"tenantID"`
}

type CreateTenantRequest struct {
	TenantID string `json:"tenantID"`
	Name     string `json:"name"`
}

type Tenant = tenant.Tenant

type ValidatePolicyRequest struct {
	TenantID string `json:"tenantID"`
	Policy   string `json:"policy"`
}

func SetupRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.JWTMiddleware)
	router.HandleFunc("/check-access", CheckAccess).Methods("POST")
	router.HandleFunc("/reload", ReloadPolicies).Methods("POST")
	router.HandleFunc("/compile", CompileRule).Methods("POST")
	router.HandleFunc("/validate-policy", ValidatePolicy).Methods("POST")
	router.HandleFunc("/tenant/create", CreateTenant).Methods("POST")
	router.HandleFunc("/tenant/delete", DeleteTenant).Methods("POST")
	router.HandleFunc("/tenant/list", ListTenants).Methods("GET")
	return router
}

func CheckAccess(w http.ResponseWriter, r *http.Request) {
	// Extract access request details from the request body
	var req AccessRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	engine, ok := policyEngines[req.TenantID]
	if !ok {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}

	// Evaluate permissions using the PolicyEngine
	decision := engine.Evaluate(req.Subject, req.Resource, req.Action, req.Conditions)

	// Respond with the authorization decision
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// ReloadPolicies reloads policies from the YAML file.
func ReloadPolicies(w http.ResponseWriter, r *http.Request) {
	var req TenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	store, ok := policyStores[req.TenantID]
	if !ok {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}
	file, ok := policyFiles[req.TenantID]
	if !ok {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}
	if err := store.LoadPolicies(file); err != nil {
		log.Printf("policy reload failed: %v", err)
		http.Error(w, "failed to reload policies", http.StatusInternalServerError)
		return
	}
	log.Print("policy reload successful")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("policies reloaded"))
}

// CompileRule compiles a natural language rule into a YAML policy.
func CompileRule(w http.ResponseWriter, r *http.Request) {
	var req CompileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if _, ok := policyStores[req.TenantID]; !ok {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}
	policy, err := compiler.Compile(req.Rule)
	if err != nil {
		http.Error(w, "failed to compile rule: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Write([]byte(policy))
}

// ValidatePolicy validates a policy definition provided in the request body.
func ValidatePolicy(w http.ResponseWriter, r *http.Request) {
	var req ValidatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if _, ok := policyStores[req.TenantID]; !ok {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}
	if err := validator.ValidatePolicyData([]byte(req.Policy)); err != nil {
		http.Error(w, "invalid policy: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("policy is valid"))
}

// CreateTenant registers a new tenant with an empty PolicyStore.
func CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.TenantID == "" {
		http.Error(w, "tenantID is required", http.StatusBadRequest)
		return
	}
	if _, err := backend.LoadTenant(r.Context(), req.TenantID); err == nil {
		http.Error(w, "tenant already exists", http.StatusConflict)
		return
	}
	store := policy.NewPolicyStore()
	g := graph.New()
	policyStores[req.TenantID] = store
	policyGraphs[req.TenantID] = g
	policyEngines[req.TenantID] = policy.NewPolicyEngine(store, g)
	policyFiles[req.TenantID] = ""
	tenant := Tenant{ID: req.TenantID, Name: req.Name, CreatedAt: time.Now()}
	if err := backend.SaveTenant(r.Context(), tenant); err != nil {
		http.Error(w, "failed to save tenant", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenant)
}

// DeleteTenant removes a tenant and associated policy data.
func DeleteTenant(w http.ResponseWriter, r *http.Request) {
	var req TenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	tenant, err := backend.LoadTenant(r.Context(), req.TenantID)
	if err != nil {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}
	delete(policyStores, req.TenantID)
	delete(policyGraphs, req.TenantID)
	delete(policyEngines, req.TenantID)
	delete(policyFiles, req.TenantID)
	backend.DeleteTenant(r.Context(), req.TenantID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenant)
}

// ListTenants returns all registered tenants.
func ListTenants(w http.ResponseWriter, r *http.Request) {
	list, err := backend.ListTenants(r.Context())
	if err != nil {
		http.Error(w, "failed to list tenants", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
