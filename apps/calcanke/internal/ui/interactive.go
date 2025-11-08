package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
)

// RunInteractive は対話的な分析フローを実行
func RunInteractive(dbPath, table string) error {
	// Analyzerを初期化
	a, err := analyzer.NewAnalyzer(dbPath, table)
	if err != nil {
		return fmt.Errorf("failed to initialize analyzer: %w", err)
	}
	defer a.Close()

	// ウェルカムメッセージ
	count, _ := a.GetTableInfo()
	fmt.Printf("\n┌─────────────────────────────────────────┐\n")
	fmt.Printf("│  Calcanke - アンケートデータ分析ツール  │\n")
	fmt.Printf("└─────────────────────────────────────────┘\n\n")
	fmt.Printf("データベース: %s\n", dbPath)
	fmt.Printf("テーブル: %s\n", table)
	fmt.Printf("総レコード数: %s件\n\n", formatNumber(count))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// 列情報を取得
	columns, err := a.GetColumns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// 分析タイプを選択
	var analysisType string
	err = survey.AskOne(&survey.Select{
		Message: "分析タイプを選択してください:",
		Options: []string{
			"単純集計（1列）",
			"クロス集計（2列）",
			"終了",
		},
	}, &analysisType)
	if err != nil {
		return err
	}

	if analysisType == "終了" {
		fmt.Println("終了します")
		return nil
	}

	// 選択された分析タイプに応じて処理を振り分け
	if analysisType == "単純集計（1列）" {
		return runSimpletabFlow(a, columns)
	}

	// クロス集計フロー
	return runCrosstabFlow(a, columns)
}

func runCrosstabFlow(a *analyzer.Analyzer, columns analyzer.ColumnList) error {
	// X軸の列を選択
	var xSelection string
	err := survey.AskOne(&survey.Select{
		Message: "X軸の列を選択してください:",
		Options: columns.ToOptions(),
		Description: func(value string, index int) string {
			if index < len(columns) && columns[index].IsMulti {
				return "この列は複数回答を含みます"
			}
			return ""
		},
	}, &xSelection)
	if err != nil {
		return err
	}

	xIndex := parseSelectionIndex(xSelection)
	xColumn := &columns[xIndex-1]

	fmt.Printf("\n✓ X軸: %s\n\n", xColumn.Name)

	// Y軸の列を選択
	var ySelection string
	err = survey.AskOne(&survey.Select{
		Message: "Y軸の列を選択してください:",
		Options: columns.ToOptions(),
		Description: func(value string, index int) string {
			if index < len(columns) && columns[index].IsMulti {
				return "この列は複数回答を含みます"
			}
			return ""
		},
	}, &ySelection)
	if err != nil {
		return err
	}

	yIndex := parseSelectionIndex(ySelection)
	yColumn := &columns[yIndex-1]

	fmt.Printf("✓ Y軸: %s\n\n", yColumn.Name)

	// 分析設定を作成
	config := analyzer.AnalysisConfig{
		XColumn: xColumn,
		YColumn: yColumn,
	}

	// X軸が複数回答の場合、分割するか確認
	if xColumn.IsMulti {
		var splitX bool
		survey.AskOne(&survey.Confirm{
			Message: "X軸を複数回答として分割しますか？",
			Default: true,
		}, &splitX)
		config.SplitX = splitX
	}

	// Y軸が複数回答の場合、分割するか確認
	if yColumn.IsMulti {
		var splitY bool
		survey.AskOne(&survey.Confirm{
			Message: "Y軸を複数回答として分割しますか？",
			Default: true,
		}, &splitY)
		config.SplitY = splitY
	}

	// 集計実行
	fmt.Println("\n集計中...")
	result, err := a.Crosstab(config)
	if err != nil {
		return fmt.Errorf("failed to execute crosstab: %w", err)
	}

	// 結果表示
	DisplayCrosstabResult(result)

	// 次のアクション
	var nextAction string
	survey.AskOne(&survey.Select{
		Message: "次の操作を選択してください:",
		Options: []string{
			"別の集計を実行",
			"終了",
		},
	}, &nextAction)

	if nextAction == "別の集計を実行" {
		// 再度、Analyzerを使って分析
		// 現在のAnalyzerをそのまま使えるので、列情報を再取得
		cols, _ := a.GetColumns()
		return runCrosstabFlow(a, cols)
	}

	fmt.Println("\n終了します")
	return nil
}

func runSimpletabFlow(a *analyzer.Analyzer, columns analyzer.ColumnList) error {
	// 列を選択
	var selection string
	err := survey.AskOne(&survey.Select{
		Message: "集計する列を選択してください:",
		Options: columns.ToOptions(),
		Description: func(value string, index int) string {
			if index < len(columns) && columns[index].IsMulti {
				return "この列は複数回答を含みます"
			}
			return ""
		},
	}, &selection)
	if err != nil {
		return err
	}

	columnIndex := parseSelectionIndex(selection)
	column := &columns[columnIndex-1]

	fmt.Printf("\n✓ 集計列: %s\n\n", column.Name)

	// 複数回答列の場合、分割するか確認
	split := false
	if column.IsMulti {
		survey.AskOne(&survey.Confirm{
			Message: "複数回答として分割しますか？",
			Default: true,
		}, &split)
	}

	// 集計実行
	fmt.Println("\n集計中...")
	result, err := a.Simpletab(column, split)
	if err != nil {
		return fmt.Errorf("failed to execute simpletab: %w", err)
	}

	// 結果表示
	DisplaySimpletabResult(result)

	// 次のアクション
	var nextAction string
	survey.AskOne(&survey.Select{
		Message: "次の操作を選択してください:",
		Options: []string{
			"別の集計を実行",
			"終了",
		},
	}, &nextAction)

	if nextAction == "別の集計を実行" {
		return RunInteractive(a.DBPath, a.Table)
	}

	fmt.Println("\n終了します")
	return nil
}

// parseSelectionIndex は選択された文字列から列番号を抽出
func parseSelectionIndex(selection string) int {
	// " 1  列名 [複数回答]" のような形式から数字を抽出
	parts := strings.Fields(selection)
	if len(parts) == 0 {
		return 1
	}
	index, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 1
	}
	return index
}
