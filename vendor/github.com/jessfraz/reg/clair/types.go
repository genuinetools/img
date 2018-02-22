package clair

import "github.com/opencontainers/go-digest"

const (
	// EmptyLayerBlobSum is the blob sum of empty layers.
	EmptyLayerBlobSum = "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"

	// LegacyEmptyLayerBlobSum is the blob sum of empty layers used by docker
	// before it could support a truly empty layer.
	LegacyEmptyLayerBlobSum = "sha256:5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef"
)

// IsEmptyLayer determines whether the blob sum is one of the known empty
// layers.
func IsEmptyLayer(blobSum digest.Digest) bool {
	return blobSum == EmptyLayerBlobSum || blobSum == LegacyEmptyLayerBlobSum
}

var (
	// Priorities are the vulnerability priority labels.
	Priorities = []string{"Unknown", "Negligible", "Low", "Medium", "High", "Critical", "Defcon1", "Fixable"}
)

// Error describes the structure of a clair error.
type Error struct {
	Message string `json:"Message,omitempty"`
}

// Layer represents an image layer.
type Layer struct {
	Name             string            `json:"Name,omitempty"`
	NamespaceName    string            `json:"NamespaceName,omitempty"`
	Path             string            `json:"Path,omitempty"`
	Headers          map[string]string `json:"Headers,omitempty"`
	ParentName       string            `json:"ParentName,omitempty"`
	Format           string            `json:"Format,omitempty"`
	IndexedByVersion int               `json:"IndexedByVersion,omitempty"`
	Features         []feature         `json:"Features,omitempty"`
}

type layerEnvelope struct {
	Layer *Layer `json:"Layer,omitempty"`
	Error *Error `json:"Error,omitempty"`
}

// Vulnerability represents vulnerability entity returned by Clair.
type Vulnerability struct {
	Name          string                 `json:"Name,omitempty"`
	NamespaceName string                 `json:"NamespaceName,omitempty"`
	Description   string                 `json:"Description,omitempty"`
	Link          string                 `json:"Link,omitempty"`
	Severity      string                 `json:"Severity,omitempty"`
	Metadata      map[string]interface{} `json:"Metadata,omitempty"`
	FixedBy       string                 `json:"FixedBy,omitempty"`
	FixedIn       []feature              `json:"FixedIn,omitempty"`
}

// VulnerabilityReport represents the result of a vulnerability scan of a repo.
type VulnerabilityReport struct {
	RegistryURL     string
	Repo            string
	Tag             string
	Date            string
	Vulns           []Vulnerability
	VulnsBySeverity map[string][]Vulnerability
	BadVulns        int
}
type feature struct {
	Name            string          `json:"Name,omitempty"`
	NamespaceName   string          `json:"NamespaceName,omitempty"`
	VersionFormat   string          `json:"VersionFormat,omitempty"`
	Version         string          `json:"Version,omitempty"`
	Vulnerabilities []Vulnerability `json:"Vulnerabilities,omitempty"`
	AddedBy         string          `json:"AddedBy,omitempty"`
}
