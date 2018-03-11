package registry

import (
	"io/ioutil"
	"net/http"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
)

// Manifest returns the manifest for a specific repository:tag.
func (r *Registry) Manifest(repository, ref string) (distribution.Manifest, error) {
	uri := r.url("/v2/%s/manifests/%s", repository, ref)
	r.Logf("registry.manifests uri=%s repository=%s ref=%s", uri, repository, ref)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", schema2.MediaTypeManifest)
	req.Header.Add("Accept", manifestlist.MediaTypeManifestList)

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r.Logf("registry.manifests resp.Status=%s, body=%s", resp.Status, body)

	m, _, err := distribution.UnmarshalManifest(resp.Header.Get("Content-Type"), body)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// ManifestList gets the registry v2 manifest list.
func (r *Registry) ManifestList(repository, ref string) (manifestlist.ManifestList, error) {
	uri := r.url("/v2/%s/manifests/%s", repository, ref)
	r.Logf("registry.manifests uri=%s repository=%s ref=%s", uri, repository, ref)

	var m manifestlist.ManifestList
	if _, err := r.getJSON(uri, &m, true); err != nil {
		r.Logf("registry.manifests response=%v", m)
		return m, err
	}

	return m, nil
}

// ManifestV2 gets the registry v2 manifest.
func (r *Registry) ManifestV2(repository, ref string) (schema2.Manifest, error) {
	uri := r.url("/v2/%s/manifests/%s", repository, ref)
	r.Logf("registry.manifests uri=%s repository=%s ref=%s", uri, repository, ref)

	var m schema2.Manifest
	if _, err := r.getJSON(uri, &m, true); err != nil {
		r.Logf("registry.manifests response=%v", m)
		return m, err
	}

	return m, nil
}

// ManifestV1 gets the registry v1 manifest.
func (r *Registry) ManifestV1(repository, ref string) (schema1.SignedManifest, error) {
	uri := r.url("/v2/%s/manifests/%s", repository, ref)
	r.Logf("registry.manifests uri=%s repository=%s ref=%s", uri, repository, ref)

	var m schema1.SignedManifest
	if _, err := r.getJSON(uri, &m, false); err != nil {
		r.Logf("registry.manifests response=%v", m)
		return m, err
	}

	return m, nil
}
