package main

import (
	"fmt"
	"os"
	"log"
	"github.com/joho/godotenv"
)

models := map[string]string{
	"GPT-4o" : "gpt-4o",
	"GPT-4o mini": "gpt-4o-mini",
	"GPT-4 Turbo": "gpt-4-turbo",
	"GPT-3.5 Turbo": "gpt-3.5-turbo-0125",
}

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
	for modelName, modelCode := range models {
		fmt.Println(modelName, modelCode)
	}
}