package analyzer

import (
	"fmt"
)

// Simpletab は単純集計を実行
func (a *Analyzer) Simpletab(column *Column, split bool) (*SimpletabResult, error) {
	// 派生列の場合は複数回答の分割に対応しない
	if column.IsDerived {
		split = false
	}

	// SQLを動的に構築
	var query string

	if split {
		// 複数回答対応の単純集計
		query = a.buildMultiAnswerSimpletabQuery(column)
	} else {
		// シンプルな単純集計
		query = a.buildSimpleSimpletabQuery(column)
	}

	// クエリ実行
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute simpletab query: %w", err)
	}
	defer rows.Close()

	// 結果をパース
	var resultRows []SimpletabRow
	total := 0
	for rows.Next() {
		var row SimpletabRow
		err := rows.Scan(&row.Value, &row.Count, &row.Percentage)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		resultRows = append(resultRows, row)
		total += row.Count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	result := &SimpletabResult{
		Column: column.Name,
		Rows:   resultRows,
		Total:  total,
	}

	return result, nil
}

// buildSimpleSimpletabQuery はシンプルな単純集計のSQLを生成
func (a *Analyzer) buildSimpleSimpletabQuery(column *Column) string {
	colExpr := column.GetSQLExpression()

	// 派生列の場合、WHERE句は不要（CASE式でNULLハンドリングされる）
	whereClause := ""
	if !column.IsDerived {
		whereClause = fmt.Sprintf(`WHERE "%s" IS NOT NULL`, column.Name)
	}

	return fmt.Sprintf(`
		SELECT
			%s as value,
			COUNT(*) as count,
			ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as percentage
		FROM %s
		%s
		GROUP BY %s
		ORDER BY count DESC
	`,
		colExpr,
		a.Table,
		whereClause,
		colExpr,
	)
}

// buildMultiAnswerSimpletabQuery は複数回答の単純集計のSQLを生成
// 注意: 派生列の場合は呼ばれない（Simpletab関数でチェック済み）
func (a *Analyzer) buildMultiAnswerSimpletabQuery(column *Column) string {
	// 派生列でないことを前提とする
	whereClause := ""
	if !column.IsDerived {
		whereClause = fmt.Sprintf(`WHERE "%s" IS NOT NULL`, column.Name)
	}

	return fmt.Sprintf(`
		WITH split_data AS (
			SELECT
				unnest(string_split("%s", CHR(10))) as value
			FROM %s
			%s
		)
		SELECT
			value,
			COUNT(*) as count,
			ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as percentage
		FROM split_data
		GROUP BY value
		ORDER BY count DESC
	`,
		column.Name,
		a.Table,
		whereClause,
	)
}
