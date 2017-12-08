package main

import (
	"crypto/tls"
	"net/http"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/jessfraz/reg/registry"
)

var (
	reProtocol = regexp.MustCompile("^https?://")
)

func newRegistryTransport(auth types.AuthConfig, dontVerifySSL bool) http.RoundTripper {
	url := strings.TrimSuffix(auth.ServerAddress, "/")

	if !reProtocol.MatchString(url) {
		url = "https://" + url
	}

	transport := http.DefaultTransport

	// If insecure set that way.
	if dontVerifySSL {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	tokenTransport := &registry.TokenTransport{
		Transport: transport,
		Username:  auth.Username,
		Password:  auth.Password,
	}
	basicAuthTransport := &registry.BasicTransport{
		Transport: tokenTransport,
		URL:       url,
		Username:  auth.Username,
		Password:  auth.Password,
	}
	errorTransport := &registry.ErrorTransport{
		Transport: basicAuthTransport,
	}

	return errorTransport
}
