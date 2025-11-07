package model

import "time"

// Project represents a project entity
type Project struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// ProjectCreateInput represents input for creating a project
type ProjectCreateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ProjectUpdateInput represents input for updating a project
type ProjectUpdateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
