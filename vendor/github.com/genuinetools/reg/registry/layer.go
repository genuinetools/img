package registry

import (
	"io"
	"net/http"
	"net/url"

	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

// DownloadLayer downloads a specific layer by digest for a repository.
func (r *Registry) DownloadLayer(repository string, digest digest.Digest) (io.ReadCloser, error) {
	url := r.url("/v2/%s/blobs/%s", repository, digest)
	r.Logf("registry.layer.download url=%s repository=%s digest=%s", url, repository, digest)

	resp, err := r.Client.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// UploadLayer uploads a specific layer by digest for a repository.
func (r *Registry) UploadLayer(repository string, digest reference.Reference, content io.Reader) error {
	uploadURL, token, err := r.initiateUpload(repository)
	if err != nil {
		return err
	}
	q := uploadURL.Query()
	q.Set("digest", digest.String())
	uploadURL.RawQuery = q.Encode()

	r.Logf("registry.layer.upload url=%s repository=%s digest=%s", uploadURL, repository, digest)

	upload, err := http.NewRequest("PUT", uploadURL.String(), content)
	if err != nil {
		return err
	}
	upload.Header.Set("Content-Type", "application/octet-stream")
	upload.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	_, err = r.Client.Do(upload)
	return err
}

// HasLayer returns if the registry contains the specific digest for a repository.
func (r *Registry) HasLayer(repository string, digest digest.Digest) (bool, error) {
	checkURL := r.url("/v2/%s/blobs/%s", repository, digest)
	r.Logf("registry.layer.check url=%s repository=%s digest=%s", checkURL, repository, digest)

	resp, err := r.Client.Head(checkURL)
	if err == nil {
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK, nil
	}

	urlErr, ok := err.(*url.Error)
	if !ok {
		return false, err
	}
	httpErr, ok := urlErr.Err.(*httpStatusError)
	if !ok {
		return false, err
	}
	if httpErr.Response.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, err
}

func (r *Registry) initiateUpload(repository string) (*url.URL, string, error) {
	initiateURL := r.url("/v2/%s/blobs/uploads/", repository)
	r.Logf("registry.layer.initiate-upload url=%s repository=%s", initiateURL, repository)

	resp, err := r.Client.Post(initiateURL, "application/octet-stream", nil)
	if err != nil {
		return nil, "", err
	}
	token := resp.Header.Get("Request-Token")
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	locationURL, err := url.Parse(location)
	if err != nil {
		return nil, token, err
	}
	return locationURL, token, nil
}
