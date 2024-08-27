package handlers

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/FiveEightyEight/gippity-serv/auth"
	"github.com/FiveEightyEight/gippity-serv/db"
	"github.com/FiveEightyEight/gippity-serv/utils"
	"github.com/labstack/echo/v4"
)

const (
	accessTokenCookieName  = "t"
	refreshTokenCookieName = "mt"
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
			log.Printf("Error decoding credentials: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials encoding"})
		}
		credentials := strings.Split(string(decodedCreds), ":")
		if len(credentials) != 2 {
			log.Printf("Invalid credentials format: expected 2 parts, got %d", len(credentials))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials format"})
		}
		username, password := credentials[0], credentials[1]

		user, err := repo.GetUserByUsername(c.Request().Context(), username)
		if err != nil {
			log.Printf("Error getting user by username: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

		hashedPassword, err := utils.HashString(password)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while logging in"})
		}

		if hashedPassword != user.PasswordHash {
			log.Printf("Password mismatch for user: %s", username)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

		accessToken, err := auth.GenerateAccessToken(user.ID.String())
		if err != nil {
			log.Printf("Error generating access token: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while logging in"})
		}

		refreshToken, err := auth.GenerateRefreshToken(user.ID.String())
		if err != nil {
			log.Printf("Error generating refresh token: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while logging in"})
		}

		// Set refresh token as HTTP-only cookie
		c.SetCookie(&http.Cookie{
			Name:     refreshTokenCookieName,
			Value:    refreshToken,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
			HttpOnly: true,
			Secure:   true, // Ensure this is true in production (over HTTPS)
			SameSite: http.SameSiteStrictMode,
		})

		return c.JSON(http.StatusOK, map[string]string{
			"t": accessToken,
		})
	}
}

// AuthMiddleware to validate auth tokens
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || len(strings.Split(authHeader, " ")) != 2 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header"})
		}
		tokenString := strings.Split(authHeader, " ")[1]
		if tokenString == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing auth token")
		}

		claims, err := auth.ValidateToken(tokenString, false)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid auth token")
		}

		c.Set("userID", claims.UserID)
		return next(c)
	}
}

func RefreshToken(c echo.Context) error {
	cookie, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Refresh token cookie is missing"})
	}
	refreshToken := cookie.Value

	newAccessToken, newRefreshToken, err := auth.RefreshTokens(refreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid refresh token"})
	}

	c.SetCookie(&http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    newRefreshToken,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,                 // Ensure this is true in production (over HTTPS)
		SameSite: http.SameSiteNoneMode, // Ensure this is true in production (over HTTPS)
	})

	return c.JSON(http.StatusOK, map[string]string{
		"t": newAccessToken,
	})
}
