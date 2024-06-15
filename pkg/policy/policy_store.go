package policy

import (
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"
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
func (ps *PolicyStore) LoadPolicies(filePath string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var config struct {
		Roles    []Role   `yaml:"roles"`
		Users    []User   `yaml:"users"`
		Policies []Policy `yaml:"policies"`
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	for _, role := range config.Roles {
		ps.Roles[role.Name] = role
	}
	for _, user := range config.Users {
		ps.Users[user.Username] = user
	}
	for _, policy := range config.Policies {
		ps.Policies[policy.ID] = policy
	}

	return nil
}

// GetPolicy retrieves a policy by its ID.
func (ps *PolicyStore) GetPolicy(id string) (Policy, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	policy, exists := ps.Policies[id]
	return policy, exists
}
