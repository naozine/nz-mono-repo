package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/importer"
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/project"
)

// ProjectHandler はプロジェクトのハンドラ
type ProjectHandler struct {
	repo       *project.Repository
	projectDir string
}

// NewProjectHandler はプロジェクトハンドラを作成
func NewProjectHandler(repo *project.Repository, projectDir string) *ProjectHandler {
	return &ProjectHandler{
		repo:       repo,
		projectDir: projectDir,
	}
}

// List はプロジェクト一覧を表示
func (h *ProjectHandler) List(c echo.Context) error {
	projects, err := h.repo.FindAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load projects: "+err.Error())
	}

	data := map[string]interface{}{
		"Projects": projects,
	}

	return c.Render(http.StatusOK, "projects_list.html", data)
}

// ShowCreateForm はプロジェクト作成フォームを表示
func (h *ProjectHandler) ShowCreateForm(c echo.Context) error {
	return c.Render(http.StatusOK, "project_create.html", nil)
}

// Create はプロジェクトを作成
func (h *ProjectHandler) Create(c echo.Context) error {
	name := c.FormValue("name")
	description := c.FormValue("description")

	if name == "" {
		return c.String(http.StatusBadRequest, "Project name is required")
	}

	// UUIDを生成
	id := uuid.New().String()

	// プロジェクトを作成
	p := project.NewProject(id, name, description)

	// プロジェクトディレクトリを作成
	projectPath := p.GetProjectDir(h.projectDir)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create project directory: "+err.Error())
	}

	// デフォルトの設定ファイルを作成
	if err := h.createDefaultConfigFiles(p); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create config files: "+err.Error())
	}

	// データベースに保存
	if err := h.repo.Create(p); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save project: "+err.Error())
	}

	// Excelアップロード画面にリダイレクト
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/projects/%s/upload", p.ID))
}

// ShowUploadForm はExcelアップロードフォームを表示
func (h *ProjectHandler) ShowUploadForm(c echo.Context) error {
	id := c.Param("id")

	p, err := h.repo.FindByID(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load project: "+err.Error())
	}

	if p == nil {
		return c.String(http.StatusNotFound, "Project not found")
	}

	data := map[string]interface{}{
		"Project": p,
	}

	return c.Render(http.StatusOK, "project_upload.html", data)
}

// Delete はプロジェクトを削除
func (h *ProjectHandler) Delete(c echo.Context) error {
	id := c.Param("id")

	p, err := h.repo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load project"})
	}

	if p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// プロジェクトディレクトリを削除
	projectPath := p.GetProjectDir(h.projectDir)
	if err := os.RemoveAll(projectPath); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete project directory"})
	}

	// データベースから削除
	if err := h.repo.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete project"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Project deleted successfully"})
}

// createDefaultConfigFiles はデフォルトの設定ファイルを作成
func (h *ProjectHandler) createDefaultConfigFiles(p *project.Project) error {
	// derived_columns.yaml
	derivedColumnsContent := `# 派生列の定義
# この設定ファイルで、既存の列から新しい列を動的に生成できます

derived_columns: []
`
	derivedColumnsPath := p.GetDerivedColumnsPath(h.projectDir)
	if err := os.WriteFile(derivedColumnsPath, []byte(derivedColumnsContent), 0644); err != nil {
		return fmt.Errorf("failed to create derived_columns.yaml: %w", err)
	}

	// filters.yaml
	filtersContent := `# フィルタの定義
# この設定ファイルで、データをフィルタリングするための条件を定義できます

filters: []
`
	filtersPath := p.GetFiltersPath(h.projectDir)
	if err := os.WriteFile(filtersPath, []byte(filtersContent), 0644); err != nil {
		return fmt.Errorf("failed to create filters.yaml: %w", err)
	}

	// column_orders.yaml
	columnOrdersContent := `# 列の値の表示順序の定義
# この設定ファイルで、列に含まれる値の表示順序を指定できます
# グラフや表での表示順序が制御されます

column_orders: []
`
	columnOrdersPath := p.GetColumnOrdersPath(h.projectDir)
	if err := os.WriteFile(columnOrdersPath, []byte(columnOrdersContent), 0644); err != nil {
		return fmt.Errorf("failed to create column_orders.yaml: %w", err)
	}

	return nil
}

