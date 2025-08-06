package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bradtumy/authorization-service/internal/logger"
	"github.com/bradtumy/authorization-service/internal/middleware"
	"github.com/bradtumy/authorization-service/pkg/contextprovider"
	"github.com/bradtumy/authorization-service/pkg/graph"
	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/bradtumy/authorization-service/pkg/policycompiler"
	"github.com/bradtumy/authorization-service/pkg/store"
	"github.com/bradtumy/authorization-service/pkg/tenant"
	"github.com/bradtumy/authorization-service/pkg/user"
	"github.com/bradtumy/authorization-service/pkg/validator"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

var (
	policyStores  map[string]*policy.PolicyStore
	policyEngines map[string]*policy.PolicyEngine
	policyGraphs  map[string]*graph.Graph
	policyFiles   map[string]string
	backend       store.Store
	policyBackend string
	compiler      policycompiler.Compiler
	auditLogger   *logger.Logger
	policyEval    = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "policy_eval_count",
			Help: "Number of policy evaluations",
		},
		[]string{"decision", "reason"},
	)
	tracer           trace.Tracer
	contextProviders contextprovider.Chain
)

func init() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: could not load .env file: %v", err)
	}

	middleware.LoadOIDCConfig()

	policyStores = make(map[string]*policy.PolicyStore)
	policyEngines = make(map[string]*policy.PolicyEngine)
	policyGraphs = make(map[string]*graph.Graph)
	policyFiles = make(map[string]string)

	var err error
	backend, err = store.New()
	if err != nil {
		panic("failed to init store: " + err.Error())
	}

	policyBackend = os.Getenv("POLICY_BACKEND")
	if policyBackend == "" {
		policyBackend = "file"
	}

	defaultTenant := "default"
	defaultFile := os.Getenv("POLICY_FILE")
	if defaultFile == "" {
		defaultFile = "configs/policies.yaml"
	}
	if _, err := os.Stat(defaultFile); os.IsNotExist(err) {
		for _, alt := range []string{"../" + defaultFile, "../../" + defaultFile} {
			if _, err2 := os.Stat(alt); err2 == nil {
				defaultFile = alt
				break
			}
		}
	}

	store := policy.NewPolicyStore()
	g := graph.New()
	policyStores[defaultTenant] = store
	policyGraphs[defaultTenant] = g
	policyEngines[defaultTenant] = policy.NewPolicyEngine(store, g)
	policyFiles[defaultTenant] = defaultFile
	def := Tenant{ID: defaultTenant, Name: "default", CreatedAt: time.Now()}
	if err := backend.SaveTenant(context.Background(), def); err != nil {
		panic("failed to save default tenant: " + err.Error())
	}

	if policyBackend == "db" {
		if err := loadPoliciesFromDB(context.Background(), defaultTenant); err != nil {
			panic("failed to load policies from db: " + err.Error())
		}
		go watchPolicies()
	} else {
		if err := store.LoadPolicies(defaultFile); err != nil {
			panic("Failed to load policies: " + err.Error())
		}
	}

	compiler = policycompiler.NewOpenAICompiler(os.Getenv("OPENAI_API_KEY"))
	lvl := logger.ParseLevel(os.Getenv("LOG_LEVEL"))
	auditLogger = logger.New(os.Stdout, lvl)
	prometheus.MustRegister(policyEval)
	tracer = otel.Tracer("authorization-service")
	contextProviders = contextprovider.Chain{
		contextprovider.TimeProvider{},
		contextprovider.GeoIPProvider{},
		contextprovider.RiskProvider{},
	}
}

type AccessRequest struct {
	TenantID   string            `json:"tenantID"`
	Subject    string            `json:"subject"`
	Resource   string            `json:"resource"`
	Action     string            `json:"action"`
	Conditions map[string]string `json:"conditions"`
}

// SimulationRequest represents a dry-run evaluation with explicit context.
type SimulationRequest struct {
	TenantID string            `json:"tenantID"`
	Subject  string            `json:"subject"`
	Resource string            `json:"resource"`
	Action   string            `json:"action"`
	Context  map[string]string `json:"context"`
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

type CreateUserRequest struct {
	TenantID string   `json:"tenantID"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

type AssignRoleRequest struct {
	TenantID string   `json:"tenantID"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

type DeleteUserRequest struct {
	TenantID string `json:"tenantID"`
	Username string `json:"username"`
}

func subjectFromRequest(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("missing token")
	}
	tokenString := strings.TrimPrefix(auth, "Bearer ")
	claims := jwt.MapClaims{}
	parser := jwt.Parser{}
	if _, _, err := parser.ParseUnverified(tokenString, claims); err != nil {
		return "", err
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", errors.New("missing subject")
	}
	return sub, nil
}

func requireAdmin(w http.ResponseWriter, r *http.Request, tenantID string) (string, bool) {
	sub, err := subjectFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", false
	}
	if !user.HasRole(tenantID, sub, "TenantAdmin", "PolicyAdmin") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return "", false
	}
	return sub, true
}

func SetupRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.TracingMiddleware)
	router.Use(middleware.CorrelationMiddleware)
	router.Use(middleware.MetricsMiddleware)
	router.Use(middleware.JWTMiddleware)
	router.HandleFunc("/check-access", CheckAccess).Methods("POST")
	router.HandleFunc("/simulate", SimulateAccess).Methods("POST")
	router.HandleFunc("/reload", ReloadPolicies).Methods("POST")
	router.HandleFunc("/compile", CompileRule).Methods("POST")
	router.HandleFunc("/validate-policy", ValidatePolicy).Methods("POST")
	router.HandleFunc("/tenant/create", CreateTenant).Methods("POST")
	router.HandleFunc("/tenant/delete", DeleteTenant).Methods("POST")
	router.HandleFunc("/tenant/list", ListTenants).Methods("GET")
	router.HandleFunc("/user/create", CreateUser).Methods("POST")
	router.HandleFunc("/user/assign-role", AssignRole).Methods("POST")
	router.HandleFunc("/user/delete", DeleteUser).Methods("POST")
	router.HandleFunc("/user/list", ListUsers).Methods("GET")
	router.HandleFunc("/user/get", GetUser).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	return router
}

func CheckAccess(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "CheckAccess")
	defer span.End()
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

	// Gather runtime context and evaluate permissions using the PolicyEngine
	ctxVals := contextProviders.GetContext(r)
	if req.Conditions == nil {
		req.Conditions = make(map[string]string)
	}
	for k, v := range ctxVals {
		req.Conditions[k] = v
	}
	_, evalSpan := tracer.Start(ctx, "PolicyEvaluation")
	for k, v := range ctxVals {
		evalSpan.SetAttributes(attribute.String(k, v))
	}
	decision := engine.Evaluate(req.Subject, req.Resource, req.Action, req.Conditions)
	status := "deny"
	if decision.Allow {
		status = "allow"
	}
	evalSpan.SetAttributes(
		attribute.String("decision", status),
		attribute.String("reason", decision.Reason),
	)
	evalSpan.End()

	// Audit log
	cid := middleware.CorrelationIDFromContext(r.Context())
	reasonLabel := ""
	if !decision.Allow {
		switch decision.Reason {
		case "risk", "time":
			reasonLabel = decision.Reason
		default:
			reasonLabel = "other"
		}
	}
	policyEval.WithLabelValues(status, reasonLabel).Inc()
	auditLogger.Log(logger.Entry{
		Level:         "info",
		CorrelationID: cid,
		TenantID:      req.TenantID,
		Subject:       req.Subject,
		Action:        req.Action,
		Resource:      req.Resource,
		Decision:      status,
		PolicyID:      decision.PolicyID,
		Reason:        decision.Reason,
	})

	// Respond with the authorization decision
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// SimulateAccess performs a dry-run policy evaluation without audit logging.
func SimulateAccess(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "SimulateAccess")
	defer span.End()
	var req SimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	engine, ok := policyEngines[req.TenantID]
	if !ok {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}
	if req.Context == nil {
		req.Context = make(map[string]string)
	}
	decision := engine.Evaluate(req.Subject, req.Resource, req.Action, req.Context)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// ReloadPolicies reloads policies from the YAML file.
func ReloadPolicies(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "ReloadPolicies")
	defer span.End()
	var req TenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if _, ok := policyStores[req.TenantID]; !ok {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}
	if policyBackend == "db" {
		if err := loadPoliciesFromDB(r.Context(), req.TenantID); err != nil {
			http.Error(w, "failed to reload policies", http.StatusInternalServerError)
			return
		}
	} else {
		file, ok := policyFiles[req.TenantID]
		if !ok {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		if err := policyStores[req.TenantID].LoadPolicies(file); err != nil {
			auditLogger.Log(logger.Entry{
				Level:         "error",
				CorrelationID: middleware.CorrelationIDFromContext(r.Context()),
				TenantID:      req.TenantID,
				Action:        "reload",
				Resource:      file,
				Reason:        err.Error(),
			})
			http.Error(w, "failed to reload policies", http.StatusInternalServerError)
			return
		}
		auditLogger.Log(logger.Entry{
			Level:         "info",
			CorrelationID: middleware.CorrelationIDFromContext(r.Context()),
			TenantID:      req.TenantID,
			Action:        "reload",
			Resource:      file,
			Decision:      "success",
		})
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("policies reloaded"))
}

