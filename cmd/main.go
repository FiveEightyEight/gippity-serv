package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var models = map[string]string{
	"GPT-4o":        "gpt-4o",
	"GPT-4o mini":   "gpt-4o-mini",
	"GPT-4 Turbo":   "gpt-4-turbo",
	"GPT-3.5 Turbo": "gpt-3.5-turbo-0125",
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatBody struct {
	Model    string    `json:"model`
	Messages []Message `json:"messages"`
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
