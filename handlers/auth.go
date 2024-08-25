package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/FiveEightyEight/gippity-serv/db"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func HashString(input string) (string, error) {
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

// Token generations and validation to be replaced with JWT in the future (oh no is this tech debt that I'll never get to?)
func GenerateToken(userID int) (string, error) {
	token := fmt.Sprintf("%d:%d", userID, time.Now().Unix())
	return HashString(token)
}

func ValidateToken(token string) (bool, error) {
	if _, err := hex.DecodeString(token); err != nil {
		return false, err
	}
	return true, nil
}

func Login(repo *db.PostgresRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		user, err := repo.GetUserByUsername(c.Request().Context(), username)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

		hashedPassword, err := HashString(password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while logging in"})
		}

		if hashedPassword != user.PasswordHash {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

		token, err := GenerateToken(user.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while logging in"})
		}

		return c.JSON(http.StatusOK, map[string]string{"token": token})
	}
}

// AuthMiddleware to validate auth tokens
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing auth token [am-300]")
		}

		valid, err := ValidateToken(token)
		if err != nil || !valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid auth token [am-301]")
		}

		return next(c)
	}
}
