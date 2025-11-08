package analyzer

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// DerivedColumnConfig は派生列の設定全体
type DerivedColumnConfig struct {
	DerivedColumns []DerivedColumn `yaml:"derived_columns"`
}

// DerivedColumn は1つの派生列の定義
type DerivedColumn struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	SourceColumns []string `yaml:"source_columns"`
	Rules         []Rule   `yaml:"rules"`
}

// Rule は分類ルール
type Rule struct {
	Label      string      `yaml:"label"`
	Conditions []Condition `yaml:"conditions"`
	IsDefault  bool        `yaml:"is_default"`
}

// Condition は条件
type Condition struct {
	Column   string   `yaml:"column"`
	Operator string   `yaml:"operator"`
	Value    string   `yaml:"value"`
	Values   []string `yaml:"values"`
}

// LoadDerivedColumns は設定ファイルから派生列の定義を読み込む
func LoadDerivedColumns(configPath string) ([]DerivedColumn, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config DerivedColumnConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return config.DerivedColumns, nil
}

// GenerateCaseExpression は派生列のSQL CASE式を生成
func (dc *DerivedColumn) GenerateCaseExpression() string {
	var cases []string

	for _, rule := range dc.Rules {
		// デフォルトルールはELSEとして最後に処理
		if rule.IsDefault {
			continue
		}

		// 条件を生成
		var conditions []string
		for _, cond := range rule.Conditions {
			condSQL := generateConditionSQL(cond)
			if condSQL != "" {
				conditions = append(conditions, condSQL)
			}
		}

		// 条件が複数ある場合はANDで結合
		var whenClause string
		if len(conditions) > 0 {
			whenClause = strings.Join(conditions, " AND ")
			cases = append(cases, fmt.Sprintf("WHEN %s THEN '%s'", whenClause, rule.Label))
		}
	}

	// デフォルトルールを探す
	var elseClause string
	for _, rule := range dc.Rules {
		if rule.IsDefault {
			elseClause = fmt.Sprintf("ELSE '%s'", rule.Label)
			break
		}
	}

	// CASE式を組み立て
	caseExpr := "CASE\n  " + strings.Join(cases, "\n  ") + "\n  " + elseClause + "\nEND"
	return caseExpr
}

// generateConditionSQL は条件からSQL文を生成
func generateConditionSQL(cond Condition) string {
	switch cond.Operator {
	case "equals":
		return fmt.Sprintf(`TRIM("%s") = '%s'`, cond.Column, cond.Value)

	case "starts_with":
		return fmt.Sprintf(`TRIM("%s") LIKE '%s%%'`, cond.Column, cond.Value)

	case "starts_with_any":
		var conditions []string
		for _, val := range cond.Values {
			conditions = append(conditions, fmt.Sprintf(`TRIM("%s") LIKE '%s%%'`, cond.Column, val))
		}
		return "(" + strings.Join(conditions, " OR ") + ")"

	case "contains":
		return fmt.Sprintf(`TRIM("%s") LIKE '%%%s%%'`, cond.Column, cond.Value)

	default:
		return ""
	}
}

// GetDerivedColumn は派生列を仮想的なColumnとして返す
func (dc *DerivedColumn) GetDerivedColumn(index int) Column {
	return Column{
		Index:     index,
		Name:      dc.Name,
		Type:      "VARCHAR (派生列)",
		IsMulti:   false,
		IsDerived: true,
		SQLExpr:   dc.GenerateCaseExpression(),
	}
}
