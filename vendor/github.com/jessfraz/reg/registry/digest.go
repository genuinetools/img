package registry

import (
	"fmt"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
)

// Digest returns the digest for a repository and reference.
func (r *Registry) Digest(repository, ref string) (string, error) {
	url := r.url("/v2/%s/manifests/%s", repository, ref)
	r.Logf("registry.manifests.get url=%s repository=%s ref=%s",
		url, repository, ref)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", schema2.MediaTypeManifest)

	resp, err := r.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return "", fmt.Errorf("Got status code: %d", resp.StatusCode)
	}

	digest := resp.Header.Get("Docker-Content-Digest")
	return digest, nil
}
