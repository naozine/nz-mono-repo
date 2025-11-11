package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/web"
)

var (
	dbPath      = flag.String("db", "data/app.duckdb", "DuckDBデータベースのパス")
	table       = flag.String("table", "excel_import", "テーブル名")
	projectsDir = flag.String("projects", "projects", "プロジェクトディレクトリのパス")
	port        = flag.String("port", "8080", "サーバーのポート番号")
)

func main() {
	flag.Parse()

	// プロジェクトディレクトリを作成
	if err := os.MkdirAll(*projectsDir, 0755); err != nil {
		fmt.Printf("Failed to create projects directory: %v\n", err)
		os.Exit(1)
	}

	// Webサーバーを起動
	e := web.NewServer(*dbPath, *table, *projectsDir)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Starting Calcanke Web UI on http://localhost%s\n", addr)
	fmt.Printf("Database: %s\n", *dbPath)
	fmt.Printf("Table: %s\n", *table)
	fmt.Printf("Projects Directory: %s\n", *projectsDir)

	e.Logger.Fatal(e.Start(addr))
}
