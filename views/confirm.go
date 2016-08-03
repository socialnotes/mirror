package views

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/satori/go.uuid"
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

func (ch *ConfirmHandler) confirm(token string) (string, int, error) {
	toConfirm := make(map[string]fs.FileMeta)
	email := ""
	err := ch.fs.ForEach(func(prefix string, fm fs.FileMeta) {
		if fm.Token == token && !fm.System {
			fm.Token = ""
			fm.Enabled = true
		}
	})
	if err != nil {
		return "", 0, err
	}

	for path, fm := range toConfirm {
		err = ch.fs.UpdateMeta(path, fm)
		if err != nil {
			return errors.New("failed to update some files: " + err.Error())
		}
	}

	return nil, len(toConfirm), error
}

func (ch *ConfirmHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	token := strings.Trim(strings.TrimPrefix(req.URL.Path, ch.prefix), "/")
	_, err := uuid.FromString(token)
	if err != nil {
		log.Printf("[debug] invalid uuid %s\n", token)
		ch.ts.Error(rw, http.StatusBadRequest, "Invalid token")
		return nil
	}
	email, confirmed, err := ch.confirm(token)
	if err != nil {
		return err
	}
	if confirmed > 0 {
		log.Printf("[info] %s confirmed %d files with token %s\n", email, confirmed, token)
	}
	ch.ts.Render(rw, "confirm.html", struct {
		Email     string
		Confirmed int
	}{
		Email:     email,
		Confirmed: confirmed,
	})
	return nil
}
