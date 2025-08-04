package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	jwt "github.com/golang-jwt/jwt/v4"
	"gopkg.in/yaml.v2"
)

type oidcProvider struct {
	Issuer   string
	Audience string
	JWKS     *keyfunc.JWKS
}

var providers []oidcProvider

// LoadOIDCConfig loads OIDC provider configuration from environment variables or a YAML file.
func LoadOIDCConfig() {
	providers = nil
	loadFromEnv()
	if len(providers) == 0 {
		loadFromFile()
	}
}

func loadFromEnv() {
	issuers := strings.Split(os.Getenv("OIDC_ISSUERS"), ",")
	audiences := strings.Split(os.Getenv("OIDC_AUDIENCES"), ",")
	if len(issuers) == 0 || issuers[0] == "" {
		return
	}
	for i := range issuers {
		iss := strings.TrimSpace(issuers[i])
		aud := ""
		if i < len(audiences) {
			aud = strings.TrimSpace(audiences[i])
		}
		if iss == "" {
			continue
		}
		if jwks, err := fetchJWKS(iss); err == nil {
			providers = append(providers, oidcProvider{Issuer: iss, Audience: aud, JWKS: jwks})
		}
	}
}

func loadFromFile() {
	path := os.Getenv("OIDC_CONFIG_FILE")
	if path == "" {
		path = "configs/oidc.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg struct {
		Providers []struct {
			Issuer   string `yaml:"issuer"`
			Audience string `yaml:"audience"`
		} `yaml:"providers"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return
	}
	for _, p := range cfg.Providers {
		if p.Issuer == "" {
			continue
		}
		if jwks, err := fetchJWKS(p.Issuer); err == nil {
			providers = append(providers, oidcProvider{Issuer: p.Issuer, Audience: p.Audience, JWKS: jwks})
		}
	}
}

func fetchJWKS(issuer string) (*keyfunc.JWKS, error) {
	configURL := strings.TrimSuffix(issuer, "/") + "/.well-known/openid-configuration"
	resp, err := http.Get(configURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var conf struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&conf); err != nil {
		return nil, err
	}
	interval := time.Hour
	if s := os.Getenv("OIDC_JWKS_REFRESH_INTERVAL"); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			interval = d
		}
	}
	return keyfunc.Get(conf.JWKSURI, keyfunc.Options{RefreshInterval: interval, RefreshTimeout: 5 * time.Second, RefreshUnknownKID: true})
}

// JWTMiddleware validates ID tokens using OIDC providers and JWKS.
func JWTMiddleware(next http.Handler) http.Handler {
	if len(providers) == 0 {
		LoadOIDCConfig()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if len(providers) > 0 {
			parser := jwt.Parser{}
			unverified := jwt.MapClaims{}
			if _, _, err := parser.ParseUnverified(tokenString, unverified); err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			iss, _ := unverified["iss"].(string)
			audClaim := unverified["aud"]
			var prov *oidcProvider
			for i := range providers {
				if providers[i].Issuer == iss && audienceMatch(audClaim, providers[i].Audience) {
					prov = &providers[i]
					break
				}
			}
			if prov == nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, prov.JWKS.Keyfunc)
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			if issClaim, _ := claims["iss"].(string); issClaim != prov.Issuer {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			if !audienceMatch(claims["aud"], prov.Audience) {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			if sub, _ := claims["sub"].(string); sub != "" {
				ctx := context.WithValue(r.Context(), "subject", sub)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func audienceMatch(claim interface{}, aud string) bool {
	if aud == "" {
		return true
	}
	switch v := claim.(type) {
	case string:
		return v == aud
	case []interface{}:
		for _, a := range v {
			if s, ok := a.(string); ok && s == aud {
				return true
			}
		}
	case []string:
		for _, s := range v {
			if s == aud {
				return true
			}
		}
	}
	return false
}
