package main

import (
	"fmt"
	"os"

	"github.com/bradtumy/authorization-service/pkg/policycompiler"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "compile" {
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
}
