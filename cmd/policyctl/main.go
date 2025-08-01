package main

import (
	"fmt"
	"os"

	"github.com/bradtumy/authorization-service/pkg/policycompiler"
	"github.com/bradtumy/authorization-service/pkg/validator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: policyctl <command> [args]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "compile":
		if len(os.Args) < 3 {
			fmt.Println("usage: policyctl compile \"<rule>\"")
			os.Exit(1)
		}
		rule := os.Args[2]
		compiler := policycompiler.NewOpenAICompiler(os.Getenv("OPENAI_API_KEY"))
		yaml, err := compiler.Compile(rule)
		if err != nil {
			fmt.Println("compile error:", err)
			os.Exit(1)
		}
		fmt.Println(yaml)
	case "validate":
		if len(os.Args) < 3 {
			fmt.Println("usage: policyctl validate <file.yaml>")
			os.Exit(1)
		}
		if err := validator.ValidatePolicyFile(os.Args[2]); err != nil {
			fmt.Println("invalid policy:", err)
			os.Exit(1)
		}
		fmt.Println("policy is valid")
	default:
		fmt.Println("usage: policyctl <compile|validate> ...")
		os.Exit(1)
	}
}
