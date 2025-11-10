package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
)

// ColumnsData は列選択UIのテンプレートデータ
type ColumnsData struct {
	AnalysisType string
	Columns      analyzer.ColumnList
}

// GetColumns は列選択UIを返す（htmx用）
func (h *Handler) GetColumns(c echo.Context) error {
	analysisType := c.QueryParam("analysis_type")

	a, err := h.getAnalyzer()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to initialize analyzer")
	}
	defer a.Close()

	columns, err := a.GetColumns()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get columns")
	}

	data := ColumnsData{
		AnalysisType: analysisType,
		Columns:      columns,
	}

	return c.Render(http.StatusOK, "column_selector.html", data)
}

// GetColumnsJSON は列のリストをJSON形式で返す（API用）
func (h *Handler) GetColumnsJSON(c echo.Context) error {
	a, err := h.getAnalyzer()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to initialize analyzer"})
	}
	defer a.Close()

	columns, err := a.GetColumns()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get columns"})
	}

	return c.JSON(http.StatusOK, columns)
}
