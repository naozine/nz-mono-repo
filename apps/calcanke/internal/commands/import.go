package commands

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/spf13/cobra"
)

var (
	excelPath    string
	sheetName    string
	importDBPath string
	targetTable  string
)

// NewImportCmd はimportコマンドを作成
func NewImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "ExcelファイルをDuckDBにインポート",
		Long:  "ExcelファイルをDuckDBデータベースにインポートします",
		RunE:  runImport,
	}

	cmd.Flags().StringVar(&excelPath, "excel", "", "Excelファイルのパス（.envから読み込み可）")
	cmd.Flags().StringVar(&sheetName, "sheet", "", "シート名（.envから読み込み可）")
	cmd.Flags().StringVar(&importDBPath, "db", "", "DuckDBデータベースのパス（.envから読み込み可）")
	cmd.Flags().StringVar(&targetTable, "table", "", "テーブル名（.envから読み込み可）")

	return cmd
}

func runImport(cmd *cobra.Command, args []string) error {
	// .env 読み込み
	_ = godotenv.Load()

	// フラグまたは環境変数から値を取得
	if excelPath == "" {
		excelPath = os.Getenv("EXCEL_PATH")
	}
	if sheetName == "" {
		sheetName = os.Getenv("SHEET_NAME")
	}
	if importDBPath == "" {
		importDBPath = os.Getenv("DUCKDB_PATH")
		if importDBPath == "" {
			importDBPath = "data/app.duckdb"
		}
	}
	if targetTable == "" {
		targetTable = os.Getenv("TARGET_TABLE")
		if targetTable == "" {
			targetTable = "excel_import"
		}
	}

	if excelPath == "" {
		return fmt.Errorf("EXCEL_PATH is required (set via flag or .env)")
	}

	// DuckDB接続
	dsn := importDBPath
	db, err := sql.Open("duckdb", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// 拡張機能の自動インストールを有効化
	if _, err := db.Exec("SET autoinstall_known_extensions = true;"); err != nil {
		return fmt.Errorf("set autoinstall failed: %w", err)
	}
	if _, err := db.Exec("SET autoload_known_extensions = true;"); err != nil {
		return fmt.Errorf("set autoload failed: %w", err)
	}

	// INSTALL と LOAD spatial 拡張機能
	if _, err := db.Exec("INSTALL spatial;"); err != nil {
		return fmt.Errorf("install spatial failed: %w", err)
	}
	if _, err := db.Exec("LOAD spatial;"); err != nil {
		return fmt.Errorf("load spatial failed: %w", err)
	}

	// テーブルが無ければスキーマを作る
	createSQL := "CREATE TABLE IF NOT EXISTS " + targetTable + " AS SELECT * FROM st_read(?" +
		func() string {
			if sheetName != "" {
				return ", layer:=?"
			}
			return ""
		}() +
		") LIMIT 0"

	var errExec error
	if sheetName != "" {
		_, errExec = db.Exec(createSQL, excelPath, sheetName)
	} else {
		_, errExec = db.Exec(createSQL, excelPath)
	}
	if errExec != nil {
		return fmt.Errorf("create table failed: %w", errExec)
	}

	// 追記
	insertSQL := "INSERT INTO " + targetTable + " SELECT * FROM st_read(?" +
		func() string {
			if sheetName != "" {
				return ", layer:=?"
			}
			return ""
		}() + ")"

	if sheetName != "" {
		_, err = db.Exec(insertSQL, excelPath, sheetName)
	} else {
		_, err = db.Exec(insertSQL, excelPath)
	}
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	sheet := sheetName
	if sheet == "" {
		sheet = "<auto>"
	}
	fmt.Printf("✓ Imported %s (sheet=%s) into table %s\n", excelPath, sheet, targetTable)

	return nil
}
