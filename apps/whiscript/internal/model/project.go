package model

import (
	"errors"
	"strings"
	"time"
)

// Project represents a project entity
type Project struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// CreateProjectRequest represents the request payload for creating a project
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateProjectRequest represents the request payload for updating a project
type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Validate validates the CreateProjectRequest
func (r *CreateProjectRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)

	if r.Name == "" {
		return errors.New("name is required")
	}

	if len(r.Name) > 100 {
		return errors.New("name must be less than 100 characters")
	}

	if len(r.Description) > 500 {
		return errors.New("description must be less than 500 characters")
	}

	return nil
}

// Validate validates the UpdateProjectRequest
func (r *UpdateProjectRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)

	if r.Name == "" {
		return errors.New("name is required")
	}

	if len(r.Name) > 100 {
		return errors.New("name must be less than 100 characters")
	}

	if len(r.Description) > 500 {
		return errors.New("description must be less than 500 characters")
	}

	return nil
}
