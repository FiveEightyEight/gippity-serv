package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/FiveEightyEight/gippity-serv/models"
	"github.com/FiveEightyEight/gippity-serv/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewDatabaseConnection() (*PostgresRepository, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Close() {
	r.db.Close()
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	return nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, username, email, password_hash FROM users WHERE id = $1`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %v", err)
	}
	return user, nil
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash FROM users WHERE username = $1`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %v", err)
	}
	return user, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `UPDATE users SET username = $1, email = $2, password_hash = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, user.Username, user.Email, user.PasswordHash, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

func (r *PostgresRepository) CreateAIModel(ctx context.Context, model *models.AIModel) error {
	query := `INSERT INTO ai_models (name, version, description, is_active) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(ctx, query, model.Name, model.Version, model.Description, model.IsActive).Scan(&model.ID)
	if err != nil {
		return fmt.Errorf("failed to create AI model: %v", err)
	}
	return nil
}

func (r *PostgresRepository) GetAIModelByID(ctx context.Context, id int) (*models.AIModel, error) {
	query := `SELECT id, name, version, description, is_active FROM ai_models WHERE id = $1`
	model := &models.AIModel{}
	err := r.db.QueryRow(ctx, query, id).Scan(&model.ID, &model.Name, &model.Version, &model.Description, &model.IsActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI model by ID: %v", err)
	}
	return model, nil
}

func (r *PostgresRepository) GetAllAIModels(ctx context.Context) ([]*models.AIModel, error) {
	query := `SELECT id, name, version, description, is_active FROM ai_models`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all AI models: %v", err)
	}
	defer rows.Close()

	var list []*models.AIModel
	for rows.Next() {
		model := &models.AIModel{}
		if err := rows.Scan(&model.ID, &model.Name, &model.Version, &model.Description, &model.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan AI model: %v", err)
		}
		list = append(list, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over AI models: %v", err)
	}

	return list, nil
}

func (r *PostgresRepository) UpdateAIModel(ctx context.Context, model *models.AIModel) error {
	query := `UPDATE ai_models SET name = $1, version = $2, description = $3, is_active = $4 WHERE id = $5`
	_, err := r.db.Exec(ctx, query, model.Name, model.Version, model.Description, model.IsActive, model.ID)
	if err != nil {
		return fmt.Errorf("failed to update AI model: %v", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteAIModel(ctx context.Context, id int) error {
	query := `DELETE FROM ai_models WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete AI model: %v", err)
	}
	return nil
}

func (r *PostgresRepository) CreateChat(ctx context.Context, chat *models.Chat) (*models.Chat, error) {
	query := `INSERT INTO chats (user_id, title, created_at, last_updated, is_archived) 
              VALUES ($1, $2, $3, $4, $5) 
              RETURNING id`
	err := r.db.QueryRow(ctx, query,
		chat.UserID,
		chat.Title,
		chat.CreatedAt,
		chat.LastUpdated,
		chat.IsArchived).Scan(&chat.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %v", err)
	}
	return chat, nil
}

func (r *PostgresRepository) GetChatByID(ctx context.Context, id uuid.UUID) (*models.Chat, error) {
	query := `SELECT id, user_id, title, created_at, last_updated, is_archived 
              FROM chats 
              WHERE id = $1`
	chat := &models.Chat{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&chat.ID,
		&chat.UserID,
		&chat.Title,
		&chat.CreatedAt,
		&chat.LastUpdated,
		&chat.IsArchived)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat by ID: %v", err)
	}
	return chat, nil
}

func (r *PostgresRepository) UpdateChat(ctx context.Context, chat *models.Chat) error {
	query := `UPDATE chats 
              SET user_id = $1, title = $2, last_updated = $3, is_archived = $4 
              WHERE id = $5`
	_, err := r.db.Exec(ctx, query,
		chat.UserID,
		chat.Title,
		chat.LastUpdated,
		chat.IsArchived,
		chat.ID)
	if err != nil {
		return fmt.Errorf("failed to update chat: %v", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteChat(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM chats WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %v", err)
	}
	return nil
}

func (r *PostgresRepository) GetChatsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Chat, error) {
	query := `SELECT id, user_id, title, created_at, last_updated, is_archived 
              FROM chats 
              WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chats by user ID: %v", err)
	}
	defer rows.Close()

	var chats []*models.Chat
	for rows.Next() {
		chat := &models.Chat{}
		if err := rows.Scan(
			&chat.ID,
			&chat.UserID,
			&chat.Title,
			&chat.CreatedAt,
			&chat.LastUpdated,
			&chat.IsArchived); err != nil {
			return nil, fmt.Errorf("failed to scan chat: %v", err)
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over chats: %v", err)
	}

	return chats, nil
}

var _ repository.UserRepository = (*PostgresRepository)(nil)

// Implement other repository methods (UserMetadata, Chat, Message, ChatAIModel, UserPreferences) similarly...

// Ensure other interfaces are implemented...
