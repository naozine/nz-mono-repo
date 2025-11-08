# クロス集計結果の2次元表示機能 設計

## 概要

クロス集計結果を、従来のリスト形式に加えて、2次元のクロス表（ピボットテーブル）形式でも表示できるようにする。

## 現在の表示形式（リスト形式）

```
┌──────────────────────────────────┐
│ X軸 \ Y軸  │ 値     │ 件数 │ 割合 │
├──────────────────────────────────┤
│ 中3                              │
│              小学生   100   10%  │
│              中学生   200   20%  │
│              その他   50    5%   │
│ 中2                              │
│              小学生   150   15%  │
│              中学生   180   18%  │
│              その他   60    6%   │
└──────────────────────────────────┘
```

**メリット:**
- シンプルで実装が簡単
- 長いY軸値でもスクロールで対応可能
- データ構造をそのまま表示

**デメリット:**
- 全体像が把握しづらい
- X値とY値の関係性が見えにくい
- 比較がしづらい

## 新しい表示形式（クロス表/ピボット形式）

```
┌─────────────────────────────────────────┐
│       │ 小学生      │ 中学生      │ その他      │
│       │ 件数 │ 割合 │ 件数 │ 割合 │ 件数 │ 割合 │
├───────┼──────────────────────────────────┤
│ 中3   │ 100  │ 10% │ 200  │ 20% │ 50   │  5% │
│ 中2   │ 150  │ 15% │ 180  │ 18% │ 60   │  6% │
│ 中1   │ 120  │ 12% │ 160  │ 16% │ 40   │  4% │
└─────────────────────────────────────────┘
```

**メリット:**
- 全体像を一目で把握できる
- X値とY値の関係性が明確
- 値の比較がしやすい
- Excelのピボットテーブルと同じ感覚

**デメリット:**
- Y値が多い場合、横に長くなる
- 実装が複雑

## UI設計

### 表示切り替えオプション

集計結果の上部に、表示形式の切り替えボタンを配置：

```
┌─────────────────────────────────────────┐
│ クロス集計結果                          │
│ ─────────────────────────────────────   │
│ X軸: 学年 / Y軸: 学校種別               │
│                                         │
│ 表示形式: ○ リスト形式  ○ クロス表形式 │
│ ─────────────────────────────────────   │
│                                         │
│ [ここに表が表示される]                  │
└─────────────────────────────────────────┘
```

### 実装方式

#### Option 1: htmxで切り替え（推奨）

```html
<!-- 表示形式選択 -->
<div class="mb-4">
  <label class="font-medium text-gray-700">表示形式:</label>
  <div class="inline-flex rounded-md shadow-sm ml-2" role="group">
    <button type="button"
            class="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded-l-lg"
            hx-post="/api/crosstab/render"
            hx-vals='{"format": "list"}'
            hx-target="#crosstab-table"
            hx-swap="innerHTML">
      リスト形式
    </button>
    <button type="button"
            class="px-4 py-2 text-sm font-medium bg-gray-200 text-gray-700 rounded-r-lg"
            hx-post="/api/crosstab/render"
            hx-vals='{"format": "pivot"}'
            hx-target="#crosstab-table"
            hx-swap="innerHTML">
      クロス表形式
    </button>
  </div>
</div>

<div id="crosstab-table">
  <!-- ここに表が表示される -->
</div>
```

#### Option 2: _hyperscriptで切り替え（クライアントサイド）

```html
<div _="on click
         if #list-view.classList.contains('hidden')
           remove .hidden from #list-view
           add .hidden to #pivot-view
         else
           add .hidden to #list-view
           remove .hidden from #pivot-view">

  <div id="list-view">
    <!-- リスト形式の表 -->
  </div>

  <div id="pivot-view" class="hidden">
    <!-- クロス表形式の表 -->
  </div>
</div>
```

**推奨:** Option 2（クライアントサイド切り替え）
- サーバーリクエスト不要
- 高速な切り替え
- 両方のデータを最初のレスポンスに含める

