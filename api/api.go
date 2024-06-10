package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/bradtumy/authorization-service/internal/middleware"
	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var policyEngine *policy.PolicyEngine

// CustomClaims defines the structure for our custom JWT claims
type CustomClaims struct {
	ClientID string `json:"client_id"`
	jwt.StandardClaims
}

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file: " + err.Error())
	}
}

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
	// Extract the JWT from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	// Parse and validate the JWT
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Decode the access request
	var req struct {
		Subject    string   `json:"subject"`
		Resource   string   `json:"resource"`
		Action     string   `json:"action"`
		Conditions []string `json:"conditions"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	allowed := policyEngine.Evaluate(req.Subject, req.Resource, req.Action, req.Conditions)
	res := map[string]bool{"allowed": allowed}

	json.NewEncoder(w).Encode(res)
}
