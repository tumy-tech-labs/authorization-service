package policy

import "fmt"

// PolicyEngine represents an engine for policy evaluation.
type PolicyEngine struct {
	store *PolicyStore
}

// NewPolicyEngine creates a new PolicyEngine instance.
func NewPolicyEngine(store *PolicyStore) *PolicyEngine {
	return &PolicyEngine{store: store}
}

// Evaluate evaluates the permission for the given subject, resource, action, and conditions.
func (pe *PolicyEngine) Evaluate(subject, resource, action string, conditions []string) bool {
	user, exists := pe.store.Users[subject]
	fmt.Println("::Policy Engine: Subject:", subject)
	fmt.Println("::Policy Engine: Subject:", user, exists)
	if !exists {
		return false // User not found
	}

	for _, roleName := range user.Roles {
		role, exists := pe.store.Roles[roleName]
		fmt.Println("::Policy Engine: Roles", role, exists)
		if !exists {
			continue
		}

		for _, policyID := range role.Policies {
			policy, exists := pe.store.Policies[policyID]
			fmt.Println("::Policy Engine: Policy", policy, exists)
			if !exists {
				continue
			}
			// debug statements
			fmt.Println("Resources: ", resource)
			fmt.Println("Action: ", action)
			fmt.Println("Policy: ", policy)
			fmt.Println("Policy Resource: ", policy.Resource)
			fmt.Println("Policy Action: ", policy.Action)

			for _, polResource := range policy.Resource {
				for _, polAction := range policy.Action {
					if polResource == "*" || polResource == resource {
						if polAction == "*" || polAction == action {
							return policy.Effect == "allow"
						}
					}
				}
			}
		}
	}

	return false // No matching policy found
}
