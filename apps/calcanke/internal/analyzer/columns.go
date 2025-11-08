package analyzer

import (
	"database/sql"
	"fmt"
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

	return columns, nil
}

// detectMultiAnswer は列が複数回答かどうかを判定
func (a *Analyzer) detectMultiAnswer(columnName string) (bool, error) {
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

	// 10%以上が改行を含む場合は複数回答と判定
	ratio := float64(multiCount) / float64(total)
	return ratio > 0.1, nil
}
