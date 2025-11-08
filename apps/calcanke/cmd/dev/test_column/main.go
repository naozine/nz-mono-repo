package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/marcboeker/go-duckdb"
)

func main() {
	db, err := sql.Open("duckdb", "data/app.duckdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	columnName := "本イベントを何でお知りになりましたか？（複数回答可）"

	// サンプルデータを取得
	query := fmt.Sprintf(`SELECT "%s" FROM excel_import WHERE "%s" IS NOT NULL LIMIT 10`, columnName, columnName)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("=== サンプルデータ ===")
	i := 1
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\n%d: %q\n", i, value)
		fmt.Printf("   長さ: %d文字\n", len(value))
		fmt.Printf("   改行(\\n)含む: %v\n", strings.Contains(value, "\n"))
		fmt.Printf("   改行(\\r\\n)含む: %v\n", strings.Contains(value, "\r\n"))
		fmt.Printf("   カンマ含む: %v\n", strings.Contains(value, ","))
		fmt.Printf("   セミコロン含む: %v\n", strings.Contains(value, ";"))
		i++
	}

	// 統計情報を取得
	var total, newlineCount, crlfCount, commaCount int
	statsQuery := fmt.Sprintf(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN POSITION(CHR(10) IN "%s") > 0 THEN 1 ELSE 0 END) as newline_count,
			SUM(CASE WHEN POSITION(CHR(13)||CHR(10) IN "%s") > 0 THEN 1 ELSE 0 END) as crlf_count,
			SUM(CASE WHEN POSITION(',' IN "%s") > 0 THEN 1 ELSE 0 END) as comma_count
		FROM excel_import
		WHERE "%s" IS NOT NULL
	`, columnName, columnName, columnName, columnName)

	err = db.QueryRow(statsQuery).Scan(&total, &newlineCount, &crlfCount, &commaCount)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n=== 統計情報 ===\n")
	fmt.Printf("総件数: %d\n", total)
	fmt.Printf("改行(LF)含む: %d (%.1f%%)\n", newlineCount, float64(newlineCount)*100/float64(total))
	fmt.Printf("改行(CRLF)含む: %d (%.1f%%)\n", crlfCount, float64(crlfCount)*100/float64(total))
	fmt.Printf("カンマ含む: %d (%.1f%%)\n", commaCount, float64(commaCount)*100/float64(total))
}
