package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	IsActive     bool       `json:"is_active"`
	LastChatID   *uuid.UUID `json:"last_chat_id,omitempty"`
}

type UserMetadata struct {
	UserID            uuid.UUID `json:"user_id"`
	PreferredLanguage string    `json:"preferred_language"`
	Timezone          string    `json:"timezone"`
	Interests         []string  `json:"interests"`
	Profession        string    `json:"profession"`
	EducationLevel    string    `json:"education_level"`
	BirthYear         int       `json:"birth_year"`
	Country           string    `json:"country"`
	LastUpdated       time.Time `json:"last_updated"`
}

type Chat struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Title          string    `json:"title"`
	CreatedAt      time.Time `json:"created_at"`
	LastUpdated    time.Time `json:"last_updated"`
	IsArchived     bool      `json:"is_archived"`
	AIModelVersion string    `json:"ai_model_version"`
}

type Message struct {
	ID        uuid.UUID `json:"id"`
	ChatID    uuid.UUID `json:"chat_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	IsEdited  bool      `json:"is_edited"`
}

type AIModel struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
}

type ChatAIModel struct {
	ChatID    uuid.UUID `json:"chat_id"`
	AIModelID uuid.UUID `json:"ai_model_id"`
}

type UserPreferences struct {
	UserID               uuid.UUID `json:"user_id"`
	DefaultAIModel       uuid.UUID `json:"default_ai_model"`
	Theme                string    `json:"theme"`
	MessageDisplayCount  int       `json:"message_display_count"`
	NotificationsEnabled bool      `json:"notifications_enabled"`
}

type MessageContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
