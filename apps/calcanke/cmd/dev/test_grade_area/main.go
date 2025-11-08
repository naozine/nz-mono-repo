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

	// 学年列を探す
	var gradeCol *analyzer.Column
	var areaCol *analyzer.Column
	for i := range columns {
		if columns[i].Name == "学年" {
			gradeCol = &columns[i]
		}
		if columns[i].Name == "エリア分類" {
			areaCol = &columns[i]
		}
	}

	if gradeCol == nil {
		log.Fatal("派生列「学年」が見つかりません")
	}
	if areaCol == nil {
		log.Fatal("派生列「エリア分類」が見つかりません")
	}

	// クロス集計を実行（学年 × エリア分類）
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("学年 × エリア分類のクロス集計")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	config := analyzer.AnalysisConfig{
		XColumn: gradeCol,
		YColumn: areaCol,
		SplitX:  false, // 派生列なので分割不可
		SplitY:  false, // 派生列なので分割不可
	}

	result, err := a.Crosstab(config)
	if err != nil {
		log.Fatal(err)
	}

	// 結果を表示
	currentX := ""
	for _, row := range result.Rows {
		if row.XValue != currentX {
			if currentX != "" {
				fmt.Println()
			}
			fmt.Printf("[%s]\n", row.XValue)
			currentX = row.XValue
		}
		fmt.Printf("  %s: %d件 (%.1f%%)\n", row.YValue, row.Count, row.Percentage)
	}
	fmt.Printf("\n総件数: %d\n", result.Total)
}
