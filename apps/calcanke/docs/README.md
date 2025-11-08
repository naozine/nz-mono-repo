# Calcanke ドキュメント

ExcelファイルをDuckDBにインポートして集計・分析するためのガイド

## 概要

Calcankeは、Excelファイル（特にアンケートデータ）をDuckDBにインポートし、SQLで集計・分析するためのツールです。

## セットアップ

### 1. 環境構築

```bash
# DuckDB CLIのインストール
brew install duckdb

# プロジェクトのビルド
go build -o calcanke cmd/calcanke/main.go
```

### 2. 環境変数の設定

`.env` ファイルを作成：

```env
EXCEL_PATH=/path/to/your/excel/file.xlsx
SHEET_NAME=Sheet0
DUCKDB_PATH=./data/app.duckdb
TARGET_TABLE=excel_import
```

### 3. データのインポート

```bash
# データディレクトリを作成
mkdir -p data

# Excelファイルをインポート
go run cmd/calcanke/main.go
```

## 基本的な使い方

### データの確認

```bash
# テーブル構造を確認
duckdb data/app.duckdb -c "DESCRIBE excel_import;"

# データ件数を確認
duckdb data/app.duckdb -c "SELECT COUNT(*) FROM excel_import;"

# 先頭5件を表示
duckdb data/app.duckdb -c "SELECT * FROM excel_import LIMIT 5;"
```

### 単純集計

```sql
-- 列の値を集計
SELECT
  列名,
  COUNT(*) as 件数,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as 割合
FROM excel_import
WHERE 列名 IS NOT NULL
GROUP BY 列名
ORDER BY 件数 DESC;
```

### クロス集計

```sql
-- 2つの列のクロス集計
SELECT
  列A,
  列B,
  COUNT(*) as 件数,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(PARTITION BY 列A), 1) as 列A内割合
FROM excel_import
WHERE 列A IS NOT NULL AND 列B IS NOT NULL
GROUP BY 列A, 列B
ORDER BY 列A, 件数 DESC;
```

## 複数回答の集計

改行（`\n`）区切りの複数回答を個別に集計する方法

### 基本パターン

```sql
-- 複数回答を分割して集計
WITH split_data AS (
  SELECT
    unnest(string_split(複数回答列, CHR(10))) as 分割値
  FROM excel_import
  WHERE 複数回答列 IS NOT NULL
)
SELECT
  分割値,
  COUNT(*) as 回答数,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as 割合
FROM split_data
GROUP BY 分割値
ORDER BY 回答数 DESC;
```

### 複数回答 × 単一回答のクロス集計

```sql
WITH split_data AS (
  SELECT
    unnest(string_split(複数回答列, CHR(10))) as 分割値,
    単一回答列
  FROM excel_import
  WHERE 複数回答列 IS NOT NULL AND 単一回答列 IS NOT NULL
)
SELECT
  分割値,
  単一回答列,
  COUNT(*) as 件数,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(PARTITION BY 分割値), 1) as 構成比
FROM split_data
GROUP BY 分割値, 単一回答列
ORDER BY 分割値, 件数 DESC;
```

### 主要な関数

- `string_split(文字列, 区切り文字)` - 文字列を配列に分割
- `CHR(10)` - 改行文字（`\n`）
- `unnest(配列)` - 配列の各要素を行に展開

## データのエクスポート

### CSVエクスポート

```bash
duckdb data/app.duckdb -c "
COPY (
  SELECT * FROM excel_import
) TO 'data/export.csv' (HEADER, DELIMITER ',');
"
```

### 集計結果のエクスポート

```bash
duckdb data/app.duckdb -c "
COPY (
  SELECT 列名, COUNT(*) as 件数
  FROM excel_import
  GROUP BY 列名
  ORDER BY 件数 DESC
) TO 'data/集計結果.csv' (HEADER, DELIMITER ',');
"
```

## SQLテンプレート

汎用的なSQLテンプレートは `sql/` ディレクトリに配置してください。

プロジェクト固有の集計例は `data/examples/` ディレクトリに配置することを推奨します（Git管理外）。

## トラブルシューティング

### データベースロックエラー

```
Error: Could not set lock on file "data/app.duckdb"
```

→ IDEのデータベースビューアを閉じてください

### 拡張機能エラー

```
IO Error: Extension not found
```

→ `main.go` で以下の設定が有効になっているか確認：
```go
db.Exec("SET autoinstall_known_extensions = true;")
db.Exec("SET autoload_known_extensions = true;")
```

### シート名が見つからない

```
Layer 'SheetName' could not be found
```

→ Excelファイルの実際のシート名を確認：
```bash
python3 -c "
import zipfile, xml.etree.ElementTree as ET
with zipfile.ZipFile('path/to/file.xlsx') as z:
    wb = ET.fromstring(z.read('xl/workbook.xml'))
    for sheet in wb.find('sheets'):
        print(sheet.get('name'))
"
```

## 参考資料

- [DuckDB Documentation](https://duckdb.org/docs/)
- [DuckDB String Functions](https://duckdb.org/docs/sql/functions/char.html)
- [DuckDB Array Functions](https://duckdb.org/docs/sql/functions/nested.html)
- [DuckDB Spatial Extension](https://duckdb.org/docs/extensions/spatial.html)
