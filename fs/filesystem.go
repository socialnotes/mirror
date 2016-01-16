package fs

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func checkName(name string) error {
	return nil
}

// A Dir implements FileSystem using the native file system restricted to a
// specific directory tree.
//
// While the OpenFile method takes '/'-separated paths, a Dir's string
// value is a filename on the native file system, not a URL, so it is separated
// by filepath.Separator, which isn't necessarily '/'.
//
// An empty Dir is treated as ".".
type Dir string

// Open opens a file in read-only mode
func (d Dir) Open(name string) (*os.File, error) {
	return d.openFile(name, os.O_RDONLY, 0)
}

// Create creates a file truncating it if it already exists
func (d Dir) Create(name string) (*os.File, error) {
	return d.openFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Exists checks whether a file exists
func (d Dir) Exists(name string) bool {
	f, err := d.Open(name)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func (d Dir) openFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return nil, errors.New("http: invalid character in file path")
	}
	dir := string(d)
	if dir == "" {
		dir = "."
	}
	f, err := os.OpenFile(
		filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name))),
		flag,
		perm,
	)
	if err != nil {
		return nil, err
	}
	return f, nil
}
