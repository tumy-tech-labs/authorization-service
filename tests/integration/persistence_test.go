package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	api "github.com/bradtumy/authorization-service/api"
	"github.com/bradtumy/authorization-service/internal/logger"
	"github.com/bradtumy/authorization-service/internal/middleware"
	"github.com/bradtumy/authorization-service/pkg/graph"
	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/bradtumy/authorization-service/pkg/policycompiler"
	"github.com/bradtumy/authorization-service/pkg/store"
	"github.com/bradtumy/authorization-service/pkg/tenant"
	"github.com/prometheus/client_golang/prometheus"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

import _ "unsafe"

// link to unexported symbols in api package
//
//go:linkname backend github.com/bradtumy/authorization-service/api.backend
var backend store.Store

//go:linkname policyBackend github.com/bradtumy/authorization-service/api.policyBackend
var policyBackend string

//go:linkname compiler github.com/bradtumy/authorization-service/api.compiler
var compiler policycompiler.Compiler

//go:linkname auditLogger github.com/bradtumy/authorization-service/api.auditLogger
var auditLogger *logger.Logger

//go:linkname tracer github.com/bradtumy/authorization-service/api.tracer
var tracer trace.Tracer

//go:linkname loadPoliciesFromDB github.com/bradtumy/authorization-service/api.loadPoliciesFromDB
func loadPoliciesFromDB(ctx context.Context, tenantID string) error

// startServer initializes global state similar to api.init and returns a new test server.
func startServer(t *testing.T) *httptest.Server {
	t.Helper()
	middleware.LoadOIDCConfig()

	policyStores = make(map[string]*policy.PolicyStore)
	policyEngines = make(map[string]*policy.PolicyEngine)
	policyGraphs = make(map[string]*graph.Graph)
	policyFiles = make(map[string]string)

	if policyEval != nil {
		prometheus.Unregister(policyEval)
	}
	policyEval = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "policy_eval_count",
		Help: "Number of policy evaluations",
	}, []string{"decision"})
	prometheus.MustRegister(policyEval)

	var err error
	backend, err = store.New()
	if err != nil {
		t.Fatalf("store.New: %v", err)
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

	st := policy.NewPolicyStore()
	if err := st.LoadPolicies(defaultFile); err != nil {
		t.Fatalf("load policies: %v", err)
	}
	g := graph.New()
	policyStores[defaultTenant] = st
	policyGraphs[defaultTenant] = g
	policyEngines[defaultTenant] = policy.NewPolicyEngine(st, g)
	policyFiles[defaultTenant] = defaultFile
	def := tenant.Tenant{ID: defaultTenant, Name: "default", CreatedAt: time.Now()}
	if err := backend.SaveTenant(context.Background(), def); err != nil {
		t.Fatalf("SaveTenant: %v", err)
	}

	if policyBackend == "db" {
		if err := loadPoliciesFromDB(context.Background(), defaultTenant); err != nil {
			t.Fatalf("loadPoliciesFromDB: %v", err)
		}
	}

	compiler = policycompiler.NewOpenAICompiler("")
	auditLogger = logger.New(io.Discard, logger.ParseLevel("info"))
	tracer = otel.Tracer("authorization-service")

	router := api.SetupRouter()
	srv := httptest.NewServer(router)
	return srv
}

// startPostgres spins up a postgres container and applies migrations.
func startPostgres(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_DB":       "authz",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	if err != nil {
		t.Skipf("postgres container: %v", err)
	}
	host, err := pgC.Host(ctx)
	if err != nil {
		t.Fatalf("host: %v", err)
	}
	port, err := pgC.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("port: %v", err)
	}
	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/authz?sslmode=disable", host, port.Port())

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	mig, err := os.ReadFile("migrations/001_init.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := db.Exec(string(mig)); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	terminate := func() {
		_ = pgC.Terminate(ctx)
	}
	return dsn, terminate
}

func TestPostgresPersistence(t *testing.T) {
	dsn, terminate := startPostgres(t)
	defer terminate()

	os.Setenv("STORE_BACKEND", "postgres")
	os.Setenv("STORE_PG_DSN", dsn)
	os.Setenv("POLICY_BACKEND", "db")
	os.Setenv("OIDC_CONFIG_FILE", "/dev/null")

	srv := startServer(t)

	pol := policy.Policy{ID: "persist", Subjects: []policy.Subject{{Role: "admin"}}, Resource: []string{"fileX"}, Action: []string{"read"}, Effect: "allow"}
	if err := backend.SavePolicy(context.Background(), "default", pol); err != nil {
		t.Fatalf("save policy: %v", err)
	}

	srv.Close()
	srv = startServer(t)
	defer srv.Close()

	policies, err := backend.LoadPolicies(context.Background(), "default")
	if err != nil {
		t.Fatalf("load policies: %v", err)
	}
	found := false
	for _, p := range policies {
		if p.ID == pol.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("policy not persisted after restart")
	}

	// switch to memory backend and ensure service still works
	srv.Close()
	os.Setenv("STORE_BACKEND", "memory")
	os.Setenv("POLICY_BACKEND", "file")
	srv = startServer(t)
	defer srv.Close()

	tok := token(t)
	body := `{"tenantID":"default","subject":"user1","resource":"file1","action":"read","conditions":{}}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/check-access", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var dec struct {
		Allow bool `json:"allow"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dec); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !dec.Allow {
		t.Fatalf("expected allow with memory backend")
	}
}
