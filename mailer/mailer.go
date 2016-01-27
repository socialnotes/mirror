package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/mail"
	"net/url"
	"text/template"
)

const (
	mailgunEndpoint = "https://api.mailgun.net/v3"
)

var (
	emailTemplate = template.Must(template.New("email").Parse(`
  Hi {{ .Email }},
  You uploaded {{ .Filename }} on {{ .Domain }}.
  To confirm the upload please visit the following link
  https://{{ .Domain }}/confirm/{{ .Token }}

  If you didn't upload files on {{ .Domain }} please ignore this email.

  Best regards,
  {{ .Domain }}
  `))
)

type M struct {
	domain  string
	from    string
	apiKey  string
	subject string

	endpoint string
	c        *http.Client
}

func New(domain, sender, apiKey string) (*M, error) {
	if apiKey == "" {
		return nil, errors.New("api key is invalid")
	}

	_, err := mail.ParseAddress(sender)
	if err != nil {
		return nil, errors.New("sender address is invalid")
	}

	return &M{
		domain: domain,
		from:   sender,
		apiKey: apiKey,

		endpoint: fmt.Sprintf("%s/%s/messages", mailgunEndpoint, domain),
		c:        &http.Client{},
	}, nil
}

func (m *M) ConfirmUpload(to, filename, token string) error {
	b := new(bytes.Buffer)
	err := emailTemplate.Execute(b, struct {
		Domain   string
		Email    string
		Filename string
		Token    string
	}{
		Domain:   m.domain,
		Email:    to,
		Filename: filename,
		Token:    token,
	})
	if err != nil {
		return err
	}
	v := url.Values{}
	v.Add("from", m.from)
	v.Add("to", to)
	v.Add("subject", "confirm upload of "+filename)
	v.Add("text", b.String())

	b.Reset()
	b.WriteString(v.Encode())

	req, err := http.NewRequest("POST", m.endpoint, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", m.apiKey)
	req.Header.Set("content-type", "multipart/form-data")
	req.Body = ioutil.NopCloser(b)

	res, err := m.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b.Reset()
		io.Copy(b, res.Body)
		return fmt.Errorf("email send failed, server responded [%d] %s", res.StatusCode, b.String())
	}

	return nil
}
