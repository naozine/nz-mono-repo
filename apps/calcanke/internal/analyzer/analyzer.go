package analyzer

import (
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
)

// Analyzer はデータ分析を実行
type Analyzer struct {
	db             *sql.DB
	DBPath         string
	Table          string
	DerivedColumns []DerivedColumn
	derivedColsMap map[string]*DerivedColumn // 名前から派生列を引くマップ
	Filters        []Filter                  // 利用可能なフィルタ
}

// NewAnalyzer はAnalyzerを作成
func NewAnalyzer(dbPath, table string) (*Analyzer, error) {
	// DuckDB拡張機能の自動インストールを有効化
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 拡張機能の自動インストール設定
	if _, err := db.Exec("SET autoinstall_known_extensions = true;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set autoinstall: %w", err)
	}
	if _, err := db.Exec("SET autoload_known_extensions = true;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set autoload: %w", err)
	}

	// テーブルの存在確認
	var count int
	err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s LIMIT 1", table)).Scan(&count)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("table %s not found or cannot be accessed: %w", table, err)
	}

	// 派生列の設定を読み込み（オプショナル）
	derivedCols, err := LoadDerivedColumns("configs/derived_columns.yaml")
	if err != nil {
		// 設定ファイルがなくてもエラーにしない
		derivedCols = []DerivedColumn{}
	}

	// 派生列のマップを作成
	derivedColsMap := make(map[string]*DerivedColumn)
	for i := range derivedCols {
		derivedColsMap[derivedCols[i].Name] = &derivedCols[i]
	}

	// フィルタの設定を読み込み（オプショナル）
	filters, err := LoadFilters("configs/filters.yaml")
	if err != nil {
		// 設定ファイルがなくてもエラーにしない
		filters = []Filter{}
	}

	return &Analyzer{
		db:             db,
		DBPath:         dbPath,
		Table:          table,
		DerivedColumns: derivedCols,
		derivedColsMap: derivedColsMap,
		Filters:        filters,
	}, nil
}

// Close はデータベース接続を閉じる
func (a *Analyzer) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// GetTableInfo はテーブルの基本情報を取得
func (a *Analyzer) GetTableInfo() (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", a.Table)
	err := a.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get table info: %w", err)
	}
	return count, nil
}
