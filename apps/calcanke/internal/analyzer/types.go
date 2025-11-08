package analyzer

import "fmt"

// Column は列の情報を保持
type Column struct {
	Index       int    // 1始まりの列番号
	Name        string // 列名
	Type        string // データ型（VARCHAR, INTEGER等）
	IsMulti     bool   // 複数回答かどうか（改行含む割合で判定）
	UniqueCount int    // ユニーク値の数
}

// ColumnList は列の一覧
type ColumnList []Column

// ToOptions は survey 用のオプション文字列に変換
func (cl ColumnList) ToOptions() []string {
	options := make([]string, len(cl))
	for i, col := range cl {
		marker := ""
		if col.IsMulti {
			marker = " [複数回答]"
		}
		options[i] = fmt.Sprintf("%2d  %s%s", col.Index, col.Name, marker)
	}
	return options
}

// AnalysisConfig は集計の設定
type AnalysisConfig struct {
	DBPath       string
	Table        string
	AnalysisType string // "simple", "crosstab", "multi_crosstab"

	// 集計対象列
	XColumn *Column
	YColumn *Column

	// オプション
	SplitX bool // X軸を分割（複数回答）
	SplitY bool // Y軸を分割

	// 出力
	OutputPath string // CSVエクスポート先（空なら画面表示のみ）
}

// CrosstabResult はクロス集計の結果
type CrosstabResult struct {
	XColumn string
	YColumn string
	Rows    []CrosstabRow
	Total   int
}

// CrosstabRow はクロス集計の1行
type CrosstabRow struct {
	XValue     string
	YValue     string
	Count      int
	Percentage float64 // X値内での割合
}

// SimpletabResult は単純集計の結果
type SimpletabResult struct {
	Column string
	Rows   []SimpletabRow
	Total  int
}

// SimpletabRow は単純集計の1行
type SimpletabRow struct {
	Value      string
	Count      int
	Percentage float64
}