## データ構造の変換

### 現在のデータ構造（行形式）

```go
type CrosstabResult struct {
    XColumn string
    YColumn string
    Rows    []CrosstabRow  // [中3→小学生, 中3→中学生, 中2→小学生, ...]
    Total   int
}

type CrosstabRow struct {
    XValue     string
    YValue     string
    Count      int
    Percentage float64
}
```

### クロス表用データ構造（マトリックス形式）

```go
// テンプレート用のピボット表示データ
type CrosstabPivot struct {
    XColumn  string
    YColumn  string
    XValues  []string              // ["中3", "中2", "中1"]
    YValues  []string              // ["小学生", "中学生", "その他"]
    Matrix   map[string]map[string]CrosstabCell  // [X値][Y値] -> Cell
    Total    int
}

type CrosstabCell struct {
    Count      int
    Percentage float64
    Exists     bool  // データが存在するか（0件とデータなしを区別）
}
```

### 変換ロジック

```go
// CrosstabResultからCrosstabPivotへ変換
func (r *CrosstabResult) ToPivot() *CrosstabPivot {
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
    sort.Strings(pivot.XValues)
    sort.Strings(pivot.YValues)

    // マトリックスを初期化
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
```

## テンプレート実装

### 新しいテンプレート: `crosstab_pivot.html`

```html
{{define "crosstab_pivot.html"}}
<div class="overflow-x-auto">
    <table class="min-w-full divide-y divide-gray-200 border border-gray-300">
        <thead class="bg-gray-50">
            <tr>
                <!-- 左上の空セル -->
                <th class="px-4 py-3 border border-gray-300 bg-gray-100">
                    <div class="text-xs font-medium text-gray-500">
                        {{.XColumn}} \ {{.YColumn}}
                    </div>
                </th>

                <!-- Y値のヘッダー（各Y値に2列: 件数と割合） -->
                {{range .YValues}}
                <th colspan="2" class="px-4 py-3 border border-gray-300 text-center">
                    <div class="text-sm font-medium text-gray-900">{{.}}</div>
                </th>
                {{end}}
            </tr>
            <tr class="bg-gray-50">
                <th class="px-4 py-2 border border-gray-300"></th>
                {{range .YValues}}
                <th class="px-3 py-2 border border-gray-300 text-xs text-gray-500 text-right">件数</th>
                <th class="px-3 py-2 border border-gray-300 text-xs text-gray-500 text-right">割合</th>
                {{end}}
            </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
            {{range $x := .XValues}}
            <tr>
                <!-- X値（行ヘッダー） -->
                <td class="px-4 py-3 border border-gray-300 font-medium text-gray-900 bg-gray-50">
                    {{$x}}
                </td>

                <!-- 各Y値のセル -->
                {{range $y := $.YValues}}
                {{$cell := index (index $.Matrix $x) $y}}
                {{if $cell.Exists}}
                <td class="px-3 py-3 border border-gray-300 text-right text-sm">
                    {{$cell.Count}}
                </td>
                <td class="px-3 py-3 border border-gray-300 text-right text-sm">
                    {{printf "%.1f%%" $cell.Percentage}}
                </td>
                {{else}}
                <td class="px-3 py-3 border border-gray-300 text-right text-sm text-gray-400">
                    -
                </td>
                <td class="px-3 py-3 border border-gray-300 text-right text-sm text-gray-400">
                    -
                </td>
                {{end}}
                {{end}}
            </tr>
            {{end}}
        </tbody>
    </table>
</div>
{{end}}
```

### 既存テンプレートの更新: `crosstab_result.html`

