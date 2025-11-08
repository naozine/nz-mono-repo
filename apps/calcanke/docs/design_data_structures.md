# データ構造設計

## 主要な型定義

### 列情報

```go
// Column は列の情報を保持
type Column struct {
    Index       int    // 1始まりの列番号
    Name        string // 列名
    Type        string // データ型（VARCHAR, INTEGER等）
    IsMulti     bool   // 複数回答かどうか（改行含む割合で判定）
    SampleValue string // サンプル値
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
```

### 集計設定

```go
// AnalysisConfig は集計の設定
type AnalysisConfig struct {
    DBPath      string
    Table       string
    AnalysisType string // "simple", "crosstab", "multi_crosstab"

    // 集計対象列
    XColumn     *Column
    YColumn     *Column

    // オプション
    SplitX      bool   // X軸を分割（複数回答）
    SplitY      bool   // Y軸を分割

    // 出力
    OutputPath  string // CSVエクスポート先（空なら画面表示のみ）
}
```

### 集計結果

```go
// CrosstabResult はクロス集計の結果
type CrosstabResult struct {
    XColumn string
    YColumn string
    Rows    []CrosstabRow
    Total   int
}

// CrosstabRow はクロス集計の1行
type CrosstabRow struct {
    XValue      string
    YValue      string
    Count       int
    Percentage  float64 // X値内での割合
}

// ToTable は tablewriter 用のデータに変換
func (r CrosstabResult) ToTable() [][]string {
    rows := make([][]string, len(r.Rows))
    for i, row := range r.Rows {
        rows[i] = []string{
            row.XValue,
            row.YValue,
            strconv.Itoa(row.Count),
            fmt.Sprintf("%.1f%%", row.Percentage),
        }
    }
    return rows
}
```

## DuckDB操作

```go
// Analyzer はデータ分析を実行
type Analyzer struct {
    db     *sql.DB
    dbPath string
    table  string
}

// NewAnalyzer はAnalyzerを作成
func NewAnalyzer(dbPath, table string) (*Analyzer, error) {
    db, err := sql.Open("duckdb", dbPath)
    if err != nil {
        return nil, err
    }
    return &Analyzer{
        db:     db,
        dbPath: dbPath,
        table:  table,
    }, nil
}

// GetColumns は全列の情報を取得
func (a *Analyzer) GetColumns() (ColumnList, error) {
    // 1. DESCRIBE でスキーマ取得
    // 2. サンプルデータで複数回答判定
    // 3. ユニーク数を計算
}

// Crosstab はクロス集計を実行
func (a *Analyzer) Crosstab(config AnalysisConfig) (*CrosstabResult, error) {
    // SQLを動的に構築して実行
    var query string

    if config.SplitX || config.SplitY {
        // WITH句で分割処理
        query = buildMultiAnswerCrosstabQuery(config)
    } else {
        // シンプルなGROUP BY
        query = buildSimpleCrosstabQuery(config)
    }

    rows, err := a.db.Query(query)
    // 結果をパース
}
```

## SQL生成ロジック

```go
// buildSimpleCrosstabQuery はシンプルなクロス集計のSQLを生成
func buildSimpleCrosstabQuery(config AnalysisConfig) string {
    return fmt.Sprintf(`
        SELECT
            "%s" as x_value,
            "%s" as y_value,
            COUNT(*) as count,
            ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(PARTITION BY "%s"), 1) as percentage
        FROM %s
        WHERE "%s" IS NOT NULL AND "%s" IS NOT NULL
        GROUP BY "%s", "%s"
        ORDER BY "%s", count DESC
    `,
        config.XColumn.Name,
        config.YColumn.Name,
        config.XColumn.Name,
        config.Table,
        config.XColumn.Name,
        config.YColumn.Name,
        config.XColumn.Name,
        config.YColumn.Name,
        config.XColumn.Name,
    )
}

// buildMultiAnswerCrosstabQuery は複数回答のクロス集計のSQLを生成
func buildMultiAnswerCrosstabQuery(config AnalysisConfig) string {
    xExpr := config.XColumn.Name
    if config.SplitX {
        xExpr = fmt.Sprintf("unnest(string_split(\"%s\", CHR(10)))", config.XColumn.Name)
    }

    yExpr := config.YColumn.Name
    if config.SplitY {
        yExpr = fmt.Sprintf("unnest(string_split(\"%s\", CHR(10)))", config.YColumn.Name)
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
        config.Table,
        config.XColumn.Name,
        config.YColumn.Name,
    )
}
```

