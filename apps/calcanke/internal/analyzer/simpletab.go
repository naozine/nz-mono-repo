package analyzer

import (
	"fmt"
	"strings"
)

// Simpletab は単純集計を実行
func (a *Analyzer) Simpletab(column *Column, split bool) (*SimpletabResult, error) {
	return a.SimpletabWithFilter(column, split, nil)
}

// SimpletabWithFilter はフィルタを適用して単純集計を実行
func (a *Analyzer) SimpletabWithFilter(column *Column, split bool, filter *Filter) (*SimpletabResult, error) {
	// 派生列の場合は、merge タイプ以外は複数回答の分割に対応しない
	if column.IsDerived && !column.IsMulti {
		split = false
	}

	// SQLを動的に構築
	var query string

	if split {
		// 複数回答対応の単純集計
		query = a.buildMultiAnswerSimpletabQuery(column, filter)
	} else {
		// シンプルな単純集計
		query = a.buildSimpleSimpletabQuery(column, filter)
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

	// カスタム順序でソート
	result.SortByAnalyzer(a)

	return result, nil
}

// buildSimpleSimpletabQuery はシンプルな単純集計のSQLを生成
func (a *Analyzer) buildSimpleSimpletabQuery(column *Column, filter *Filter) string {
	colExpr := column.GetSQLExpression()

	// WHERE句の構築
	var whereClauses []string

	// 派生列でない場合はNULL除外
	if !column.IsDerived {
		whereClauses = append(whereClauses, fmt.Sprintf(`"%s" IS NOT NULL`, column.Name))
	}

	// フィルタがある場合は追加
	if filter != nil {
		filterWhere := filter.GenerateWhereClause(a)
		if filterWhere != "" {
			whereClauses = append(whereClauses, filterWhere)
		}
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
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
func (a *Analyzer) buildMultiAnswerSimpletabQuery(column *Column, filter *Filter) string {
	// WHERE句の構築
	var whereClauses []string

	// 派生列でない場合はNULL除外
	if !column.IsDerived {
		whereClauses = append(whereClauses, fmt.Sprintf(`"%s" IS NOT NULL`, column.Name))
	}

	// フィルタがある場合は追加
	if filter != nil {
		filterWhere := filter.GenerateWhereClause(a)
		if filterWhere != "" {
			whereClauses = append(whereClauses, filterWhere)
		}
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 列式の取得
	var valueExpr string
	if column.IsDerived && column.IsMulti {
		// merge派生列の場合は '|||' で分割
		valueExpr = fmt.Sprintf(`unnest(string_split(%s, '|||'))`, column.GetSQLExpression())
	} else {
		// 通常列の場合は改行で分割
		valueExpr = fmt.Sprintf(`unnest(string_split("%s", CHR(10)))`, column.Name)
	}

	return fmt.Sprintf(`
		WITH split_data AS (
			SELECT
				%s as value
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
		valueExpr,
		a.Table,
		whereClause,
	)
}
