package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"log"

	"github.com/boltdb/bolt"
	"github.com/socialnotes/mirror/fs"
)

var (
	boltFile = flag.String("bolt-file", "bolt.db", "bolt database file")
	sqlFile  = flag.String("sql-file", "mirror.db", "sqlite output file")
)

var (
	insertQuery = `insert into files (name, uploader, token, authorized, uploader) values (?, ?, ?, ?, ?);`
)

func main() {
	flag.Parse()
	db, err := bolt.Open(*boltFile, 0600, nil)
	if err != nil {
		log.Fatalf("[crit] opening database file %s: %s\n", *boltFile, err)
	}
	defer db.Close()

	sdb, err := sql.Open("sqlite3", *sqlFile)
	if err != nil {
		log.Fatalf("[crit] opening database file %s: %s\n", *sqlFile, err)
	}
	defer sdb.Close()

	insert, err := sdb.Prepare(insertQuery)
	if err != nil {
		log.Fatalf("[crit] preparing query %s\n", err)
	}
	defer insert.Close()

	err = forEach(db, func(path string, fm fs.FileMeta) {
		enabled := 0
		if fm.Enabled {
			enabled = 1
		}
		_, ierr := insert.Exec(path, fm.Email, fm.Token, enabled, fm.Info.ModTime)
		if ierr != nil {
			log.Printf("[err] inserting data: %s\n", ierr)
		}
		// fmt.Printf("%s\t%+v\n", path, fm)
	})
	if err != nil {
		log.Fatalf("[crit] reading fileInfos: %s", err)
	}

}

func forEach(DB *bolt.DB, cb func(filePath string, fm fs.FileMeta)) error {
	return DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("files"))
		fm := fs.FileMeta{}
		return bucket.ForEach(func(k, v []byte) error {
			err := json.Unmarshal(v, &fm)
			if err != nil {
				return err
			}
			cb(string(k), fm)
			return nil
		})
	})
}
