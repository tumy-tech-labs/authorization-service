package policycompiler

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/bradtumy/authorization-service/pkg/policy"
)

// Compiler defines the interface for natural language policy compilers.
type Compiler interface {
	Compile(rule string) (string, error)
}

// OpenAICompiler is a stub implementation that would call the OpenAI API to
// translate natural language rules into YAML policies. If no API key is
// provided, it falls back to a very simple local parser.
type OpenAICompiler struct {
	apiKey string
}

// NewOpenAICompiler creates a new OpenAI-backed policy compiler.
// The apiKey should be provided via configuration or environment variables.
func NewOpenAICompiler(apiKey string) Compiler {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &OpenAICompiler{apiKey: apiKey}
}

type compiledPolicy struct {
	Subjects []policy.Subject `yaml:"subjects"`
	Action   []string         `yaml:"action"`
	Resource []string         `yaml:"resource"`
	Effect   string           `yaml:"effect"`
}

// Compile converts a natural language rule into a YAML policy. When an API key
// is available, this method would invoke the OpenAI API. For now, it falls back
// to a basic heuristic parser.
func (c *OpenAICompiler) Compile(rule string) (string, error) {
	if c.apiKey != "" {
		// Pseudocode for calling the OpenAI API.
		// client := openai.NewClient(c.apiKey)
		// prompt := fmt.Sprintf("Translate the following rule into the YAML policy schema: %q", rule)
		// resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{ ... })
		// if err != nil { return "", err }
		// return resp.Choices[0].Message.Content, nil
	}

	// Fallback simple parser expects pattern: "<subject> can <action> <resource>".
	lower := strings.ToLower(rule)
	idx := strings.Index(lower, " can ")
	if idx == -1 {
		return "", fmt.Errorf("unsupported rule format")
	}
	subject := strings.TrimSpace(rule[:idx])
	rest := strings.TrimSpace(rule[idx+len(" can "):])
	parts := strings.SplitN(rest, " ", 2)
	action := parts[0]
	resource := ""
	if len(parts) > 1 {
		resource = strings.TrimSpace(parts[1])
	}

	p := compiledPolicy{
		Subjects: []policy.Subject{{Role: subject}},
		Action:   []string{action},
		Resource: []string{resource},
		Effect:   "allow",
	}
	out, err := yaml.Marshal(&p)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
