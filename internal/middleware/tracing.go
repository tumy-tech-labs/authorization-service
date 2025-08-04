package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("authorization-service")

// TracingMiddleware starts a root span for each incoming request.
func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
