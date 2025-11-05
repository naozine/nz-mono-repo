package handler

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/whiscript/internal/service"
	"github.com/yourusername/whiscript/internal/ui"
)

// AudioHandler handles HTTP requests for audio files
type AudioHandler struct {
	projectService *service.ProjectService
	audioService   *service.AudioFileService
	corpusService  *service.CorpusService
	templates      *template.Template
}

// NewAudioHandler creates a new audio handler
func NewAudioHandler(projectService *service.ProjectService, audioService *service.AudioFileService, corpusService *service.CorpusService) (*AudioHandler, error) {
	// Create template with helper functions
	funcMap := template.FuncMap{
		"add":     func(a, b int) int { return a + b },
		"sub":     func(a, b int) int { return a - b },
		"eq":      func(a, b string) bool { return a == b },
		"div":     func(a, b float64) float64 { return a / b },
		"float64": func(i int64) float64 { return float64(i) },
		"deref": func(p *float64) float64 {
			if p == nil {
				return 0
			}
			return *p
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(ui.TemplatesFS, "templates/*.html", "templates/projects/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &AudioHandler{
		projectService: projectService,
		audioService:   audioService,
		corpusService:  corpusService,
		templates:      tmpl,
	}, nil
}

// Detail handles GET /projects/:id
func (h *AudioHandler) Detail(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid project ID")
	}

	project, err := h.projectService.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Project not found")
	}

	audioFiles, err := h.audioService.ListByProjectID(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load audio files")
	}

	corpusFiles, err := h.corpusService.ListFilesByProjectID(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load corpus files")
	}

	data := map[string]interface{}{
		"Project":     project,
		"AudioFiles":  audioFiles,
		"CorpusFiles": corpusFiles,
	}

	return h.renderTemplate(c, "projects/detail.html", data)
}

// Upload handles POST /projects/:id/audio
func (h *AudioHandler) Upload(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid project ID")
	}

	// Get uploaded file
	file, err := c.FormFile("audio_file")
	if err != nil {
		return c.String(http.StatusBadRequest, "No file uploaded")
	}

	// Upload file
	_, err = h.audioService.Upload(projectID, file)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Return updated audio files list
	audioFiles, err := h.audioService.ListByProjectID(projectID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load audio files")
	}

	return h.renderTemplate(c, "projects/_audio_list.html", audioFiles)
}

// Delete handles DELETE /projects/audio/:id
func (h *AudioHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid audio file ID")
	}

	if err := h.audioService.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete audio file")
	}

	return c.NoContent(http.StatusOK)
}

// Serve handles GET /uploads/:id
func (h *AudioHandler) Serve(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid audio file ID")
	}

	audioFile, err := h.audioService.GetByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Audio file not found")
	}

	// Open file
	file, err := os.Open(audioFile.FilePath)
	if err != nil {
		return c.String(http.StatusNotFound, "File not found")
	}
	defer file.Close()

	// Set content type
	c.Response().Header().Set("Content-Type", audioFile.MimeType)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", audioFile.OriginalFilename))

	// Stream file
	if _, err := io.Copy(c.Response().Writer, file); err != nil {
		return err
	}

	return nil
}

// renderTemplate renders a template with the given data
func (h *AudioHandler) renderTemplate(c echo.Context, name string, data interface{}) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)
	return h.templates.ExecuteTemplate(c.Response().Writer, name, data)
}
