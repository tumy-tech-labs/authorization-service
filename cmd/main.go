package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/bradtumy/authorization-service/api"
	"github.com/bradtumy/authorization-service/internal/telemetry"
	"github.com/bradtumy/authorization-service/pkg/user"
	"github.com/joho/godotenv"
)

func main() {
	persistUsers := flag.Bool("persist-users", false, "persist users to configs/<tenantID>/users.yaml")
	flag.Parse()

	// Load environment variables from .env file if present.
	// Missing files are ignored to avoid noisy startup warnings.
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: could not load .env file: %v", err)
	}

	// Get the port from the environment variable
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	ctx := context.Background()
	shutdown, err := telemetry.InitTracer(ctx)
	if err != nil {
		log.Fatalf("failed to init tracing: %v", err)
	}
	defer func() { _ = shutdown(ctx) }()

	user.EnablePersistence(*persistUsers)
	router := api.SetupRouter()
	log.Println("Starting server on :", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
