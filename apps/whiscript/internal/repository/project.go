package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/whiscript/internal/model"
)

// ProjectRepository handles database operations for projects
type ProjectRepository struct {
	db *sqlx.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// List retrieves projects with pagination, search, and sorting
func (r *ProjectRepository) List(page, limit int, query, sort string) ([]*model.Project, int, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := ""
	args := make(map[string]interface{})
	if query != "" {
		whereClause = "WHERE name LIKE :query OR description LIKE :query"
		args["query"] = "%" + query + "%"
	}

	// Build ORDER BY clause
	orderClause := "ORDER BY created_at DESC"
	if sort != "" {
		parts := strings.Split(sort, ".")
		if len(parts) == 2 {
			column := parts[0]
			direction := strings.ToUpper(parts[1])
			if direction == "ASC" || direction == "DESC" {
				// Validate column name to prevent SQL injection
				validColumns := map[string]bool{"name": true, "created_at": true, "updated_at": true}
				if validColumns[column] {
					orderClause = fmt.Sprintf("ORDER BY %s %s", column, direction)
				}
			}
		}
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM projects %s", whereClause)
	var total int
	countStmt, err := r.db.PrepareNamed(countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to prepare count query: %w", err)
	}
	defer countStmt.Close()

	if err := countStmt.Get(&total, args); err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// Get projects
	args["limit"] = limit
	args["offset"] = offset
	selectQuery := fmt.Sprintf("SELECT * FROM projects %s %s LIMIT :limit OFFSET :offset", whereClause, orderClause)

	stmt, err := r.db.PrepareNamed(selectQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to prepare select query: %w", err)
	}
	defer stmt.Close()

	var projects []*model.Project
	if err := stmt.Select(&projects, args); err != nil {
		return nil, 0, fmt.Errorf("failed to select projects: %w", err)
	}

	return projects, total, nil
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(id int64) (*model.Project, error) {
	var project model.Project
	query := "SELECT * FROM projects WHERE id = ?"
	if err := r.db.Get(&project, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}

// Create creates a new project
func (r *ProjectRepository) Create(input *model.ProjectCreateInput) (*model.Project, error) {
	now := time.Now()
	query := `INSERT INTO projects (name, description, created_at, updated_at)
	          VALUES (:name, :description, :created_at, :updated_at)`

	args := map[string]interface{}{
		"name":        input.Name,
		"description": input.Description,
		"created_at":  now,
		"updated_at":  now,
	}

	result, err := r.db.NamedExec(query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetByID(id)
}

// Update updates an existing project
func (r *ProjectRepository) Update(id int64, input *model.ProjectUpdateInput) (*model.Project, error) {
	query := `UPDATE projects
	          SET name = :name, description = :description, updated_at = :updated_at
	          WHERE id = :id`

	args := map[string]interface{}{
		"id":          id,
		"name":        input.Name,
		"description": input.Description,
		"updated_at":  time.Now(),
	}

	result, err := r.db.NamedExec(query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("project not found")
	}

	return r.GetByID(id)
}

// Delete deletes a project by ID
func (r *ProjectRepository) Delete(id int64) error {
	query := "DELETE FROM projects WHERE id = ?"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}
