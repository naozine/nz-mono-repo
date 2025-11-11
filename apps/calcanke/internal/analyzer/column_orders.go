package analyzer

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// ColumnOrderConfig は列の値の表示順序設定全体
type ColumnOrderConfig struct {
	ColumnOrders []ColumnOrder `yaml:"column_orders"`
}

// ColumnOrder は1つの列の値の表示順序定義
type ColumnOrder struct {
	Column      string   `yaml:"column" json:"column"`
	Description string   `yaml:"description" json:"description"`
	Values      []string `yaml:"values" json:"values"`
}

// LoadColumnOrders は設定ファイルから列の値の表示順序を読み込む
func LoadColumnOrders(configPath string) ([]ColumnOrder, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ColumnOrderConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return config.ColumnOrders, nil
}

// SaveColumnOrders は列の値の表示順序を設定ファイルに書き込む
func SaveColumnOrders(configPath string, orders []ColumnOrder) error {
	config := ColumnOrderConfig{
		ColumnOrders: orders,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	// ヘッダーコメントを追加
	header := "# 列の値の表示順序の定義\n# この設定ファイルで、列に含まれる値の表示順序を指定できます\n# グラフや表での表示順序が制御されます\n\n"
	data = append([]byte(header), data...)

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetOrderForColumn は指定された列の値の順序を取得（マップ: 値 -> 順序インデックス）
func (co *ColumnOrder) GetOrderForColumn() map[string]int {
	orderMap := make(map[string]int)
	for i, value := range co.Values {
		orderMap[value] = i
	}
	return orderMap
}

// GetValueOrder は値の表示順序を取得（優先順位: 明示的設定 > 派生列ルール > デフォルト）
func (a *Analyzer) GetValueOrder(columnName string) map[string]int {
	// 優先度1: 明示的な列順序設定
	if colOrder, exists := a.columnOrdersMap[columnName]; exists {
		return colOrder.GetOrderForColumn()
	}

	// 優先度2: 派生列のルール順序
	if derivedCol, exists := a.derivedColsMap[columnName]; exists {
		if derivedCol.CalculationType == "rules" && len(derivedCol.Rules) > 0 {
			orderMap := make(map[string]int)
			for i, rule := range derivedCol.Rules {
				orderMap[rule.Label] = i
			}
			return orderMap
		}
	}

	// 優先度3: デフォルト（順序なし = 空のマップ）
	return make(map[string]int)
}

// sortByOrder は指定された順序マップに従って値をソートする
// orderMapが空の場合は文字列順にソートする
func sortByOrder(values []string, orderMap map[string]int) {
	if len(orderMap) == 0 {
		// 順序が定義されていない場合は文字列順
		sort.Strings(values)
		return
	}

	// カスタム順序でソート
	sort.Slice(values, func(i, j int) bool {
		orderI, existsI := orderMap[values[i]]
		orderJ, existsJ := orderMap[values[j]]

		// 両方とも順序が定義されている場合
		if existsI && existsJ {
			return orderI < orderJ
		}

		// iのみ順序が定義されている場合（iを先に）
		if existsI {
			return true
		}

		// jのみ順序が定義されている場合（jを先に）
		if existsJ {
			return false
		}

		// 両方とも順序が定義されていない場合は文字列順
		return values[i] < values[j]
	})
}

// sortByOrderForSimpletab は単純集計の行を指定された順序マップに従ってソートする
// orderMapが空の場合は元の順序（カウント順）を維持する
func sortByOrderForSimpletab(rows []SimpletabRow, orderMap map[string]int) {
	if len(orderMap) == 0 {
		// 順序が定義されていない場合は元の順序を維持（すでにカウント順）
		return
	}

	// カスタム順序でソート
	sort.Slice(rows, func(i, j int) bool {
		orderI, existsI := orderMap[rows[i].Value]
		orderJ, existsJ := orderMap[rows[j].Value]

		// 両方とも順序が定義されている場合
		if existsI && existsJ {
			return orderI < orderJ
		}

		// iのみ順序が定義されている場合（iを先に）
		if existsI {
			return true
		}

		// jのみ順序が定義されている場合（jを先に）
		if existsJ {
			return false
		}

		// 両方とも順序が定義されていない場合はカウント順
		return rows[i].Count > rows[j].Count
	})
}
