package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gigaroby/mirror/views"
)

var (
	baseDir     = flag.String("base-dir", ".", "directory where files will be hosted, must be an absolute path")
	templateDir = flag.String("template-dir", "templates/", "directory containing templates")
)

func main() {
	flag.Parse()
	ts, err := views.NewTemplates(*templateDir, "*.html")
	if err != nil {
		log.Fatalf("[crit] parsing templates in %s: %s\n", *templateDir, err)
	}
	fs := http.Dir(*baseDir)
	sh := views.NewServerHandler(fs, ts)
	http.Handle("/", sh)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
