package views

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/socialnotes/mirror/fs"
)

func NewServerHandler(fs fs.Dir, ts *Templates, db *bolt.DB) *ServerHandler {
	return &ServerHandler{
		fs: fs,
		ts: ts,
		db: db,
	}
}

type ServerHandler struct {
	fs fs.Dir
	ts *Templates
	db *bolt.DB
}

func (sh *ServerHandler) list(rw http.ResponseWriter, req *http.Request, path string) error {
	dirs, files, err := directoryContent(sh.db, path)

	if path != "/" {
		dirs = append([]string{".."}, dirs...)
	}

	if err != nil {
		return errors.New("rendering list template: " + err.Error())
	}

	authorizedFiles := make([]fs.DBFile, 0, len(files))
	for _, f := range files {
		if !f.Authorized {
			continue
		}
		authorizedFiles = append(authorizedFiles, f)
	}

	rw.WriteHeader(http.StatusOK)
	sh.ts.Render(rw, "list.html", struct {
		Path        string
		Directories []string
		Files       []fs.DBFile
	}{
		Path:        path,
		Directories: dirs,
		Files:       authorizedFiles,
	})
	return nil
}

func (sh *ServerHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	var (
		path  = req.URL.Path
		found = true
		isDir = false
		file  = fs.DBFile{}
	)

	err := sh.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(fs.FilesBucket)
		prefix := []byte(path)
		c := bucket.Cursor()
		k, v := c.Seek(prefix)
		if k == nil || !bytes.HasPrefix(k, prefix) {
			found = false
			return nil
		}

		// it's a file
		if bytes.Equal(k, prefix) {
			return json.Unmarshal(v, &file)
		}
		isDir = true
		return nil
	})

	if err != nil {
		return err
	}

	if !found {
		sh.ts.Error(rw, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return nil
	}

	if isDir {
		if !strings.HasSuffix(path, "/") {
			http.Redirect(rw, req, path+"/", http.StatusTemporaryRedirect)
			return nil
		}
		return sh.list(rw, req, path)
	}

	if !file.Authorized {
		sh.ts.Error(rw, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return nil
	}

	f, err := sh.fs.Open(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return ViewErr(err, http.StatusNotFound)
		case os.IsPermission(err):
			return ViewErr(err, http.StatusForbidden)
		default:
			return err
		}
	}
	defer f.Close()
	http.ServeContent(rw, req, file.Name, file.ModTime, f)
	return nil
}
