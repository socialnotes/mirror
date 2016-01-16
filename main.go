package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gigaroby/mirror/fs"
	"github.com/gigaroby/mirror/views"
)

var (
	baseDir     = flag.String("base-dir", ".", "directory where files will be hosted, must be an absolute path")
	templateDir = flag.String("template-dir", "templates/", "directory containing templates")
	addr        = flag.String("port", ":8080", "bind to <address:port>")
)

func main() {
	flag.Parse()
	ts, err := views.NewTemplates(*templateDir, "*.html")
	if err != nil {
		log.Fatalf("[crit] parsing templates in %s: %s\n", *templateDir, err)
	}
	fs := fs.Dir(*baseDir)
	sh := views.NewServerHandler(fs, ts)
	uh := views.NewUploadHandler(fs, ts, "/upload")
	http.Handle("/", sh)
	http.Handle("/upload/", uh)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
