package registry

import (
	"testing"

	"github.com/genuinetools/reg/repoutils"
)

func TestDigestFromDockerHub(t *testing.T) {
	auth, err := repoutils.GetAuthConfig("", "", "docker.io")
	if err != nil {
		t.Fatalf("Could not get auth config: %s", err)
	}

	r, err := New(auth, Opt{})
	if err != nil {
		t.Fatalf("Could not create registry instance: %s", err)
	}

	d, err := r.Digest(Image{Domain: "docker.io", Path: "library/alpine", Tag: "latest"})
	if err != nil {
		t.Fatalf("Could not get digest: %s", err)
	}

	if d == "" {
		t.Error("Empty digest received")
	}
}

func TestDigestFromGCR(t *testing.T) {
	auth, err := repoutils.GetAuthConfig("", "", "gcr.io")
	if err != nil {
		t.Fatalf("Could not get auth config: %s", err)
	}

	r, err := New(auth, Opt{})
	if err != nil {
		t.Fatalf("Could not create registry instance: %s", err)
	}

	d, err := r.Digest(Image{Domain: "gcr.io", Path: "google_containers/hyperkube", Tag: "v1.9.9"})
	if err != nil {
		t.Fatalf("Could not get digest: %s", err)
	}

	if d == "" {
		t.Error("Empty digest received")
	}
}
