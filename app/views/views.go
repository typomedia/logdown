// Package views holds the embedded HTML templates. It exists so the templates
// can live in their own directory: go:embed can only reach files within its
// own package directory, so the embed lives here and app/web consumes FS.
package views

import "embed"

// FS contains every page and layout template (*.html).
//
//go:embed *.html
var FS embed.FS
