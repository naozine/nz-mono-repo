// Go
package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/marcboeker/go-duckdb"
)

func main() {
	// .env 読み込み（カレントに .env が無ければ環境既定値だけ使う）
	_ = godotenv.Load()

	excelPath := os.Getenv("EXCEL_PATH") // 例: ./data/input.xlsx
	sheet := os.Getenv("SHEET_NAME")     // 例: Sheet1（未指定ならデフォルト）
	dbPath := os.Getenv("DUCKDB_PATH")   // 例: ./data/app.duckdb（未指定ならメモリ）

	if excelPath == "" {
		log.Fatal("EXCEL_PATH is required")
	}

	dsn := ""
	if dbPath != "" {
		dsn = dbPath
	}
	db, err := sql.Open("duckdb", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 拡張機能の自動インストールを有効化
	if _, err := db.Exec("SET autoinstall_known_extensions = true;"); err != nil {
		log.Fatalf("set autoinstall failed: %v", err)
	}
	if _, err := db.Exec("SET autoload_known_extensions = true;"); err != nil {
		log.Fatalf("set autoload failed: %v", err)
	}

	// 取り込み先テーブル名（環境変数で上書き可）
	table := os.Getenv("TARGET_TABLE")
	if table == "" {
		table = "excel_import"
	}

	// INSTALL と LOAD spatial 拡張機能
	if _, err := db.Exec("INSTALL spatial;"); err != nil {
		log.Fatalf("install spatial failed: %v", err)
	}
	if _, err := db.Exec("LOAD spatial;"); err != nil {
		log.Fatalf("load spatial failed: %v", err)
	}

	// テーブルが無ければスキーマを作る（最初の読み出し結果で定義）
	createSQL := "CREATE TABLE IF NOT EXISTS " + table + " AS SELECT * FROM st_read(?" +
		func() string {
			if sheet != "" {
				return ", layer:=?"
			}
			return ""
		}() +
		") LIMIT 0"

	var errExec error
	if sheet != "" {
		_, errExec = db.Exec(createSQL, excelPath, sheet)
	} else {
		_, errExec = db.Exec(createSQL, excelPath)
	}
	if errExec != nil {
		log.Fatalf("create table failed: %v", errExec)
	}

	// 追記
	insertSQL := "INSERT INTO " + table + " SELECT * FROM st_read(?" +
		func() string {
			if sheet != "" {
				return ", layer:=?"
			}
			return ""
		}() + ")"

	if sheet != "" {
		_, err = db.Exec(insertSQL, excelPath, sheet)
	} else {
		_, err = db.Exec(insertSQL, excelPath)
	}
	if err != nil {
		log.Fatalf("insert failed: %v", err)
	}

	log.Printf("Imported %s (sheet=%s) into table %s", excelPath, valOr(sheet, "<auto>"), table)
}

func valOr(s, alt string) string {
	if s == "" {
		return alt
	}
	return s
}
