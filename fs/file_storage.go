package fs

import (
	"database/sql"
	"errors"
	"io"
	"path"
	"strings"
	"sync"
	"time"
)

var (
	FileExists = errors.New("file exists")

	metaByName = `select (email, token, authorized, uploaded) from files where name = ?`
)

// FileStorage provides access to a file database
type FileStorage struct {
	FS FS

	// DB provides metadata storage
	DB *sql.DB
	// dbLock protects writes on DB.
	// this should not be necessary but sqlite requires it
	dbLock sync.RWMutex
}

//     name TEXT PRIMARY KEY,
//     email TEXT, -- email of uploader
//     token TEXT, -- token sent via email
//     authorized INTEGER, -- either 0 or 1
//     uploaded INTEGER -- datetime the file was uploaded

func metaFromRow(row *sql.Row, fm *FileMeta) error {
	err := row.Scan(&fm.Email, &fm.Token, &fm.Authorized, &fm.UploadTime)
	switch err {
	case sql.ErrNoRows:
		fm.Authorized = true
		fm.Email = "-"
	case nil:
	default:
		return err
	}
	return nil
}

type File interface {
	io.Reader
	io.Seeker
	io.Closer
}

func (fs *FileStorage) Open(filePath string) (File, error) {
	return nil, nil
}

// Stat returns information on the file with a given path
// Exists is set if the path eixsts and is either a directory or a file
// If the path is not a directory, fm contains the *FileMeta associated
// Returns err in case of problems with the bolt database
func (fs *FileStorage) Stat(filePath string) (bool, FileMeta, error) {
	fm := FileMeta{}
	fs.dbLock.RLock()
	defer fs.dbLock.RUnlock()
	err := metaFromRow(fs.DB.QueryRow(metaByName, filePath), &fm)
	if err != nil {
		return false, fm, err
	}

	fi, err := fs.FS.Stat(filePath)
	if err != nil {
		return false, fm, err
	}

	fm.Info = fi
	return true, fm, nil
}

// List returns all children of path.
// If path is not a directory no files will be returned.
func (fs *FileStorage) List(filePath string) ([]string, []FileMeta, error) {
	dirContent, err := fs.FS.Readdir(filePath)
	if err != nil {
		return nil, nil, err
	}

	dirs := make([]string, 0)
	files := make([]FileMeta, 0)

	for _, fi := range dirContent {
		if fi.IsDir() {
			dirs = append(dirs, fi.Name())
			continue
		}
		// it's a file
		files = append(files, FileMeta{
			Info: fi,
		})
	}

	filePath = strings.TrimRight(filePath, "/")
	// TODO[rob]: make this one query
	fs.dbLock.RLock()
	defer fs.dbLock.RUnlock()
	query, err := fs.DB.Prepare(metaByName)
	if err != nil {
		return nil, nil, err
	}
	defer query.Close()

	// fill authorized and uploaded for all files in dir
	for i, _ := range files {
		name := filePath + "/" + strings.TrimLeft(files[i].Info.Name(), "/")
		if err := metaFromRow(query.QueryRow(name), &files[i]); err != nil {
			return nil, nil, err
		}
	}

	return dirs, files, nil
}

func (fs *FileStorage) writeTx(txFunc func(tx *sql.Tx) error) error {
	fs.dbLock.Lock()
	defer fs.dbLock.Unlock()
	tx, err := fs.DB.Begin()
	if err != nil {
		return err
	}
	err = txFunc(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Create
func (fs *FileStorage) Create(dir, filename string, fm FileMeta, content io.Reader) error {
	filePath := path.Join(dir, filename)
	exists, _, err := fs.Stat(filePath)
	if err != nil {
		return err
	}
	if exists {
		return FileExists
	}

	return fs.writeTx(func(tx *sql.Tx) error {
		modTime := time.Now()
		f, err := fs.FS.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, content)
		if err != nil {
			return err
		}
		_, err = tx.Exec(
			`insert into files (name, email, token, authorized, uploaded) values (?, ?, ?, ?, ?)`,
			filePath, fm.Email, fm.Token, fm.Authorized, modTime,
		)
		return err
	})
}

// Confirm confirms all files with a given token
// returns the number of confirmed files or an error
func (fs *FileStorage) Confirm(token string) (int64, error) {
	confirmed := int64(0)
	return confirmed, fs.writeTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(
			`update files set confirmed = ? where token = ?`,
			true, token,
		)
		if err != nil {
			return err
		}
		confirmed, err = res.RowsAffected()
		return err
	})
}
