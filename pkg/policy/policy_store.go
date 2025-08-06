package policy

import (
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/bradtumy/authorization-service/pkg/validator"
)

// PolicyStore represents a store for policies, roles, and users.
type PolicyStore struct {
	Policies map[string]Policy
	Roles    map[string]Role
	Users    map[string]User
	mu       sync.RWMutex
}

// NewPolicyStore creates a new PolicyStore instance.
func NewPolicyStore() *PolicyStore {
	return &PolicyStore{
		Policies: make(map[string]Policy),
		Roles:    make(map[string]Role),
		Users:    make(map[string]User),
	}
}

// LoadPolicies loads policies, roles, and users from the specified file.
// The configuration is validated before being swapped into the store.
func (ps *PolicyStore) LoadPolicies(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err = validator.ValidatePolicyData(data); err != nil {
		return err
	}

	var config struct {
		Roles    []Role   `yaml:"roles"`
		Users    []User   `yaml:"users"`
		Policies []Policy `yaml:"policies"`
	}

	if err = yaml.UnmarshalStrict(data, &config); err != nil {
		return err
	}

	newRoles := make(map[string]Role)
	newUsers := make(map[string]User)
	newPolicies := make(map[string]Policy)

	for _, role := range config.Roles {
		newRoles[role.Name] = role
	}
	for _, user := range config.Users {
		newUsers[user.Username] = user
	}
	for _, policy := range config.Policies {
		newPolicies[policy.ID] = policy
	}

	ps.mu.Lock()
	ps.Roles = newRoles
	ps.Users = newUsers
	ps.Policies = newPolicies
	ps.mu.Unlock()

	return nil
}

// ReplacePolicies swaps the current policies with the provided list. Roles and
// users remain untouched. This is primarily used when loading policies from a
// database backend.
func (ps *PolicyStore) ReplacePolicies(policies []Policy) {
	newPolicies := make(map[string]Policy)
	for _, p := range policies {
		newPolicies[p.ID] = p
	}
	ps.mu.Lock()
	ps.Policies = newPolicies
	ps.mu.Unlock()
}

// GetPolicy retrieves a policy by its ID.
func (ps *PolicyStore) GetPolicy(id string) (Policy, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	policy, exists := ps.Policies[id]
	return policy, exists
}
