package clair

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Clair defines the client for retriving information from the clair API.
type Clair struct {
	URL    string
	Client *http.Client
	Logf   LogfCallback
}

// LogfCallback is the callback for formatting logs.
type LogfCallback func(format string, args ...interface{})

// Quiet discards logs silently.
func Quiet(format string, args ...interface{}) {}

// Log passes log messages to the logging package.
func Log(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// New creates a new Clair struct with the given URL and credentials.
func New(url string, debug bool) (*Clair, error) {
	transport := http.DefaultTransport

	errorTransport := &ErrorTransport{
		Transport: transport,
	}

	// set the logging
	logf := Quiet
	if debug {
		logf = Log
	}

	registry := &Clair{
		URL: url,
		Client: &http.Client{
			Timeout:   5 * time.Minute,
			Transport: errorTransport,
		},
		Logf: logf,
	}

	return registry, nil
}

// url returns a clair URL with the passed arguements concatenated.
func (c *Clair) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", c.URL, pathSuffix)
	return url
}

func (c *Clair) getJSON(url string, response interface{}) (http.Header, error) {
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	c.Logf("clair.clair resp.Status=%s", resp.Status)

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		c.Logf("clair.clair resp.Status=%s, body=%s", resp.Status, response)
		return nil, err
	}

	return resp.Header, nil
}
