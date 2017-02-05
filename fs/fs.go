package fs

import (
	"io"
	"time"
)

// FileInfo contains basic file information
type FileInfo struct {
	// Name is the name of the file
	Name string
	// Size is the size of the file in bytes
	Size int64
	// ModTime is the modification time for the file
	ModTime time.Time
}

// FS provides an interface for file content storage
type FS interface {
	// Open opens a file for reading.
	Open(path string) (io.ReadCloser, error)
	// Create creates a file with the specified path. If the file exists it is truncated.
	Create(path string) (io.WriteCloser, error)
	// Stat returns information for the file.
	Stat(path string) (*FileInfo, error)
}
