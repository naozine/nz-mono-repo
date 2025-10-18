package ui

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templatesFS embed.FS

// Templates returns the parsed templates
func Templates() (*template.Template, error) {
	return template.ParseFS(templatesFS, "templates/*.html")
}
