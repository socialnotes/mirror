package views

import (
	"log"
	"net/http"
)

// ViewError contains both the original error and the status code to
// be returned to the client
type ViewError struct {
	Err    error
	Status int
}

func (v *ViewError) Error() string {
	return v.Err.Error()
}

func ViewErr(err error, status int) *ViewError {
	return &ViewError{
		Err:    err,
		Status: status,
	}
}

type ViewHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

func ToHandler(v ViewHandler, ts *Templates) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := v.ServeHTTP(rw, req)
		log.Printf("here\n")
		if err == nil {
			return
		}

		log.Printf("[err] %s\n", err)
		status := http.StatusInternalServerError
		if verr, ok := err.(*ViewError); ok {
			status = verr.Status
		}

		ts.Error(rw, status, http.StatusText(status))
	})
}
