package policy

type Policy struct {
	ID         string   `yaml:"id"`
	Resource   string   `yaml:"resource"`
	Action     string   `yaml:"action"`
	Effect     string   `yaml:"effect"`
	Conditions []string `yaml:"conditions"`
}
