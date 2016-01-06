package views

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
)

var (
	// ErrNoTemplate is returned when a requested template is not found
	ErrNoTemplate = errors.New("no such template")

	funcs = template.FuncMap{
		"humanizeBytes": humanizeBytes,
	}
)

// Templates handles operations on templates.
type Templates struct {
	main *template.Template
}

// Render renders the named template passing the provided data as context.
func (t *Templates) Render(rw http.ResponseWriter, templateName string, data interface{}) error {
	rw.Header().Add("Content-Type", "text/html")
	template := t.main.Lookup(templateName)
	if template == nil {
		return ErrNoTemplate
	}
	return template.Execute(rw, data)
}

// NewTemplates creates a new instance of Templates from the provided folder and glob
func NewTemplates(path, glob string) (*Templates, error) {
	completePath := filepath.Join(path, glob)
	main, err := template.New("main").Funcs(funcs).ParseGlob(completePath)
	if err != nil {
		return nil, err
	}

	return &Templates{main: main}, nil
}
