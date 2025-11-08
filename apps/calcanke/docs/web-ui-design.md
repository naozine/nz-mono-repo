# Calcanke Web UI 設計計画

## 概要

CalcankeにWeb UIを追加し、ブラウザからアンケートデータの集計・分析を実行できるようにする。

## 技術スタック

### Backend
- **言語**: Go 1.21+
- **Webフレームワーク**: Echo v4
- **テンプレートエンジン**: `html/template`
- **既存ロジック**: `internal/analyzer` パッケージを再利用

### Frontend
- **htmx**: 2.0.x (CDN) - サーバーサイドレンダリングの動的更新
- **_hyperscript**: 0.9.x (CDN) - クライアントサイドのインタラクティブ処理
- **Tailwind CSS**: 3.x (CDN) - スタイリング
- **Alpine.js**: 使用しない

### データベース
- **DuckDB**: 既存のデータベース（`data/app.duckdb`）を使用

## 基本方針

1. **認証なし**: とりあえず認証機能は実装しない（ローカル環境での利用を想定）
2. **サーバーサイドレンダリング**: htmxでHTMLフラグメントを動的に取得
3. **既存ロジックの再利用**: CLIで実装済みの分析ロジックを最大限活用
4. **段階的実装**: 基本機能から順次拡張

## 機能要件

### Phase 1: 基本機能（MVP）

#### 1.1 データベース選択
- データベースファイルとテーブル名の選択
- デフォルト: `data/app.duckdb` / `excel_import`

#### 1.2 単純集計
- 集計列の選択
- フィルタの選択（オプション）
- 複数回答分割の設定
- 集計結果のテーブル表示

#### 1.3 クロス集計
- X軸列とY軸列の選択
- フィルタの選択（オプション）
- 複数回答分割の設定（X軸・Y軸それぞれ）
- 集計結果のテーブル表示

### Phase 2: 拡張機能

#### 2.1 データエクスポート
- 集計結果のCSVダウンロード

#### 2.2 集計履歴
- 実行した集計の履歴を表示
- 過去の集計設定を再実行

#### 2.3 可視化
- 簡単なグラフ表示（Chart.js等）

### Phase 3: 運用機能

#### 3.1 データインポート
- Web UIからExcelファイルをアップロード
- DuckDBへのインポート実行

#### 3.2 設定管理
- 派生列・フィルタ設定のブラウザ上での編集

## 画面設計

### メイン画面（分析画面）

```
┌────────────────────────────────────────────────────────────┐
│ Calcanke - アンケートデータ分析ツール              [設定] │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────────┐  ┌──────────────────────────────┐  │
│  │  設定パネル      │  │  結果表示エリア               │  │
│  │                 │  │                              │  │
│  │ [ 分析タイプ ]  │  │  ┌────────────────────────┐  │  │
│  │  ○ 単純集計    │  │  │ 総件数: 19,061件       │  │  │
│  │  ○ クロス集計  │  │  └────────────────────────┘  │  │
│  │                 │  │                              │  │
│  │ [ 集計列 ]      │  │  ┌────────────────────────┐  │  │
│  │  ▼ 学年         │  │  │ 値      | 件数 | 割合  │  │  │
│  │                 │  │  ├────────────────────────┤  │  │
│  │ [ フィルタ ]    │  │  │ 中3     | 7,282| 38.2% │  │  │
│  │  ▼ なし         │  │  │ 中2     | 4,432| 23.3% │  │  │
│  │                 │  │  │ 小4     | 1,645|  8.6% │  │  │
│  │ [ オプション ]  │  │  │ ...     | ...  | ...   │  │  │
│  │  □ 複数回答分割 │  │  └────────────────────────┘  │  │
│  │                 │  │                              │  │
│  │  [集計実行]     │  │  [CSV エクスポート]          │  │
│  │                 │  │                              │  │
│  └─────────────────┘  └──────────────────────────────┘  │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### レイアウト

- **2カラムレイアウト**: 左側に設定パネル、右側に結果表示
- **レスポンシブ**: スマホでは1カラム（設定→結果の順）
- **固定ヘッダー**: アプリケーション名と基本情報

## ディレクトリ構造

```
apps/calcanke/
├── cmd/
│   ├── calcanke/        # CLI（既存）
│   └── calcanke-web/    # Webサーバー（新規）
│       └── main.go
├── internal/
│   ├── analyzer/        # 分析ロジック（既存）
│   ├── ui/             # CLI UI（既存）
│   └── web/            # Web UI（新規）
│       ├── handlers/   # HTTPハンドラー
│       │   ├── home.go
│       │   ├── simpletab.go
│       │   ├── crosstab.go
│       │   └── export.go
│       ├── middleware/ # ミドルウェア
│       │   └── logger.go
│       └── server.go   # サーバー初期化
├── web/                # Webアセット（新規）
│   └── templates/      # HTMLテンプレート
│       ├── base.html   # ベースレイアウト
│       ├── index.html  # メイン画面
│       ├── partials/   # htmx用の部分テンプレート
│       │   ├── column_selector.html
│       │   ├── filter_selector.html
│       │   ├── simpletab_result.html
│       │   └── crosstab_result.html
│       └── components/ # 再利用可能なコンポーネント
│           ├── header.html
│           └── table.html
├── configs/            # 設定ファイル（既存）
├── data/              # データファイル（既存）
└── docs/              # ドキュメント
    └── web-ui-design.md  # この設計書
