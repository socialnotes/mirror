package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/boltdb/bolt"
)

var (
	boltFile = flag.String("bolt-file", "bolt.db", "bolt database file")
	sqlFile  = flag.String("sql-file", "mirror.db", "sqlite output file")
)

var (
	insertQuery = `insert into files (name, email, token, authorized, uploaded) values (?, ?, ?, ?, ?);`
)

type oldMeta struct {
	Name       string
	Size       int
	ModTime    time.Time
	Email      string
	Authorized bool
	Token      string
}

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
	fm := oldMeta{}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("files"))
		return b.ForEach(func(k, v []byte) error {
			ierr := json.Unmarshal(v, &fm)
			if ierr != nil {
				return err
			}
			_, ierr = insert.Exec(string(k), fm.Email, fm.Token, fm.Authorized, fm.ModTime)
			return ierr
		})
	})
	if err != nil {
		log.Fatalf("[crit] converting file: %s\n", err)
	}

	return
}
