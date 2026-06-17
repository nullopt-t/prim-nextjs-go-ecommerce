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

	err := r.templates.ExecuteTemplate(&buf, name+".html", data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
