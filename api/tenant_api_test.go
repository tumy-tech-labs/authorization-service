package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateTenant(t *testing.T) {
	id := "tenantCreate"
	body := fmt.Sprintf(`{"tenantID":"%s","name":"%s"}`, id, id)
	r := httptest.NewRequest(http.MethodPost, "/tenant/create", strings.NewReader(body))
	w := httptest.NewRecorder()
	CreateTenant(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var tenant Tenant
	if err := json.NewDecoder(w.Body).Decode(&tenant); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if tenant.ID != id {
		t.Fatalf("expected id %s, got %s", id, tenant.ID)
	}
	if store, ok := policyStores[id]; !ok || len(store.Policies) != 0 {
		t.Fatalf("expected empty policy store for tenant")
	}
	// cleanup
	delBody := fmt.Sprintf(`{"tenantID":"%s"}`, id)
	dr := httptest.NewRequest(http.MethodPost, "/tenant/delete", strings.NewReader(delBody))
	dw := httptest.NewRecorder()
	DeleteTenant(dw, dr)
}

func TestListTenants(t *testing.T) {
	id1 := "tenantList1"
	id2 := "tenantList2"
	for _, id := range []string{id1, id2} {
		body := fmt.Sprintf(`{"tenantID":"%s","name":"%s"}`, id, id)
		r := httptest.NewRequest(http.MethodPost, "/tenant/create", strings.NewReader(body))
		w := httptest.NewRecorder()
		CreateTenant(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("create tenant %s failed", id)
		}
	}
	r := httptest.NewRequest(http.MethodGet, "/tenant/list", nil)
	w := httptest.NewRecorder()
	ListTenants(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var list []Tenant
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found1, found2 := false, false
	for _, tnt := range list {
		if tnt.ID == id1 {
			found1 = true
		}
		if tnt.ID == id2 {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Fatalf("expected tenants in list")
	}
	// cleanup
	for _, id := range []string{id1, id2} {
		delBody := fmt.Sprintf(`{"tenantID":"%s"}`, id)
		dr := httptest.NewRequest(http.MethodPost, "/tenant/delete", strings.NewReader(delBody))
		dw := httptest.NewRecorder()
		DeleteTenant(dw, dr)
	}
}

func TestDeleteTenant(t *testing.T) {
	id := "tenantDelete"
	createBody := fmt.Sprintf(`{"tenantID":"%s","name":"%s"}`, id, id)
	cr := httptest.NewRequest(http.MethodPost, "/tenant/create", strings.NewReader(createBody))
	cw := httptest.NewRecorder()
	CreateTenant(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("create tenant failed")
	}
	delBody := fmt.Sprintf(`{"tenantID":"%s"}`, id)
	r := httptest.NewRequest(http.MethodPost, "/tenant/delete", strings.NewReader(delBody))
	w := httptest.NewRecorder()
	DeleteTenant(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if _, err := backend.LoadTenant(r.Context(), id); err == nil {
		t.Fatalf("tenant not deleted")
	}
}
