package commands

import (
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/ui"
	"github.com/spf13/cobra"
)

var (
	analyzeDBPath string
	analyzeTable  string
)

// NewAnalyzeCmd はanalyzeコマンドを作成
func NewAnalyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "対話的にデータ分析",
		Long:  "対話的にデータを分析します。列を選択してクロス集計を実行できます。",
		RunE:  runAnalyze,
	}

	cmd.Flags().StringVar(&analyzeDBPath, "db", "data/app.duckdb", "DuckDBデータベースのパス")
	cmd.Flags().StringVar(&analyzeTable, "table", "excel_import", "テーブル名")

	return cmd
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	return ui.RunInteractive(analyzeDBPath, analyzeTable)
}
