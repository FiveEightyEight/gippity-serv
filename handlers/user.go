package handlers

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/FiveEightyEight/gippity-serv/db"
	"github.com/FiveEightyEight/gippity-serv/models"
	"github.com/FiveEightyEight/gippity-serv/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func CreateUser(userRepo *db.PostgresRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		// Check if username or email already exists
		existingUser, err := userRepo.GetUserByUsername(context.Background(), username)
		if err == nil && existingUser != nil {
			return echo.NewHTTPError(http.StatusConflict, "username or email already exists [cu-101]")
		}

		hashedPassword, err := utils.HashString(password)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "an error occurred while creating the user [cu-102]")
		}

		user := &models.User{
			Username:     username,
			Email:        email,
			PasswordHash: hashedPassword,
		}

		if err := userRepo.CreateUser(context.Background(), user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user [cu-103]")
		}

		// Generate and return a login token
		token, err := utils.GenerateToken(user.ID.String())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "an error occurred while creating the user [cu-104]")
		}

		return c.JSON(http.StatusCreated, map[string]string{"token": token})
	}
}

func GetUser(userRepo *db.PostgresRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
		}

		user, err := userRepo.GetUserByID(context.Background(), id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}

		return c.JSON(http.StatusOK, user)
	}
}

func UpdateUser(userRepo *db.PostgresRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
		}

		user := new(models.User)
		if err := c.Bind(user); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
		}

		parsedID, err := uuid.Parse(id)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID format")
		}
		user.ID = parsedID

		if err := userRepo.UpdateUser(context.Background(), user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user")
		}

		return c.JSON(http.StatusOK, user)
	}
}

func DeleteUser(userRepo *db.PostgresRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
		}

		if err := userRepo.DeleteUser(context.Background(), id); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete user")
		}

		return c.NoContent(http.StatusNoContent)
	}
}
