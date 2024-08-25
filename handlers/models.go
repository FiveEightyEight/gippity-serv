package handlers

import (
	"net/http"

	"github.com/FiveEightyEight/gippity-serv/db"
	"github.com/labstack/echo/v4"
)

// GetAllAIModels returns a handler function to fetch all AI models
func GetAllAIModels(repo *db.PostgresRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		models, err := repo.GetAllAIModels(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch AI models")
		}

		return c.JSON(http.StatusOK, models)
	}
}