// GetProjectListAPI はプロジェクト一覧をJSONで返す
func (h *ProjectHandler) GetProjectListAPI(c echo.Context) error {
	projects, err := h.repo.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load projects"})
	}

	return c.JSON(http.StatusOK, projects)
}

// GetProjectAPI はプロジェクト詳細をJSONで返す
func (h *ProjectHandler) GetProjectAPI(c echo.Context) error {
	id := c.Param("id")

	p, err := h.repo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load project"})
	}

	if p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	return c.JSON(http.StatusOK, p)
}

// Upload はExcelファイルをアップロードしてインポートする
func (h *ProjectHandler) Upload(c echo.Context) error {
	id := c.Param("id")

	// プロジェクトを取得
	p, err := h.repo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load project"})
	}

	if p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// ファイルを取得
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No file uploaded"})
	}

	// ファイルを開く
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to open file"})
	}
	defer src.Close()

	// 保存先パス
	excelPath := p.GetExcelPath(h.projectDir)

	// ファイルを保存
	dst, err := os.Create(excelPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create file"})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save file"})
	}

	// DuckDBにインポート
	duckdbPath := p.GetDuckDBPath(h.projectDir)
	tableName := "excel_import"

	if err := importer.ImportExcel(excelPath, duckdbPath, tableName); err != nil {
		// インポート失敗時はステータスをエラーに
		p.Status = string(project.StatusError)
		h.repo.Update(p)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to import Excel: " + err.Error()})
	}

	// プロジェクト情報を更新
	p.TableName = tableName
	p.ExcelFilename = filepath.Base(file.Filename)
	p.Status = string(project.StatusReady)

	if err := h.repo.Update(p); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update project"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Upload successful", "project_id": p.ID})
}

// ShowAnalysis はプロジェクトの分析画面を表示
func (h *ProjectHandler) ShowAnalysis(c echo.Context) error {
	id := c.Param("id")

	// プロジェクトを取得
	p, err := h.repo.FindByID(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load project: "+err.Error())
	}

	if p == nil {
		return c.String(http.StatusNotFound, "Project not found")
	}

	// プロジェクトがready状態でない場合はエラー
	if p.Status != string(project.StatusReady) {
		return c.String(http.StatusBadRequest, "Project is not ready for analysis")
	}

	// プロジェクトのDuckDBパスと設定ファイルパスを取得
	dbPath := p.GetDuckDBPath(h.projectDir)
	derivedColumnsPath := p.GetDerivedColumnsPath(h.projectDir)
	filtersPath := p.GetFiltersPath(h.projectDir)
	columnOrdersPath := p.GetColumnOrdersPath(h.projectDir)

	// Analyzerを作成してテーブル情報を取得
	a, err := analyzer.NewAnalyzerWithConfigs(dbPath, p.TableName, derivedColumnsPath, filtersPath, columnOrdersPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to initialize analyzer")
	}
	defer a.Close()

	total, err := a.GetTableInfo()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get table info")
	}

	data := map[string]interface{}{
		"Project": p,
		"DBPath":  dbPath,
		"Table":   p.TableName,
		"Total":   total,
	}

	return c.Render(http.StatusOK, "project_analysis.html", data)
}

// getProjectHandler はプロジェクト用のHandlerを作成する
// 既存のHandlerメソッドを再利用するため
func (h *ProjectHandler) getProjectHandler(projectID string) (*Handler, error) {
	p, err := h.repo.FindByID(projectID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("project not found")
	}
	if p.Status != string(project.StatusReady) {
		return nil, fmt.Errorf("project is not ready")
	}

	// プロジェクト固有のパスを取得
	dbPath := p.GetDuckDBPath(h.projectDir)
	derivedColumnsPath := p.GetDerivedColumnsPath(h.projectDir)
	filtersPath := p.GetFiltersPath(h.projectDir)
	columnOrdersPath := p.GetColumnOrdersPath(h.projectDir)

	return NewHandlerWithConfigs(dbPath, p.TableName, derivedColumnsPath, filtersPath, columnOrdersPath), nil
}

