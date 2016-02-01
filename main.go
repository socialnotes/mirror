package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/socialnotes/mirror/fs"
	"github.com/socialnotes/mirror/mailer"
	"github.com/socialnotes/mirror/views"
)

var (
	addr = flag.String("addr", ":8080", "bind to <address:port>")

	baseDir     = flag.String("base-dir", ".", "directory where files will be hosted, must be an absolute path")
	dbFile      = flag.String("db-file", "db.bolt", "bolt database file")
	templateDir = flag.String("template-dir", "templates/", "directory containing templates")

	mailgunDomain = flag.String("mailgun-domain", "socialnotes.eu", "mailgun domain to send emails from")
	mailgunSender = flag.String("mailgun-sender", "SocialNotes <files@socialnotes.eu>", "name of the email address that will be used to send emails")
	mailgunAPIKey = flag.String("mailgun-api-key", "", "mailgun api key")
)

func main() {
	flag.Parse()
	ts, err := views.NewTemplates(*templateDir, "*.html")
	if err != nil {
		log.Fatalf("[crit] parsing templates in %s: %s\n", *templateDir, err)
	}
	db, err := bolt.Open(*dbFile, 0600, nil)
	if err != nil {
		log.Fatalf("[crit] opening database file %s: %s\n", *dbFile, err)
	}
	defer db.Close()
	err = views.CheckDatabase(db)
	if err != nil {
		log.Fatalf("[crit] checking database %s: %s\n", *dbFile, err)
	}

	m, err := mailer.New(*mailgunDomain, *mailgunSender, *mailgunAPIKey)
	if err != nil {
		log.Fatalf("[crit] initializing mailer: %s\n", err)
	}

	fs := fs.Dir(*baseDir)
	sh := views.ToHandler(views.NewServerHandler(fs, ts, db), ts)
	uh := views.ToHandler(views.NewUploadHandler(fs, ts, db, m, "/upload"), ts)
	ch := views.ToHandler(views.NewConfirmHandler(ts, db, "/confirm"), ts)
	tos := views.ToHandler(views.NewStaticPageHandler(ts, "tos.html"), ts)
	http.Handle("/", sh)
	http.Handle("/tos.html", tos)
	http.Handle("/upload/", uh)
	http.Handle("/confirm/", ch)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
