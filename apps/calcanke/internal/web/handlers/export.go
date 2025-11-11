package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Export はCSVエクスポートを実行する（Phase 2で実装予定）
func (h *Handler) Export(c echo.Context) error {
	// TODO: Phase 2で実装
	return c.String(http.StatusNotImplemented, "Export feature is not implemented yet")
}
