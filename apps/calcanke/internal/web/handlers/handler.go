package handlers

import (
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
)

// Handler はHTTPハンドラーの基底構造
type Handler struct {
	analyzer           *analyzer.Analyzer
	dbPath             string
	table              string
	derivedColumnsPath string // オプショナル：プロジェクト固有の派生列設定パス
	filtersPath        string // オプショナル：プロジェクト固有のフィルタ設定パス
	columnOrdersPath   string // オプショナル：プロジェクト固有の列順序設定パス
}

// NewHandler はハンドラーを作成する（デフォルトの設定パスを使用）
func NewHandler(dbPath, table string) *Handler {
	return &Handler{
		dbPath: dbPath,
		table:  table,
	}
}

// NewHandlerWithConfigs は設定ファイルパスを指定してハンドラーを作成する
func NewHandlerWithConfigs(dbPath, table, derivedColumnsPath, filtersPath, columnOrdersPath string) *Handler {
	return &Handler{
		dbPath:             dbPath,
		table:              table,
		derivedColumnsPath: derivedColumnsPath,
		filtersPath:        filtersPath,
		columnOrdersPath:   columnOrdersPath,
	}
}

// getAnalyzer はAnalyzerのインスタンスを取得する
// 各リクエストごとに新しいインスタンスを作成
func (h *Handler) getAnalyzer() (*analyzer.Analyzer, error) {
	// 設定ファイルパスが指定されている場合はそれを使用
	if h.derivedColumnsPath != "" && h.filtersPath != "" && h.columnOrdersPath != "" {
		return analyzer.NewAnalyzerWithConfigs(h.dbPath, h.table, h.derivedColumnsPath, h.filtersPath, h.columnOrdersPath)
	}
	return analyzer.NewAnalyzer(h.dbPath, h.table)
}
