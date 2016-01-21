package views

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path"
	"strings"

	"github.com/gigaroby/mirror/fs"
)

const (
	megaByte = 1 << 20

	maxInMemoryFormData = 10 * megaByte
)

var (
	errFileExists = errors.New("file exists")
)

type UploadHandler struct {
	ts *Templates
	fs fs.Dir

	prefix string
}

func NewUploadHandler(fs fs.Dir, ts *Templates, prefix string) *UploadHandler {
	return &UploadHandler{
		fs: fs,
		ts: ts,

		prefix: prefix,
	}
}

func checkEmail(email string) (string, error) {
	ma, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(ma.Address, "unitn.it") {
		return "", errors.New("not a @unitn.it address")
	}

	return ma.Address, nil
}

func (uh *UploadHandler) handleUpload(rw http.ResponseWriter, req *http.Request) {
	var status int

	email, err := checkEmail(req.FormValue("email"))
	if err != nil {
		status = http.StatusBadRequest
		uh.ts.Error(rw, status, "Address not valid. Email must be a @unitn.it address.")
		return
	}

	directory := strings.TrimPrefix(req.URL.Path, uh.prefix)
	filename, err := uh.copyFiles(req, directory)
	if err != nil {
		log.Printf("[err] processing upload for %s: %s\n", path.Join(directory, filename), err)
		status := http.StatusInternalServerError
		uh.ts.Error(rw, status, statusString[status])
		return
	}

	// TODO: insert into bolt a record for the upload
	// TODO: send email and validate upload

	rw.WriteHeader(http.StatusOK)
	uh.ts.Render(rw, "report.html", struct {
		Path     string
		Filename string
		Email    string
	}{
		Path:     directory,
		Filename: filename,
		Email:    email,
	})
}

func (uh *UploadHandler) copyFiles(req *http.Request, dir string) (string, error) {
	_, err := uh.fs.Open(dir)
	if err != nil {
		return "", err
	}

	f, fh, err := req.FormFile("document")
	filePath := path.Join(dir, fh.Filename)
	_, err = uh.fs.Open(filePath)
	if !os.IsNotExist(err) {
		if err == nil {
			return "", errFileExists
		}
		return "", err
	}

	fsf, err := uh.fs.Create(filePath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(fsf, f)
	if err != nil {
		return "", err
	}
	return fh.Filename, nil
}

func (uh *UploadHandler) handleForm(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	uh.ts.Render(rw, "upload.html", struct {
		UploadPath string
		FilePath   string
	}{
		UploadPath: req.URL.Path,
		FilePath:   strings.TrimPrefix(req.URL.Path, uh.prefix),
	})
}

func (uh *UploadHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		uh.handleForm(rw, req)
	case "POST":
		uh.handleUpload(rw, req)
	}
}
