package handlers

import (
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
)

// Handler はHTTPハンドラーの基底構造
type Handler struct {
	analyzer *analyzer.Analyzer
	dbPath   string
	table    string
}

// NewHandler はハンドラーを作成する
func NewHandler(dbPath, table string) *Handler {
	return &Handler{
		dbPath: dbPath,
		table:  table,
	}
}

// getAnalyzer はAnalyzerのインスタンスを取得する
// 各リクエストごとに新しいインスタンスを作成
func (h *Handler) getAnalyzer() (*analyzer.Analyzer, error) {
	return analyzer.NewAnalyzer(h.dbPath, h.table)
}
