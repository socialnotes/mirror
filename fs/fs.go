package fs

import (
	"io"
	"os"
)

// FS provides an interface for file content storage
type FS interface {
	// Open opens a file for reading.
	Open(path string) (io.ReadCloser, error)
	// Create creates a file with the specified path. If the file exists it is truncated.
	Create(path string) (io.WriteCloser, error)
	// Stat returns information for the file.
	Stat(path string) (os.FileInfo, error)
	// Readdir returns all fileInfos from a directory
	Readdir(dirPath string) ([]os.FileInfo, error)
}
