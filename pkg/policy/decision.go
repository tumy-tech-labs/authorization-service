package policy

// Decision represents the outcome of a policy evaluation.
type Decision struct {
	Allow    bool              `json:"allow"`
	PolicyID string            `json:"policy_id,omitempty"`
	Reason   string            `json:"reason"`
	Context  map[string]string `json:"context,omitempty"`
}
