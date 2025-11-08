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

	fmt.Printf("総列数: %d\n\n", len(columns))

	// 学年列を探す
	var gradeCol *analyzer.Column
	for i := range columns {
		if columns[i].Name == "学年" {
			gradeCol = &columns[i]
			break
		}
	}

	if gradeCol == nil {
		log.Fatal("派生列「学年」が見つかりません")
	}

	fmt.Printf("派生列: %s\n", gradeCol.Name)
	fmt.Printf("説明: %s\n", gradeCol.Type)
	fmt.Printf("派生列?: %v\n", gradeCol.IsDerived)
	fmt.Printf("\nSQL式:\n%s\n\n", gradeCol.SQLExpr)

	// 単純集計を実行
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("学年の集計結果")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	result, err := a.Simpletab(gradeCol, false)
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range result.Rows {
		fmt.Printf("%s: %d件 (%.1f%%)\n", row.Value, row.Count, row.Percentage)
	}
	fmt.Printf("\n総件数: %d\n", result.Total)
}
