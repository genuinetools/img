package registry

import (
	"fmt"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
)

// Delete removes a repository reference from the registry.
// https://docs.docker.com/registry/spec/api/#deleting-an-image
func (r *Registry) Delete(repository, ref string) error {
	// get digest first
	url := r.url("/v2/%s/manifests/%s", repository, ref)
	r.Logf("registry.manifests.get url=%s repository=%s ref=%s",
		url, repository, ref)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", schema2.MediaTypeManifest)

	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	digest := resp.Header.Get("Docker-Content-Digest")
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("Got status code: %d", resp.StatusCode)
	}

	// delete image
	url = r.url("/v2/%s/manifests/%s", repository, digest)
	r.Logf("registry.manifests.delete url=%s repository=%s ref=%s",
		url, repository, digest)

	req, err = http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", schema2.MediaTypeManifest)
	resp, err = r.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return fmt.Errorf("Got status code: %d", resp.StatusCode)
}
