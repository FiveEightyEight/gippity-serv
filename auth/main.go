package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/FiveEightyEight/gippity-serv/db"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
}

type AuthService struct {
	DB *db.Database
}

// HashString hashes a given string using SHA-256 and a salt from the environment
func (as *AuthService) HashString(input string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}

	salt := os.Getenv("HASH_SALT")
	if salt == "" {
		return "", errors.New("HASH_SALT is not set in the environment")
	}

	saltedInput := input + salt
	hasher := sha256.New()
	hasher.Write([]byte(saltedInput))
	hashedBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashedBytes), nil
}

// GenerateToken generates a simple token (to be replaced with JWT in the future)
func (as *AuthService) GenerateToken(userID int) (string, error) {
	token := fmt.Sprintf("%d:%d", userID, time.Now().Unix())
	return as.HashString(token)
}

// ValidateToken validates the given token (to be replaced with JWT in the future)
func (as *AuthService) ValidateToken(token string) (bool, error) {
	// In this simple implementation, we just check if the token is a valid hash
	_, err := hex.DecodeString(token)
	return err == nil, nil
}

// CreateUser creates a new user in the database
func (as *AuthService) CreateUser(username, email, password string) (string, error) {
	// Check if username or email already exists
	var count int
	err := as.DB.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2", username, email).Scan(&count)
	if err != nil {
		log.Printf("Database error [cu-100] when checking existing user: %v", err)
		return "", errors.New("an error occurred while creating the user [cu-100]")
	}
	if count > 0 {
		return "", errors.New("username or email already exists [cu-101]")
	}

	// Hash the password
	hashedPassword, err := as.HashString(password)
	if err != nil {
		return "", errors.New("an error occurred while creating the user [cu-102]")
	}

	// Insert the new user
	var userID int
	err = as.DB.Pool.QueryRow(context.Background(), "INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id", username, email, hashedPassword).Scan(&userID)
	if err != nil {
		log.Printf("Database error [cu-103] when inserting new user: %v", err)
		return "", errors.New("an error occurred while creating the user [cu-103]")
	}

	// Generate and return a login token
	token, err := as.GenerateToken(userID)
	if err != nil {
		return "", errors.New("an error occurred while creating the user [cu-104]")
	}

	return token, nil
}

// Login handles user login and returns an auth token
func (as *AuthService) Login(username, password string) (string, error) {
	var user User
	err := as.DB.Pool.QueryRow(context.Background(), "SELECT id, password_hash FROM users WHERE username = $1", username).Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		log.Printf("Database error [ln-200] during login: %v", err)
		return "", errors.New("invalid login [ln-200]")
	}

	hashedPassword, err := as.HashString(password)
	if err != nil {
		return "", errors.New("an error occurred while logging in [ln-201]")
	}

	if hashedPassword != user.PasswordHash {
		return "", errors.New("invalid login [ln-202]")
	}

	return as.GenerateToken(user.ID)
}

// AuthMiddleware is an Echo middleware to validate auth tokens
func (as *AuthService) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing auth token [am-300]")
		}

		valid, err := as.ValidateToken(token)
		if err != nil || !valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid auth token [am-301]")
		}

		return next(c)
	}
}
