package views

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/socialnotes/mirror/fs"
	"github.com/socialnotes/mirror/mailer"
	"github.com/satori/go.uuid"
)

var (
	errFileExists = errors.New("file exists")
)

type UploadHandler struct {
	ts *Templates
	fs fs.Dir
	db *bolt.DB
	m  *mailer.M

	prefix string
}

func NewUploadHandler(fs fs.Dir, ts *Templates, db *bolt.DB, m *mailer.M, prefix string) *UploadHandler {
	return &UploadHandler{
		fs: fs,
		ts: ts,
		db: db,
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
	var (
		status    int
		info      os.FileInfo
		directory = path.Clean(strings.TrimPrefix(req.URL.Path, uh.prefix))
		dbf       fs.DBFile
	)

	email, err := checkEmail(req.FormValue("email"))
	if err != nil {
		status = http.StatusBadRequest
		uh.ts.Error(rw, status, "The email provided was not valid. Remember that only unitn.it email addresses are accepted.")
		return nil
	}

	err = uh.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(fs.FilesBucket)
		if exists := bucket.Get([]byte(directory)) != nil; exists {
			return errFileExists
		}

		if info, err = uh.copyFiles(req, directory); err != nil {
			return err
		}
		dbf = fs.FromFileInfo(info)
		dbf.Authorized = false
		dbf.Email = email
		dbf.Token = uuid.Must(uuid.NewV4()).String()
		data, err := json.Marshal(dbf)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(path.Join(directory, info.Name())), data)
	})

	if err != nil {
		if err == errFileExists {
			status = http.StatusConflict
			uh.ts.Error(rw, status, "A file with the same name already exists")
			return nil
		}
		return fmt.Errorf("processing upload for %s: %s", path.Join(directory, info.Name()), err)
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
		Path:     directory,
		Filename: info.Name(),
		Email:    email,
	})
	return nil
}

func (uh *UploadHandler) copyFiles(req *http.Request, dir string) (os.FileInfo, error) {
	// TODO: create intermediate directories
	// TODO: add support for multiple upload
	f, fh, err := req.FormFile("document")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	filePath := path.Join(dir, fh.Filename)
	_, err = uh.fs.Stat(filePath)
	if !os.IsNotExist(err) {
		if err == nil {
			return nil, errFileExists
		}
		return nil, err
	}

	fsf, err := uh.fs.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer fsf.Close()

	_, err = io.Copy(fsf, f)
	if err != nil {
		return nil, err
	}

	fi, err := fsf.Stat()
	if err != nil {
		return nil, err
	}

	return fi, nil
}

func (uh *UploadHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	if req.Method != "POST" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}

	return uh.handleUpload(rw, req)
}
