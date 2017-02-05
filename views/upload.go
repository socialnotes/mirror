package views

import (
	"errors"
	"log"
	"net/http"
	"net/mail"
	"path"
	"strings"

	"github.com/satori/go.uuid"
	"github.com/socialnotes/mirror/fs"
	"github.com/socialnotes/mirror/mailer"
)

var (
	errFileExists = errors.New("file exists")
)

type UploadHandler struct {
	ts *Templates
	fs *fs.FileStorage
	m  *mailer.M

	prefix string
}

func NewUploadHandler(fs *fs.FileStorage, ts *Templates, m *mailer.M, prefix string) *UploadHandler {
	return &UploadHandler{
		fs: fs,
		ts: ts,
		m:  m,

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
		Email:   email,
		Enabled: false,
		Token:   uuid.NewV4().String(),
		System:  false,
	}
	err := uh.fs.Create(dir, fh.Filename, f)
	if err != nil {
		if err == fs.FileExists {
			uh.ts.Error(rw, status, "A file with the same name already exists")
			return nil
		}
		return err
	}

	go func() {
		if err := uh.m.ConfirmUpload(dbf.Email, info.Name(), dbf.Token); err != nil {
			log.Printf("sending confirmation email for %s: %s\n", path.Join(directory, info.Name()), err)
		}
	}()

	rw.WriteHeader(http.StatusOK)
	uh.ts.Render(rw, "report.html", struct {
		Path     string
		Filename string
		Email    string
	}{
		Path:     dir,
		Filename: info.Name(),
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
