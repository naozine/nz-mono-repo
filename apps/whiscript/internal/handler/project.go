package handler

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/hnao/nz-mono-repo/apps/whiscript/internal/model"
	"github.com/hnao/nz-mono-repo/apps/whiscript/internal/service"
	"github.com/labstack/echo/v4"
)

// ProjectHandler handles HTTP requests for projects
type ProjectHandler struct {
	service   *service.ProjectService
	templates *template.Template
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(service *service.ProjectService, templates *template.Template) *ProjectHandler {
	return &ProjectHandler{
		service:   service,
		templates: templates,
	}
}

// Index renders the index page with all projects
func (h *ProjectHandler) Index(c echo.Context) error {
	projects, err := h.service.ListProjects()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch projects")
	}

	data := map[string]interface{}{
		"Projects": projects,
	}

	return h.templates.ExecuteTemplate(c.Response().Writer, "index.html", data)
}

// Create handles project creation
func (h *ProjectHandler) Create(c echo.Context) error {
	req := &model.CreateProjectRequest{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
	}

	project, err := h.service.CreateProject(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	c.Response().Header().Set("HX-Trigger", "projectCreated")
	return h.templates.ExecuteTemplate(c.Response().Writer, "project-item", project)
}

// Update handles project updates
func (h *ProjectHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	req := &model.UpdateProjectRequest{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
	}

	project, err := h.service.UpdateProject(id, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	c.Response().Header().Set("HX-Trigger", "projectUpdated")
	return h.templates.ExecuteTemplate(c.Response().Writer, "project-item", project)
}

// Delete handles project deletion
func (h *ProjectHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	err = h.service.DeleteProject(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete project")
	}

	c.Response().Header().Set("HX-Trigger", "projectDeleted")
	return c.NoContent(http.StatusOK)
}
