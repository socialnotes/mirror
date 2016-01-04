package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	baseDir     = flag.String("base-dir", ".", "directory where files will be hosted, must be an absolute path")
	templateDir = flag.String("template-dir", ".", "directory containing templates")

	errNoTemplate = errors.New("no such template")

	funcs = template.FuncMap{
		"humanizeBytes": humanizeBytes,
	}
)

func humanizeBytes(bytes int64) string {
	var (
		scale = []string{"b", "kb", "Mb", "Gb", "Tb", "Pb", "Eb", "Zb", "Yb"}
		rest  = bytes
		n     = 0
		s     = "mathisbroken"
	)
	for rest > 1000 {
		rest /= 1000
		n++
	}
	if n < len(scale) {
		s = scale[n]
	}

	return fmt.Sprintf("%d%s", rest, s)
}

type mirrorHandler struct {
	BaseDir string

	fs http.FileSystem
	ts *template.Template
}

func (m *mirrorHandler) list(rw http.ResponseWriter, req *http.Request, path string, dir http.File) {
	children, err := dir.Readdir(128)
	if err != nil && err != io.EOF {
		log.Printf("[err] fetching children of %s: %s\n", path, err)
		m.Error(rw, req, err)
		return
	}

	rw.WriteHeader(http.StatusOK)
	err = m.render(rw, "list.html", struct {
		Path     string
		Children []os.FileInfo
	}{
		Path:     path,
		Children: children,
	})
	if err != nil {
		log.Printf("[err] while rendering list template: %s\n", err)
	}
}

func (m *mirrorHandler) render(rw http.ResponseWriter, tn string, data interface{}) error {
	rw.Header().Add("Content-Type", "text/html")
	t := m.ts.Lookup(tn)
	if t == nil {
		return errNoTemplate
	}
	return t.Execute(rw, data)
}

func (m *mirrorHandler) Error(rw http.ResponseWriter, req *http.Request, err error) {
	var (
		msg    string
		status int
	)
	switch {
	case os.IsNotExist(err):
		msg, status = "404 page not found", http.StatusNotFound
	case os.IsPermission(err):
		msg, status = "403 Forbidden", http.StatusForbidden
	default:
		msg, status = "500 Internal Server Error", http.StatusInternalServerError
	}
	// TODO: render error template
	http.Error(rw, msg, status)
}

func (m *mirrorHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fpath := req.URL.Path
	f, err := m.fs.Open(fpath)
	if err != nil {
		log.Printf("[err] opening file %s: %s", fpath, err)
		m.Error(rw, req, err)
		return
	}
	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		log.Printf("[err] getting metadata for %s: %s", fpath, err)
		m.Error(rw, req, err)
		return
	}

	if s.IsDir() {
		m.list(rw, req, fpath, f)
		return
	}

	http.ServeContent(rw, req, s.Name(), s.ModTime(), f)
}

func main() {
	flag.Parse()
	ts, err := template.New("main").Funcs(funcs).ParseGlob(filepath.Join(*templateDir, "*.html"))
	if err != nil {
		log.Fatalf("[crit] parsing templates in %s: %s\n", *templateDir, err)
	}
	fs := http.Dir(*baseDir)
	h := &mirrorHandler{*baseDir, fs, ts}
	http.Handle("/", h)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
