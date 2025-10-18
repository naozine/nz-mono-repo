package main

import (
	"log"
	"os"

	"github.com/hnao/nz-mono-repo/apps/whiscript/internal/db"
	"github.com/hnao/nz-mono-repo/apps/whiscript/internal/handler"
	"github.com/hnao/nz-mono-repo/apps/whiscript/internal/repository"
	"github.com/hnao/nz-mono-repo/apps/whiscript/internal/service"
	"github.com/hnao/nz-mono-repo/apps/whiscript/internal/ui"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Database setup
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./whiscript.db"
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Run migrations (goose requires *sql.DB)
	if err := db.RunMigrations(database.DB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Load templates
	tmpl, err := ui.Templates()
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Initialize layers
	projectRepo := repository.NewProjectRepository(database)
	projectService := service.NewProjectService(projectRepo)
	projectHandler := handler.NewProjectHandler(projectService, tmpl)

	// Echo setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", projectHandler.Index)
	e.POST("/projects", projectHandler.Create)
	e.PUT("/projects/:id", projectHandler.Update)
	e.DELETE("/projects/:id", projectHandler.Delete)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
