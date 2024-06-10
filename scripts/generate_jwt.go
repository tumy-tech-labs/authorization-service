package main

import (
	"fmt"
	"log"
	"time"

	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)

// CustomClaims defines the structure for our custom JWT claims
type CustomClaims struct {
	ClientID string `json:"client_id"`
	jwt.StandardClaims
}

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	clientID := os.Getenv("CLIENT_ID")
	if clientID == "" {
		log.Fatal("CLIENT_ID environment variable is not set")
	}

	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("CLIENT_SECRET environment variable is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	// Create the JWT claims, which includes the client ID and standard claims
	claims := CustomClaims{
		ClientID: clientID,
		StandardClaims: jwt.StandardClaims{
			Issuer:    "authorization-service",
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		},
	}

	// Create the token using the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Fatalf("Error signing token: %v", err)
	}

	fmt.Printf("Generated JWT Token: %s\n", tokenString)
}
