package contextprovider

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// ContextProvider retrieves context values from an HTTP request.
type ContextProvider interface {
	GetContext(req *http.Request) (map[string]string, error)
}

// Chain executes multiple providers and merges their context values.
type Chain []ContextProvider

// GetContext gathers context values from all providers in the chain.
func (c Chain) GetContext(req *http.Request) map[string]string {
	ctxVals := make(map[string]string)
	tracer := otel.Tracer("authorization-service")
	ctx, span := tracer.Start(req.Context(), "ContextEvaluation")
	defer span.End()
	for _, p := range c {
		vals, err := p.GetContext(req.WithContext(ctx))
		if err != nil {
			continue
		}
		for k, v := range vals {
			ctxVals[k] = v
			span.SetAttributes(attribute.String(k, v))
		}
	}
	return ctxVals
}
