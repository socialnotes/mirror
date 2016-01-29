package views

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/gigaroby/mirror/fs"
	"github.com/satori/go.uuid"
)

type ConfirmHandler struct {
	ts *Templates
	db *bolt.DB

	prefix string
}

func NewConfirmHandler(ts *Templates, db *bolt.DB, prefix string) *ConfirmHandler {
	return &ConfirmHandler{
		ts: ts,
		db: db,

		prefix: prefix,
	}
}

func (ch *ConfirmHandler) confirm(token string) (string, int, error) {
	processed := 0
	email := ""
	return email, processed, ch.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(fs.FilesBucket)
		toConfirm := make(map[string]fs.DBFile)
		dbf := fs.DBFile{}

		err := bucket.ForEach(func(k, v []byte) error {
			err := json.Unmarshal(v, &dbf)
			if err != nil {
				return err
			}
			if dbf.Token == token && !dbf.Authorized {
				toConfirm[string(k)] = dbf
			}
			return nil
		})
		if err != nil {
			return err
		}
		for path, dbf := range toConfirm {
			dbf.Authorized = true
			v, err := json.Marshal(dbf)
			if err != nil {
				return err
			}
			bucket.Put([]byte(path), v)
			email = dbf.Email
			processed++
		}
		return nil
	})
}

func (ch *ConfirmHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	token := strings.Trim(strings.TrimPrefix(req.URL.Path, ch.prefix), "/")
	_, err := uuid.FromString(token)
	if err != nil {
		log.Printf("[debug] invalid uuid %s\n", token)
		ch.ts.Error(rw, http.StatusBadRequest, "Invalid token")
		return nil
	}
	email, confirmed, err := ch.confirm(token)
	if err != nil {
		return err
	}
	if confirmed > 0 {
		log.Printf("[info] %s confirmed %d files with token %s\n", email, confirmed, token)
	}
	ch.ts.Render(rw, "confirm.html", struct {
		Email     string
		Confirmed int
	}{
		Email:     email,
		Confirmed: confirmed,
	})
	return nil
}
