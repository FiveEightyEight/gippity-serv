package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
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

func GenerateToken(userID string) (string, error) {
	token := fmt.Sprintf("%s:%d", userID, time.Now().Unix())
	return HashString(token)
}

func ValidateToken(token string) (bool, error) {
	if _, err := hex.DecodeString(token); err != nil {
		return false, err
	}
	return true, nil
}
