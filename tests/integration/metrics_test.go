package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	api "github.com/bradtumy/authorization-service/api"
	"github.com/bradtumy/authorization-service/internal/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

import _ "unsafe"

//go:linkname policyEval github.com/bradtumy/authorization-service/api.policyEval
var policyEval *prometheus.CounterVec

//go:linkname httpRequests github.com/bradtumy/authorization-service/internal/middleware.httpRequests
var httpRequests *prometheus.CounterVec

func setupServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	os.Setenv("OIDC_CONFIG_FILE", "/dev/null")
	middleware.LoadOIDCConfig()
	router := api.SetupRouter()
	srv := httptest.NewServer(router)
	tok := token(t)
	t.Cleanup(srv.Close)
	return srv, tok
}

func makeCheckRequest(t *testing.T, srv *httptest.Server, tok, body string) {
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/check-access", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
}

func TestCheckAccessRequestCounter(t *testing.T) {
	srv, tok := setupServer(t)
	before := testutil.ToFloat64(httpRequests.WithLabelValues("/check-access"))
	body := `{"tenantID":"default","subject":"user1","resource":"file1","action":"read","conditions":{}}`
	makeCheckRequest(t, srv, tok, body)
	after := testutil.ToFloat64(httpRequests.WithLabelValues("/check-access"))
	if after != before+1 {
		t.Fatalf("expected http_requests_total to increment by 1, before %v after %v", before, after)
	}
	// metrics endpoint should be reachable
	mreq, _ := http.NewRequest(http.MethodGet, srv.URL+"/metrics", nil)
	mreq.Header.Set("Authorization", "Bearer "+tok)
	mresp, err := http.DefaultClient.Do(mreq)
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	mresp.Body.Close()
	if mresp.StatusCode != http.StatusOK {
		t.Fatalf("metrics status %d", mresp.StatusCode)
	}
}

func TestPolicyEvalCounters(t *testing.T) {
	srv, tok := setupServer(t)
	allowBefore := testutil.ToFloat64(policyEval.WithLabelValues("allow", ""))
	denyBefore := testutil.ToFloat64(policyEval.WithLabelValues("deny", "other"))

	allowBody := `{"tenantID":"default","subject":"user1","resource":"file1","action":"read","conditions":{}}`
	denyBody := `{"tenantID":"default","subject":"user2","resource":"file3","action":"edit","conditions":{}}`
	makeCheckRequest(t, srv, tok, allowBody)
	makeCheckRequest(t, srv, tok, denyBody)

	allowAfter := testutil.ToFloat64(policyEval.WithLabelValues("allow", ""))
	denyAfter := testutil.ToFloat64(policyEval.WithLabelValues("deny", "other"))
	if allowAfter != allowBefore+1 {
		t.Fatalf("expected policy_eval_count allow to increment, before %v after %v", allowBefore, allowAfter)
	}
	if denyAfter != denyBefore+1 {
		t.Fatalf("expected policy_eval_count deny to increment, before %v after %v", denyBefore, denyAfter)
	}
}
