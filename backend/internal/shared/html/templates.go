// Package html provides embedded HTML template rendering for localized transactional email bodies.
package html

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
)

// templateFS embeds all HTML templates located in the templates subdirectory at compile-time.
//
//go:embed templates/*.html
var templateFS embed.FS

// Renderer manages parsed HTML templates and renders them with contextual data.
type Renderer struct {
	templates *template.Template
}

// NewRenderer parses all embedded template files matching templates/*.html and returns a Renderer.
func NewRenderer() (*Renderer, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Renderer{
		templates: tmpl,
	}, nil
}

// Render executes the named template with the provided data payload and returns the rendered HTML string.
// If the localized template cannot be found, it strips any locale suffix and falls back to .en.html.
func (r *Renderer) Render(name string, data any) (string, error) {
	if r == nil || r.templates == nil {
		return "", fmt.Errorf("html renderer or template set is uninitialized")
	}

	var buf bytes.Buffer

	target := name + ".html"
	if r.templates.Lookup(target) == nil {
		// Strip any existing locale suffix (e.g. ".ar", ".en") before appending ".en.html"
		baseName := name
		if idx := strings.LastIndex(baseName, "."); idx != -1 {
			baseName = baseName[:idx]
		}
		fallback := baseName + ".en.html"
		if r.templates.Lookup(fallback) != nil {
			target = fallback
		}
	}

	err := r.templates.ExecuteTemplate(&buf, target, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

