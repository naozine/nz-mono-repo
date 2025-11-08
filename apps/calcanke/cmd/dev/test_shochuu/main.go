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

	fmt.Printf("派生列の数: %d個\n\n", len(a.DerivedColumns))

	// 派生列を表示
	for i, dc := range a.DerivedColumns {
		fmt.Printf("%d. %s\n", i+1, dc.Name)
		fmt.Printf("   説明: %s\n", dc.Description)
		fmt.Printf("   SQL: %s\n\n", dc.GenerateCaseExpression())
	}

	// 列を取得
	columns, err := a.GetColumns()
	if err != nil {
		log.Fatal(err)
	}

	// 「学校種別」列を探す
	var schoolTypeCol *analyzer.Column
	for i := range columns {
		if columns[i].Name == "学校種別" {
			schoolTypeCol = &columns[i]
			break
		}
	}

	if schoolTypeCol == nil {
		log.Fatal("派生列「学校種別」が見つかりません")
	}

	// 「学校種別」列で集計
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("【学校種別列の集計】")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	result, err := a.Simpletab(schoolTypeCol, false)
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range result.Rows {
		fmt.Printf("%s: %d件 (%.1f%%)\n", row.Value, row.Count, row.Percentage)
	}
	fmt.Printf("総件数: %d\n\n", result.Total)

	// 学年列で集計（比較用）
	var gradeCol *analyzer.Column
	for i := range columns {
		if columns[i].Name == "学年" {
			gradeCol = &columns[i]
			break
		}
	}

	if gradeCol != nil {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("【学年列の集計（参考）】")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		result2, err := a.Simpletab(gradeCol, false)
		if err != nil {
			log.Fatal(err)
		}

		for _, row := range result2.Rows {
			fmt.Printf("%s: %d件 (%.1f%%)\n", row.Value, row.Count, row.Percentage)
		}
		fmt.Printf("総件数: %d\n\n", result2.Total)
	}
}
