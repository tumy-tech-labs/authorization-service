package integration

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	api "github.com/bradtumy/authorization-service/api"
	"github.com/bradtumy/authorization-service/internal/middleware"
	jwt "github.com/golang-jwt/jwt/v4"
	jose "gopkg.in/go-jose/go-jose.v2"
)

func TestOIDCTokenValidation(t *testing.T) {
	// Generate RSA key and JWKS
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	kid := "testkid"
	jwk := jose.JSONWebKey{Key: &priv.PublicKey, KeyID: kid, Algorithm: "RS256", Use: "sig"}
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}
	jwksBytes, _ := json.Marshal(jwks)

	// Mock OIDC provider with JWKS endpoint
	var oidc *httptest.Server
	oidc = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]string{"jwks_uri": oidc.URL + "/keys"})
		case "/keys":
			w.Header().Set("Content-Type", "application/json")
			w.Write(jwksBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer oidc.Close()

	// Configure middleware to use the mock issuer
	os.Setenv("OIDC_ISSUERS", oidc.URL)
	os.Setenv("OIDC_AUDIENCES", "test-aud")
	middleware.LoadOIDCConfig()

	// Start API server
	router := api.SetupRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	makeToken := func(iss, aud string, exp time.Time) string {
		claims := jwt.MapClaims{
			"iss": iss,
			"sub": "tester",
			"aud": aud,
			"exp": exp.Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = kid
		str, err := token.SignedString(priv)
		if err != nil {
			t.Fatalf("sign token: %v", err)
		}
		return str
	}

	call := func(tok string) *http.Response {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/metrics", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		return resp
	}

	t.Run("valid token", func(t *testing.T) {
		tok := makeToken(oidc.URL, "test-aud", time.Now().Add(time.Hour))
		resp := call(tok)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 got %d", resp.StatusCode)
		}
	})

	t.Run("wrong audience", func(t *testing.T) {
		tok := makeToken(oidc.URL, "other-aud", time.Now().Add(time.Hour))
		resp := call(tok)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 got %d", resp.StatusCode)
		}
	})

	t.Run("wrong issuer", func(t *testing.T) {
		tok := makeToken("http://wrong", "test-aud", time.Now().Add(time.Hour))
		resp := call(tok)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 got %d", resp.StatusCode)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		tok := makeToken(oidc.URL, "test-aud", time.Now().Add(-time.Hour))
		resp := call(tok)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 got %d", resp.StatusCode)
		}
	})
}
