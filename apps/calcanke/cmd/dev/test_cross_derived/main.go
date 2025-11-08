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

	// 複数回答列を探す
	var eventInfoCol *analyzer.Column
	for i := range columns {
		if columns[i].Name == "本イベントを何でお知りになりましたか？（複数回答可）" {
			eventInfoCol = &columns[i]
			break
		}
	}

	if eventInfoCol == nil {
		log.Fatal("列が見つかりません")
	}

	// クロス集計を実行（派生列 × 複数回答）
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("エリア分類 × イベント情報源（分割あり）")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	config := analyzer.AnalysisConfig{
		XColumn: areaCol,
		YColumn: eventInfoCol,
		SplitX:  false, // 派生列なので分割不可
		SplitY:  true,  // 複数回答を分割
	}

	result, err := a.Crosstab(config)
	if err != nil {
		log.Fatal(err)
	}

	// 結果を表示（最初の30行のみ）
	count := 0
	for _, row := range result.Rows {
		if count >= 30 {
			fmt.Printf("... 他%d行\n", len(result.Rows)-count)
			break
		}
		fmt.Printf("%s × %s: %d件 (%.1f%%)\n", row.XValue, row.YValue, row.Count, row.Percentage)
		count++
	}
	fmt.Printf("\n総件数: %d\n", result.Total)
}
