package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/analyzer"
)

// FiltersData はフィルタ選択UIのテンプレートデータ
type FiltersData struct {
	Filters []analyzer.Filter
}

// GetFilters はフィルタ選択UIを返す（htmx用）
func (h *Handler) GetFilters(c echo.Context) error {
	a, err := h.getAnalyzer()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to initialize analyzer")
	}
	defer a.Close()

	data := FiltersData{
		Filters: a.Filters,
	}

	return c.Render(http.StatusOK, "filter_selector.html", data)
}
