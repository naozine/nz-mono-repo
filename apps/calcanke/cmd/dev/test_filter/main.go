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

	fmt.Printf("利用可能なフィルタ: %d個\n\n", len(a.Filters))

	// フィルタを表示
	for i, filter := range a.Filters {
		fmt.Printf("%d. %s\n", i+1, filter.Name)
		fmt.Printf("   説明: %s\n", filter.Description)
		fmt.Printf("   SQL: %s\n\n", filter.GenerateWhereClause(a))
	}

	// 列を取得
	columns, err := a.GetColumns()
	if err != nil {
		log.Fatal(err)
	}

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

	// フィルタなしで集計
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("【フィルタなし】学年の集計")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	result1, err := a.SimpletabWithFilter(gradeCol, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range result1.Rows {
		fmt.Printf("%s: %d件 (%.1f%%)\n", row.Value, row.Count, row.Percentage)
	}
	fmt.Printf("総件数: %d\n\n", result1.Total)

	// 「小中学生のみ」フィルタを適用
	var smallMidFilter *analyzer.Filter
	for i := range a.Filters {
		if a.Filters[i].Name == "小中学生のみ" {
			smallMidFilter = &a.Filters[i]
			break
		}
	}

	if smallMidFilter != nil {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("【フィルタ: %s】学年の集計\n", smallMidFilter.Name)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		result2, err := a.SimpletabWithFilter(gradeCol, false, smallMidFilter)
		if err != nil {
			log.Fatal(err)
		}

		for _, row := range result2.Rows {
			fmt.Printf("%s: %d件 (%.1f%%)\n", row.Value, row.Count, row.Percentage)
		}
		fmt.Printf("総件数: %d\n", result2.Total)
		fmt.Printf("除外件数: %d件\n\n", result1.Total-result2.Total)
	}
}
