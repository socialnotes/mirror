package views

import "net/http"

var (
	statusString = map[int]string{
		http.StatusBadRequest:          "Bad Request",
		http.StatusForbidden:           "Forbidden",
		http.StatusNotFound:            "404 Page Not Found",
		http.StatusInternalServerError: "500 Internal Server Error",
	}
)
