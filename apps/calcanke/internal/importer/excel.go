package importer

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/marcboeker/go-duckdb"
)

// ImportExcel はExcelファイルをDuckDBにインポートする
func ImportExcel(excelPath, dbPath, tableName string) error {
	// DuckDB接続
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// 拡張機能の自動インストールを有効化
	if _, err := db.Exec("SET autoinstall_known_extensions = true;"); err != nil {
		return fmt.Errorf("set autoinstall failed: %w", err)
	}
	if _, err := db.Exec("SET autoload_known_extensions = true;"); err != nil {
		return fmt.Errorf("set autoload failed: %w", err)
	}

	// spatial 拡張機能をインストール・ロード
	if _, err := db.Exec("INSTALL spatial;"); err != nil {
		return fmt.Errorf("install spatial failed: %w", err)
	}
	if _, err := db.Exec("LOAD spatial;"); err != nil {
		return fmt.Errorf("load spatial failed: %w", err)
	}

	// 絶対パスに変換
	absPath, err := filepath.Abs(excelPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// テーブルが既に存在する場合は削除
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	if _, err := db.Exec(dropSQL); err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	// Excelファイルからテーブルを作成
	// open_options=['HEADERS=FORCE'] で1行目をヘッダーとして扱う
	createSQL := fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM st_read(?, open_options=['HEADERS=FORCE'])", tableName)
	if _, err := db.Exec(createSQL, absPath); err != nil {
		return fmt.Errorf("failed to create table from excel: %w", err)
	}

	// データが正常にインポートされたか確認
	var count int
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	if err := db.QueryRow(countSQL).Scan(&count); err != nil {
		return fmt.Errorf("failed to count rows: %w", err)
	}

	fmt.Printf("Successfully imported %d rows into table '%s'\n", count, tableName)

	return nil
}
