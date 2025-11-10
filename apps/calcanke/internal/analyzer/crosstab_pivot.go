package analyzer

import "sort"

// ToPivot はクロス集計結果をピボット表形式に変換する
func (r *CrosstabResult) ToPivot() *CrosstabPivot {
	return r.ToPivotWithAnalyzer(nil)
}

// ToPivotWithAnalyzer は列の値の順序を考慮してピボット表形式に変換する
func (r *CrosstabResult) ToPivotWithAnalyzer(a *Analyzer) *CrosstabPivot {
	pivot := &CrosstabPivot{
		XColumn: r.XColumn,
		YColumn: r.YColumn,
		Matrix:  make(map[string]map[string]CrosstabCell),
		Total:   r.Total,
	}

	// X値とY値のユニークリストを作成
	xSet := make(map[string]bool)
	ySet := make(map[string]bool)
	for _, row := range r.Rows {
		xSet[row.XValue] = true
		ySet[row.YValue] = true
	}

	// ソートしてスライスに
	for x := range xSet {
		pivot.XValues = append(pivot.XValues, x)
	}
	for y := range ySet {
		pivot.YValues = append(pivot.YValues, y)
	}

	// ソート（カスタム順序または文字列順）
	if a != nil {
		sortByOrder(pivot.XValues, a.GetValueOrder(r.XColumn))
		sortByOrder(pivot.YValues, a.GetValueOrder(r.YColumn))
	} else {
		sort.Strings(pivot.XValues)
		sort.Strings(pivot.YValues)
	}

	// マトリックスを初期化（全てのセルをExists=falseで初期化）
	for _, x := range pivot.XValues {
		pivot.Matrix[x] = make(map[string]CrosstabCell)
		for _, y := range pivot.YValues {
			pivot.Matrix[x][y] = CrosstabCell{Exists: false}
		}
	}

	// データを埋める
	for _, row := range r.Rows {
		pivot.Matrix[row.XValue][row.YValue] = CrosstabCell{
			Count:      row.Count,
			Percentage: row.Percentage,
			Exists:     true,
		}
	}

	return pivot
}
