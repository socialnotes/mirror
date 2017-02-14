package fs

import (
	"os"
	"time"
)

// FileMeta contains both file information and metadata
type FileMeta struct {
	// Info contains information on the concrete file as a os.FileInfo
	Info os.FileInfo

	// UploadTime contains the datetime of the file upload
	UploadTime time.Time
	// Email contains the email of the user uploading the document
	Email string
	// Token contains the token for verification purposes
	Token string
	// Authorized indicates whether or not the file can be served
	Authorized bool
}

func (fm FileMeta) IsPublic() bool {
	return fm.Authorized
}
