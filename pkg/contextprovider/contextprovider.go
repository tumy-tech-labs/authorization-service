package contextprovider

import "net/http"

// ContextProvider retrieves context values from an HTTP request.
type ContextProvider interface {
	GetContext(req *http.Request) (map[string]string, error)
}

// Chain executes multiple providers and merges their context values.
type Chain []ContextProvider

// GetContext gathers context values from all providers in the chain.
func (c Chain) GetContext(req *http.Request) map[string]string {
	ctx := make(map[string]string)
	for _, p := range c {
		vals, err := p.GetContext(req)
		if err != nil {
			continue
		}
		for k, v := range vals {
			ctx[k] = v
		}
	}
	return ctx
}
