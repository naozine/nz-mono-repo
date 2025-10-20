package service

import (
	"fmt"
	"strings"

	"github.com/yourusername/whiscript/internal/model"
	"github.com/yourusername/whiscript/internal/repository"
)

// ProjectService handles business logic for projects
type ProjectService struct {
	repo *repository.ProjectRepository
}

// NewProjectService creates a new project service
func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

// List retrieves projects with pagination, search, and sorting
func (s *ProjectService) List(page, limit int, query, sort string) ([]*model.Project, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.repo.List(page, limit, query, sort)
}

// GetByID retrieves a project by ID
func (s *ProjectService) GetByID(id int64) (*model.Project, error) {
	if id < 1 {
		return nil, fmt.Errorf("invalid project id")
	}
	return s.repo.GetByID(id)
}

// Create creates a new project with validation
func (s *ProjectService) Create(input *model.ProjectCreateInput) (*model.Project, error) {
	if err := s.validateProjectInput(input.Name, input.Description); err != nil {
		return nil, err
	}
	return s.repo.Create(input)
}

// Update updates an existing project with validation
func (s *ProjectService) Update(id int64, input *model.ProjectUpdateInput) (*model.Project, error) {
	if id < 1 {
		return nil, fmt.Errorf("invalid project id")
	}

	if err := s.validateProjectInput(input.Name, input.Description); err != nil {
		return nil, err
	}

	return s.repo.Update(id, input)
}

// Delete deletes a project by ID
func (s *ProjectService) Delete(id int64) error {
	if id < 1 {
		return fmt.Errorf("invalid project id")
	}
	return s.repo.Delete(id)
}

// validateProjectInput validates project input fields
func (s *ProjectService) validateProjectInput(name, description string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("project name is required")
	}
	if len(name) > 255 {
		return fmt.Errorf("project name must be less than 255 characters")
	}
	if len(description) > 1000 {
		return fmt.Errorf("project description must be less than 1000 characters")
	}
	return nil
}
