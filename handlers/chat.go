package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/FiveEightyEight/gippity-serv/auth"
	"github.com/FiveEightyEight/gippity-serv/db"
	"github.com/FiveEightyEight/gippity-serv/models"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	openai "github.com/sashabaranov/go-openai"
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

func ChatCompletionStream(ctx context.Context, messages []models.MessageContent, aiModelVersion string) (*openai.ChatCompletionStream, error) {
	c := openai.NewClient(getEnvKey())

	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: openaiMessages,
		Stream:   true,
	}

	if aiModelVersion != "" {
		req.Model = aiModelVersion
	}

	return c.CreateChatCompletionStream(ctx, req)
}

func Conversation(repo *db.PostgresRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		var message models.Message
		if err := c.Bind(&message); err != nil {
			log.Println("Failed to bind message [c-00]", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body [c-00]"})
		}
		if message.Content == "" {
			log.Println("Content is required [c-0]")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Content is required [c-0]"})
		}

		// Get user ID from JWT
		tokenString := c.Request().Header.Get("Authorization")
		claims, err := auth.ValidateToken(tokenString, false)
		if err != nil {
			log.Println("Failed to validate token [c-1]", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token [c-1]"})
		}
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			log.Println("Failed to parse user ID [c-2]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-2]"})
		}
		message.UserID = userID
		messages := []models.MessageContent{}
		// If no chat ID, create a new chat
		isNewChat := message.ChatID == uuid.Nil
		if isNewChat {
			newChat := &models.Chat{
				UserID:         userID,
				Title:          message.Content[:min(50, len(message.Content))],
				CreatedAt:      time.Now(),
				LastUpdated:    time.Now(),
				IsArchived:     false,
				AIModelVersion: *message.AIModelVersion,
			}
			createdChat, err := repo.CreateChat(c.Request().Context(), newChat)
			if err != nil {
				log.Println("Failed to create new chat [c-3]", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-3]"})
			}
			message.ChatID = createdChat.ID
			messages = append(messages, models.MessageContent{
				Role:    "user",
				Content: message.Content,
			})
		} else {
			chat, err := repo.GetChatByID(c.Request().Context(), message.ChatID)
			if err != nil {
				log.Println("Failed to get chat [c-3]", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-3]"})
			}
			message.AIModelVersion = &chat.AIModelVersion
			messages, err = repo.GetMessageContentsByChatID(c.Request().Context(), message.ChatID)
			if err != nil {
				log.Println("Failed to get message contents [c-4]", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-4]"})
			}
			messages = append(messages, models.MessageContent{
				Role:    "user",
				Content: message.Content,
			})
		}

		// Insert message into database
		err = repo.CreateMessage(c.Request().Context(), &message)
		if err != nil {
			log.Println("Failed to save message [c-5]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-5]"})
		}

		stream, err := ChatCompletionStream(c.Request().Context(), messages, *message.AIModelVersion)
		if err != nil {
			log.Println("Failed to create chat completion stream [c-6]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-6]"})
		}
		defer stream.Close()

		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().WriteHeader(http.StatusOK)

		assistantResponse := ""
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				log.Println("Stream error [c-7]:", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-7]"})
			}

			content := response.Choices[0].Delta.Content
			assistantResponse += content
			_, err = c.Response().Write([]byte(content))
			if err != nil {
				log.Println("Failed to write response [c-8]:", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-8]"})
			}
			c.Response().Flush()
		}

		// Save assistant's response as a new message
		assistantMessage := &models.Message{
			ChatID:         message.ChatID,
			UserID:         message.UserID,
			Content:        assistantResponse,
			Role:           "assistant",
			AIModelVersion: message.AIModelVersion,
		}
		err = repo.CreateMessage(c.Request().Context(), assistantMessage)
		if err != nil {
			log.Println("Failed to save assistant message [c-9]", err)
			// Note: We don't return an error here because the response has already been sent to the client
		}

		return nil
	}
}
