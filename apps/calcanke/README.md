# Calcanke

Excelファイル（アンケート等）をDuckDBにインポートして集計・分析するツール

## 概要

このプロジェクトは、ExcelファイルをDuckDBデータベースにインポートし、SQLで集計・分析を行うためのツールです。特に複数回答（改行区切り）を含むアンケートデータの分析に対応しています。

## プロジェクト構成

```
calcanke/
├── cmd/calcanke/              # メインアプリケーション
│   └── main.go                # Excelインポート処理
├── docs/                      # ドキュメント（汎用）
│   └── README.md              # ガイド
├── sql/                       # SQLテンプレート（汎用）
│   └── multi_answer_template.sql
├── data/                      # データファイル（Git管理外）
│   ├── app.duckdb             # DuckDBデータベース
│   ├── examples/              # プロジェクト固有のドキュメント・SQL
│   │   ├── survey_analysis_guide.md
│   │   ├── multi_answer_guide.md
│   │   ├── multi_answer_examples.sql
│   │   └── 1st_research.txt
│   └── *.csv                  # エクスポート結果
├── .env                       # 環境変数設定
├── .gitignore                 # Git除外設定（data/を除外）
├── go.mod                     # Go依存関係
└── go.sum                     # Go依存関係チェックサム
```

### ディレクトリの使い分け

- **`docs/` `sql/`**: 汎用的なドキュメント・SQLテンプレート（Git管理）
- **`data/`**: データベース、プロジェクト固有のファイル（Git管理外）
  - `data/examples/`: プロジェクト固有のドキュメント・集計例

**重要**: このリポジトリはpublicなので、データ固有の情報は `data/` 配下に配置してください。

## セットアップ

### 前提条件

- Go 1.25.3以上
- DuckDB CLI（Homebrewでインストール可能）

```bash
brew install duckdb
```

### 環境変数設定

`.env` ファイルを編集：

```env
EXCEL_PATH=/path/to/your/excel/file.xlsx
SHEET_NAME=Sheet0
DUCKDB_PATH=./data/app.duckdb
```

### データのインポート

```bash
# データディレクトリを作成
mkdir -p data

# Excelファイルをインポート
go run cmd/calcanke/main.go
```

## 使い方

### 基本的な集計

```bash
# DuckDB CLIで直接クエリ
duckdb data/app.duckdb -c "SELECT COUNT(*) FROM excel_import;"

# テンプレートSQLを参考に集計
cat sql/multi_answer_template.sql
```

### 複数回答の集計

複数回答（改行区切り）の項目を個別にカウントして集計：

```bash
# テンプレートを編集して実行
duckdb data/app.duckdb < sql/multi_answer_template.sql
```

詳細は `docs/README.md` を参照してください。

プロジェクト固有の集計例は `data/examples/` を参照してください（Git管理外）。

### CSVエクスポート

```bash
duckdb data/app.duckdb -c "
COPY (SELECT * FROM excel_import LIMIT 100)
TO 'data/export.csv' (HEADER, DELIMITER ',');
"
```

## ドキュメント

- **[docs/README.md](docs/README.md)** - 使い方ガイド
- **[sql/multi_answer_template.sql](sql/multi_answer_template.sql)** - SQLテンプレート
- **`data/examples/`** - プロジェクト固有の集計例（Git管理外）

## 主な機能

### 1. Excelインポート
- DuckDBのspatial拡張を使用してExcelファイルを読み込み
- シート名の指定が可能
- 自動的にテーブルスキーマを検出

### 2. 複数回答の分析
- 改行区切りの複数回答を個別に集計
- クロス集計（2軸、3軸以上）に対応
- 回答数の分布分析

### 3. SQLによる柔軟な集計
- DuckDBの強力なSQL機能を活用
- 複雑な集計やフィルタリングが可能
- CSV/Parquetへのエクスポート

## 技術スタック

- **Go**: メインアプリケーション
- **DuckDB**: 分析用データベース
- **DuckDB spatial extension**: Excel読み込み
- **go-duckdb**: DuckDBのGoドライバー

## トラブルシューティング

### データベースがロックされている

```
Error: Could not set lock on file "data/app.duckdb"
```

→ IDEのデータベースビューアを閉じてください

### 拡張機能のエラー

```
IO Error: Extension not found
```

→ `main.go` で自動インストールが有効になっているか確認してください

### シート名が見つからない

```
Layer 'Sheet1' could not be found
```

→ Excelファイルの実際のシート名を確認して `.env` を更新してください

```bash
# シート名を確認する方法
python3 -c "
import zipfile, xml.etree.ElementTree as ET
with zipfile.ZipFile('path/to/file.xlsx') as z:
    wb = ET.fromstring(z.read('xl/workbook.xml'))
    for sheet in wb.find('sheets'):
        print(sheet.get('name'))
"
```

## ライセンス

（適宜設定してください）

## 参考資料

- [DuckDB Documentation](https://duckdb.org/docs/)
- [DuckDB Spatial Extension](https://duckdb.org/docs/extensions/spatial.html)
- [go-duckdb](https://github.com/marcboeker/go-duckdb)
