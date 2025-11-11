package project

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Repository はプロジェクトのリポジトリ
type Repository struct {
	db *sql.DB
}

// NewRepository はリポジトリを作成
func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// テーブル作成
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Repository{db: db}, nil
}

// Close はデータベース接続を閉じる
func (r *Repository) Close() error {
	return r.db.Close()
}

// createTables はテーブルを作成
func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		table_name TEXT,
		excel_filename TEXT,
		status TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
	`

	_, err := db.Exec(schema)
	return err
}

// Create はプロジェクトを作成
func (r *Repository) Create(p *Project) error {
	query := `
		INSERT INTO projects (id, name, description, created_at, updated_at, table_name, excel_filename, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		p.ID,
		p.Name,
		p.Description,
		p.CreatedAt,
		p.UpdatedAt,
		p.TableName,
		p.ExcelFilename,
		p.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// FindByID はIDでプロジェクトを取得
func (r *Repository) FindByID(id string) (*Project, error) {
	query := `
		SELECT id, name, description, created_at, updated_at, table_name, excel_filename, status
		FROM projects
		WHERE id = ?
	`

	p := &Project{}
	err := r.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.TableName,
		&p.ExcelFilename,
		&p.Status,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	return p, nil
}

// FindAll は全てのプロジェクトを取得
func (r *Repository) FindAll() ([]*Project, error) {
	query := `
		SELECT id, name, description, created_at, updated_at, table_name, excel_filename, status
		FROM projects
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.TableName,
			&p.ExcelFilename,
			&p.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return projects, nil
}

// Update はプロジェクトを更新
func (r *Repository) Update(p *Project) error {
	p.UpdatedAt = time.Now()

	query := `
		UPDATE projects
		SET name = ?, description = ?, updated_at = ?, table_name = ?, excel_filename = ?, status = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		p.Name,
		p.Description,
		p.UpdatedAt,
		p.TableName,
		p.ExcelFilename,
		p.Status,
		p.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// Delete はプロジェクトを削除
func (r *Repository) Delete(id string) error {
	query := "DELETE FROM projects WHERE id = ?"

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}
