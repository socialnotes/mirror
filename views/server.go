package views

import (
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

func NewServerHandler(fs http.FileSystem, ts *Templates) *ServerHandler {
	return &ServerHandler{
		fs: fs,
		ts: ts,
	}
}

type ServerHandler struct {
	fs http.FileSystem
	ts *Templates
}

func (sh *ServerHandler) list(rw http.ResponseWriter, req *http.Request, path string, dir http.File) {
	c, err := dir.Readdir(128)
	if err != nil && err != io.EOF {
		log.Printf("[err] fetching children of %s: %s\n", path, err)
		sh.error(rw, req, err)
		return
	}

	children := fileInfos(c)
	sort.Sort(children)

	rw.WriteHeader(http.StatusOK)
	err = sh.ts.Render(rw, "list.html", struct {
		Path     string
		Children []os.FileInfo
	}{
		Path:     path,
		Children: children,
	})
	if err != nil {
		log.Printf("[err] while rendering list template: %s\n", err)
	}
}

func (sh *ServerHandler) error(rw http.ResponseWriter, req *http.Request, err error) {
	var (
		msg    string
		status int
	)
	switch {
	case os.IsNotExist(err):
		msg, status = "404 page not found", http.StatusNotFound
	case os.IsPermission(err):
		msg, status = "403 Forbidden", http.StatusForbidden
	default:
		msg, status = "500 Internal Server Error", http.StatusInternalServerError
	}

	rw.WriteHeader(status)
	err = sh.ts.Render(rw, "error.html", struct {
		Status  int
		Message string
	}{
		Status:  status,
		Message: msg,
	})
	if err != nil {
		log.Printf("[err] while rendering error template: %s\n", err)
	}
}

func (sh *ServerHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fpath := req.URL.Path
	f, err := sh.fs.Open(fpath)
	if err != nil {
		log.Printf("[err] opening file %s: %s", fpath, err)
		sh.error(rw, req, err)
		return
	}
	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		log.Printf("[err] getting metadata for %s: %s", fpath, err)
		sh.error(rw, req, err)
		return
	}

	if s.IsDir() {
		sh.list(rw, req, fpath, f)
		return
	}

	http.ServeContent(rw, req, s.Name(), s.ModTime(), f)
}

type fileInfos []os.FileInfo

func (f fileInfos) Len() int {
	return len(f)
}

func (f fileInfos) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f fileInfos) Less(i, j int) bool {
	a, b := f[i], f[j]
	switch {
	case a.IsDir() && b.IsDir():
		return strings.Compare(a.Name(), b.Name()) <= 0
	case a.IsDir():
		return true
	case b.IsDir():
		return false
	default:
		return strings.Compare(a.Name(), b.Name()) <= 0
	}
}
