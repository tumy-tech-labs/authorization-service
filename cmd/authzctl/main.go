package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bradtumy/authorization-service/pkg/validator"
	"github.com/joho/godotenv"
)

func main() {
	// Load configuration from .env if present
	_ = godotenv.Load(".env")

	addrEnv := os.Getenv("AUTHZCTL_ADDR")
	if addrEnv == "" {
		addrEnv = "http://localhost:8080"
	}
	tokenEnv := os.Getenv("AUTHZCTL_TOKEN")

	addr := flag.String("addr", addrEnv, "service base address")
	token := flag.String("token", tokenEnv, "bearer token for authorization")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
	}
	switch args[0] {
	case "tenant":
		handleTenant(args[1:], *addr, *token)
	case "policy":
		handlePolicy(args[1:])
	case "check-access":
		handleCheckAccess(args[1:], *addr, *token)
	default:
		usage()
	}
}

func usage() {
	fmt.Println("usage: authzctl [--addr URL] [--token TOKEN] <command> [args]")
	fmt.Println("commands: tenant, policy, check-access")
	os.Exit(1)
}

func handleTenant(args []string, addr, token string) {
	if len(args) < 1 {
		fmt.Println("usage: authzctl tenant <create|delete> <id>")
		os.Exit(1)
	}
	client := &http.Client{}
	switch args[0] {
	case "create":
		if len(args) < 2 {
			fmt.Println("usage: authzctl tenant create <id>")
			os.Exit(1)
		}
		data, _ := json.Marshal(map[string]string{"tenantID": args[1], "name": args[1]})
		req, _ := http.NewRequest(http.MethodPost, addr+"/tenant/create", bytes.NewReader(data))
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
		if resp.StatusCode >= 300 {
			os.Exit(1)
		}
	case "delete":
		if len(args) < 2 {
			fmt.Println("usage: authzctl tenant delete <id>")
			os.Exit(1)
		}
		data, _ := json.Marshal(map[string]string{"tenantID": args[1]})
		req, _ := http.NewRequest(http.MethodPost, addr+"/tenant/delete", bytes.NewReader(data))
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
		if resp.StatusCode >= 300 {
			os.Exit(1)
		}
	default:
		fmt.Println("usage: authzctl tenant <create|delete> <id>")
		os.Exit(1)
	}
}

func handlePolicy(args []string) {
	if len(args) < 2 || args[0] != "validate" {
		fmt.Println("usage: authzctl policy validate <file>")
		os.Exit(1)
	}
	if err := validator.ValidatePolicyFile(args[1]); err != nil {
		fmt.Println("invalid policy:", err)
		os.Exit(1)
	}
	fmt.Println("policy is valid")
}

func handleCheckAccess(args []string, addr, token string) {
	fs := flag.NewFlagSet("check-access", flag.ExitOnError)
	tenant := fs.String("tenant", "", "tenant ID")
	subject := fs.String("subject", "", "subject performing the action")
	resource := fs.String("resource", "", "resource being accessed")
	action := fs.String("action", "", "action to check")
	fs.Parse(args)
	if *tenant == "" || *subject == "" || *resource == "" || *action == "" {
		fmt.Println("usage: authzctl check-access --tenant TENANT --subject SUBJECT --resource RESOURCE --action ACTION")
		os.Exit(1)
	}
	payload, _ := json.Marshal(map[string]any{
		"tenantID":   *tenant,
		"subject":    *subject,
		"resource":   *resource,
		"action":     *action,
		"conditions": map[string]any{},
	})
	req, _ := http.NewRequest(http.MethodPost, addr+"/check-access", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("request error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	if resp.StatusCode >= 300 {
		os.Exit(1)
	}
	var result struct {
		Allow bool `json:"allow"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		os.Exit(1)
	}
	if !result.Allow {
		os.Exit(1)
	}
}
