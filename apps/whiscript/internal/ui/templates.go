package ui

import "embed"

//go:embed templates/*.html templates/projects/*.html
var TemplatesFS embed.FS
