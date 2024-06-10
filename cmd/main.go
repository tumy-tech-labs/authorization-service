package main

import (
	"log"
	"net/http"

	"github.com/bradtumy/authorization-service/api"
)

func main() {
	router := api.SetupRouter()
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
