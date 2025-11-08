package main

import (
	"fmt"
	"log"

	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
)

func main() {
	// Analyzerを初期化
	a, err := analyzer.NewAnalyzer("data/app.duckdb", "excel_import")
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	// 列を取得
	columns, err := a.GetColumns()
	if err != nil {
		log.Fatal(err)
	}

	// 派生列（エリア分類）を探す
	var areaCol *analyzer.Column
	for i := range columns {
		if columns[i].Name == "エリア分類" {
			areaCol = &columns[i]
			break
		}
	}

	if areaCol == nil {
		log.Fatal("派生列「エリア分類」が見つかりません")
	}

	fmt.Printf("派生列: %s\n", areaCol.Name)
	fmt.Printf("タイプ: %s\n", areaCol.Type)
	fmt.Printf("派生列?: %v\n", areaCol.IsDerived)
	fmt.Printf("\nSQL式:\n%s\n\n", areaCol.SQLExpr)

	// 単純集計を実行
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("エリア分類の集計結果")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	result, err := a.Simpletab(areaCol, false)
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range result.Rows {
		fmt.Printf("%s: %d件 (%.1f%%)\n", row.Value, row.Count, row.Percentage)
	}
	fmt.Printf("\n総件数: %d\n", result.Total)
}
