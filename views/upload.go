package views

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gigaroby/mirror/fs"
)

const (
	megaByte = 1 << 20

	maxInMemoryFormData = 10 * megaByte
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

func (uh *UploadHandler) error(rw http.ResponseWriter, req *http.Request, status int, message string) error {
	rw.WriteHeader(status)
	return uh.ts.Render(rw, "error.html", struct {
		Status  int
		Message string
	}{
		Status:  status,
		Message: message,
	})
}

func (uh *UploadHandler) handleUpload(rw http.ResponseWriter, req *http.Request) (completed bool, code int, message string) {
	dir := req.URL.Path
	if !uh.fs.Exists(dir) {
		return false, http.StatusNotFound, fmt.Sprintf("can not upload in %s, the directory does not exist", dir)
	}
	err := req.ParseMultipartForm(maxInMemoryFormData)
	if err != nil {
		return false, http.StatusInternalServerError, "Internal server error"
	}
	form := req.MultipartForm
	f := form.File["document"]
	return true, 0, ""
}

func (uh *UploadHandler) handleForm(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	err := uh.ts.Render(rw, "upload.html", struct {
		Path string
	}{
		Path: strings.TrimSuffix(req.URL.Path, "/"),
	})
	if err != nil {
		log.Printf("[err] while rendering upload template: %s\n", err)
	}
}

func (uh *UploadHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.URL.Path = strings.TrimPrefix(req.URL.Path, uh.prefix)
	switch req.Method {
	case "GET":
		uh.handleForm(rw, req)
	case "POST":
		completed, status, message := uh.handleUpload(rw, req)
		if !completed {
			err := uh.error(rw, req, status, message)
			if err != nil {
				log.Printf("[err] while rendering error template: %s\n", err)
			}
		}
	}
}
