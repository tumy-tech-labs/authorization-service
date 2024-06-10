package api

import (
	"encoding/json"
	"net/http"

	"github.com/bradtumy/authorization-service/internal/middleware"
	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/gorilla/mux"
)

var policyEngine *policy.PolicyEngine

func SetupRouter() *mux.Router {
	store := policy.NewPolicyStore()
	err := store.LoadPolicies("configs/policies.yaml")
	if err != nil {
		panic("Failed to load policies: " + err.Error())
	}
	policyEngine = policy.NewPolicyEngine(store)

	router := mux.NewRouter()
	router.Use(middleware.JWTMiddleware)
	router.HandleFunc("/check-access", CheckAccess).Methods("POST")

	return router
}

func CheckAccess(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Subject    string   `json:"subject"`
		Resource   string   `json:"resource"`
		Action     string   `json:"action"`
		Conditions []string `json:"conditions"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	allowed := policyEngine.Evaluate(req.Subject, req.Resource, req.Action, req.Conditions)
	res := map[string]bool{"allowed": allowed}

	json.NewEncoder(w).Encode(res)
}
