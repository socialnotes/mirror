package views

import (
	"net/http"
	"path"
	"strings"

	"github.com/socialnotes/mirror/fs"
)

func NewServerHandler(ts *Templates, fs *fs.FileStorage) *ServerHandler {
	return &ServerHandler{
		fs: fs,
		ts: ts,
	}
}

type ServerHandler struct {
	fs *fs.FileStorage
	ts *Templates
}

func (sh *ServerHandler) list(path string, rw http.ResponseWriter) error {
	dirs, files, err := sh.fs.List(path)
	if err != nil {
		return err
	}

	if path != "/" {
		dirs = append([]string{".."}, dirs...)
	}

	publicFiles := make([]fs.FileMeta, 0, len(files))
	for _, f := range files {
		if f.IsPublic() {
			publicFiles = append(publicFiles, f)
		}
	}

	rw.WriteHeader(http.StatusOK)
	sh.ts.Render(rw, "list.html", struct {
		Path        string
		Directories []string
		Files       []fs.FileMeta
	}{
		Path:        path,
		Directories: dirs,
		Files:       publicFiles,
	})
	return nil
}

func (sh *ServerHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	path := path.Clean(req.URL.Path)
	exists, isDir, fm, err := sh.fs.Stat(path)
	if err != nil {
		return err
	}

	if !exists {
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

	if !fm.IsPublic() {
		sh.ts.Error(rw, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return nil
	}

	f, err := sh.fs.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	http.ServeContent(rw, req, file.Name, file.ModTime, f)
	return nil
}