## 複数回答の自動判定

```go
// detectMultiAnswer は列が複数回答かどうかを判定
func (a *Analyzer) detectMultiAnswer(columnName string) (bool, error) {
    query := fmt.Sprintf(`
        SELECT
            COUNT(*) as total,
            SUM(CASE WHEN POSITION(CHR(10) IN "%s") > 0 THEN 1 ELSE 0 END) as multi_count
        FROM %s
        WHERE "%s" IS NOT NULL
    `, columnName, a.table, columnName)

    var total, multiCount int
    err := a.db.QueryRow(query).Scan(&total, &multiCount)
    if err != nil {
        return false, err
    }

    // 10%以上が改行を含む場合は複数回答と判定
    ratio := float64(multiCount) / float64(total)
    return ratio > 0.1, nil
}
```

## エクスポート

```go
// ExportCSV は結果をCSVにエクスポート
func (r CrosstabResult) ExportCSV(path string) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    w := csv.NewWriter(f)
    defer w.Flush()

    // ヘッダー
    w.Write([]string{r.XColumn, r.YColumn, "件数", "割合"})

    // データ
    for _, row := range r.Rows {
        w.Write([]string{
            row.XValue,
            row.YValue,
            strconv.Itoa(row.Count),
            fmt.Sprintf("%.1f%%", row.Percentage),
        })
    }

    return nil
}
```

## メイン処理フロー

```go
func runInteractiveAnalysis() error {
    // 1. Analyzerを初期化
    analyzer, err := NewAnalyzer("data/app.duckdb", "excel_import")
    if err != nil {
        return err
    }
    defer analyzer.Close()

    // 2. 列情報を取得
    columns, err := analyzer.GetColumns()
    if err != nil {
        return err
    }

    // 3. 分析タイプを選択
    var analysisType string
    survey.AskOne(&survey.Select{
        Message: "分析タイプを選択:",
        Options: []string{"単純集計", "クロス集計", "複数回答のクロス集計"},
    }, &analysisType)

    // 4. 列を選択
    config := AnalysisConfig{}
    if analysisType == "クロス集計" || analysisType == "複数回答のクロス集計" {
        // X軸選択
        var xIndex int
        survey.AskOne(&survey.Select{
            Message: "X軸の列を選択:",
            Options: columns.ToOptions(),
        }, &xIndex)
        config.XColumn = &columns[xIndex]

        // Y軸選択
        var yIndex int
        survey.AskOne(&survey.Select{
            Message: "Y軸の列を選択:",
            Options: columns.ToOptions(),
        }, &yIndex)
        config.YColumn = &columns[yIndex]

        // 複数回答確認
        if config.XColumn.IsMulti {
            survey.AskOne(&survey.Confirm{
                Message: "X軸を複数回答として分割しますか？",
            }, &config.SplitX)
        }
        if config.YColumn.IsMulti {
            survey.AskOne(&survey.Confirm{
                Message: "Y軸を複数回答として分割しますか？",
            }, &config.SplitY)
        }
    }

    // 5. 集計実行
    result, err := analyzer.Crosstab(config)
    if err != nil {
        return err
    }

    // 6. 結果表示
    displayTable(result)

    // 7. エクスポート確認
    var doExport bool
    survey.AskOne(&survey.Confirm{
        Message: "CSVでエクスポートしますか？",
    }, &doExport)

    if doExport {
        var outputPath string
        survey.AskOne(&survey.Input{
            Message: "出力ファイル名:",
            Default: "data/crosstab_result.csv",
        }, &outputPath)

        result.ExportCSV(outputPath)
        fmt.Printf("✓ エクスポート完了: %s\n", outputPath)
    }

    return nil
}
```

## エラーハンドリング

```go
// AppError はアプリケーションエラー
type AppError struct {
    Type    string // "db", "user_input", "export"
    Message string
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}
```
