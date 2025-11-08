package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
)

// CrosstabResultData はクロス集計結果のテンプレートデータ
type CrosstabResultData struct {
	Result *analyzer.CrosstabResult
	Filter *analyzer.Filter
}

// Crosstab はクロス集計を実行する
func (h *Handler) Crosstab(c echo.Context) error {
	a, err := h.getAnalyzer()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to initialize analyzer")
	}
	defer a.Close()

	// パラメータ取得
	xColumnIndexStr := c.FormValue("x_column")
	yColumnIndexStr := c.FormValue("y_column")
	splitXStr := c.FormValue("split_x")
	splitYStr := c.FormValue("split_y")
	filterName := c.FormValue("filter")

	// 列インデックスをパース
	xColumnIndex, err := strconv.Atoi(xColumnIndexStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid X column index")
	}

	yColumnIndex, err := strconv.Atoi(yColumnIndexStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid Y column index")
	}

	// 列を取得
	columns, err := a.GetColumns()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get columns")
	}

	if xColumnIndex < 1 || xColumnIndex > len(columns) {
		return c.String(http.StatusBadRequest, "X column index out of range")
	}

	if yColumnIndex < 1 || yColumnIndex > len(columns) {
		return c.String(http.StatusBadRequest, "Y column index out of range")
	}

	xColumn := &columns[xColumnIndex-1]
	yColumn := &columns[yColumnIndex-1]

	// 分割フラグ
	splitX := splitXStr == "true" || splitXStr == "on"
	splitY := splitYStr == "true" || splitYStr == "on"

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

	// 集計設定
	config := analyzer.AnalysisConfig{
		XColumn: xColumn,
		YColumn: yColumn,
		SplitX:  splitX,
		SplitY:  splitY,
	}

	// 集計実行
	result, err := a.CrosstabWithFilter(config, filter)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to execute crosstab: "+err.Error())
	}

	data := CrosstabResultData{
		Result: result,
		Filter: filter,
	}

	return c.Render(http.StatusOK, "crosstab_result.html", data)
}
