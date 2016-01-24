package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/gigaroby/mirror/fs"
)

var (
	email   = flag.String("email", "admin@mirror.marvinware.com", "email to upload files as")
	baseDir = flag.String("base-dir", ".", "directory of files to index")
	dbPath  = flag.String("db-path", "db.bolt", "bolt database file")

	verbose = flag.Bool("verbose", true, "be verbose during indexing")
)

func walker(b *bolt.Bucket, prefix, email string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("[err] indexing file %s: %s\n", path, err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		dbf := fs.DBFile{
			Name:    info.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),

			Email:      email,
			Authorized: true,
		}
		value, err := json.Marshal(dbf)
		if err != nil {
			log.Printf("[err] serializing file %s to json: %s\n", path, err)
			return nil
		}
		path = strings.TrimPrefix(path, prefix)
		return b.Put([]byte(path), value)
	}
}

func main() {
	flag.Parse()
	err := os.Remove(*dbPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("[crit] removing %s: %s\n", *dbPath, err)
	}
	db, err := bolt.Open(*dbPath, 0600, nil)
	if err != nil {
		log.Fatalf("[crit] opening database file %s: %s\n", *dbPath, err)
	}
	defer db.Close()
	prefix, err := filepath.Abs(filepath.Clean(*baseDir))
	if err != nil {
		log.Fatalf("[crit] obtaining absolute path for baseDir %s: %s\n", *baseDir, err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucket(fs.FilesBucket)
		if err != nil {
			return err
		}
		walkFn := walker(bucket, prefix, *email)
		if *verbose {
			wf := walkFn
			walkFn = func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					log.Printf("[info] indexing %s\n", path)
				}
				return wf(path, info, err)
			}
		}
		return filepath.Walk(prefix, walkFn)
	})
	if err != nil {
		log.Fatalf("[crit] while building index: %s\n", err)
	}
}
