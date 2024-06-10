package policy

import (
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"
)

type PolicyStore struct {
	policies map[string]Policy
	mu       sync.RWMutex
}

func NewPolicyStore() *PolicyStore {
	return &PolicyStore{
		policies: make(map[string]Policy),
	}
}

func (ps *PolicyStore) LoadPolicies(filePath string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var policies []Policy
	err = yaml.Unmarshal(data, &policies)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		ps.policies[policy.ID] = policy
	}

	return nil
}

func (ps *PolicyStore) GetPolicy(id string) (Policy, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	policy, exists := ps.policies[id]
	return policy, exists
}
