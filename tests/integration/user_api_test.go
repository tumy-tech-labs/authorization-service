package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	api "github.com/bradtumy/authorization-service/api"
	"github.com/bradtumy/authorization-service/internal/middleware"
	"github.com/bradtumy/authorization-service/pkg/user"
	jwt "github.com/golang-jwt/jwt/v4"
)

func tokenFor(t *testing.T, sub string) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": sub})
	str, err := tok.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return str
}

func TestUserAPI(t *testing.T) {
	os.Setenv("OIDC_CONFIG_FILE", "/dev/null")
	middleware.LoadOIDCConfig()
	user.Reset()
	user.EnablePersistence(false)
	if _, err := user.Create("default", "admin", []string{"TenantAdmin"}); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	router := api.SetupRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	adminTok := tokenFor(t, "admin")

	// create bob
	body := `{"tenantID":"default","username":"bob","roles":["viewer"]}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/user/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminTok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("create user: %v status %d", err, resp.StatusCode)
	}
	resp.Body.Close()

	// assign role
	body = `{"tenantID":"default","username":"bob","roles":["editor"]}`
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/user/assign-role", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminTok)
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("assign role: %v status %d", err, resp.StatusCode)
	}
	resp.Body.Close()

	// list users
	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/user/list?tenantID=default", nil)
	req.Header.Set("Authorization", "Bearer "+adminTok)
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("list users: %v status %d", err, resp.StatusCode)
	}
	var list []user.User
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	resp.Body.Close()
	if len(list) != 2 { // admin and bob
		t.Fatalf("expected 2 users, got %d", len(list))
	}

	// get user bob
	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/user/get?tenantID=default&username=bob", nil)
	req.Header.Set("Authorization", "Bearer "+adminTok)
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("get user: %v status %d", err, resp.StatusCode)
	}
	var u user.User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	resp.Body.Close()
	if u.Username != "bob" || len(u.Roles) != 1 || u.Roles[0] != "editor" {
		t.Fatalf("unexpected user %+v", u)
	}

	// unauthorized with bob token
	bobTok := tokenFor(t, "bob")
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/user/delete", strings.NewReader(`{"tenantID":"default","username":"bob"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bobTok)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("bob delete request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// delete bob as admin
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/user/delete", strings.NewReader(`{"tenantID":"default","username":"bob"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminTok)
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("delete user: %v status %d", err, resp.StatusCode)
	}
	resp.Body.Close()
}
