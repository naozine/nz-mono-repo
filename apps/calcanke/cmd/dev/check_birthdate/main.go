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

	// 生年月日列のサンプルデータを取得
	query := `
		SELECT "生年月日"
		FROM excel_import
		WHERE "生年月日" IS NOT NULL
		LIMIT 50
	`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("生年月日のサンプルデータ（50件）")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	i := 1
	birthdates := make(map[string]int)
	var samples []string

	for rows.Next() {
		var birthdate string
		if err := rows.Scan(&birthdate); err != nil {
			log.Fatal(err)
		}

		if i <= 20 {
			fmt.Printf("%2d: %q (長さ: %d)\n", i, birthdate, len(birthdate))
			samples = append(samples, birthdate)
		}

		birthdates[birthdate]++
		i++
	}

	// 統計情報
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("統計情報\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// データ型を確認
	query2 := `
		SELECT column_name, column_type
		FROM (DESCRIBE excel_import)
		WHERE column_name = '生年月日'
	`
	var colName, colType string
	err = db.QueryRow(query2).Scan(&colName, &colType)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("列名: %s\n", colName)
	fmt.Printf("データ型: %s\n\n", colType)

	// 全体の統計
	var total, nullCount int
	query3 := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN "生年月日" IS NULL THEN 1 ELSE 0 END) as null_count
		FROM excel_import
	`
	err = db.QueryRow(query3).Scan(&total, &nullCount)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("総レコード数: %d\n", total)
	fmt.Printf("NULL: %d件 (%.1f%%)\n", nullCount, float64(nullCount)*100/float64(total))
	fmt.Printf("データあり: %d件 (%.1f%%)\n", total-nullCount, float64(total-nullCount)*100/float64(total))

	// フォーマットパターンの検出
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("フォーマットパターンの分析\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	patterns := make(map[string]int)
	for _, bd := range samples {
		pattern := detectPattern(bd)
		patterns[pattern]++
	}

	for pattern, count := range patterns {
		fmt.Printf("%s: %d件\n", pattern, count)
	}

	// ユニーク値の数
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("ユニーク値の数\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	var uniqueCount int
	query4 := `
		SELECT COUNT(DISTINCT "生年月日")
		FROM excel_import
		WHERE "生年月日" IS NOT NULL
	`
	err = db.QueryRow(query4).Scan(&uniqueCount)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ユニーク値: %d種類\n", uniqueCount)

	// 頻度が高い値
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("頻度の高い生年月日（上位10件）\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	query5 := `
		SELECT "生年月日", COUNT(*) as cnt
		FROM excel_import
		WHERE "生年月日" IS NOT NULL
		GROUP BY "生年月日"
		ORDER BY cnt DESC
		LIMIT 10
	`
	rows2, err := db.Query(query5)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	i = 1
	for rows2.Next() {
		var bd string
		var cnt int
		if err := rows2.Scan(&bd, &cnt); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%2d. %q: %d件\n", i, bd, cnt)
		i++
	}
}

func detectPattern(s string) string {
	// フォーマットパターンを推測
	if len(s) == 0 {
		return "空文字"
	}

	// YYYY/MM/DD
	if len(s) == 10 && s[4] == '/' && s[7] == '/' {
		return "YYYY/MM/DD"
	}

	// YYYY-MM-DD
	if len(s) == 10 && s[4] == '-' && s[7] == '-' {
		return "YYYY-MM-DD"
	}

	// YYYYMMDD
	if len(s) == 8 {
		return "YYYYMMDD"
	}

	// その他
	return fmt.Sprintf("その他(長さ%d)", len(s))
}
