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
	for k, v := range env {
		ctx[k] = v
	}

	// Collect candidate subjects including delegation chain.
	subjects := []string{subject}
	if pe.graph != nil {
		queue := []string{subject}
		visited := map[string]struct{}{subject: struct{}{}}
		for len(queue) > 0 {
			s := queue[0]
			queue = queue[1:]
			for _, t := range pe.graph.Targets("user:" + s) {
				if strings.HasPrefix(t, "user:") {
					u := strings.TrimPrefix(t, "user:")
					if _, ok := visited[u]; !ok {
						visited[u] = struct{}{}
						subjects = append(subjects, u)
						queue = append(queue, u)
					}
				}
			}
		}
	}

	for idx, subj := range subjects {
		user, exists := pe.store.Users[subj]
		if !exists {
			if idx == 0 {
				return Decision{Allow: false, Reason: "user not found", Context: ctx}
			}
			continue
		}

		// Gather roles from user definition and graph-based group memberships.
		roles := append([]string{}, user.Roles...)
		if pe.graph != nil {
			for _, target := range pe.graph.Targets("user:" + subj) {
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
					for _, subjRole := range policy.Subjects {
						if subjRole.Role == roleName {
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
								dec := Decision{Allow: false, PolicyID: policy.ID, Reason: "conditions not satisfied", Context: ctx}
								if subj != subject {
									dec.Delegator = subj
								}
								return dec
							}
							if ok := evaluateWhen(policy.When, env); !ok {
								dec := Decision{Allow: false, PolicyID: policy.ID, Reason: "conditions not satisfied", Context: ctx}
								if subj != subject {
									dec.Delegator = subj
								}
								return dec
							}
							dec := Decision{PolicyID: policy.ID, Context: ctx}
							if subj != subject {
								dec.Delegator = subj
							}
							switch policy.Effect {
							case "allow":
								dec.Allow = true
								dec.Reason = "allowed by policy"
							case "deny":
								dec.Allow = false
								dec.Reason = "denied by policy"
							}
							return dec
						}
					}
				}
			}
		}
	}

	return Decision{Allow: false, Reason: "no matching policy", Context: ctx}
}