// GetProjectColumns はプロジェクトのカラム一覧をHTML形式で返す（htmx用）
func (h *ProjectHandler) GetProjectColumns(c echo.Context) error {
	projectID := c.Param("id")
	handler, err := h.getProjectHandler(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return handler.GetColumns(c)
}

// GetProjectColumnsJSON はプロジェクトのカラム一覧をJSON形式で返す（API用）
func (h *ProjectHandler) GetProjectColumnsJSON(c echo.Context) error {
	projectID := c.Param("id")
	handler, err := h.getProjectHandler(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return handler.GetColumnsJSON(c)
}

// GetProjectFilters はプロジェクトのフィルタ一覧を返す
func (h *ProjectHandler) GetProjectFilters(c echo.Context) error {
	projectID := c.Param("id")
	handler, err := h.getProjectHandler(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return handler.GetFilters(c)
}

// ProjectSimpletab はプロジェクトの単純集計を実行
func (h *ProjectHandler) ProjectSimpletab(c echo.Context) error {
	projectID := c.Param("id")
	handler, err := h.getProjectHandler(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return handler.Simpletab(c)
}

// ProjectCrosstab はプロジェクトのクロス集計を実行
func (h *ProjectHandler) ProjectCrosstab(c echo.Context) error {
	projectID := c.Param("id")
	handler, err := h.getProjectHandler(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return handler.Crosstab(c)
}

// ProjectExport はプロジェクトのエクスポートを実行
func (h *ProjectHandler) ProjectExport(c echo.Context) error {
	projectID := c.Param("id")
	handler, err := h.getProjectHandler(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return handler.Export(c)
}

// GetDerivedColumns は派生列の一覧を取得
func (h *ProjectHandler) GetDerivedColumns(c echo.Context) error {
	id := c.Param("id")

	p, err := h.repo.FindByID(id)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	derivedColumnsPath := p.GetDerivedColumnsPath(h.projectDir)
	columns, err := analyzer.LoadDerivedColumns(derivedColumnsPath)
	if err != nil {
		// ファイルがない場合は空配列を返す
		return c.JSON(http.StatusOK, []analyzer.DerivedColumn{})
	}

	return c.JSON(http.StatusOK, columns)
}

// AddDerivedColumn は新しい派生列を追加
func (h *ProjectHandler) AddDerivedColumn(c echo.Context) error {
	id := c.Param("id")

	p, err := h.repo.FindByID(id)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// リクエストボディから派生列を取得
	var newColumn analyzer.DerivedColumn
	if err := c.Bind(&newColumn); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// 既存の派生列を読み込み
	derivedColumnsPath := p.GetDerivedColumnsPath(h.projectDir)
	columns, err := analyzer.LoadDerivedColumns(derivedColumnsPath)
	if err != nil {
		columns = []analyzer.DerivedColumn{}
	}

	// 新しい派生列を追加
	columns = append(columns, newColumn)

	// 保存
	if err := analyzer.SaveDerivedColumns(derivedColumnsPath, columns); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save derived columns"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Derived column added successfully"})
}

// UpdateDerivedColumn は派生列を更新
func (h *ProjectHandler) UpdateDerivedColumn(c echo.Context) error {
	id := c.Param("id")
	indexStr := c.Param("index")

	p, err := h.repo.FindByID(id)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// インデックスを整数に変換
	index := 0
	fmt.Sscanf(indexStr, "%d", &index)

	// リクエストボディから派生列を取得
	var updatedColumn analyzer.DerivedColumn
	if err := c.Bind(&updatedColumn); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// 既存の派生列を読み込み
	derivedColumnsPath := p.GetDerivedColumnsPath(h.projectDir)
	columns, err := analyzer.LoadDerivedColumns(derivedColumnsPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load derived columns"})
	}

	// インデックスチェック
	if index < 0 || index >= len(columns) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid index"})
	}

	// 更新
	columns[index] = updatedColumn

	// 保存
	if err := analyzer.SaveDerivedColumns(derivedColumnsPath, columns); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save derived columns"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Derived column updated successfully"})
}

// DeleteDerivedColumn は派生列を削除
func (h *ProjectHandler) DeleteDerivedColumn(c echo.Context) error {
	id := c.Param("id")
	indexStr := c.Param("index")

	p, err := h.repo.FindByID(id)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// インデックスを整数に変換
	index := 0
	fmt.Sscanf(indexStr, "%d", &index)

	// 既存の派生列を読み込み
	derivedColumnsPath := p.GetDerivedColumnsPath(h.projectDir)
	columns, err := analyzer.LoadDerivedColumns(derivedColumnsPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load derived columns"})
	}

	// インデックスチェック
	if index < 0 || index >= len(columns) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid index"})
	}

	// 削除
	columns = append(columns[:index], columns[index+1:]...)

	// 保存
	if err := analyzer.SaveDerivedColumns(derivedColumnsPath, columns); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save derived columns"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Derived column deleted successfully"})
}

