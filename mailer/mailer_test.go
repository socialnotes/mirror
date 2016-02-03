package mailer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	domain = "example.com"
	sender = "Test Account <test@example.com>"
	apiKey = "test-api-key"
)

func SetMailgunEndpoint(m *M, endpoint string) {
	m.endpoint = endpoint
}

func TestNew(t *testing.T) {
	_, err := New(domain, sender, "")
	if err == nil {
		t.Error("an empty key should be invalid but it was allowed")
	}
	_, err = New(domain, "invalid", apiKey)
	if err == nil {
		t.Error("an invalid sender was provided but did not raise an error")
	}
	m, _ := New(domain, sender, apiKey)
	expected := mailgunApiDomain + "/" + domain + "/messages"
	if m.endpoint != expected {
		t.Errorf("invalid endpoint: expected %s, got %s", expected, m.endpoint)
	}
}

func expected(t *testing.T, field, expected, got string) {
	if expected != got {
		t.Errorf("expected %s=%s, got %s", field, expected, got)
	}
}

func TestConfirmUpload(t *testing.T) {
	reqCh := make(chan *http.Request, 1)
	statuses := make(chan int, 1)

	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqCh <- req
		req.ParseMultipartForm(1 * 1 << 10) // 1kb
		rw.WriteHeader(<-statuses)
	}))
	defer s.Close()

	m, _ := New(domain, sender, apiKey)
	SetMailgunEndpoint(m, s.URL)
	statuses <- http.StatusBadRequest
	if err := m.ConfirmUpload("test2@example.com", "testfile", "testtoken"); err == nil {
		t.Error("expected ConfirmUpload to return error")
	}
	<-reqCh // discard request, we don't need to look at it

	statuses <- http.StatusOK
	err := m.ConfirmUpload("test2@example.com", "testfile", "testtoken")
	if err != nil {
		t.Errorf("expected ConfirmUpload not to return errors, got %s", err)
	}
	req := <-reqCh
	defer req.Body.Close()
	user, pwd, ok := req.BasicAuth()
	if !ok || user != "api" || pwd != apiKey {
		t.Error("missing http basic auth or wrong credentials")
	}

	err = req.ParseMultipartForm(1000)
	if err != nil {
		t.Errorf("could not parse form: %s", err)
	}

	from := req.FormValue("from")
	expected(t, "from", sender, from)
	to := req.FormValue("to")
	expected(t, "to", "test2@example.com", to)
	subject := req.FormValue("subject")
	expected(t, "subject", "confirm upload of testfile", subject)
	text := req.FormValue("text")
	if !strings.Contains(text, fmt.Sprintf("https://%s/confirm/%s", domain, "testtoken")) {
		t.Error("email does not contain confirmation link")
	}
}
