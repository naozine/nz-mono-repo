package main

import (
	"fmt"
	"os"

	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "calcanke",
	Short: "Excelデータ分析ツール",
	Long: `Calcanke - ExcelファイルをDuckDBにインポートして分析するツール

使い方:
  calcanke import   - ExcelファイルをDuckDBにインポート
  calcanke columns  - テーブルの列一覧を表示
  calcanke analyze  - 対話的にデータ分析（予定）`,
}

func main() {
	// サブコマンドを登録
	rootCmd.AddCommand(commands.NewImportCmd())
	rootCmd.AddCommand(commands.NewColumnsCmd())
	rootCmd.AddCommand(commands.NewAnalyzeCmd())

	// 実行
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
