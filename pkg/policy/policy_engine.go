package policy

import (
	"strings"

	"github.com/bradtumy/authorization-service/pkg/graph"
)

// PolicyEngine evaluates policies to determine access decisions.
//
// The engine performs simple matching on resource and action attributes. Policies
// may optionally scope themselves to specific roles via the `Subjects` field.
// Evaluation stops at the first matching policy and returns a structured
// decision describing the result.
type PolicyEngine struct {
	store *PolicyStore
	graph *graph.Graph
}

// NewPolicyEngine creates a new PolicyEngine instance.
func NewPolicyEngine(store *PolicyStore, g *graph.Graph) *PolicyEngine {
	return &PolicyEngine{store: store, graph: g}
}

// Evaluate determines whether the given subject is allowed to perform the
// specified action on the resource. It returns a Decision describing the
// outcome and does not log sensitive data.
func (pe *PolicyEngine) Evaluate(subject, resource, action string, env map[string]string) Decision {
	ctx := map[string]string{
		"subject":  subject,
		"resource": resource,
		"action":   action,
	}

	user, exists := pe.store.Users[subject]
	if !exists {
		return Decision{Allow: false, Reason: "user not found", Context: ctx}
	}

	// Gather roles from user definition and graph-based group memberships.
	roles := append([]string{}, user.Roles...)
	if pe.graph != nil {
		for _, target := range pe.graph.Targets("user:" + subject) {
			if strings.HasPrefix(target, "group:") {
				roles = append(roles, strings.TrimPrefix(target, "group:"))
			}
		}
	}

	for _, roleName := range roles {
		role, exists := pe.store.Roles[roleName]
		if !exists {
			continue
		}

		for _, policyID := range role.Policies {
			policy, exists := pe.store.Policies[policyID]
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
				matchResource := polResource == "*" || polResource == resource
				if !matchResource && pe.graph != nil {
					if pe.graph.HasPath("group:"+polResource, "resource:"+resource) {
						matchResource = true
					}
				}
				for _, polAction := range policy.Action {
					if matchResource && (polAction == "*" || polAction == action) {
						if ok := evaluateConditions(policy.Conditions, env); !ok {
							return Decision{Allow: false, PolicyID: policy.ID, Reason: "conditions not satisfied", Context: ctx}
						}
						switch policy.Effect {
						case "allow":
							return Decision{Allow: true, PolicyID: policy.ID, Reason: "allowed by policy", Context: ctx}
						case "deny":
							return Decision{Allow: false, PolicyID: policy.ID, Reason: "denied by policy", Context: ctx}
						}
					}
				}
			}
		}
	}

	return Decision{Allow: false, Reason: "no matching policy", Context: ctx}
}
