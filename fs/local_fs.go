package fs

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// A Dir implements FileSystem using the native file system restricted to a
// specific directory tree.
//
// While the OpenFile method takes '/'-separated paths, a Dir's string
// value is a filename on the native file system, not a URL, so it is separated
// by filepath.Separator, which isn't necessarily '/'.
//
// An empty Dir is treated as ".".
type Dir string

func (d Dir) openFile(name string) (*os.File, error) {
	path, err := d.cleanPath(name)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

// Open opens a file in read-only mode
func (d Dir) Open(name string) (io.ReadCloser, error) {
	return d.openFile(name)
}

// Create creates a file truncating it if it already exists.
// It also creates intermediate directories.
func (d Dir) Create(name string) (io.WriteCloser, error) {
	path, err := d.cleanPath(name)
	if err != nil {
		return nil, err
	}
	dir, _ := filepath.Split(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}

// Stat returns a FileInfo describing the named file.
func (d Dir) Stat(name string) (os.FileInfo, error) {
	path, err := d.cleanPath(name)
	if err != nil {
		return nil, err
	}
	return os.Stat(path)
}

func (d Dir) cleanPath(name string) (string, error) {
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return "", errors.New("invalid character in file path")
	}
	dir := string(d)
	if dir == "" {
		dir = "."
	}

	return filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name))), nil
}

func (d Dir) Readdir(path string) ([]os.FileInfo, error) {
	wd, err := d.openFile(path)
	if err != nil {
		return nil, err
	}

	return wd.Readdir(-1)
}
