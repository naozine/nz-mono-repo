package analyzer

import (
	"fmt"
	"strings"
)

// Crosstab はクロス集計を実行
func (a *Analyzer) Crosstab(config AnalysisConfig) (*CrosstabResult, error) {
	// 派生列の場合は複数回答の分割に対応しない
	if config.XColumn.IsDerived {
		config.SplitX = false
	}
	if config.YColumn.IsDerived {
		config.SplitY = false
	}

	// SQLを動的に構築
	var query string

	if config.SplitX || config.SplitY {
		// 複数回答対応のクロス集計
		query = a.buildMultiAnswerCrosstabQuery(config)
	} else {
		// シンプルなクロス集計
		query = a.buildSimpleCrosstabQuery(config)
	}

	// クエリ実行
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute crosstab query: %w", err)
	}
	defer rows.Close()

	// 結果をパース
	var resultRows []CrosstabRow
	for rows.Next() {
		var row CrosstabRow
		err := rows.Scan(&row.XValue, &row.YValue, &row.Count, &row.Percentage)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		resultRows = append(resultRows, row)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// 総件数を計算
	total := 0
	for _, row := range resultRows {
		total += row.Count
	}

	result := &CrosstabResult{
		XColumn: config.XColumn.Name,
		YColumn: config.YColumn.Name,
		Rows:    resultRows,
		Total:   total,
	}

	return result, nil
}

// buildSimpleCrosstabQuery はシンプルなクロス集計のSQLを生成
func (a *Analyzer) buildSimpleCrosstabQuery(config AnalysisConfig) string {
	xExpr := config.XColumn.GetSQLExpression()
	yExpr := config.YColumn.GetSQLExpression()

	// WHERE句の構築（派生列の場合は不要）
	var whereClauses []string
	if !config.XColumn.IsDerived {
		whereClauses = append(whereClauses, fmt.Sprintf(`"%s" IS NOT NULL`, config.XColumn.Name))
	}
	if !config.YColumn.IsDerived {
		whereClauses = append(whereClauses, fmt.Sprintf(`"%s" IS NOT NULL`, config.YColumn.Name))
	}
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	return fmt.Sprintf(`
		SELECT
			%s as x_value,
			%s as y_value,
			COUNT(*) as count,
			ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(PARTITION BY %s), 1) as percentage
		FROM %s
		%s
		GROUP BY %s, %s
		ORDER BY %s, count DESC
	`,
		xExpr,
		yExpr,
		xExpr,
		a.Table,
		whereClause,
		xExpr,
		yExpr,
		xExpr,
	)
}

// buildMultiAnswerCrosstabQuery は複数回答のクロス集計のSQLを生成
func (a *Analyzer) buildMultiAnswerCrosstabQuery(config AnalysisConfig) string {
	xExpr := fmt.Sprintf(`"%s"`, config.XColumn.Name)
	if config.SplitX {
		xExpr = fmt.Sprintf(`unnest(string_split("%s", CHR(10)))`, config.XColumn.Name)
	}

	yExpr := fmt.Sprintf(`"%s"`, config.YColumn.Name)
	if config.SplitY {
		yExpr = fmt.Sprintf(`unnest(string_split("%s", CHR(10)))`, config.YColumn.Name)
	}

	return fmt.Sprintf(`
		WITH split_data AS (
			SELECT
				%s as x_value,
				%s as y_value
			FROM %s
			WHERE "%s" IS NOT NULL AND "%s" IS NOT NULL
		)
		SELECT
			x_value,
			y_value,
			COUNT(*) as count,
			ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(PARTITION BY x_value), 1) as percentage
		FROM split_data
		GROUP BY x_value, y_value
		ORDER BY x_value, count DESC
	`,
		xExpr,
		yExpr,
		a.Table,
		config.XColumn.Name,
		config.YColumn.Name,
	)
}
