package contextprovider

import "net/http"

// RiskProvider reads a static risk score from the X-Risk-Score header.
type RiskProvider struct{}

// GetContext extracts the risk score header if present.
func (RiskProvider) GetContext(req *http.Request) (map[string]string, error) {
	score := req.Header.Get("X-Risk-Score")
	if score == "" {
		score = "0"
	}
	return map[string]string{"risk_score": score}, nil
}
