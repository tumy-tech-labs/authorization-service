package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/bradtumy/authorization-service/internal/middleware"
	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/bradtumy/authorization-service/pkg/policycompiler"
	"github.com/bradtumy/authorization-service/pkg/validator"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var (
	policyStores  map[string]*policy.PolicyStore
	policyEngines map[string]*policy.PolicyEngine
	policyFiles   map[string]string
	compiler      policycompiler.Compiler
)

func init() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: could not load .env file: %v", err)
	}

	policyStores = make(map[string]*policy.PolicyStore)
	policyEngines = make(map[string]*policy.PolicyEngine)
	policyFiles = make(map[string]string)

	defaultTenant := "default"
	defaultFile := os.Getenv("POLICY_FILE")
	if defaultFile == "" {
		defaultFile = "configs/policies.yaml"
	}

	store := policy.NewPolicyStore()
	if err := store.LoadPolicies(defaultFile); err != nil {
		panic("Failed to load policies: " + err.Error())
	}
	policyStores[defaultTenant] = store
	policyEngines[defaultTenant] = policy.NewPolicyEngine(store)
	policyFiles[defaultTenant] = defaultFile

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