// GetDerivedColumnTemplates はテンプレートライブラリから派生列テンプレートを取得
func (h *ProjectHandler) GetDerivedColumnTemplates(c echo.Context) error {
	// configs/derived_columns.yaml からテンプレートを読み込む
	templatePath := filepath.Join("configs", "derived_columns.yaml")
	templates, err := analyzer.LoadDerivedColumns(templatePath)
	if err != nil {
		// ファイルがない場合は空配列を返す
		return c.JSON(http.StatusOK, []analyzer.DerivedColumn{})
	}

	return c.JSON(http.StatusOK, templates)
}

// ImportDerivedColumnTemplates はテンプレートライブラリから派生列をインポート
func (h *ProjectHandler) ImportDerivedColumnTemplates(c echo.Context) error {
	id := c.Param("id")

	p, err := h.repo.FindByID(id)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// リクエストボディから選択されたテンプレートのインデックスを取得
	var requestBody struct {
		Indices []int `json:"indices"`
	}
	if err := c.Bind(&requestBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// configs/derived_columns.yaml からテンプレートを読み込む
	templatePath := filepath.Join("configs", "derived_columns.yaml")
	templates, err := analyzer.LoadDerivedColumns(templatePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load templates"})
	}

	// 既存の派生列を読み込み
	derivedColumnsPath := p.GetDerivedColumnsPath(h.projectDir)
	columns, err := analyzer.LoadDerivedColumns(derivedColumnsPath)
	if err != nil {
		columns = []analyzer.DerivedColumn{}
	}

	// 既存の列名マップを作成
	existingNames := make(map[string]bool)
	for _, col := range columns {
		existingNames[col.Name] = true
	}

	// 選択されたテンプレートをインポート
	importedCount := 0
	for _, idx := range requestBody.Indices {
		if idx < 0 || idx >= len(templates) {
			continue
		}

		template := templates[idx]
		originalName := template.Name

		// 名前の重複チェック（重複している場合は (2), (3)... を付ける）
		newName := originalName
		counter := 2
		for existingNames[newName] {
			newName = fmt.Sprintf("%s (%d)", originalName, counter)
			counter++
		}
		template.Name = newName
		existingNames[newName] = true

		// 派生列を追加
		columns = append(columns, template)
		importedCount++
	}

	// 保存
	if err := analyzer.SaveDerivedColumns(derivedColumnsPath, columns); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save derived columns"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":        "Templates imported successfully",
		"imported_count": importedCount,
	})
}

