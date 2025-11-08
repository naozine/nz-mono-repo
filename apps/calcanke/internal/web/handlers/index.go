package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// IndexData はメイン画面のテンプレートデータ
type IndexData struct {
	DBPath string
	Table  string
	Total  int
}

// Index はメイン画面を表示する
func (h *Handler) Index(c echo.Context) error {
	a, err := h.getAnalyzer()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to initialize analyzer")
	}
	defer a.Close()

	total, err := a.GetTableInfo()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get table info")
	}

	data := IndexData{
		DBPath: h.dbPath,
		Table:  h.table,
		Total:  total,
	}

	return c.Render(http.StatusOK, "index.html", data)
}