```

## htmx の使用方法

### 基本パターン

#### 1. 列選択の動的更新

```html
<!-- 分析タイプが変更されたら、列選択UIを更新 -->
<select name="analysis_type"
        hx-get="/api/columns"
        hx-target="#column-selector"
        hx-trigger="change">
  <option value="simple">単純集計</option>
  <option value="cross">クロス集計</option>
</select>

<div id="column-selector">
  <!-- サーバーからHTMLフラグメントが挿入される -->
</div>
```

#### 2. 集計実行

```html
<form hx-post="/api/analyze"
      hx-target="#result-area"
      hx-indicator="#loading">
  <!-- フォームフィールド -->
  <button type="submit">集計実行</button>
</form>

<div id="loading" class="htmx-indicator">
  読み込み中...
</div>

<div id="result-area">
  <!-- 集計結果が表示される -->
</div>
```

#### 3. フィルタ選択

```html
<select name="filter"
        hx-get="/api/filter-preview"
        hx-target="#filter-info"
        hx-swap="innerHTML">
  <option value="">なし</option>
  <option value="elementary_junior">小中学生のみ</option>
  <!-- ... -->
</select>

<div id="filter-info">
  <!-- フィルタの説明が表示される -->
</div>
```

## _hyperscript の使用方法

### 基本パターン

#### 1. 複数回答分割チェックボックスの表示/非表示

```html
<div _="on change from #column-select
        if #column-select.value contains '[複数回答]'
          show #split-option
        else
          hide #split-option">

  <select id="column-select">
    <option>学年</option>
    <option>趣味 [複数回答]</option>
  </select>

  <div id="split-option" style="display:none;">
    <label>
      <input type="checkbox" name="split">
      複数回答として分割
    </label>
  </div>
</div>
```

#### 2. 集計ボタンの無効化制御

```html
<button type="submit"
        _="on htmx:beforeRequest add [@disabled='disabled']
           on htmx:afterRequest remove [@disabled]">
  集計実行
</button>
```

## API エンドポイント設計

### HTMLページ

| Method | Path | 説明 |
|--------|------|------|
| GET | `/` | メイン画面 |

### API（HTMLフラグメント返却）

| Method | Path | 説明 |
|--------|------|------|
| GET | `/api/columns` | 列選択UIのHTMLフラグメント |
| GET | `/api/filters` | フィルタ選択UIのHTMLフラグメント |
| GET | `/api/filter-preview` | フィルタ説明のHTMLフラグメント |
| POST | `/api/simpletab` | 単純集計実行（結果HTMLを返す） |
| POST | `/api/crosstab` | クロス集計実行（結果HTMLを返す） |
| POST | `/api/export` | CSV エクスポート（ファイルダウンロード） |

### リクエスト/レスポンス例

#### 単純集計リクエスト

```
POST /api/simpletab
Content-Type: application/x-www-form-urlencoded

db_path=data/app.duckdb&
table=excel_import&
column=学年&
split=false&
filter=elementary_junior
```

#### 単純集計レスポンス（HTMLフラグメント）

```html
<div class="result-container">
  <div class="result-header">
    <h3>集計結果：学年</h3>
    <p>総件数: 18,711件</p>
    <p>フィルタ: 小中学生のみ</p>
  </div>

  <table class="result-table">
    <thead>
      <tr>
        <th>値</th>
        <th>件数</th>
        <th>割合</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>中3</td>
        <td>7,282</td>
        <td>38.9%</td>
      </tr>
      <!-- ... -->
    </tbody>
  </table>

  <form hx-post="/api/export" hx-target="this" hx-swap="none">
    <input type="hidden" name="result_id" value="...">
    <button type="submit">CSV エクスポート</button>
  </form>
