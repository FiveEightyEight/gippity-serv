package main

import (
	"log"
	"net/http"
	"os"

	"github.com/FiveEightyEight/gippity-serv/db"
	"github.com/FiveEightyEight/gippity-serv/handlers"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func homePath(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome home")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	patternsStorage := &db.Storage{
		Label:         "Patterns",
		Dir:           "./patterns", // Adjust this path as needed
		ItemIsDir:     true,
		FileExtension: "",
	}
	if err := patternsStorage.Configure(); err != nil {
		log.Fatalf("Error configuring patterns storage: %v", err)
	}
	// Initialize database connection
	db, err := db.NewDatabaseConnection()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())

	// Update the CORS middleware configuration when deploying to prod
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:4321"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	e.GET("/", homePath)
	e.POST("/create_account", handlers.CreateUser(db))
	e.POST("/login", handlers.Login(db))
	e.POST("/register", handlers.CreateUser(db))
	e.POST("/refresh", handlers.RefreshToken)

	// Protected routes
	authGroup := e.Group("/api/v1")
	authGroup.Use(handlers.AuthMiddleware)
	authGroup.GET("/models", handlers.GetAllAIModels(db))
	authGroup.GET("/patterns", handlers.GetPatterns(patternsStorage))
	authGroup.GET("/chat", handlers.GetConversation(db))
	authGroup.POST("/conversation", handlers.Conversation(db))
	authGroup.GET("/chat-history", handlers.GetChatHistory(db))
	authGroup.DELETE("/chat/:id", handlers.DeleteChat(db))
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	e.Logger.Fatal(e.Start(port))
	log.Printf("Server is running on port %s\n", port)
}
