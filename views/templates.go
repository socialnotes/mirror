package views

import (
	"errors"
	"html/template"
	"log"
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

// TODO: make Render return an error instead
func (t *Templates) Exists(templateName string) bool {
	return t.main.Lookup(templateName) != nil
}

// Render renders the named template passing the provided data as context.
// If the template does not exist, Render panics
func (t *Templates) Render(rw http.ResponseWriter, templateName string, data interface{}) {
	rw.Header().Add("Content-Type", "text/html")
	template := t.main.Lookup(templateName)
	if template == nil {
		panic("no such template " + templateName)
	}
	err := template.Execute(rw, data)
	if err != nil {
		log.Printf("[err] while rendering template %s: %s\n", templateName, err)
	}
}

// Error renders the error.html template with the given parameters
func (t *Templates) Error(rw http.ResponseWriter, status int, message string) {
	rw.WriteHeader(status)
	t.Render(rw, "error.html", struct {
		Status  int
		Message string
	}{
		Status:  status,
		Message: message,
	})
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
