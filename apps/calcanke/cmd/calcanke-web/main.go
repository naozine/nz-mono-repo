package main

import (
	"flag"
	"fmt"

	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/web"
)

var (
	dbPath = flag.String("db", "data/app.duckdb", "DuckDBデータベースのパス")
	table  = flag.String("table", "excel_import", "テーブル名")
	port   = flag.String("port", "8080", "サーバーのポート番号")
)

func main() {
	flag.Parse()

	// Webサーバーを起動
	e := web.NewServer(*dbPath, *table)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Starting Calcanke Web UI on http://localhost%s\n", addr)
	fmt.Printf("Database: %s\n", *dbPath)
	fmt.Printf("Table: %s\n", *table)

	e.Logger.Fatal(e.Start(addr))
}
