package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bradtumy/authorization-service/pkg/graph"
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
	case "tenant":
		handleTenant(os.Args[2:])
	case "graph":
		handleGraph(os.Args[2:])
	default:
		fmt.Println("usage: policyctl <compile|validate|tenant|graph> ...")
		os.Exit(1)
	}
}

func handleTenant(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: policyctl tenant <create|delete|list> ...")
		os.Exit(1)
	}
	base := os.Getenv("POLICYCTL_ADDR")
	if base == "" {
		base = "http://localhost:8080"
	}
	token := os.Getenv("POLICYCTL_TOKEN")
	client := &http.Client{}
	switch args[0] {
	case "create":
		if len(args) < 2 {
			fmt.Println("usage: policyctl tenant create <id>")
			os.Exit(1)
		}
		data, _ := json.Marshal(map[string]string{"tenantID": args[1], "name": args[1]})
		req, _ := http.NewRequest(http.MethodPost, base+"/tenant/create", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("request error:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	case "delete":
		if len(args) < 2 {
			fmt.Println("usage: policyctl tenant delete <id>")
			os.Exit(1)
		}
		data, _ := json.Marshal(map[string]string{"tenantID": args[1]})
		req, _ := http.NewRequest(http.MethodPost, base+"/tenant/delete", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("request error:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	case "list":
		req, _ := http.NewRequest(http.MethodGet, base+"/tenant/list", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("request error:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	default:
		fmt.Println("usage: policyctl tenant <create|delete|list> ...")
		os.Exit(1)
	}
}

func handleGraph(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: policyctl graph <add|list|delegate> ...")
		os.Exit(1)
	}
	file := "graph.json"
	g := loadGraphFile(file)
	switch args[0] {
	case "add":
		if len(args) < 3 {
			fmt.Println("usage: policyctl graph add <src> <dst>")
			os.Exit(1)
		}
		g.AddRelation(args[1], args[2])
		saveGraphFile(file, g)
		fmt.Println("relationship added")
	case "delegate":
		if len(args) < 3 {
			fmt.Println("usage: policyctl graph delegate <delegator> <delegatee>")
			os.Exit(1)
		}
		g.AddRelation("user:"+args[1], "user:"+args[2])
		saveGraphFile(file, g)
		fmt.Println("delegation added")
	case "list":
		rels := g.List()
		for src, targets := range rels {
			for _, dst := range targets {
				fmt.Printf("%s -> %s\n", src, dst)
			}
		}
	default:
		fmt.Println("usage: policyctl graph <add|list|delegate> ...")
		os.Exit(1)
	}
}

func loadGraphFile(path string) *graph.Graph {
	g := graph.New()
	data, err := os.ReadFile(path)
	if err != nil {
		return g
	}
	var m map[string][]string
	if err := json.Unmarshal(data, &m); err != nil {
		return g
	}
	for src, targets := range m {
		for _, dst := range targets {
			g.AddRelation(src, dst)
		}
	}
	return g
}

func saveGraphFile(path string, g *graph.Graph) {
	m := g.List()
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(path, data, 0644)
}
