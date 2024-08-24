package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/FiveEightyEight/gippity-serv/auth"
	"github.com/FiveEightyEight/gippity-serv/db"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	openai "github.com/sashabaranov/go-openai"
)

type App struct {
	DB *db.Database
}

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
	// Initialize database connection
	db, err := db.NewDatabaseConnection()
	authService := &auth.AuthService{DB: db}
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())

	// Update the CORS middleware configuration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.GET("/", homePath)
	e.POST("/create_account", func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || len(strings.Split(authHeader, " ")) != 2 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid authorization header"})
		}
		encodedCreds := strings.Split(authHeader, " ")[1]
		decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid credentials encoding"})
		}
		credentials := strings.Split(string(decodedCreds), ":")
		if len(credentials) != 3 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid credentials format"})
		}
		username, email, password := credentials[0], credentials[1], credentials[2]

		if username == "" || email == "" || password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username, email, and password are required"})
		}

		token, err := authService.CreateUser(username, email, password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusCreated, map[string]string{"token": token})
	})
	e.POST("/login", func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || len(strings.Split(authHeader, " ")) != 2 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header"})
		}
		encodedCreds := strings.Split(authHeader, " ")[1]
		decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials encoding"})
		}
		credentials := strings.Split(string(decodedCreds), ":")
		if len(credentials) != 2 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials format"})
		}
		username, password := credentials[0], credentials[1]
		token, err := authService.Login(username, password)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		return c.JSON(http.StatusOK, map[string]string{"token": token})
	})

	// Routes that require authentication
	authGroup := e.Group("")
	authGroup.Use(authService.AuthMiddleware)
	authGroup.POST("/chat", handleChatCompletion)
	authGroup.GET("/models", getAllModels)

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	e.Logger.Fatal(e.Start(port))
	log.Printf("Server is running on port %s\n", port)
}
