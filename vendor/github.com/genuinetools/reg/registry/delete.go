package registry

import (
	"fmt"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
	ocd "github.com/opencontainers/go-digest"
)

// Delete removes a repository digest or reference from the registry.
// https://docs.docker.com/registry/spec/api/#deleting-an-image
func (r *Registry) Delete(repository, digest string) error {
	// If digest is not valid try resolving it as a reference
	if _, err := ocd.Parse(digest); err != nil {
		digest, err = r.Digest(repository, digest)
		if err != nil {
			return err
		}
		if digest == "" {
			return nil
		}
	}

	// Delete the image.
	url := r.url("/v2/%s/manifests/%s", repository, digest)
	r.Logf("registry.manifests.delete url=%s repository=%s digest=%s",
		url, repository, digest)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", schema2.MediaTypeManifest)
	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return fmt.Errorf("Got status code: %d", resp.StatusCode)
}
