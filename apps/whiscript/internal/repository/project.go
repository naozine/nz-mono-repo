package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/hnao/nz-mono-repo/apps/whiscript/internal/model"
	"github.com/jmoiron/sqlx"
)

// ProjectRepository handles database operations for projects
type ProjectRepository struct {
	db *sqlx.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// FindAll retrieves all projects
func (r *ProjectRepository) FindAll() ([]*model.Project, error) {
	var projects []*model.Project
	query := `SELECT id, name, description, created_at, updated_at FROM projects ORDER BY created_at DESC`

	if err := r.db.Select(&projects, query); err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	return projects, nil
}

// FindByID retrieves a project by ID
func (r *ProjectRepository) FindByID(id int64) (*model.Project, error) {
	var project model.Project
	query := `SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ?`

	if err := r.db.Get(&project, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	return &project, nil
}

// Create creates a new project
func (r *ProjectRepository) Create(req *model.CreateProjectRequest) (*model.Project, error) {
	now := time.Now()
	query := `INSERT INTO projects (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)`

	result, err := r.db.Exec(query, req.Name, req.Description, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &model.Project{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update updates an existing project
func (r *ProjectRepository) Update(id int64, req *model.UpdateProjectRequest) (*model.Project, error) {
	now := time.Now()
	query := `UPDATE projects SET name = ?, description = ?, updated_at = ? WHERE id = ?`

	result, err := r.db.Exec(query, req.Name, req.Description, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, nil
	}

	return r.FindByID(id)
}

// Delete deletes a project by ID
func (r *ProjectRepository) Delete(id int64) error {
	query := `DELETE FROM projects WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
