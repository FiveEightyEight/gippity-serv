package main

import (
	"fmt"
	"os"
	"log"
	"github.com/joho/godotenv"
)

func getEnvKey() string {
	err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
        fmt.Println("API_KEY is not set")
    } else {
        fmt.Printf("The OpenAI API Key is: %s\n", apiKey)
    }
	return apiKey
}

func main() {
	// come back later
	key := getEnvKey()
	fmt.Println("hey this is in progress...", key)
}