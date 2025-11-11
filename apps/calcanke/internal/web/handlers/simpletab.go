package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
)

// SimpletabResultData は単純集計結果のテンプレートデータ
type SimpletabResultData struct {
	Result *analyzer.SimpletabResult
	Filter *analyzer.Filter
}

// Simpletab は単純集計を実行する
func (h *Handler) Simpletab(c echo.Context) error {
	a, err := h.getAnalyzer()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to initialize analyzer")
	}
	defer a.Close()

	// パラメータ取得
	columnIndexStr := c.FormValue("column")
	splitStr := c.FormValue("split")
	filterName := c.FormValue("filter")

	// 列インデックスをパース
	columnIndex, err := strconv.Atoi(columnIndexStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid column index")
	}

	// 列を取得
	columns, err := a.GetColumns()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get columns")
	}

	if columnIndex < 1 || columnIndex > len(columns) {
		return c.String(http.StatusBadRequest, "Column index out of range")
	}

	column := &columns[columnIndex-1]

	// 分割フラグ
	split := splitStr == "true" || splitStr == "on"

	// フィルタを取得
	var filter *analyzer.Filter
	if filterName != "" {
		for i := range a.Filters {
			if a.Filters[i].Name == filterName {
				filter = &a.Filters[i]
				break
			}
		}
	}

	// 集計実行
	result, err := a.SimpletabWithFilter(column, split, filter)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to execute simpletab: "+err.Error())
	}

	data := SimpletabResultData{
		Result: result,
		Filter: filter,
	}

	return c.Render(http.StatusOK, "simpletab_result.html", data)
}
