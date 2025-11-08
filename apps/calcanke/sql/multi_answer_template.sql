-- =====================================
-- 複数回答の集計テンプレート
-- =====================================
-- このファイルは汎用的なSQLテンプレートです
-- プロジェクト固有の例は data/examples/ に配置してください

-- =====================================
-- 1. 単純集計
-- =====================================

-- 複数回答を分割して集計
WITH split_data AS (
  SELECT
    unnest(string_split("複数回答列名", CHR(10))) as 値
  FROM excel_import
  WHERE "複数回答列名" IS NOT NULL
)
SELECT
  値,
  COUNT(*) as 回答数,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as 割合
FROM split_data
GROUP BY 値
ORDER BY 回答数 DESC;

-- =====================================
-- 2. 回答数の分布
-- =====================================

-- 1人が何個選択したかの分布
SELECT
  array_length(string_split("複数回答列名", CHR(10))) as 回答数,
  COUNT(*) as 人数,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as 割合
FROM excel_import
WHERE "複数回答列名" IS NOT NULL
GROUP BY 回答数
ORDER BY 回答数;

-- =====================================
-- 3. クロス集計（2軸）
-- =====================================

-- 複数回答 × 単一回答
WITH split_data AS (
  SELECT
    unnest(string_split("複数回答列名", CHR(10))) as 複数回答値,
    "単一回答列名" as 単一回答値
  FROM excel_import
  WHERE "複数回答列名" IS NOT NULL
    AND "単一回答列名" IS NOT NULL
)
SELECT
  複数回答値,
  単一回答値,
  COUNT(*) as 件数,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(PARTITION BY 複数回答値), 1) as 構成比
FROM split_data
GROUP BY 複数回答値, 単一回答値
ORDER BY 複数回答値, 件数 DESC;

-- =====================================
-- 4. クロス集計（3軸）
-- =====================================

-- 複数回答 × 単一回答A × 単一回答B
WITH split_data AS (
  SELECT
    unnest(string_split("複数回答列名", CHR(10))) as 複数回答値,
    "単一回答列A" as 値A,
    "単一回答列B" as 値B
  FROM excel_import
  WHERE "複数回答列名" IS NOT NULL
    AND "単一回答列A" IS NOT NULL
    AND "単一回答列B" IS NOT NULL
)
SELECT
  複数回答値,
  値A,
  値B,
  COUNT(*) as 件数,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(PARTITION BY 複数回答値, 値A), 1) as 構成比
FROM split_data
GROUP BY 複数回答値, 値A, 値B
ORDER BY 複数回答値, 値A, 件数 DESC;

-- =====================================
-- 5. 特定値でフィルタ
-- =====================================

-- 特定の値を選択した人だけを抽出
WITH has_specific_value AS (
  SELECT DISTINCT
    ROW_NUMBER() OVER() as id,
    "単一回答列名" as 単一回答値
  FROM excel_import
  WHERE POSITION('特定の値' IN "複数回答列名") > 0
    AND "単一回答列名" IS NOT NULL
)
SELECT
  単一回答値,
  COUNT(*) as 人数,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as 割合
FROM has_specific_value
GROUP BY 単一回答値
ORDER BY 人数 DESC;

-- =====================================
-- 6. 組み合わせパターンの分析
-- =====================================

-- どの組み合わせが多いか
SELECT
  "複数回答列名" as 組み合わせ,
  array_length(string_split("複数回答列名", CHR(10))) as 選択数,
  COUNT(*) as 人数,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as 割合
FROM excel_import
WHERE "複数回答列名" IS NOT NULL
  AND POSITION(CHR(10) IN "複数回答列名") > 0  -- 複数選択のみ
GROUP BY "複数回答列名", 選択数
ORDER BY 人数 DESC
LIMIT 20;

-- =====================================
-- 7. CSVエクスポート
-- =====================================

-- 結果をCSVに出力
COPY (
  WITH split_data AS (
    SELECT unnest(string_split("複数回答列名", CHR(10))) as 値
    FROM excel_import
    WHERE "複数回答列名" IS NOT NULL
  )
  SELECT 値, COUNT(*) as 回答数
  FROM split_data
  GROUP BY 値
  ORDER BY 回答数 DESC
) TO 'data/集計結果.csv' (HEADER, DELIMITER ',');
