package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// TokenTransport defines the data structure for authentication via tokens.
type TokenTransport struct {
	Transport http.RoundTripper
	Username  string
	Password  string
}

// RoundTrip defines the round tripper for token transport.
func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	authService, err := isTokenDemand(resp)
	if err != nil {
		return nil, err
	}

	if authService == nil {
		return resp, nil
	}

	return t.authAndRetry(authService, req)
}

type authToken struct {
	Token string `json:"token"`
}

func (t *TokenTransport) authAndRetry(authService *authService, req *http.Request) (*http.Response, error) {
	token, authResp, err := t.auth(authService)
	if err != nil {
		return authResp, err
	}

	return t.retry(req, token)
}

func (t *TokenTransport) auth(authService *authService) (string, *http.Response, error) {
	authReq, err := authService.Request(t.Username, t.Password)
	if err != nil {
		return "", nil, err
	}

	c := http.Client{
		Transport: t.Transport,
	}

	resp, err := c.Do(authReq)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", resp, err
	}

	var authToken authToken
	if err := json.NewDecoder(resp.Body).Decode(&authToken); err != nil {
		return "", nil, err
	}

	return authToken.Token, nil, nil
}

func (t *TokenTransport) retry(req *http.Request, token string) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return t.Transport.RoundTrip(req)
}

type authService struct {
	Realm   *url.URL
	Service string
	Scope   []string
}

func (a *authService) Request(username, password string) (*http.Request, error) {
	q := a.Realm.Query()
	q.Set("service", a.Service)
	for _, s := range a.Scope {
		q.Set("scope", s)
	}
	//	q.Set("scope", "repository:r.j3ss.co/htop:push,pull")
	a.Realm.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", a.Realm.String(), nil)

	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	return req, err
}

func isTokenDemand(resp *http.Response) (*authService, error) {
	if resp == nil {
		return nil, nil
	}
	if resp.StatusCode != http.StatusUnauthorized {
		return nil, nil
	}
	return parseAuthHeader(resp.Header)
}

// Token returns the required token for the specific resource url.
func (r *Registry) Token(url string) (string, error) {
	r.Logf("registry.token url=%s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	a, err := isTokenDemand(resp)
	if err != nil {
		return "", err
	}

	if a == nil {
		r.Logf("registry.token authService=nil")
		return "", nil
	}

	authReq, err := a.Request(r.Username, r.Password)
	if err != nil {
		return "", err
	}
	resp, err = r.Client.Do(authReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Getting token failed with StatusCode != StatusOK but %d", resp.StatusCode)
	}

	var authToken authToken
	if err := json.NewDecoder(resp.Body).Decode(&authToken); err != nil {
		return "", err
	}

	if authToken.Token == "" {
		return "", errors.New("Auth token cannot be empty")
	}

	return authToken.Token, nil
}
