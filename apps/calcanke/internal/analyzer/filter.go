package analyzer

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// FilterConfig はフィルタ設定全体
type FilterConfig struct {
	Filters []Filter `yaml:"filters"`
}

// Filter は1つのフィルタ定義
type Filter struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Conditions  []FilterCondition `yaml:"conditions"`
}

// FilterCondition はフィルタ条件
type FilterCondition struct {
	Column        string   `yaml:"column"`
	IncludeValues []string `yaml:"include_values"` // この値のみ含む
	ExcludeValues []string `yaml:"exclude_values"` // この値を除外
}

// LoadFilters は設定ファイルからフィルタを読み込む
func LoadFilters(configPath string) ([]Filter, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		// ファイルがなくてもエラーにしない
		return []Filter{}, nil
	}

	var config FilterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return config.Filters, nil
}

// SaveFilters はフィルタの定義を設定ファイルに書き込む
func SaveFilters(configPath string, filters []Filter) error {
	config := FilterConfig{
		Filters: filters,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	// ヘッダーコメントを追加
	header := "# フィルタの定義\n# この設定ファイルで、データをフィルタリングするための条件を定義できます\n\n"
	data = append([]byte(header), data...)

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GenerateWhereClause はフィルタからSQL WHERE句を生成
func (f *Filter) GenerateWhereClause(analyzer *Analyzer) string {
	var whereClauses []string

	for _, cond := range f.Conditions {
		// 列が派生列かどうか確認
		column := findColumnByName(analyzer, cond.Column)
		if column == nil {
			continue
		}

		colExpr := column.GetSQLExpression()

		// include_values がある場合
		if len(cond.IncludeValues) > 0 {
			quotedValues := make([]string, len(cond.IncludeValues))
			for i, val := range cond.IncludeValues {
				quotedValues[i] = fmt.Sprintf("'%s'", val)
			}
			whereClauses = append(whereClauses, fmt.Sprintf(
				"%s IN (%s)",
				colExpr,
				strings.Join(quotedValues, ", "),
			))
		}

		// exclude_values がある場合
		if len(cond.ExcludeValues) > 0 {
			quotedValues := make([]string, len(cond.ExcludeValues))
			for i, val := range cond.ExcludeValues {
				quotedValues[i] = fmt.Sprintf("'%s'", val)
			}
			whereClauses = append(whereClauses, fmt.Sprintf(
				"%s NOT IN (%s)",
				colExpr,
				strings.Join(quotedValues, ", "),
			))
		}
	}

	if len(whereClauses) == 0 {
		return ""
	}

	return strings.Join(whereClauses, " AND ")
}

// findColumnByName は列名からColumnを検索
func findColumnByName(analyzer *Analyzer, name string) *Column {
	columns, err := analyzer.GetColumns()
	if err != nil {
		return nil
	}

	for i := range columns {
		if columns[i].Name == name {
			return &columns[i]
		}
	}
	return nil
}
