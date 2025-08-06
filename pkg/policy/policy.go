package policy

// Role represents a user role.
type Role struct {
	Name     string   `yaml:"name"`
	Policies []string `yaml:"policies"`
}

// User represents a user and their assigned roles.
type User struct {
	Username string   `yaml:"username"`
	Roles    []string `yaml:"roles"`
}

type Subject struct {
	Role string `yaml:"role"`
}

// Policy represents an authorization policy.
type Policy struct {
	ID          string            `yaml:"id"`
	Description string            `yaml:"description"`
	Subjects    []Subject         `yaml:"subjects"`
	Resource    []string          `yaml:"resource"`
	Action      []string          `yaml:"action"`
	Effect      string            `yaml:"effect"`
	Conditions  map[string]string `yaml:"conditions"`
	When        []string          `yaml:"when"`
}
