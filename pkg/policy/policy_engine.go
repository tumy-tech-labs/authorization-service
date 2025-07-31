package policy

// PolicyEngine represents an engine for policy evaluation.
//
// The engine currently performs simple matching on resource and action
// attributes. Policies may optionally scope themselves to specific roles via
// the `Subjects` field. This check was previously ignored which meant a policy
// could be enforced even when the requesting role wasn't included in the
// policy's subjects. The evaluation now ensures the policy explicitly allows
// the role before considering resources and actions.

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
			// Ensure the policy applies to the current role
			if len(policy.Subjects) > 0 {
				allowed := false
				for _, subj := range policy.Subjects {
					if subj.Role == roleName {
						allowed = true
						break
					}
				}
				if !allowed {
					continue
				}
			}

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
