package fs

import (
	"os"
	"time"
)

var (
	// FilesBucket is the name of the bucket containing file information
	FilesBucket = []byte("files")
)

// A DBFile is the structure used to serialize file information to boltdb
type DBFile struct {
	// Name is the filename
	Name string
	// Size is the file size in bytes
	Size int64
	// ModTime is the file time of last modification
	ModTime time.Time

	// email is the email of the person who uploaded the file
	Email string
	// Token is used to verify user uploads
	Token string
	// Authorized is set to true after the user verified the upload
	Authorized bool
}

// FromFileInfo returns an instance of DBFile constructed from a os.FileInfo
// Email, Token and Authorized are left at their default value
func FromFileInfo(fi os.FileInfo) DBFile {
	return DBFile{
		Name:    fi.Name(),
		Size:    fi.Size(),
		ModTime: fi.ModTime(),
	}
}
