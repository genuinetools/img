package main

import (
	"crypto/tls"
	"net/http"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/jessfraz/distribution/registry/client/transport"
	"github.com/jessfraz/reg/registry"
)

var (
	defaultUserAgent = "jessfraz/img"
	reProtocol       = regexp.MustCompile("^https?://")
)

func newRegistryTransport(auth types.AuthConfig, dontVerifySSL bool) http.RoundTripper {
	url := strings.TrimSuffix(auth.ServerAddress, "/")

	if !reProtocol.MatchString(url) {
		url = "https://" + url
	}

	tp := http.DefaultTransport

	// If insecure set that way.
	if dontVerifySSL {
		tp = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	t := transport.NewTransport(tp, transportHeaders(defaultUserAgent, nil)...)

	tokenTransport := &registry.TokenTransport{
		Transport: t,
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

// transportHeaders returns request modifiers with a User-Agent and metaHeaders
func transportHeaders(userAgent string, metaHeaders http.Header) []transport.RequestModifier {
	modifiers := []transport.RequestModifier{}
	if userAgent != "" {
		modifiers = append(modifiers, transport.NewHeaderRequestModifier(http.Header{
			"User-Agent": []string{userAgent},
		}))
	}
	if metaHeaders != nil {
		modifiers = append(modifiers, transport.NewHeaderRequestModifier(metaHeaders))
	}
	return modifiers
}
