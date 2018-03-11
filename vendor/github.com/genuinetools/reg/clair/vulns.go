package clair

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/genuinetools/reg/registry"
)

// Vulnerabilities scans the given repo and tag
func (c *Clair) Vulnerabilities(r *registry.Registry, repo, tag string) (VulnerabilityReport, error) {
	report := VulnerabilityReport{
		RegistryURL:     r.Domain,
		Repo:            repo,
		Tag:             tag,
		Date:            time.Now().Local().Format(time.RFC1123),
		VulnsBySeverity: make(map[string][]Vulnerability),
	}

	// Get the v1 manifest to pass to clair.
	m, err := r.ManifestV1(repo, tag)
	if err != nil {
		return report, fmt.Errorf("getting the v1 manifest for %s:%s failed: %v", repo, tag, err)
	}

	// Filter out the empty layers.
	var filteredLayers []schema1.FSLayer
	for _, layer := range m.FSLayers {
		if layer.BlobSum != EmptyLayerBlobSum {
			filteredLayers = append(filteredLayers, layer)
		}
	}

	m.FSLayers = filteredLayers
	if len(m.FSLayers) == 0 {
		fmt.Printf("No need to analyse image %s:%s as there is no non-emtpy layer", repo, tag)
		return report, nil
	}

	for i := len(m.FSLayers) - 1; i >= 0; i-- {
		// Form the clair layer.
		l, err := c.NewClairLayer(r, repo, m.FSLayers, i)
		if err != nil {
			return report, err
		}

		// Post the layer.
		if _, err := c.PostLayer(l); err != nil {
			return report, err
		}
	}

	vl, err := c.GetLayer(m.FSLayers[0].BlobSum.String(), false, true)
	if err != nil {
		return report, err
	}

	// Get the vulns.
	for _, f := range vl.Features {
		report.Vulns = append(report.Vulns, f.Vulnerabilities...)
	}

	vulnsBy := func(sev string, store map[string][]Vulnerability) []Vulnerability {
		items, found := store[sev]
		if !found {
			items = make([]Vulnerability, 0)
			store[sev] = items
		}
		return items
	}

	// group by severity
	for _, v := range report.Vulns {
		sevRow := vulnsBy(v.Severity, report.VulnsBySeverity)
		report.VulnsBySeverity[v.Severity] = append(sevRow, v)
	}

	// calculate number of bad vulns
	report.BadVulns = len(report.VulnsBySeverity["High"]) + len(report.VulnsBySeverity["Critical"]) + len(report.VulnsBySeverity["Defcon1"])

	return report, nil
}

// NewClairLayer will form a layer struct required for a clar scan
func (c *Clair) NewClairLayer(r *registry.Registry, image string, fsLayers []schema1.FSLayer, index int) (*Layer, error) {
	var parentName string
	if index < len(fsLayers)-1 {
		parentName = fsLayers[index+1].BlobSum.String()
	}

	// form the path
	p := strings.Join([]string{r.URL, "v2", image, "blobs", fsLayers[index].BlobSum.String()}, "/")

	useBasicAuth := false

	// get the token
	token, err := r.Token(p)
	if err != nil {
		// if we get an error here of type: malformed auth challenge header: 'Basic realm="Registry Realm"'
		// we need to use basic auth for the registry
		if !strings.Contains(err.Error(), `malformed auth challenge header: 'Basic realm="Registry`) {
			return nil, err
		}
		useBasicAuth = true
	}

	h := make(map[string]string)
	if token != "" && !useBasicAuth {
		h = map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}
	}

	if token == "" || useBasicAuth {
		c.Logf("clair.vulns using basic auth")
		h = map[string]string{
			"Authorization": fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(r.Username+":"+r.Password))),
		}
	}

	return &Layer{
		Name:       fsLayers[index].BlobSum.String(),
		Path:       p,
		ParentName: parentName,
		Format:     "Docker",
		Headers:    h,
	}, nil
}