// CompileRule compiles a natural language rule into a YAML policy.
func CompileRule(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "CompileRule")
	defer span.End()
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
		auditLogger.Log(logger.Entry{
			Level:         "error",
			CorrelationID: middleware.CorrelationIDFromContext(r.Context()),
			TenantID:      req.TenantID,
			Action:        "compile",
			Reason:        err.Error(),
		})
		http.Error(w, "failed to compile rule: "+err.Error(), http.StatusInternalServerError)
		return
	}
	auditLogger.Log(logger.Entry{
		Level:         "info",
		CorrelationID: middleware.CorrelationIDFromContext(r.Context()),
		TenantID:      req.TenantID,
		Action:        "compile",
		Decision:      "success",
	})
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Write([]byte(policy))
}

// ValidatePolicy validates a policy definition provided in the request body.
func ValidatePolicy(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "ValidatePolicy")
	defer span.End()
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
		auditLogger.Log(logger.Entry{
			Level:         "warn",
			CorrelationID: middleware.CorrelationIDFromContext(r.Context()),
			TenantID:      req.TenantID,
			Action:        "validate",
			Reason:        err.Error(),
		})
		http.Error(w, "invalid policy: "+err.Error(), http.StatusBadRequest)
		return
	}
	auditLogger.Log(logger.Entry{
		Level:         "info",
		CorrelationID: middleware.CorrelationIDFromContext(r.Context()),
		TenantID:      req.TenantID,
		Action:        "validate",
		Decision:      "success",
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("policy is valid"))
}

func loadPoliciesFromDB(ctx context.Context, tenantID string) error {
	policies, err := backend.LoadPolicies(ctx, tenantID)
	if err != nil {
		return err
	}
	store, ok := policyStores[tenantID]
	if !ok {
		store = policy.NewPolicyStore()
		policyStores[tenantID] = store
	}
	store.ReplacePolicies(policies)
	return nil
}

func watchPolicies() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		tenants, err := backend.ListTenants(context.Background())
		if err != nil {
			continue
		}
		for _, t := range tenants {
			if _, ok := policyStores[t.ID]; ok {
				loadPoliciesFromDB(context.Background(), t.ID)
			}
		}
	}
}

// CreateTenant registers a new tenant with an empty PolicyStore.
func CreateTenant(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "CreateTenant")
	defer span.End()
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
	if policyBackend == "db" {
		loadPoliciesFromDB(r.Context(), req.TenantID)
	}
	auditLogger.Log(logger.Entry{
		Level:         "info",
		CorrelationID: middleware.CorrelationIDFromContext(r.Context()),
		TenantID:      req.TenantID,
		Action:        "tenant_create",
		Decision:      "success",
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenant)
}

// DeleteTenant removes a tenant and associated policy data.
func DeleteTenant(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "DeleteTenant")
	defer span.End()
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
	auditLogger.Log(logger.Entry{
		Level:         "info",
		CorrelationID: middleware.CorrelationIDFromContext(r.Context()),
		TenantID:      req.TenantID,
		Action:        "tenant_delete",
		Decision:      "success",
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenant)
}

// ListTenants returns all registered tenants.
func ListTenants(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ListTenants")
	defer span.End()
	list, err := backend.ListTenants(ctx)
	if err != nil {
		auditLogger.Log(logger.Entry{
			Level:         "error",
			CorrelationID: middleware.CorrelationIDFromContext(r.Context()),
			Action:        "tenant_list",
			Reason:        err.Error(),
		})
		http.Error(w, "failed to list tenants", http.StatusInternalServerError)
		return
	}
	auditLogger.Log(logger.Entry{
		Level:         "info",
		CorrelationID: middleware.CorrelationIDFromContext(r.Context()),
		Action:        "tenant_list",
		Decision:      "success",
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// CreateUser creates a new user under a tenant.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if _, ok := requireAdmin(w, r, req.TenantID); !ok {
		return
	}
	u, err := user.Create(req.TenantID, req.Username, req.Roles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// AssignRole assigns roles to an existing user.
func AssignRole(w http.ResponseWriter, r *http.Request) {
	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if _, ok := requireAdmin(w, r, req.TenantID); !ok {
		return
	}
	if err := user.AssignRoles(req.TenantID, req.Username, req.Roles); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteUser removes a user from the system.
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	var req DeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if _, ok := requireAdmin(w, r, req.TenantID); !ok {
		return
	}
	if err := user.Delete(req.TenantID, req.Username); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ListUsers returns all users for a tenant.
func ListUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantID")
	if tenantID == "" {
		http.Error(w, "missing tenantID", http.StatusBadRequest)
		return
	}
	if _, ok := requireAdmin(w, r, tenantID); !ok {
		return
	}
	list := user.List(tenantID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// GetUser returns information for a specific user.
func GetUser(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantID")
	username := r.URL.Query().Get("username")
	if tenantID == "" || username == "" {
		http.Error(w, "missing tenantID or username", http.StatusBadRequest)
		return
	}
	if _, ok := requireAdmin(w, r, tenantID); !ok {
		return
	}
	u, err := user.Get(tenantID, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}
