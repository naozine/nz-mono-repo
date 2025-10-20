package main

import (
	"fmt"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yourusername/whiscript/internal/handler"
	"github.com/yourusername/whiscript/internal/repository"
	"github.com/yourusername/whiscript/internal/service"
)

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./whiscript.db"
	}

	// Initialize database and run migrations
	database, err := repository.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize repository, service, and handler
	projectRepo := repository.NewProjectRepository(database)
	projectService := service.NewProjectService(projectRepo)
	projectHandler, err := handler.NewProjectHandler(projectService)
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(302, "/projects")
	})

	e.GET("/projects", projectHandler.Index)
	e.POST("/projects", projectHandler.Create)
	e.GET("/projects/:id/edit", projectHandler.Edit)
	e.POST("/projects/:id", projectHandler.Update)
	e.DELETE("/projects/:id", projectHandler.Delete)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting server on :%s", port)
	if err := e.Start(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
