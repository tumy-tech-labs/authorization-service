package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bradtumy/authorization-service/internal/middleware"
	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var (
	policyEngine *policy.PolicyEngine
	policyStore  *policy.PolicyStore
	policyFile   = "configs/policies.yaml"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		panic("Failed to load .env file")
	}

	policyStore = policy.NewPolicyStore()
	err = policyStore.LoadPolicies(policyFile)
	if err != nil {
		panic("Failed to load policies: " + err.Error())
	}
	policyEngine = policy.NewPolicyEngine(policyStore)
}

type AccessRequest struct {
	Subject    string            `json:"subject"`
	Resource   string            `json:"resource"`
	Action     string            `json:"action"`
	Conditions map[string]string `json:"conditions"`
}

func SetupRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.JWTMiddleware)
	router.HandleFunc("/check-access", CheckAccess).Methods("POST")
	router.HandleFunc("/reload", ReloadPolicies).Methods("POST")
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

	// Evaluate permissions using the PolicyEngine
	decision := policyEngine.Evaluate(req.Subject, req.Resource, req.Action, req.Conditions)

	// Respond with the authorization decision
	json.NewEncoder(w).Encode(decision)
}

// ReloadPolicies reloads policies from the YAML file.
func ReloadPolicies(w http.ResponseWriter, r *http.Request) {
	if err := policyStore.LoadPolicies(policyFile); err != nil {
		log.Printf("policy reload failed: %v", err)
		http.Error(w, "failed to reload policies", http.StatusInternalServerError)
		return
	}
	log.Print("policy reload successful")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("policies reloaded"))
}
