package analyzer

import (
	"database/sql"
	"fmt"
	"strings"
)

// GetColumns は全列の情報を取得
func (a *Analyzer) GetColumns() (ColumnList, error) {
	// 1. DESCRIBE でスキーマ取得
	query := fmt.Sprintf("DESCRIBE %s", a.Table)
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to describe table: %w", err)
	}
	defer rows.Close()

	var columns ColumnList
	index := 1

	for rows.Next() {
		var columnName, columnType string
		var null, key, defaultVal, extra sql.NullString
		err := rows.Scan(&columnName, &columnType, &null, &key, &defaultVal, &extra)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		column := Column{
			Index: index,
			Name:  columnName,
			Type:  columnType,
		}

		// 2. 複数回答かどうか判定（改行を含むデータの割合をチェック）
		isMulti, err := a.detectMultiAnswer(columnName)
		if err != nil {
			// エラーは無視して続行（複数回答判定は参考情報）
			isMulti = false
		}
		column.IsMulti = isMulti

		columns = append(columns, column)
		index++
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// 派生列を追加
	for _, derivedCol := range a.DerivedColumns {
		col := derivedCol.GetDerivedColumn(index)
		columns = append(columns, col)
		index++
	}

	return columns, nil
}

// detectMultiAnswer は列が複数回答かどうかを判定
func (a *Analyzer) detectMultiAnswer(columnName string) (bool, error) {
	// 列名に複数回答を示すキーワードが含まれている場合は複数回答と判定
	multiKeywords := []string{
		"複数回答可",
		"複数回答",
		"（複数選択可）",
		"複数選択可",
	}
	for _, keyword := range multiKeywords {
		if strings.Contains(columnName, keyword) {
			return true, nil
		}
	}

	// データの内容から判定
	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN POSITION(CHR(10) IN "%s") > 0 THEN 1 ELSE 0 END) as multi_count
		FROM %s
		WHERE "%s" IS NOT NULL
	`, columnName, a.Table, columnName)

	var total, multiCount int
	err := a.db.QueryRow(query).Scan(&total, &multiCount)
	if err != nil {
		return false, err
	}

	// データが0件の場合は複数回答ではない
	if total == 0 {
		return false, nil
	}

	// 5%以上が改行を含む場合は複数回答と判定（閾値を10%から5%に下げる）
	ratio := float64(multiCount) / float64(total)
	return ratio > 0.05, nil
}
