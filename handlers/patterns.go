package handlers

import (
	"log"
	"net/http"

	"github.com/FiveEightyEight/gippity-serv/db"
	"github.com/labstack/echo/v4"
)

func GetPatterns(storage *db.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		patterns, err := storage.GetNames()
		if err != nil {
			log.Println("[gp-001]", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve patterns [gp-001]"})
		}
		return c.JSON(http.StatusOK, patterns)
	}
}
