package views

import (
	"errors"
	"log"
	"net/http"
	"net/mail"
	"path"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/socialnotes/mirror/fs"
	"github.com/socialnotes/mirror/mailer"
)

var (
	errFileExists = errors.New("file exists")
)

type UploadHandler struct {
	ts      *Templates
	storage *fs.FileStorage
	m       *mailer.M

	prefix string
}

func NewUploadHandler(ts *Templates, storage *fs.FileStorage, m *mailer.M, prefix string) *UploadHandler {
	return &UploadHandler{
		ts:      ts,
		storage: storage,
		m:       m,

		prefix: prefix,
	}
}

func checkEmail(email string) (string, error) {
	ma, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(ma.Address, "unitn.it") && !strings.HasSuffix(ma.Address, "unitn.eu") {
		return "", errors.New("not an unitn address")
	}

	return ma.Address, nil
}

func (uh *UploadHandler) handleUpload(rw http.ResponseWriter, req *http.Request) error {
	dir := path.Clean(strings.TrimPrefix(req.URL.Path, uh.prefix))

	email, err := checkEmail(req.FormValue("email"))
	if err != nil {
		uh.ts.Error(rw, http.StatusBadRequest, "The email provided was not valid. Remember that only unitn.it email addresses are accepted.")
		return nil
	}

	f, fh, err := req.FormFile("document")
	if err != nil {
		return err
	}
	defer f.Close()

	fm := fs.FileMeta{
		Email:      email,
		Authorized: false,
		Token:      uuid.NewV4().String(),
	}
	err = uh.storage.Create(dir, fh.Filename, fm, f)
	if err != nil {
		if err == fs.FileExists {
			uh.ts.Error(rw, http.StatusBadRequest, "A file with the same name already exists")
			return nil
		}
		return err
	}

	go func() {
		if err := uh.m.ConfirmUpload(fm.Email, fm.Info.Name(), fm.Token); err != nil {
			log.Printf("sending confirmation email for %s: %s\n", path.Join(dir, fm.Info.Name()), err)
		}
	}()

	rw.WriteHeader(http.StatusOK)
	uh.ts.Render(rw, "report.html", struct {
		Path     string
		Filename string
		Email    string
	}{
		Path:     dir,
		Filename: fm.Info.Name(),
		Email:    email,
	})
	return nil
}

func (uh *UploadHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	if req.Method != "POST" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}

	return uh.handleUpload(rw, req)
}
