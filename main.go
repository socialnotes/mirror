package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gigaroby/mirror/fs"
	"github.com/gigaroby/mirror/views"
)

var (
	addr        = flag.String("port", ":8080", "bind to <address:port>")
	baseDir     = flag.String("base-dir", ".", "directory where files will be hosted, must be an absolute path")
	dbPath      = flag.String("db-path", "db.bolt", "bolt database file")
	templateDir = flag.String("template-dir", "templates/", "directory containing templates")
)

func main() {
	flag.Parse()
	ts, err := views.NewTemplates(*templateDir, "*.html")
	if err != nil {
		log.Fatalf("[crit] parsing templates in %s: %s\n", *templateDir, err)
	}
	db, err := bolt.Open(*dbPath, 0600, nil)
	if err != nil {
		log.Fatalf("[crit] opening database file %s: %s\n", *dbPath, err)
	}
	fs := fs.Dir(*baseDir)
	sh := views.ToHandler(views.NewServerHandler(fs, ts, db), ts)
	uh := views.NewUploadHandler(fs, ts, db, "/upload")
	http.Handle("/", sh)
	http.Handle("/upload/", uh)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
