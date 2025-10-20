package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/whiscript/internal/model"
	"github.com/yourusername/whiscript/internal/service"
	"github.com/yourusername/whiscript/internal/ui"
)

// ProjectHandler handles HTTP requests for projects
type ProjectHandler struct {
	service   *service.ProjectService
	templates *template.Template
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(service *service.ProjectService) (*ProjectHandler, error) {
	// Create template with helper functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"eq":  func(a, b string) bool { return a == b },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(ui.TemplatesFS, "templates/*.html", "templates/projects/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &ProjectHandler{
		service:   service,
		templates: tmpl,
	}, nil
}

// isHTMXRequest checks if the request is from HTMX
func (h *ProjectHandler) isHTMXRequest(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

// Index handles GET /projects
func (h *ProjectHandler) Index(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	query := c.QueryParam("q")
	sort := c.QueryParam("sort")

	projects, total, err := h.service.List(page, 20, query, sort)
	if err != nil {
		return h.renderError(c, "Failed to load projects", err)
	}

	totalPages := (total + 19) / 20

	data := map[string]interface{}{
		"Projects":    projects,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Query":       query,
		"Sort":        sort,
		"Total":       total,
	}

	if h.isHTMXRequest(c) {
		// Return partial template for HTMX requests (pass projects directly for range)
		return h.renderTemplate(c, "projects/_tbody.html", projects)
	}

	// Return full page for non-HTMX requests
	return h.renderTemplate(c, "projects/index.html", data)
}

// Create handles POST /projects
func (h *ProjectHandler) Create(c echo.Context) error {
	var input model.ProjectCreateInput
	input.Name = c.FormValue("name")
	input.Description = c.FormValue("description")

	project, err := h.service.Create(&input)
	if err != nil {
		return h.renderError(c, "Failed to create project", err)
	}

	return h.renderTemplate(c, "projects/_row.html", project)
}

// Edit handles GET /projects/:id/edit
func (h *ProjectHandler) Edit(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return h.renderError(c, "Invalid project ID", err)
	}

	project, err := h.service.GetByID(id)
	if err != nil {
		return h.renderError(c, "Failed to load project", err)
	}

	data := map[string]interface{}{
		"Project": project,
	}

	return h.renderTemplate(c, "projects/_form.html", data)
}

// Update handles POST /projects/:id
func (h *ProjectHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return h.renderError(c, "Invalid project ID", err)
	}

	var input model.ProjectUpdateInput
	input.Name = c.FormValue("name")
	input.Description = c.FormValue("description")

	project, err := h.service.Update(id, &input)
	if err != nil {
		return h.renderError(c, "Failed to update project", err)
	}

	return h.renderTemplate(c, "projects/_row.html", project)
}

// Delete handles DELETE /projects/:id
func (h *ProjectHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid project ID")
	}

	if err := h.service.Delete(id); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete project")
	}

	// Return empty response with 200 status to trigger HTMX swap
	return c.NoContent(http.StatusOK)
}

// renderTemplate renders a template with the given data
func (h *ProjectHandler) renderTemplate(c echo.Context, name string, data interface{}) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)
	return h.templates.ExecuteTemplate(c.Response().Writer, name, data)
}

// renderError renders an error message
func (h *ProjectHandler) renderError(c echo.Context, message string, err error) error {
	if h.isHTMXRequest(c) {
		data := map[string]interface{}{
			"Message": fmt.Sprintf("%s: %v", message, err),
		}
		return h.renderTemplate(c, "projects/_toast.html", data)
	}
	return c.String(http.StatusInternalServerError, fmt.Sprintf("%s: %v", message, err))
}