// GetFiltersConfig はフィルタ設定の一覧を取得
func (h *ProjectHandler) GetFiltersConfig(c echo.Context) error {
	id := c.Param("id")

	p, err := h.repo.FindByID(id)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	filtersPath := p.GetFiltersPath(h.projectDir)
	filters, err := analyzer.LoadFilters(filtersPath)
	if err != nil {
		// ファイルがない場合は空配列を返す
		return c.JSON(http.StatusOK, []analyzer.Filter{})
	}

	return c.JSON(http.StatusOK, filters)
}

// AddFilterConfig は新しいフィルタ設定を追加
func (h *ProjectHandler) AddFilterConfig(c echo.Context) error {
	id := c.Param("id")

	p, err := h.repo.FindByID(id)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// リクエストボディからフィルタを取得
	var newFilter analyzer.Filter
	if err := c.Bind(&newFilter); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// 既存のフィルタを読み込み
	filtersPath := p.GetFiltersPath(h.projectDir)
	filters, err := analyzer.LoadFilters(filtersPath)
	if err != nil {
		filters = []analyzer.Filter{}
	}

	// 新しいフィルタを追加
	filters = append(filters, newFilter)

	// 保存
	if err := analyzer.SaveFilters(filtersPath, filters); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save filters"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Filter added successfully"})
}

// UpdateFilterConfig はフィルタ設定を更新
func (h *ProjectHandler) UpdateFilterConfig(c echo.Context) error {
	id := c.Param("id")
	indexStr := c.Param("index")

	p, err := h.repo.FindByID(id)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// インデックスを整数に変換
	index := 0
	fmt.Sscanf(indexStr, "%d", &index)

	// リクエストボディからフィルタを取得
	var updatedFilter analyzer.Filter
	if err := c.Bind(&updatedFilter); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// 既存のフィルタを読み込み
	filtersPath := p.GetFiltersPath(h.projectDir)
	filters, err := analyzer.LoadFilters(filtersPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load filters"})
	}

	// インデックスチェック
	if index < 0 || index >= len(filters) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid index"})
	}

	// 更新
	filters[index] = updatedFilter

	// 保存
	if err := analyzer.SaveFilters(filtersPath, filters); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save filters"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Filter updated successfully"})
}

// DeleteFilterConfig はフィルタ設定を削除
func (h *ProjectHandler) DeleteFilterConfig(c echo.Context) error {
	id := c.Param("id")
	indexStr := c.Param("index")

	p, err := h.repo.FindByID(id)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// インデックスを整数に変換
	index := 0
	fmt.Sscanf(indexStr, "%d", &index)

	// 既存のフィルタを読み込み
	filtersPath := p.GetFiltersPath(h.projectDir)
	filters, err := analyzer.LoadFilters(filtersPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load filters"})
	}

	// インデックスチェック
	if index < 0 || index >= len(filters) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid index"})
	}

	// 削除
	filters = append(filters[:index], filters[index+1:]...)

	// 保存
	if err := analyzer.SaveFilters(filtersPath, filters); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save filters"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Filter deleted successfully"})
}

// GetColumnOrders は列の値の表示順序設定を取得する
func (h *ProjectHandler) GetColumnOrders(c echo.Context) error {
	projectID := c.Param("id")

	// プロジェクトを取得
	p, err := h.repo.FindByID(projectID)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// 列順序設定を読み込み
	columnOrdersPath := p.GetColumnOrdersPath(h.projectDir)
	columnOrders, err := analyzer.LoadColumnOrders(columnOrdersPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load column orders"})
	}

	return c.JSON(http.StatusOK, columnOrders)
}

// UpdateColumnOrders は列の値の表示順序設定を更新する
func (h *ProjectHandler) UpdateColumnOrders(c echo.Context) error {
	projectID := c.Param("id")

	// プロジェクトを取得
	p, err := h.repo.FindByID(projectID)
	if err != nil || p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// リクエストボディをパース
	var columnOrders []analyzer.ColumnOrder
	if err := c.Bind(&columnOrders); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// 保存
	columnOrdersPath := p.GetColumnOrdersPath(h.projectDir)
	if err := analyzer.SaveColumnOrders(columnOrdersPath, columnOrders); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save column orders"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Column orders updated successfully"})
}
