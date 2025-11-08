package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/marcboeker/go-duckdb"
)

func main() {
	db, err := sql.Open("duckdb", "data/app.duckdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 都道府県と市区町村のサンプルデータを取得
	query := `
		SELECT "都道府県", "市区町村"
		FROM excel_import
		WHERE "都道府県" IS NOT NULL AND "市区町村" IS NOT NULL
		LIMIT 30
	`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("=== サンプルデータ ===")
	i := 1
	for rows.Next() {
		var pref, city string
		if err := rows.Scan(&pref, &city); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%2d: 都道府県=%q, 市区町村=%q\n", i, pref, city)
		i++
	}

	// 東京都のデータパターンを確認
	fmt.Println("\n=== 東京都の市区町村パターン ===")
	query2 := `
		SELECT DISTINCT "市区町村"
		FROM excel_import
		WHERE "都道府県" = '東京都'
		ORDER BY "市区町村"
		LIMIT 50
	`
	rows2, err := db.Query(query2)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	i = 1
	for rows2.Next() {
		var city string
		if err := rows2.Scan(&city); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%2d: %q\n", i, city)
		i++
	}
}
