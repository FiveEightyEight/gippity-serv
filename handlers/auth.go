package handlers

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/FiveEightyEight/gippity-serv/db"
	"github.com/FiveEightyEight/gippity-serv/utils"
	"github.com/labstack/echo/v4"
)

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

		hashedPassword, err := utils.HashString(password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while logging in"})
		}

		if hashedPassword != user.PasswordHash {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

		token, err := utils.GenerateToken(user.ID.String())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while logging in"})
		}

		return c.JSON(http.StatusOK, map[string]string{"token": token})
	}
}

// AuthMiddleware to validate auth tokens
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || len(strings.Split(authHeader, " ")) != 2 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header"})
		}
		token := strings.Split(authHeader, " ")[1]
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing auth token [am-300]")
		}

		valid, err := utils.ValidateToken(token)
		if err != nil || !valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid auth token [am-301]")
		}

		return next(c)
	}
}
