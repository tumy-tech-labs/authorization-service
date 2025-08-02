package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bradtumy/authorization-service/api"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if present
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning: could not load .env file: %v", err)
	}

	// Get the port from the environment variable
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	router := api.SetupRouter()
	log.Println("Starting server on :", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
