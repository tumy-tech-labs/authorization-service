package main

import (
	"fmt"

	sdk "github.com/bradtumy/authorization-service/sdk/go"
)

func main() {
	client := sdk.NewClient("http://localhost:8080")
	decision, err := client.CheckAccess(sdk.AccessRequest{
		TenantID: "default",
		Subject:  "alice",
		Resource: "file1",
		Action:   "read",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("allow:", decision.Allow)
}
