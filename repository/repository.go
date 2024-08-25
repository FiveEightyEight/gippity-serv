package repository

import (
	"context"

	"github.com/FiveEightyEight/gippity-serv/models"
)

// UserRepository defines the interface for user-related database operations
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id int) error
}

// UserMetadataRepository defines the interface for user metadata-related database operations
type UserMetadataRepository interface {
	CreateUserMetadata(ctx context.Context, metadata *models.UserMetadata) error
	GetUserMetadataByUserID(ctx context.Context, userID int) (*models.UserMetadata, error)
	UpdateUserMetadata(ctx context.Context, metadata *models.UserMetadata) error
	DeleteUserMetadata(ctx context.Context, userID int) error
}

// ChatRepository defines the interface for chat-related database operations
type ChatRepository interface {
	CreateChat(ctx context.Context, chat *models.Chat) error
	GetChatByID(ctx context.Context, id int) (*models.Chat, error)
	GetChatsByUserID(ctx context.Context, userID int) ([]*models.Chat, error)
	UpdateChat(ctx context.Context, chat *models.Chat) error
	DeleteChat(ctx context.Context, id int) error
}

// MessageRepository defines the interface for message-related database operations
type MessageRepository interface {
	CreateMessage(ctx context.Context, message *models.Message) error
	GetMessageByID(ctx context.Context, id int) (*models.Message, error)
	GetMessagesByChatID(ctx context.Context, chatID int) ([]*models.Message, error)
	UpdateMessage(ctx context.Context, message *models.Message) error
	DeleteMessage(ctx context.Context, id int) error
}

// AIModelRepository defines the interface for AI model-related database operations
type AIModelRepository interface {
	CreateAIModel(ctx context.Context, model *models.AIModel) error
	GetAIModelByID(ctx context.Context, id int) (*models.AIModel, error)
	GetAllActiveAIModels(ctx context.Context) ([]*models.AIModel, error)
	UpdateAIModel(ctx context.Context, model *models.AIModel) error
	DeleteAIModel(ctx context.Context, id int) error
}

// ChatAIModelRepository defines the interface for chat-AI model association-related database operations
type ChatAIModelRepository interface {
	AssociateChatWithAIModel(ctx context.Context, chatID, aiModelID int) error
	GetAIModelsByChatID(ctx context.Context, chatID int) ([]*models.AIModel, error)
	RemoveChatAIModelAssociation(ctx context.Context, chatID, aiModelID int) error
}

// UserPreferencesRepository defines the interface for user preferences-related database operations
type UserPreferencesRepository interface {
	CreateUserPreferences(ctx context.Context, preferences *models.UserPreferences) error
	GetUserPreferencesByUserID(ctx context.Context, userID int) (*models.UserPreferences, error)
	UpdateUserPreferences(ctx context.Context, preferences *models.UserPreferences) error
	DeleteUserPreferences(ctx context.Context, userID int) error
}
