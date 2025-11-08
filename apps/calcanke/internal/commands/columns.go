package commands

import (
	"fmt"

	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
	"github.com/spf13/cobra"
)

var (
	dbPath string
	table  string
)

// NewColumnsCmd はcolumnsコマンドを作成
func NewColumnsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "columns",
		Short: "テーブルの列一覧を表示",
		Long:  "DuckDBデータベースのテーブルの列一覧を番号付きで表示します",
		RunE:  runColumns,
	}

	cmd.Flags().StringVar(&dbPath, "db", "data/app.duckdb", "DuckDBデータベースのパス")
	cmd.Flags().StringVar(&table, "table", "excel_import", "テーブル名")

	return cmd
}

func runColumns(cmd *cobra.Command, args []string) error {
	// Analyzerを初期化
	a, err := analyzer.NewAnalyzer(dbPath, table)
	if err != nil {
		return fmt.Errorf("failed to initialize analyzer: %w", err)
	}
	defer a.Close()

	// テーブル情報を取得
	count, err := a.GetTableInfo()
	if err != nil {
		return fmt.Errorf("failed to get table info: %w", err)
	}

	// 列情報を取得
	columns, err := a.GetColumns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// 結果を表示
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("データベース: %s\n", dbPath)
	fmt.Printf("テーブル: %s\n", table)
	fmt.Printf("総件数: %s件\n", formatNumber(count))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// 列一覧を表示
	fmt.Printf("%-3s %-50s %-12s %s\n", "No", "列名", "型", "")
	fmt.Printf("%-3s %-50s %-12s %s\n", "---", "------------------------------------------------", "------------", "")

	for _, col := range columns {
		marker := ""
		if col.IsMulti {
			marker = "[複数回答]"
		}
		fmt.Printf("%3d  %-50s %-12s %s\n", col.Index, col.Name, col.Type, marker)
	}

	fmt.Printf("\n総列数: %d\n\n", len(columns))

	return nil
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
