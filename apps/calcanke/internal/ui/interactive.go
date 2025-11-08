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
			if index < len(columns) {
				if columns[index].IsDerived {
					return "派生列（設定ファイルで定義）"
				} else if columns[index].IsMulti {
					return "この列は複数回答を含みます"
				}
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
			if index < len(columns) {
				if columns[index].IsDerived {
					return "派生列（設定ファイルで定義）"
				} else if columns[index].IsMulti {
					return "この列は複数回答を含みます"
				}
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

	// X軸が複数回答の場合、分割するか確認（派生列は除く）
	if xColumn.IsMulti && !xColumn.IsDerived {
		var splitX bool
		survey.AskOne(&survey.Confirm{
			Message: "X軸を複数回答として分割しますか？",
			Default: true,
		}, &splitX)
		config.SplitX = splitX
	}

	// Y軸が複数回答の場合、分割するか確認（派生列は除く）
	if yColumn.IsMulti && !yColumn.IsDerived {
		var splitY bool
		survey.AskOne(&survey.Confirm{
			Message: "Y軸を複数回答として分割しますか？",
			Default: true,
		}, &splitY)
		config.SplitY = splitY
	}

	// フィルタを選択
	selectedFilter, err := selectFilter(a)
	if err != nil {
		return err
	}

	// 集計実行
	fmt.Println("\n集計中...")
	result, err := a.CrosstabWithFilter(config, selectedFilter)
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
			if index < len(columns) {
				if columns[index].IsDerived {
					return "派生列（設定ファイルで定義）"
				} else if columns[index].IsMulti {
					return "この列は複数回答を含みます"
				}
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

	// 複数回答列の場合、分割するか確認（派生列は除く）
	split := false
	if column.IsMulti && !column.IsDerived {
		survey.AskOne(&survey.Confirm{
			Message: "複数回答として分割しますか？",
			Default: true,
		}, &split)
	}

	// フィルタを選択
	selectedFilter, err := selectFilter(a)
	if err != nil {
		return err
	}

	// 集計実行
	fmt.Println("\n集計中...")
	result, err := a.SimpletabWithFilter(column, split, selectedFilter)
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

// selectFilter はフィルタを選択するプロンプトを表示
func selectFilter(a *analyzer.Analyzer) (*analyzer.Filter, error) {
	if len(a.Filters) == 0 {
		// フィルタが定義されていない場合はnilを返す
		return nil, nil
	}

	// フィルタオプションを作成
	options := []string{"フィルタなし（全データ）"}
	for _, filter := range a.Filters {
		options = append(options, filter.Name)
	}

	// フィルタを選択
	var selection string
	err := survey.AskOne(&survey.Select{
		Message: "フィルタを選択してください:",
		Options: options,
		Description: func(value string, index int) string {
			if index == 0 {
				return "フィルタを適用せず、全データを集計します"
			}
			if index-1 < len(a.Filters) {
				return a.Filters[index-1].Description
			}
			return ""
		},
	}, &selection)
	if err != nil {
		return nil, err
	}

	// 「フィルタなし」が選択された場合はnilを返す
	if selection == "フィルタなし（全データ）" {
		fmt.Println("\n✓ フィルタ: なし\n")
		return nil, nil
	}

	// 選択されたフィルタを探す
	for i := range a.Filters {
		if a.Filters[i].Name == selection {
			fmt.Printf("\n✓ フィルタ: %s\n", a.Filters[i].Name)
			fmt.Printf("  条件: %s\n\n", a.Filters[i].Description)
			return &a.Filters[i], nil
		}
	}

	// 見つからない場合はnilを返す（通常は発生しない）
	return nil, nil
}
