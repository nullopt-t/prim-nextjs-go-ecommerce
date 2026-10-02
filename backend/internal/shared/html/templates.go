package html

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

type Renderer struct {
	templates *template.Template
}

func NewRenderer() (*Renderer, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Renderer{
		templates: tmpl,
	}, nil
}

func (r *Renderer) Render(name string, data any) (string, error) {
	var buf bytes.Buffer

	target := name + ".html"
	if r.templates.Lookup(target) == nil {
		// Fallback to .en.html if specific localized template is missing
		fallback := name + ".en.html"
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
