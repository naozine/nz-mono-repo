package web

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/web/handlers"
)

// TemplateRenderer はhtml/templateをEchoで使用するためのレンダラー
type TemplateRenderer struct {
	templates *template.Template
}

// Render はテンプレートをレンダリングする
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// NewServer はWebサーバーを作成する
func NewServer(dbPath, table string) *echo.Echo {
	e := echo.New()

	// ミドルウェアの設定
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// テンプレートの読み込み
	templates := template.Must(template.ParseGlob("web/templates/*.html"))
	templates = template.Must(templates.ParseGlob("web/templates/partials/*.html"))
	// componentsは将来の拡張用（現在は未使用）
	// templates = template.Must(templates.ParseGlob("web/templates/components/*.html"))

	e.Renderer = &TemplateRenderer{
		templates: templates,
	}

	// ハンドラーの初期化
	h := handlers.NewHandler(dbPath, table)

	// ルーティング
	e.GET("/", h.Index)
	e.GET("/api/columns", h.GetColumns)
	e.GET("/api/filters", h.GetFilters)
	e.POST("/api/simpletab", h.Simpletab)
	e.POST("/api/crosstab", h.Crosstab)
	e.POST("/api/export", h.Export)

	return e
}
