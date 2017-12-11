package registry

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

var (
	authChallengeRegex = regexp.MustCompile(
		`^\s*Bearer\s+realm="([^"]+)",service="([^"]+)"\s*$`)
	basicRegex     = regexp.MustCompile(`^\s*Basic\s+.*$`)
	challengeRegex = regexp.MustCompile(
		`^\s*Bearer\s+realm="([^"]+)",service="([^"]+)",scope="([^"]+)"\s*$`)

	scopeSeparatorRegex = regexp.MustCompile(`\s+`)
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
		return nil, nil
	}

	match := challengeRegex.FindAllStringSubmatch(challengeHeader, -1)

	if len(match) != 1 {
		match = authChallengeRegex.FindAllStringSubmatch(challengeHeader, -1)
		if len(match) != 1 {
			return nil, fmt.Errorf("malformed auth challenge header: '%s'", challengeHeader)
		}
	}

	parsedRealm, err := url.Parse(match[0][1])
	if err != nil {
		return nil, err
	}

	a := &authService{
		Realm:   parsedRealm,
		Service: match[0][2],
	}

	if len(match[0]) >= 4 {
		a.Scope = scopeSeparatorRegex.Split(match[0][3], -1)
	}

	return a, nil
}
