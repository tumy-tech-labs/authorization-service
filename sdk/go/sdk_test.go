package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type checkAccessResponse struct {
	Allow    bool   `json:"allow"`
	PolicyID string `json:"policyID"`
	Reason   string `json:"reason"`
}

func TestClient(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/check-access", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(checkAccessResponse{Allow: true, PolicyID: "p1", Reason: "ok"})
	})
	mux.HandleFunc("/compile", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("policy: allow"))
	})
	mux.HandleFunc("/validate-policy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("policy is valid"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(srv.URL)
	dec, err := c.CheckAccess(AccessRequest{TenantID: "t", Subject: "s", Resource: "r", Action: "a"})
	if err != nil || !dec.Allow {
		t.Fatalf("CheckAccess failed: %v", err)
	}
	if _, err := c.CompileRule("t", "rule"); err != nil {
		t.Fatalf("CompileRule failed: %v", err)
	}
	if err := c.ValidatePolicy("t", "policy"); err != nil {
		t.Fatalf("ValidatePolicy failed: %v", err)
	}
}
