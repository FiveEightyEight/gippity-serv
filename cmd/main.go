package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func handleChatCompletion(c echo.Context) error {
	var chatBody ChatBody
	err := json.NewDecoder(c.Request().Body).Decode(&chatBody)
	if err != nil {
		return errors.New("invalid request body")
	}

	requestedModel, exists := models[chatBody.Model]
	if !exists {
		return errors.New("invalid model")
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
		return c.JSON(http.StatusCreated, response)
	case <-time.After(35 * time.Second):
		return errors.New("chat completion request timeout")
	}
}

func getAllModels(c echo.Context) error {
	response := createModelResponse()
	log.Println("Successfully sent response for /models")
	return c.JSON(http.StatusCreated, response)
}

func homePath(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome home")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/", homePath)
	e.POST("/chat", handleChatCompletion)
	e.GET("/models", getAllModels)

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	e.Logger.Fatal(e.Start(port))
	log.Printf("Server is running on port %s\n", port)
}
