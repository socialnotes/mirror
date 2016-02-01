package views

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
	"github.com/socialnotes/mirror/fs"
)

// CheckDatabase performs sanity checks on the database provided
func CheckDatabase(db *bolt.DB) error {
	return db.View(func(tx *bolt.Tx) error {
		if tx.Bucket(fs.FilesBucket) == nil {
			return errors.New("no bucket named files")
		}
		return nil
	})
}

// directoryContent returns the files and directories contained in the indicated subdirectory
// errors are returned only in case of malformed records in the database
func directoryContent(db *bolt.DB, path string) (dirs []string, files []fs.DBFile, err error) {
	dirs = make([]string, 0)
	files = make([]fs.DBFile, 0)
	prefix := []byte(path)

	return dirs, files, db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(fs.FilesBucket)
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
			dbf := fs.DBFile{}
			if err := json.Unmarshal(v, &dbf); err != nil {
				return err
			}
			files = append(files, dbf)
		}

		return nil
	})
}
