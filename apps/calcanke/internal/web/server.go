package web

import (
	"html/template"
	"io"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/naozine/nz-mono-repo/apps/calcanke/internal/project"
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
func NewServer(dbPath, table, projectsDir string) *echo.Echo {
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

	// プロジェクトリポジトリの初期化
	projectRepo, err := project.NewRepository(projectsDir + "/projects.db")
	if err != nil {
		log.Fatalf("Failed to initialize project repository: %v", err)
	}

	// ハンドラーの初期化
	h := handlers.NewHandler(dbPath, table)
	projectHandler := handlers.NewProjectHandler(projectRepo, projectsDir)

	// ルーティング - プロジェクト管理
	e.GET("/", projectHandler.List)
	e.GET("/projects/new", projectHandler.ShowCreateForm)
	e.POST("/projects", projectHandler.Create)
	e.GET("/projects/:id/upload", projectHandler.ShowUploadForm)
	e.POST("/api/projects/:id/upload", projectHandler.Upload)
	e.GET("/api/projects", projectHandler.GetProjectListAPI)
	e.GET("/api/projects/:id", projectHandler.GetProjectAPI)
	e.DELETE("/api/projects/:id", projectHandler.Delete)

	// ルーティング - プロジェクトごとの集計機能
	e.GET("/projects/:id", projectHandler.ShowAnalysis)
	e.GET("/api/projects/:id/columns", projectHandler.GetProjectColumns)
	e.GET("/api/projects/:id/filters", projectHandler.GetProjectFilters)
	e.POST("/api/projects/:id/simpletab", projectHandler.ProjectSimpletab)
	e.POST("/api/projects/:id/crosstab", projectHandler.ProjectCrosstab)
	e.POST("/api/projects/:id/export", projectHandler.ProjectExport)

	// ルーティング - 派生列管理
	e.GET("/api/projects/:id/derived-columns", projectHandler.GetDerivedColumns)
	e.POST("/api/projects/:id/derived-columns", projectHandler.AddDerivedColumn)
	e.PUT("/api/projects/:id/derived-columns/:index", projectHandler.UpdateDerivedColumn)
	e.DELETE("/api/projects/:id/derived-columns/:index", projectHandler.DeleteDerivedColumn)

	// ルーティング - 派生列テンプレート
	e.GET("/api/projects/:id/derived-columns/templates", projectHandler.GetDerivedColumnTemplates)
	e.POST("/api/projects/:id/derived-columns/import", projectHandler.ImportDerivedColumnTemplates)

	// ルーティング - フィルタ管理
	e.GET("/api/projects/:id/filters-config", projectHandler.GetFiltersConfig)
	e.POST("/api/projects/:id/filters-config", projectHandler.AddFilterConfig)
	e.PUT("/api/projects/:id/filters-config/:index", projectHandler.UpdateFilterConfig)
	e.DELETE("/api/projects/:id/filters-config/:index", projectHandler.DeleteFilterConfig)

	// ルーティング - 列順序管理
	e.GET("/api/projects/:id/column-orders", projectHandler.GetColumnOrders)
	e.PUT("/api/projects/:id/column-orders", projectHandler.UpdateColumnOrders)

	// ルーティング - 集計機能（既存、後でプロジェクトIDベースに変更予定）
	e.GET("/analysis", h.Index) // 一時的に /analysis に移動
	e.GET("/api/columns", h.GetColumns)
	e.GET("/api/filters", h.GetFilters)
	e.POST("/api/simpletab", h.Simpletab)
	e.POST("/api/crosstab", h.Crosstab)
	e.POST("/api/export", h.Export)

	return e
}
