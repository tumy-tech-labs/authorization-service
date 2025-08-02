package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	oidc "github.com/coreos/go-oidc"
	"github.com/joho/godotenv"
)

var (
	verifier *oidc.IDTokenVerifier
)

func init() {
	godotenv.Load()
	issuer := os.Getenv("OIDC_ISSUER")
	audience := os.Getenv("OIDC_AUDIENCE")
	if issuer == "" {
		return
	}
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		panic("failed to create OIDC provider: " + err.Error())
	}
	cfg := &oidc.Config{ClientID: audience}
	verifier = provider.Verifier(cfg)
}

// JWTMiddleware validates ID tokens using OIDC and JWKS.
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if verifier != nil {
			ctx := r.Context()
			idToken, err := verifier.Verify(ctx, tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			var claims struct {
				Subject string `json:"sub"`
			}
			if err := idToken.Claims(&claims); err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			ctx = context.WithValue(ctx, "subject", claims.Subject)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
