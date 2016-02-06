package views

import (
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
