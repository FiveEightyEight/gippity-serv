package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

func callOpenAI() {
	// Load the OpenAI API key from environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	// Set up the request payload
	requestBody := ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: "Write a haiku that explains the concept of recursion.",
			},
		},
	}

	// Convert the request payload to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set the required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Output the response
	fmt.Println("Response status:", resp.Status)
	fmt.Println("Response body:", string(body))
}

func main() {
	// come back later
	key := getEnvKey()
	fmt.Println("hey this is in progress...", key)
	for modelName, modelCode := range models {
		fmt.Println(modelName, modelCode)
	}
}
