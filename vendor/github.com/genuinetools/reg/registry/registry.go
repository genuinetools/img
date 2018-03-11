package registry

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/docker/api/types"
)

// Registry defines the client for retriving information from the registry API.
type Registry struct {
	URL      string
	Domain   string
	Username string
	Password string
	Client   *http.Client
	Logf     LogfCallback
}

var reProtocol = regexp.MustCompile("^https?://")

// LogfCallback is the callback for formatting logs.
type LogfCallback func(format string, args ...interface{})

// Quiet discards logs silently.
func Quiet(format string, args ...interface{}) {}

// Log passes log messages to the logging package.
func Log(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// New creates a new Registry struct with the given URL and credentials.
func New(auth types.AuthConfig, debug bool) (*Registry, error) {
	transport := http.DefaultTransport

	return newFromTransport(auth, transport, debug)
}

// NewInsecure creates a new Registry struct with the given URL and credentials,
// using a http.Transport that will not verify an SSL certificate.
func NewInsecure(auth types.AuthConfig, debug bool) (*Registry, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return newFromTransport(auth, transport, debug)
}

func newFromTransport(auth types.AuthConfig, transport http.RoundTripper, debug bool) (*Registry, error) {
	url := strings.TrimSuffix(auth.ServerAddress, "/")

	if !reProtocol.MatchString(url) {
		url = "https://" + url
	}

	tokenTransport := &TokenTransport{
		Transport: transport,
		Username:  auth.Username,
		Password:  auth.Password,
	}
	basicAuthTransport := &BasicTransport{
		Transport: tokenTransport,
		URL:       url,
		Username:  auth.Username,
		Password:  auth.Password,
	}
	errorTransport := &ErrorTransport{
		Transport: basicAuthTransport,
	}

	// set the logging
	logf := Quiet
	if debug {
		logf = Log
	}

	registry := &Registry{
		URL:    url,
		Domain: reProtocol.ReplaceAllString(url, ""),
		Client: &http.Client{
			Transport: errorTransport,
		},
		Username: auth.Username,
		Password: auth.Password,
		Logf:     logf,
	}

	if err := registry.Ping(); err != nil {
		return nil, err
	}

	return registry, nil
}

// url returns a registry URL with the passed arguements concatenated.
func (r *Registry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

func (r *Registry) getJSON(url string, response interface{}, addV2Header bool) (http.Header, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if addV2Header {
		req.Header.Add("Accept", schema2.MediaTypeManifest)
		req.Header.Add("Accept", manifestlist.MediaTypeManifestList)
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	r.Logf("registry.registry resp.Status=%s", resp.Status)

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, err
	}

	return resp.Header, nil
}
