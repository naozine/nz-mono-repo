package project

import (
	"time"
)

// Project はプロジェクトのドメインモデル
type Project struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TableName     string    `json:"table_name"`
	ExcelFilename string    `json:"excel_filename"`
	Status        string    `json:"status"` // 'ready', 'importing', 'error'
}

// ProjectStatus はプロジェクトの状態
type ProjectStatus string

const (
	StatusReady     ProjectStatus = "ready"
	StatusImporting ProjectStatus = "importing"
	StatusError     ProjectStatus = "error"
)

// NewProject は新しいプロジェクトを作成
func NewProject(id, name, description string) *Project {
	now := time.Now()
	return &Project{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		Status:      string(StatusImporting),
	}
}

// GetProjectDir はプロジェクトのディレクトリパスを返す
func (p *Project) GetProjectDir(baseDir string) string {
	return baseDir + "/" + p.ID
}

// GetDuckDBPath はDuckDBファイルのパスを返す
func (p *Project) GetDuckDBPath(baseDir string) string {
	return p.GetProjectDir(baseDir) + "/data.duckdb"
}

// GetExcelPath はExcelファイルのパスを返す
func (p *Project) GetExcelPath(baseDir string) string {
	return p.GetProjectDir(baseDir) + "/source.xlsx"
}

// GetDerivedColumnsPath は派生列設定ファイルのパスを返す
func (p *Project) GetDerivedColumnsPath(baseDir string) string {
	return p.GetProjectDir(baseDir) + "/derived_columns.yaml"
}

// GetFiltersPath はフィルタ設定ファイルのパスを返す
func (p *Project) GetFiltersPath(baseDir string) string {
	return p.GetProjectDir(baseDir) + "/filters.yaml"
}

// GetColumnOrdersPath は列の値の表示順序設定ファイルのパスを返す
func (p *Project) GetColumnOrdersPath(baseDir string) string {
	return p.GetProjectDir(baseDir) + "/column_orders.yaml"
}
