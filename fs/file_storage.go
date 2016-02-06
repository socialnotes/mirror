package fs

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

var (
	FileExists = errors.New("file exists")
)

// FileMeta contains both file information and metadata
type FileMeta struct {
	// Contains file information
	Info FileInfo

	// Email contains the email of the user uploading the document
	Email string
	// Token contains the token for verification purposes
	Token string
	// Enabled indicates whether or not the file can be served
	Enabled bool
	// System indicates a placeholder file used to show empty directories
	System bool
}

func (fm FileMeta) IsPublic() bool {
	return fm.Enabled && !fm.System
}

// FileStorage provides access to a file database
type FileStorage struct {
	// FileStore provides physical storage for files
	FileStore FS
	// DB provides metadata storage
	DB *bolt.DB
}

// Stat returns information on the file with a given path
// Exists is set if the path eixsts and is either a directory or a file
// If the path is not a directory, fm contains the *FileMeta associated
// Returns err in case of problems with the bolt database
func (f *FileStorage) Stat(path string) (exists, isDir bool, fm FileMeta, err error) {
	fm = FileMeta{}
	return exists, isDir, fm, f.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(FilesBucket)
		prefix := []byte(path)
		k, v := bucket.Cursor().Seek(prefix)
		if k == nil || !bytes.HasPrefix(k, prefix) {
			exists = false
			return nil
		}

		exists = true
		if bytes.Equal(k, prefix) {
			isDir = false
			return json.Unmarshal(v, &fm)
		}

		isDir = true
		return nil
	})
}

// List returns all children of path.
// If path is not a directory no files will be returned.
func (f *FileStorage) List(path string) (dirs []string, fms []FileMeta, err error) {
	dirs = make([]string, 0, 4)
	fms = make([]FileMeta, 0, 4)
	return dirs, fms, f.DB.View(func(tx *bolt.Tx) error {
		prefix := []byte(path)
		bucket := tx.Bucket(FilesBucket)
		c := bucket.Cursor()
		last := []byte{}
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			k = bytes.TrimPrefix(k, prefix)
			// there is still a slash after the prefix has been stripped
			// therefore it is a directory
			if pos := bytes.IndexByte(k, '/'); pos > -1 {
				dir := k[:pos]
				// if I already encountered the directory, just skip
				if bytes.Equal(last, dir) {
					continue
				}
				last = dir[:]
				dirs = append(dirs, string(dir))
				continue
			}
			// it's a file
			fm := FileMeta{}
			if err := json.Unmarshal(v, &fm); err != nil {
				return err
			}
			fms = append(fms, fm)
		}

		return nil
	})
}

// Create
func (fs *FileStorage) Create(dir, filename string, fm FileMeta, content io.Reader) error {
	path := path.Join(dir, filename)
	exists, _, _, err := fs.Stat(path)
	if err != nil {
		return err
	}
	if exists {
		return FileExists
	}
	return fs.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(FilesBucket)
		modTime := time.Now()
		f, err := fs.FileStore.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		n, err := io.Copy(f, content)
		if err != nil {
			return err
		}
		fm.Info = FileInfo{
			Name:    filename,
			Size:    n,
			ModTime: modTime,
		}

		data, err := json.Marshal(fm)
		if err != nil {
			return err
		}
		bucket.Put([]byte(path), data)
		return nil
	})
}
