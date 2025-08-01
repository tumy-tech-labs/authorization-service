package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bradtumy/authorization-service/internal/middleware"
	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var policyEngine *policy.PolicyEngine

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		panic("Failed to load .env file")
	}

	policyStore := policy.NewPolicyStore()
	err = policyStore.LoadPolicies("configs/policies.yaml")
	if err != nil {
		panic("Failed to load policies: " + err.Error())
	}
	policyEngine = policy.NewPolicyEngine(policyStore)
}

type AccessRequest struct {
	Subject    string   `json:"subject"`
	Resource   string   `json:"resource"`
	Action     string   `json:"action"`
	Conditions []string `json:"conditions"`
}

func SetupRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.JWTMiddleware)
	router.HandleFunc("/check-access", CheckAccess).Methods("POST")
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
