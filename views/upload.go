package views

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/gigaroby/mirror/fs"
)

const (
	megaByte = 1 << 20

	maxInMemoryFormData = 10 * megaByte

	StatusUnprocessableEntity = 422
)

var (
	errFileExists   = errors.New("file exists")
	unitnMailRegExp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@([a-zA-Z0-9-]+\\.)*unitn\\.(it|eu)$")
)

type UploadHandler struct {
	ts *Templates
	fs fs.Dir
	db *bolt.DB

	prefix string
}

func NewUploadHandler(fs fs.Dir, ts *Templates, db *bolt.DB, prefix string) *UploadHandler {
	return &UploadHandler{
		fs: fs,
		ts: ts,
		db: db,

		prefix: prefix,
	}
}

func checkEmail(email string) (string, error) {
	ma, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}

	if !unitnMailRegExp.MatchString(ma.Address) {
		return "", errors.New("not an Unitn address")
	}

	return ma.Address, nil
}

func (uh *UploadHandler) handleUpload(rw http.ResponseWriter, req *http.Request) {
	var status int

	email, err := checkEmail(req.FormValue("email"))
	if err != nil {
		status = StatusUnprocessableEntity
		uh.ts.Error(rw, status, "Address not valid. Email must be an Unitn address.")
		return
	}

	directory := strings.TrimPrefix(req.URL.Path, uh.prefix)
	filename, err := uh.copyFiles(req, directory)
	if err != nil {
		log.Printf("[err] processing upload for %s: %s\n", path.Join(directory, filename), err)
		status := http.StatusInternalServerError
		uh.ts.Error(rw, status, http.StatusText(status))
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

	// TODO: add support for multiple upload
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

func (uh *UploadHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		uh.handleUpload(rw, req)
	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
