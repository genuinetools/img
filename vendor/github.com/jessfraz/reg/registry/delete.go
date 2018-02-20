package registry

import (
	"fmt"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
)

// Delete removes a repository reference from the registry.
// https://docs.docker.com/registry/spec/api/#deleting-an-image
func (r *Registry) Delete(repository, ref string) error {
	// Get the digest first.
	digest, err := r.Digest(repository, ref)
	if err != nil {
		return err
	}

	// If we couldn't get the digest because it was not found just try and delete the ref they passed.
	if digest == "" {
		digest = ref
	}

	// Delete the image.
	url := r.url("/v2/%s/manifests/%s", repository, digest)
	r.Logf("registry.manifests.delete url=%s repository=%s ref=%s",
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