</div>
```

## データフロー

```
┌─────────┐   HTTP Request    ┌─────────────┐
│ Browser │ ───────────────> │   Handler   │
│ (htmx)  │                   │  (Go)       │
└─────────┘                   └─────────────┘
     ↑                              │
     │                              ↓
     │                        ┌─────────────┐
     │    HTML Fragment       │  Analyzer   │
     │ <──────────────────── │  (既存)     │
     │                        └─────────────┘
     │                              │
     │                              ↓
     │                        ┌─────────────┐
     │                        │   DuckDB    │
     │                        └─────────────┘
```

1. ブラウザ（htmx）がHTTPリクエストを送信
2. ハンドラーがリクエストを受け取り、既存のAnalyzerを呼び出し
3. AnalyzerがDuckDBにクエリを実行
4. 結果をHTMLテンプレートでレンダリング
5. HTMLフラグメントをブラウザに返却
6. htmxが指定された要素を更新

## 実装の優先順位

### Phase 1: MVP（最小実用版）
1. ✅ Webサーバーの基本構造
2. ✅ メイン画面のHTML/CSS
3. ✅ 単純集計機能
4. ✅ クロス集計機能
5. ✅ フィルタ選択

### Phase 2: 機能拡張
1. CSVエクスポート
2. エラーハンドリングとバリデーション
3. ローディングインジケーター

### Phase 3: UX改善
1. レスポンシブデザイン
2. 集計履歴
3. 可視化（グラフ）

### Phase 4: 運用機能
1. Excelアップロード機能
2. 設定ファイル編集UI

## 開発環境

### ローカル開発

```bash
# Webサーバー起動
go run cmd/calcanke-web/main.go

# デフォルトで http://localhost:8080 で起動
```

### 依存関係

```bash
# Echoフレームワークのインストール
go get github.com/labstack/echo/v4
go get github.com/labstack/echo/v4/middleware
```

### 設定

```yaml
# config.yaml
server:
  port: 8080
  host: localhost

database:
  default_path: data/app.duckdb
  default_table: excel_import

templates:
  path: web/templates
  reload: true  # 開発時はtrue、本番はfalse
```

### 基本構造例（Echo使用）

```go
// cmd/calcanke-web/main.go
package main

import (
    "html/template"
    "io"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/naozine/nz-mono-repo/apps/calcanke/internal/web/handlers"
)

type TemplateRenderer struct {
    templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
    e := echo.New()

    // ミドルウェア
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // テンプレート設定
    renderer := &TemplateRenderer{
        templates: template.Must(template.ParseGlob("web/templates/**/*.html")),
    }
    e.Renderer = renderer

    // ルーティング
    e.GET("/", handlers.Index)
    e.GET("/api/columns", handlers.GetColumns)
    e.POST("/api/simpletab", handlers.Simpletab)
    e.POST("/api/crosstab", handlers.Crosstab)
    e.POST("/api/export", handlers.Export)

    // サーバー起動
    e.Logger.Fatal(e.Start(":8080"))
}
```

## セキュリティ考慮事項

### Phase 1（認証なし）での制限
- ローカル環境での利用を想定
- 外部公開しない
- データベースファイルへのアクセス制限（読み取り専用推奨）

### 将来的な認証実装時の検討事項
- Basic認証
- セッション管理
- CSRF対策

## テスト戦略

### 単体テスト
- ハンドラーのロジックをテスト
- 既存のAnalyzerのテストを活用

### 統合テスト
- エンドポイントごとのHTTPテスト
- テンプレートレンダリングの確認

### E2Eテスト
- 将来的にはPlaywright等での自動テスト

## 次のステップ

1. この設計書をレビュー・承認
2. Phase 1のタスク分割
3. 実装開始（Webサーバー基本構造から）

## 参考資料

- [htmx Documentation](https://htmx.org/docs/)
- [_hyperscript Documentation](https://hyperscript.org/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Go html/template](https://pkg.go.dev/html/template)
