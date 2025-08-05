package validator

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the structure of the policy file.
type role struct {
	Name     string   `yaml:"name"`
	Policies []string `yaml:"policies"`
}

type subject struct {
	Role string `yaml:"role"`
}

type user struct {
	Username string   `yaml:"username"`
	Roles    []string `yaml:"roles"`
}

type policy struct {
	ID          string            `yaml:"id"`
	Description string            `yaml:"description"`
	Subjects    []subject         `yaml:"subjects"`
	Resource    []string          `yaml:"resource"`
	Action      []string          `yaml:"action"`
	Effect      string            `yaml:"effect"`
	Conditions  map[string]string `yaml:"conditions"`
	When        []string          `yaml:"when"`
}

// Config represents the structure of the policy file.
type Config struct {
	Roles    []role   `yaml:"roles"`
	Users    []user   `yaml:"users"`
	Policies []policy `yaml:"policies"`
}

// ValidateConfig performs schema validation on the provided configuration.
func ValidateConfig(cfg *Config) error {
	roleSet := make(map[string]struct{})
	for _, r := range cfg.Roles {
		roleSet[r.Name] = struct{}{}
	}

	for _, p := range cfg.Policies {
		if p.ID == "" {
			return fmt.Errorf("policy id is required")
		}
		if len(p.Action) == 0 {
			return fmt.Errorf("policy %s must have at least one action", p.ID)
		}
		if len(p.Resource) == 0 {
			return fmt.Errorf("policy %s must have at least one resource", p.ID)
		}
		if p.Effect == "" {
			return fmt.Errorf("policy %s must have an effect", p.ID)
		}
		for _, subj := range p.Subjects {
			if subj.Role == "" {
				return fmt.Errorf("policy %s has subject with empty role", p.ID)
			}
			if _, ok := roleSet[subj.Role]; !ok {
				return fmt.Errorf("policy %s references undefined role %s", p.ID, subj.Role)
			}
		}
	}
	return nil
}

// ValidatePolicyData validates the given YAML policy data.
func ValidatePolicyData(data []byte) error {
	var cfg Config
	if err := yaml.UnmarshalStrict(data, &cfg); err != nil {
		return err
	}
	return ValidateConfig(&cfg)
}

// ValidatePolicyFile validates a policy file at the given path.
func ValidatePolicyFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return ValidatePolicyData(data)
}
