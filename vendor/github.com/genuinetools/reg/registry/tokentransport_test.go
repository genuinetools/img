package registry

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docker/docker/api/types"
)

func TestErrBasicAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("www-authenticate", `Basic realm="Registry Realm",service="Docker registry"`)
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	authConfig := types.AuthConfig{
		Username:      "j3ss",
		Password:      "ss3j",
		ServerAddress: ts.URL,
	}
	r, err := New(authConfig, Opt{Insecure: true, Debug: true})
	if err != nil {
		t.Fatalf("expected no error creating client, got %v", err)
	}
	token, err := r.Token(ts.URL)
	if err != ErrBasicAuth {
		t.Fatalf("expected ErrBasicAuth getting token, got %v", err)
	}
	if token != "" {
		t.Fatalf("expected empty token, got %v", err)
	}
}
