package registry

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	bearerRegex = regexp.MustCompile(
		`^\s*Bearer\s+(.*)$`)
	basicRegex = regexp.MustCompile(`^\s*Basic\s+.*$`)

	// ErrBasicAuth indicates that the repository requires basic rather than token authentication.
	ErrBasicAuth = errors.New("basic auth required")
)

func parseAuthHeader(header http.Header) (*authService, error) {
	ch, err := parseChallenge(header.Get("www-authenticate"))
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func parseChallenge(challengeHeader string) (*authService, error) {
	if basicRegex.MatchString(challengeHeader) {
		return nil, ErrBasicAuth
	}

	match := bearerRegex.FindAllStringSubmatch(challengeHeader, -1)
	if d := len(match); d != 1 {
		return nil, fmt.Errorf("malformed auth challenge header: '%s', %d", challengeHeader, d)
	}
	parts := strings.SplitN(strings.TrimSpace(match[0][1]), ",", 3)

	var realm, service string
	var scope []string
	for _, s := range parts {
		p := strings.SplitN(s, "=", 2)
		if len(p) != 2 {
			return nil, fmt.Errorf("malformed auth challenge header: '%s'", challengeHeader)
		}
		key := p[0]
		value := strings.TrimSuffix(strings.TrimPrefix(p[1], `"`), `"`)
		switch key {
		case "realm":
			realm = value
		case "service":
			service = value
		case "scope":
			scope = strings.Fields(value)
		default:
			return nil, fmt.Errorf("unknown field in challege header %s: %v", key, challengeHeader)
		}
	}
	parsedRealm, err := url.Parse(realm)
	if err != nil {
		return nil, err
	}

	a := &authService{
		Realm:   parsedRealm,
		Service: service,
		Scope:   scope,
	}

	return a, nil
}
