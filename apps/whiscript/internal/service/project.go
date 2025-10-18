package service

import (
	"fmt"

	"github.com/hnao/nz-mono-repo/apps/whiscript/internal/model"
	"github.com/hnao/nz-mono-repo/apps/whiscript/internal/repository"
)

// ProjectService handles business logic for projects
type ProjectService struct {
	repo *repository.ProjectRepository
}

// NewProjectService creates a new project service
func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

// ListProjects retrieves all projects
func (s *ProjectService) ListProjects() ([]*model.Project, error) {
	return s.repo.FindAll()
}

// GetProject retrieves a project by ID
func (s *ProjectService) GetProject(id int64) (*model.Project, error) {
	project, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, fmt.Errorf("project not found")
	}
	return project, nil
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(req *model.CreateProjectRequest) (*model.Project, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	return s.repo.Create(req)
}

// UpdateProject updates an existing project
func (s *ProjectService) UpdateProject(id int64, req *model.UpdateProjectRequest) (*model.Project, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	project, err := s.repo.Update(id, req)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, fmt.Errorf("project not found")
	}
	return project, nil
}

// DeleteProject deletes a project by ID
func (s *ProjectService) DeleteProject(id int64) error {
	return s.repo.Delete(id)
}
