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
		// Get user ID from JWT

		// tokenString := c.Request().Header.Get("Authorization")
		// claims, err := auth.ValidateToken(tokenString, false)
		// if err != nil {
		// 	log.Println("Failed to validate token [c-1]", err)
		// 	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token [c-1]"})
		// }
		// userID, err := uuid.Parse(claims.UserID)
		// if err != nil {
		// 	log.Println("Failed to parse user ID [c-2]", err)
		// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-2]"})
		// }
		unparsedUserID := c.Get("userID")
		log.Printf("Unparsed User ID: %v", unparsedUserID)
		if unparsedUserID == nil {
			log.Println("Failed to get userID from context [c-001]")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized [c-001]"})
		}
		userIDString, ok := unparsedUserID.(string)
		if !ok {
			log.Println("Failed to convert userID to string [c-003]")
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-003]"})
		}
		userID, err := uuid.Parse(userIDString)
		if err != nil {
			log.Println("Failed to parse userID [c-002]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-002]"})
		}

		// Bind the raw payload first to check for invalid fields
		var rawPayload map[string]interface{}
		if err := c.Bind(&rawPayload); err != nil {
			log.Println("Failed to bind message [c-000]", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body [c-000]"})
		}

		if rawPayload["content"] == "" {
			log.Println("Content is required [c-0000]")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Content is required [c-0000]"})
		}

		var isNewChat bool
		var chatID uuid.UUID
		var aiModelVersion string
		messages := []models.MessageContent{}
		// If no chat ID, create a new chat
		if rawPayload["chat_id"] == "" {
			isNewChat = true
			newChat := &models.Chat{
				UserID:         userID,
				Title:          rawPayload["content"].(string)[:min(50, len(rawPayload["content"].(string)))],
				CreatedAt:      time.Now(),
				LastUpdated:    time.Now(),
				IsArchived:     false,
				AIModelVersion: rawPayload["ai_model_version"].(string),
			}

			createdChat, err := repo.CreateChat(c.Request().Context(), newChat)
			if err != nil {
				log.Println("Failed to create new chat [c-3]", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-3]"})
			}
			aiModelVersion = createdChat.AIModelVersion
			rawPayload["chat_id"] = createdChat.ID
			chatID = createdChat.ID
		} else {
			chat, err := repo.GetChatByID(c.Request().Context(), rawPayload["chat_id"].(uuid.UUID))
			if err != nil {
				log.Println("Failed to get chat [c-3]", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-3]"})
			}
			aiModelVersion = chat.AIModelVersion
			chatID = chat.ID
		}

		// Parse the created_at string into a time.Time object
		createdAtStr, ok := rawPayload["created_at"].(string)
		if !ok {
			log.Println("Failed to get created_at as string [c-004]")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid created_at format [c-004]"})
		}

		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Println("Failed to parse created_at [c-005]", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid created_at format [c-005]"})
		}

		// Use the parsed time in the message struct
		// the user message
		message := models.Message{
			ChatID:    rawPayload["chat_id"].(uuid.UUID),
			Content:   rawPayload["content"].(string),
			UserID:    userID,
			Role:      "user",
			IsEdited:  rawPayload["is_edited"].(bool),
			CreatedAt: createdAt,
		}

		if isNewChat {
			messages = append(messages, models.MessageContent{
				Role:    "user",
				Content: message.Content,
			})
		} else {
			messages, err = repo.GetMessageContentsByChatID(c.Request().Context(), chatID)
			if err != nil {
				log.Println("Failed to get message contents [c-4]", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-4]"})
			}
			messages = append(messages, models.MessageContent{
				Role:    "user",
				Content: message.Content,
			})
		}

		err = repo.CreateMessage(c.Request().Context(), &message)
		if err != nil {
			log.Println("Failed to save message [c-5]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [c-5]"})
		}

		stream, err := ChatCompletionStream(c.Request().Context(), messages, aiModelVersion)
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
			ChatID:  chatID,
			UserID:  userID,
			Content: assistantResponse,
			Role:    "assistant",
		}
		err = repo.CreateMessage(c.Request().Context(), assistantMessage)
		if err != nil {
			log.Println("Failed to save assistant message [c-9]", err)
			// Note: We don't return an error here because the response has already been sent to the client
		}

		return nil
	}
}
