package views

import "net/http"

func NewStaticPageHandler(ts *Templates, page string) *StaticPageHandler {
	return &StaticPageHandler{
		ts:   ts,
		page: page,
	}
}

// StaticPageHandler serves a single static html page
type StaticPageHandler struct {
	ts   *Templates
	page string
}

func (sh *StaticPageHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	rw.WriteHeader(http.StatusOK)
	sh.ts.Render(rw, sh.page, nil)
	return nil
}
