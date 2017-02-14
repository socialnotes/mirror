package views

import (
	"log"
	"net/http"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/socialnotes/mirror/fs"
)

type ConfirmHandler struct {
	ts *Templates
	fs *fs.FileStorage

	prefix string
}

func NewConfirmHandler(ts *Templates, fs *fs.FileStorage, prefix string) *ConfirmHandler {
	return &ConfirmHandler{
		ts: ts,
		fs: fs,

		prefix: prefix,
	}
}

func (ch *ConfirmHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	token := strings.Trim(strings.TrimPrefix(req.URL.Path, ch.prefix), "/")
	_, err := uuid.FromString(token)
	if err != nil {
		log.Printf("[debug] invalid uuid %s\n", token)
		ch.ts.Error(rw, http.StatusBadRequest, "Invalid token")
		return nil
	}
	confirmed, err := ch.fs.Confirm(token)
	if err != nil {
		return err
	}
	if confirmed > 0 {
		log.Printf("[info] confirmed %d files with token %s\n", confirmed, token)
	}
	ch.ts.Render(rw, "confirm.html", struct {
		Confirmed int64
	}{
		Confirmed: confirmed,
	})
	return nil
}
