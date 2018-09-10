package registry

import (
	"fmt"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
	digest "github.com/opencontainers/go-digest"
)

// Delete removes a repository digest from the registry.
// https://docs.docker.com/registry/spec/api/#deleting-an-image
func (r *Registry) Delete(repository string, digest digest.Digest) (err error) {
	url := r.url("/v2/%s/manifests/%s", repository, digest)
	r.Logf("registry.manifests.delete url=%s repository=%s digest=%s",
		url, repository, digest)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", fmt.Sprintf("%s;q=0.9", schema2.MediaTypeManifest))
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
