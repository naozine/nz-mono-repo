package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/whiscript/internal/service"
	"github.com/yourusername/whiscript/internal/ui"
)

// CorpusHandler handles HTTP requests for corpus files
type CorpusHandler struct {
	projectService *service.ProjectService
	corpusService  *service.CorpusService
	templates      *template.Template
}

// NewCorpusHandler creates a new corpus handler
func NewCorpusHandler(projectService *service.ProjectService, corpusService *service.CorpusService) (*CorpusHandler, error) {
	// Create template with helper functions
	funcMap := template.FuncMap{
		"add":     func(a, b int) int { return a + b },
		"sub":     func(a, b int) int { return a - b },
		"eq":      func(a, b string) bool { return a == b },
		"div":     func(a, b float64) float64 { return a / b },
		"float64": func(i int64) float64 { return float64(i) },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(ui.TemplatesFS, "templates/*.html", "templates/projects/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &CorpusHandler{
		projectService: projectService,
		corpusService:  corpusService,
		templates:      tmpl,
	}, nil
}

// Upload handles POST /projects/:id/corpus
func (h *CorpusHandler) Upload(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid project ID")
	}

	// Get audio_file_id if provided (optional)
	var audioFileID *int64
	if audioFileIDStr := c.FormValue("audio_file_id"); audioFileIDStr != "" {
		id, err := strconv.ParseInt(audioFileIDStr, 10, 64)
		if err == nil {
			audioFileID = &id
		}
	}

	// Get uploaded file
	file, err := c.FormFile("corpus_file")
	if err != nil {
		return c.String(http.StatusBadRequest, "No file uploaded")
	}

	// Upload file
	_, err = h.corpusService.Upload(projectID, audioFileID, file)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Return updated corpus files list
	corpusFiles, err := h.corpusService.ListFilesByProjectID(projectID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load corpus files")
	}

	return h.renderTemplate(c, "projects/_corpus_list.html", corpusFiles)
}

// Delete handles DELETE /projects/corpus/:id
func (h *CorpusHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid corpus file ID")
	}

	if err := h.corpusService.DeleteFile(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete corpus file")
	}

	return c.NoContent(http.StatusOK)
}

// ViewSegments handles GET /projects/corpus/:id/segments
func (h *CorpusHandler) ViewSegments(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid corpus file ID")
	}

	corpusFile, err := h.corpusService.GetFileByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Corpus file not found")
	}

	segments, err := h.corpusService.GetSegmentsByCorpusFileID(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load segments")
	}

	data := map[string]interface{}{
		"CorpusFile": corpusFile,
		"Segments":   segments,
	}

	return h.renderTemplate(c, "projects/corpus_segments.html", data)
}

// ViewEditor handles GET /projects/corpus/:id/editor
func (h *CorpusHandler) ViewEditor(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid corpus file ID")
	}

	corpusFile, err := h.corpusService.GetFileByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Corpus file not found")
	}

	// Check if audio file is associated
	if corpusFile.AudioFileID == nil {
		return c.String(http.StatusBadRequest, "No audio file associated with this corpus")
	}

	// Get gap threshold from query parameter (default: 2.0 seconds)
	gapThreshold := 2.0
	if thresholdStr := c.QueryParam("gap_threshold"); thresholdStr != "" {
		if threshold, err := strconv.ParseFloat(thresholdStr, 64); err == nil && threshold > 0 {
			gapThreshold = threshold
		}
	}

	// Get project to retrieve project info
	project, err := h.projectService.GetByID(corpusFile.ProjectID)
	if err != nil {
		return c.String(http.StatusNotFound, "Project not found")
	}

	// Get segments with gap information
	segmentsWithGaps, err := h.corpusService.GetSegmentsWithGaps(id, gapThreshold)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load segments")
	}

	data := map[string]interface{}{
		"Project":          project,
		"CorpusFile":       corpusFile,
		"SegmentsWithGaps": segmentsWithGaps,
		"GapThreshold":     gapThreshold,
	}

	return h.renderTemplate(c, "projects/corpus_editor.html", data)
}

// renderTemplate renders a template with the given data
func (h *CorpusHandler) renderTemplate(c echo.Context, name string, data interface{}) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)
	return h.templates.ExecuteTemplate(c.Response().Writer, name, data)
}
