package ui

import (
	"fmt"
	"os"

	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
	"github.com/olekukonko/tablewriter"
)

// DisplayCrosstabResult はクロス集計の結果を表形式で表示
func DisplayCrosstabResult(result *analyzer.CrosstabResult) {
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("クロス集計結果\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(result.XColumn, result.YColumn, "件数", "割合")

	for _, row := range result.Rows {
		table.Append(
			row.XValue,
			row.YValue,
			formatNumber(row.Count),
			fmt.Sprintf("%.1f%%", row.Percentage),
		)
	}

	table.Render()

	fmt.Printf("\n総件数: %s\n\n", formatNumber(result.Total))
}

// formatNumber は数値を3桁カンマ区切りにフォーマット
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	s := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}
