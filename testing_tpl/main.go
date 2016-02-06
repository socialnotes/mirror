package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/socialnotes/mirror/views"
)

var (
	addr = flag.String("addr", ":8080", "bind to <address:port>")

	templateDir = flag.String("template-dir", "templates/", "directory containing templates")
	staticDir   = flag.String("static-dir", "static/", "directory containing static files")
)

func main() {
	flag.Parse()
	ts, _ := views.NewTemplates(*templateDir, "*.html")
	sh := &StaticPageHandler{ts: ts}

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(*staticDir))))
	http.Handle("/", sh)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

type StaticPageHandler struct {
	ts *views.Templates
}

func (sh *StaticPageHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	templateName := strings.Trim(req.URL.Path, "/")
	if !sh.ts.Exists(templateName) {
		sh.ts.Error(rw, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		fmt.Println("404")
		return
	}

	rw.WriteHeader(http.StatusOK)
	sh.ts.Render(rw, templateName, nil)
}
