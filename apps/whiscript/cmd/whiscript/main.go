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

	// Get upload path from environment or use default
	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		uploadPath = "./uploads"
	}

	// Initialize database and run migrations
	database, err := repository.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize repositories
	projectRepo := repository.NewProjectRepository(database)
	audioRepo := repository.NewAudioFileRepository(database)
	corpusRepo := repository.NewCorpusRepository(database)

	// Initialize services
	projectService := service.NewProjectService(projectRepo)
	audioService := service.NewAudioFileService(audioRepo, uploadPath)
	corpusService := service.NewCorpusService(corpusRepo, uploadPath)

	// Initialize handlers
	projectHandler, err := handler.NewProjectHandler(projectService)
	if err != nil {
		log.Fatalf("Failed to initialize project handler: %v", err)
	}

	audioHandler, err := handler.NewAudioHandler(projectService, audioService, corpusService)
	if err != nil {
		log.Fatalf("Failed to initialize audio handler: %v", err)
	}

	corpusHandler, err := handler.NewCorpusHandler(projectService, corpusService)
	if err != nil {
		log.Fatalf("Failed to initialize corpus handler: %v", err)
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

	// Project routes
	e.GET("/projects", projectHandler.Index)
	e.POST("/projects", projectHandler.Create)
	e.GET("/projects/:id", audioHandler.Detail)
	e.GET("/projects/:id/edit", projectHandler.Edit)
	e.POST("/projects/:id", projectHandler.Update)
	e.DELETE("/projects/:id", projectHandler.Delete)

	// Audio routes
	e.POST("/projects/:id/audio", audioHandler.Upload)
	e.DELETE("/projects/audio/:id", audioHandler.Delete)
	e.GET("/uploads/:id", audioHandler.Serve)

	// Corpus routes
	e.POST("/projects/:id/corpus", corpusHandler.Upload)
	e.DELETE("/projects/corpus/:id", corpusHandler.Delete)
	e.GET("/projects/corpus/:id/segments", corpusHandler.ViewSegments)
	e.GET("/projects/corpus/:id/editor", corpusHandler.ViewEditor)

	// Corpus group routes
	e.POST("/projects/:id/corpus-groups", corpusHandler.CreateGroup)
	e.GET("/projects/corpus-groups/:id/editor", corpusHandler.ViewGroupEditor)
	e.POST("/projects/corpus-groups/:id/refine", corpusHandler.RefineGroupSegments)

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