```html
{{define "crosstab_result.html"}}
<div class="space-y-4">
    <!-- ヘッダー情報 -->
    <div class="border-b border-gray-200 pb-4">
        <h3 class="text-lg font-semibold text-gray-900">クロス集計結果</h3>
        <div class="mt-2 text-sm text-gray-600 space-y-1">
            <!-- 既存の情報 -->
        </div>
    </div>

    <!-- 表示形式切り替え -->
    <div class="flex items-center space-x-4">
        <span class="text-sm font-medium text-gray-700">表示形式:</span>
        <div class="inline-flex rounded-md shadow-sm" role="group">
            <button type="button" id="list-view-btn"
                    class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-l-lg hover:bg-blue-700"
                    _="on click
                       add .bg-blue-600 .text-white to me
                       remove .bg-blue-600 .text-white from #pivot-view-btn
                       add .bg-gray-200 .text-gray-700 to #pivot-view-btn
                       remove .hidden from #list-table
                       add .hidden to #pivot-table">
                リスト形式
            </button>
            <button type="button" id="pivot-view-btn"
                    class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-200 rounded-r-lg hover:bg-gray-300"
                    _="on click
                       add .bg-blue-600 .text-white to me
                       remove .bg-blue-600 .text-white from #list-view-btn
                       add .bg-gray-200 .text-gray-700 to #list-view-btn
                       add .hidden to #list-table
                       remove .hidden from #pivot-table">
                クロス表形式
            </button>
        </div>
    </div>

    <!-- リスト形式の表 -->
    <div id="list-table">
        {{template "crosstab_list.html" .}}
    </div>

    <!-- クロス表形式の表 -->
    <div id="pivot-table" class="hidden">
        {{template "crosstab_pivot.html" .Pivot}}
    </div>

    <!-- エクスポートボタン -->
    <div class="flex justify-end">
        <button type="button" class="..." disabled>
            CSV エクスポート（準備中）
        </button>
    </div>
</div>
{{end}}
```

### リスト形式を分離: `crosstab_list.html`

既存の `crosstab_result.html` の表部分を `crosstab_list.html` に分離

## ハンドラーの修正

```go
// CrosstabResultData はクロス集計結果のテンプレートデータ
type CrosstabResultData struct {
    Result *analyzer.CrosstabResult
    Pivot  *CrosstabPivot  // 追加
    Filter *analyzer.Filter
}

// Crosstab はクロス集計を実行する
func (h *Handler) Crosstab(c echo.Context) error {
    // ... 既存のコード ...

    // ピボット形式のデータも生成
    pivot := result.ToPivot()

    data := CrosstabResultData{
        Result: result,
        Pivot:  pivot,  // 追加
        Filter: filter,
    }

    return c.Render(http.StatusOK, "crosstab_result.html", data)
}
```

## 実装の優先順位

### Phase 1: 基本実装
1. ✅ `CrosstabPivot` データ構造の定義
2. ✅ `ToPivot()` 変換メソッドの実装
3. ✅ `crosstab_pivot.html` テンプレート作成
4. ✅ `crosstab_list.html` テンプレート作成（既存を分離）
5. ✅ 表示切り替えボタンの実装（_hyperscript）
6. ✅ ハンドラーの修正

### Phase 2: 改善
1. ソート順の改善（数値順、カスタムソート等）
2. 合計行/合計列の追加
3. ヒートマップ表示（割合に応じて背景色を変更）
4. 列幅の自動調整

### Phase 3: 拡張
1. セルクリックで詳細表示
2. エクスポート時のフォーマット選択

## 注意点

1. **Y値が多い場合の対応**
   - 横スクロール可能にする（`overflow-x-auto`）
   - Y値が10個以上の場合は警告表示？

2. **パフォーマンス**
   - 大量データの場合、両方の形式を同時に生成すると重い
   - 初期表示はリスト形式のみ、切り替え時にピボット形式を生成する選択肢も

3. **空セルの表示**
   - データが存在しないセルは `-` で表示
   - 0件のセルと区別する

## まとめ

- **実装方式:** クライアントサイド切り替え（_hyperscript）
- **データ変換:** `ToPivot()` メソッドでマトリックス形式に変換
- **表示:** Tailwind CSSでスタイリングされたHTMLテーブル
- **既存機能:** リスト形式は残し、切り替え可能に

この設計で進めてよろしいですか？
