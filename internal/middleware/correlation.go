package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// correlationIDKey is the context key for the correlation ID.
type correlationIDKey struct{}

// CorrelationIDFromContext retrieves the correlation ID from context.
func CorrelationIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(correlationIDKey{}).(string); ok {
		return v
	}
	return ""
}

// CorrelationMiddleware generates a correlation ID for each request and stores it in the context.
func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		ctx := context.WithValue(r.Context(), correlationIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
