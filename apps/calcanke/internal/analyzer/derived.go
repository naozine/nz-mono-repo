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
	Name            string                 `yaml:"name"`
	Description     string                 `yaml:"description"`
	SourceColumns   []string               `yaml:"source_columns"`
	CalculationType string                 `yaml:"calculation_type"` // "rules" または "grade_from_birthdate"
	Parameters      map[string]interface{} `yaml:"parameters"`       // 計算パラメータ
	Rules           []Rule                 `yaml:"rules"`            // calculation_type="rules"の場合
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

// SaveDerivedColumns は派生列の定義を設定ファイルに書き込む
func SaveDerivedColumns(configPath string, columns []DerivedColumn) error {
	config := DerivedColumnConfig{
		DerivedColumns: columns,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	// ヘッダーコメントを追加
	header := "# 派生列の定義\n# この設定ファイルで、既存の列から新しい列を動的に生成できます\n\n"
	data = append([]byte(header), data...)

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GenerateCaseExpression は派生列のSQL CASE式を生成
func (dc *DerivedColumn) GenerateCaseExpression() string {
	// calculation_typeに応じて処理を分岐
	switch dc.CalculationType {
	case "grade_from_birthdate":
		return dc.generateGradeCalculation()
	case "school_type_from_birthdate":
		return dc.generateSchoolTypeCalculation()
	case "merge":
		return dc.generateMergeExpression()
	case "rules", "":
		// デフォルトはルールベース
		return dc.generateRuleBasedExpression()
	default:
		return dc.generateRuleBasedExpression()
	}
}

// generateRuleBasedExpression はルールベースのCASE式を生成
func (dc *DerivedColumn) generateRuleBasedExpression() string {
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

	case "between":
		// Values[0]: 開始値, Values[1]: 終了値
		if len(cond.Values) >= 2 {
			return fmt.Sprintf(`"%s" >= '%s' AND "%s" <= '%s'`,
				cond.Column, cond.Values[0], cond.Column, cond.Values[1])
		}
		return ""

	case "in":
		// 複数値のいずれかに一致（OR条件）
		if len(cond.Values) > 0 {
			quotedValues := make([]string, len(cond.Values))
			for i, val := range cond.Values {
				quotedValues[i] = fmt.Sprintf("'%s'", val)
			}
			return fmt.Sprintf(`"%s" IN (%s)`, cond.Column, strings.Join(quotedValues, ", "))
		}
		return ""

	default:
		return ""
	}
}

// generateGradeCalculation は生年月日から学年を計算するSQL式を生成
func (dc *DerivedColumn) generateGradeCalculation() string {
	// パラメータから対象年度を取得（デフォルトは2025）
	targetYear := 2025
	if year, ok := dc.Parameters["target_year"]; ok {
		if y, ok := year.(int); ok {
			targetYear = y
		}
	}

	// 生年月日列名を取得
	birthdateColumn := "生年月日"
	if col, ok := dc.Parameters["birthdate_column"]; ok {
		if c, ok := col.(string); ok {
			birthdateColumn = c
		}
	}

	// 各学年の生年月日範囲を計算
	// 2025年度の小1 = 2018/04/02 〜 2019/04/01
	// YYYYMMDD形式で比較
	type GradeRange struct {
		Label     string
		StartDate string // YYYYMMDD形式（この日以降）
		EndDate   string // YYYYMMDD形式（この日以前）
	}

	var ranges []GradeRange

	// 小1〜小6
	for grade := 1; grade <= 6; grade++ {
		birthYear := targetYear - grade - 6 // 小1は7歳になる年度
		startDate := fmt.Sprintf("%04d0402", birthYear)
		endDate := fmt.Sprintf("%04d0401", birthYear+1)
		ranges = append(ranges, GradeRange{
			Label:     fmt.Sprintf("小%d", grade),
			StartDate: startDate,
			EndDate:   endDate,
		})
	}

	// 中1〜中3
	for grade := 1; grade <= 3; grade++ {
		birthYear := targetYear - grade - 12 // 中1は13歳になる年度
		startDate := fmt.Sprintf("%04d0402", birthYear)
		endDate := fmt.Sprintf("%04d0401", birthYear+1)
		ranges = append(ranges, GradeRange{
			Label:     fmt.Sprintf("中%d", grade),
			StartDate: startDate,
			EndDate:   endDate,
		})
	}

	// CASE式を構築
	var cases []string

	for _, r := range ranges {
		cases = append(cases, fmt.Sprintf(
			"WHEN \"%s\" >= '%s' AND \"%s\" <= '%s' THEN '%s'",
			birthdateColumn, r.StartDate, birthdateColumn, r.EndDate, r.Label,
		))
	}

	// 小1未満（最も新しい小1の範囲より後）
	youngest := ranges[0] // 小1
	cases = append(cases, fmt.Sprintf(
		"WHEN \"%s\" > '%s' THEN '小1未満'",
		birthdateColumn, youngest.EndDate,
	))

	// 高1以上（最も古い中3の範囲より前）
	oldest := ranges[len(ranges)-1] // 中3
	cases = append(cases, fmt.Sprintf(
		"WHEN \"%s\" < '%s' THEN '高1以上'",
		birthdateColumn, oldest.StartDate,
	))

	// NULL対応
	cases = append(cases, fmt.Sprintf(
		"WHEN \"%s\" IS NULL THEN 'データなし'",
		birthdateColumn,
	))

	// CASE式を組み立て
	caseExpr := "CASE\n  " + strings.Join(cases, "\n  ") + "\n  ELSE 'データ不正'\nEND"
	return caseExpr
}

// generateSchoolTypeCalculation は生年月日から学校種別を計算するSQL式を生成
func (dc *DerivedColumn) generateSchoolTypeCalculation() string {
	// パラメータから対象年度を取得（デフォルトは2025）
	targetYear := 2025
	if year, ok := dc.Parameters["target_year"]; ok {
		if y, ok := year.(int); ok {
			targetYear = y
		}
	}

	// 生年月日列名を取得
	birthdateColumn := "生年月日"
	if col, ok := dc.Parameters["birthdate_column"]; ok {
		if c, ok := col.(string); ok {
			birthdateColumn = c
		}
	}

	// カスタムラベルを取得（オプション）
	elementaryLabel := "小学生"
	juniorHighLabel := "中学生"
	otherLabel := "その他"

	if labels, ok := dc.Parameters["labels"]; ok {
		if labelMap, ok := labels.(map[string]interface{}); ok {
			if el, ok := labelMap["elementary"].(string); ok {
				elementaryLabel = el
			}
			if jh, ok := labelMap["junior_high"].(string); ok {
				juniorHighLabel = jh
			}
			if ot, ok := labelMap["other"].(string); ok {
				otherLabel = ot
			}
		}
	}

	// 小学校の範囲を計算（小1〜小6）
	// 小1は7歳になる年度、小6は12歳になる年度
	// 小1: targetYear - 1 - 6 = targetYear - 7 歳の生年月日
	// 小6: targetYear - 6 - 6 = targetYear - 12 歳の生年月日
	elemStart := fmt.Sprintf("%04d0402", targetYear-12) // 小6の開始日（最も古い）
	elemEnd := fmt.Sprintf("%04d0401", targetYear-6)    // 小1の終了日（最も新しい）

	// 中学校の範囲を計算（中1〜中3）
	// 中1は13歳になる年度、中3は15歳になる年度
	juniorStart := fmt.Sprintf("%04d0402", targetYear-15) // 中3の開始日（最も古い）
	juniorEnd := fmt.Sprintf("%04d0401", targetYear-12)   // 中1の終了日（最も新しい）

	// CASE式を構築
	var cases []string

	// 小学生
	cases = append(cases, fmt.Sprintf(
		"WHEN \"%s\" >= '%s' AND \"%s\" <= '%s' THEN '%s'",
		birthdateColumn, elemStart, birthdateColumn, elemEnd, elementaryLabel,
	))

	// 中学生
	cases = append(cases, fmt.Sprintf(
		"WHEN \"%s\" >= '%s' AND \"%s\" <= '%s' THEN '%s'",
		birthdateColumn, juniorStart, birthdateColumn, juniorEnd, juniorHighLabel,
	))

	// NULL対応
	cases = append(cases, fmt.Sprintf(
		"WHEN \"%s\" IS NULL THEN NULL",
		birthdateColumn,
	))

	// CASE式を組み立て
	caseExpr := fmt.Sprintf("CASE\n  %s\n  ELSE '%s'\nEND", strings.Join(cases, "\n  "), otherLabel)
	return caseExpr
}

// generateMergeExpression は複数列を結合するSQL式を生成
func (dc *DerivedColumn) generateMergeExpression() string {
	// パラメータからセパレータを取得（デフォルトは"|||"）
	separator := "|||"
	if sep, ok := dc.Parameters["separator"]; ok {
		if s, ok := sep.(string); ok {
			separator = s
		}
	}

	// ソース列が指定されていない場合はエラー
	if len(dc.SourceColumns) == 0 {
		return "NULL"
	}

	// 各列をNULLIF(TRIM(列), '')でラップして空文字を除外
	var columns []string
	for _, colName := range dc.SourceColumns {
		columns = append(columns, fmt.Sprintf(`NULLIF(TRIM("%s"), '')`, colName))
	}

	// CONCAT_WSで結合（NULLは自動的にスキップされる）
	// 結果が空文字列の場合はNULLに変換
	expr := fmt.Sprintf("NULLIF(TRIM(CONCAT_WS('%s', %s)), '')",
		separator,
		strings.Join(columns, ", "))

	return expr
}

// GetDerivedColumn は派生列を仮想的なColumnとして返す
func (dc *DerivedColumn) GetDerivedColumn(index int) Column {
	// mergeタイプの場合は複数回答として扱う
	isMulti := dc.CalculationType == "merge"

	return Column{
		Index:     index,
		Name:      dc.Name,
		Type:      "VARCHAR (派生列)",
		IsMulti:   isMulti,
		IsDerived: true,
		SQLExpr:   dc.GenerateCaseExpression(),
	}
}
