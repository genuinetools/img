package registry

import (
	"net/http"
	"strings"
)

// BasicTransport defines the data structure for authentication via basic auth.
type BasicTransport struct {
	Transport http.RoundTripper
	URL       string
	Username  string
	Password  string
}

// RoundTrip defines the round tripper for basic auth transport.
func (t *BasicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.String(), t.URL) && req.Header.Get("Authorization") == "" {
		if t.Username != "" || t.Password != "" {
			req.SetBasicAuth(t.Username, t.Password)
		}
	}
	resp, err := t.Transport.RoundTrip(req)
	return resp, err
}
