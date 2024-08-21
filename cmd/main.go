package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// "github.com/FiveEightyEight/gippity-serv/api"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

var models = map[string]string{
	"GPT-4o":        "gpt-4o",
	"GPT-4o mini":   "gpt-4o-mini",
	"GPT-4 Turbo":   "gpt-4-turbo",
	"GPT-3.5 Turbo": "gpt-3.5-turbo-0125",
}

var modelsResponse = []string{"GPT-4o", "GPT-4o mini", "GPT-4 Turbo", "GPT-3.5 Turbo"}

type ModelResponse struct {
	Models []string `json:"models"`
}

func createModelResponse() ModelResponse {
	return ModelResponse{
		Models: modelsResponse,
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatBody struct {
	Model string `json:"model"`
	// Messages []Message `json:"messages"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

type ChatResponse struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
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

func handleChatCompletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var chatBody ChatBody
	err := json.NewDecoder(r.Body).Decode(&chatBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	requestedModel, exists := models[chatBody.Model]
	if !exists {
		http.Error(w, "Invalid Model", http.StatusBadRequest)
		return
	}

	apiKey := getEnvKey()
	client := openai.NewClient(apiKey)

	// Create a channel to receive the response
	responseChan := make(chan ChatResponse)

	// Start a goroutine to handle the API call
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		resp, err := client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:    requestedModel,
				Messages: chatBody.Messages,
			},
		)

		if err != nil {
			responseChan <- ChatResponse{Error: fmt.Sprintf("ChatCompletion error: %v", err)}
			return
		}

		responseChan <- ChatResponse{Content: resp.Choices[0].Message.Content}
	}()

	// Wait for the response or timeout
	select {
	case response := <-responseChan:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case <-time.After(35 * time.Second):
		http.Error(w, "Request timeout", http.StatusGatewayTimeout)
	}
}

func getAllModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	response := createModelResponse()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Println("Successfully sent response for /models")
}

func homePath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, "Welcome home")

}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", homePath)
	mux.HandleFunc("POST /chat", handleChatCompletion)
	mux.HandleFunc("GET /models", getAllModels)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
