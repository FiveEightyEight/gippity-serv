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

func loadTZLocation() *time.Location {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	locationENV := os.Getenv("LOCATION")
	location, err := time.LoadLocation(locationENV)
	if err != nil {
		log.Fatalf("Error loading location %s", err)
	}
	return location
}

func getUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userID := c.Get("userID")
	if userID == nil {
		return uuid.Nil, errors.New("userID not found in context")
	}
	userIDString, ok := userID.(string)
	if !ok {
		return uuid.Nil, errors.New("userID is not a string")
	}
	userUUID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, errors.New("userID is not a valid UUID")
	}
	return userUUID, nil
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
		userID, err := getUserIDFromContext(c)
		if err != nil {
			log.Println("Failed to get userID from context [c-001]", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized [c-001]"})
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
		timeLocation := loadTZLocation()
		// If no chat ID, create a new chat
		if rawPayload["chat_id"] == "" {
			isNewChat = true
			currentTime := time.Now().In(timeLocation)
			newChat := &models.Chat{
				UserID:         userID,
				Title:          rawPayload["content"].(string)[:min(50, len(rawPayload["content"].(string)))],
				CreatedAt:      currentTime,
				LastUpdated:    currentTime,
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
			rawChatID := rawPayload["chat_id"]
			chatIDString, ok := rawChatID.(string)
			if !ok {
				log.Println("Failed to get chat_id as string [c-004]")
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid chat_id format [c-004]"})
			}
			chatIDParsed, err := uuid.Parse(chatIDString)
			if err != nil {
				log.Println("Failed to parse chat_id [c-005]", err)
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid chat_id format [c-005]"})
			}
			chat, err := repo.GetChatByID(c.Request().Context(), chatIDParsed)
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
			ChatID:    chatID,
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
		c.Response().Header().Set("X-Chat-Id", chatID.String())
		c.Response().Header().Set("Access-Control-Expose-Headers", "X-Chat-Id")
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
		lastUpdated := time.Now().In(timeLocation)
		assistantMessage := &models.Message{
			ChatID:    chatID,
			UserID:    userID,
			Content:   assistantResponse,
			Role:      "assistant",
			CreatedAt: lastUpdated,
		}
		err = repo.CreateMessage(c.Request().Context(), assistantMessage)
		if err != nil {
			log.Println("Failed to save assistant message [c-9]", err)
			// Note: We don't return an error here because the response has already been sent to the client
		}

		// Update the Chat's LastUpdated value
		chat, err := repo.GetChatByID(c.Request().Context(), chatID)
		if err != nil {
			log.Println("Failed to get chat for updating LastUpdated [c-10]", err)
		} else {
			chat.LastUpdated = lastUpdated
			err = repo.UpdateChat(c.Request().Context(), chat)
			if err != nil {
				log.Println("Failed to update chat's LastUpdated [c-11]", err)
			}
		}

		return nil
	}
}

func GetConversation(repo *db.PostgresRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		chatID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			log.Println("Invalid chat ID [gc-001]", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid chat ID [gc-001]"})
		}

		userID := c.Get("userID").(string)
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			log.Println("Invalid user ID [gc-002]", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID [gc-002]"})
		}

		chat, err := repo.GetChatByID(c.Request().Context(), chatID)
		if err != nil {
			log.Println("Failed to get chat [gc-003]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [gc-003]"})
		}

		if chat.UserID != userUUID {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied [gc-004]"})
		}

		messages, err := repo.GetMessagesByChatID(c.Request().Context(), chatID)
		if err != nil {
			log.Println("Failed to get messages [gc-005]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [gc-005]"})
		}

		return c.JSON(http.StatusOK, messages)
	}
}

func GetChatHistory(repo *db.PostgresRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := getUserIDFromContext(c)
		if err != nil {
			log.Println("Failed to get userID from context [gch-001]", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized [gch-001]"})
		}
		sortByDate := true
		chats, err := repo.GetChatsByUserID(c.Request().Context(), userID, sortByDate)
		if err != nil {
			log.Println("Failed to get chat history [gch-002]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [gch-002]"})
		}

		return c.JSON(http.StatusOK, chats)
	}
}

func DeleteChat(repo *db.PostgresRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		chatID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			log.Println("Invalid chat ID [dc-001]", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid chat ID [dc-001]"})
		}

		userID, err := getUserIDFromContext(c)
		if err != nil {
			log.Println("Failed to get userID from context [dc-002]", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized [dc-002]"})
		}

		chat, err := repo.GetChatByID(c.Request().Context(), chatID)
		if err != nil {
			log.Println("Failed to get chat [dc-003]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [dc-003]"})
		}

		if chat.UserID != userID {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied [dc-004]"})
		}

		err = repo.DeleteChat(c.Request().Context(), chatID)
		if err != nil {
			log.Println("Failed to delete chat [dc-005]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error [dc-005]"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Chat deleted successfully"})
	}
}
