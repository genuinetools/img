package registry

import (
	"net/url"

	"github.com/peterhellberg/link"
)

type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

// Catalog returns the repositories in a registry.
func (r *Registry) Catalog(u string) ([]string, error) {
	if u == "" {
		u = "/v2/_catalog"
	}
	uri := r.url(u)
	r.Logf("registry.catalog url=%s", uri)

	var response catalogResponse
	h, err := r.getJSON(uri, &response)
	if err != nil {
		return nil, err
	}

	for _, l := range link.ParseHeader(h) {
		if l.Rel == "next" {
			unescaped, _ := url.QueryUnescape(l.URI)
			repos, err := r.Catalog(unescaped)
			if err != nil {
				return nil, err
			}
			response.Repositories = append(response.Repositories, repos...)
		}
	}

	return response.Repositories, nil
}
