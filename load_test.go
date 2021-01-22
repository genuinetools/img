package main

import (
	"testing"
)

func TestLoadImage(t *testing.T) {
	output := run(t, "load", "-i", "testdata/img-load/oci.tar")
	if output != "Loaded image: docker.io/library/dummy:latest\n" {
		t.Fatalf("Unexpected output: %s", output)
	}

	output = run(t, "load", "-i", "testdata/img-load/oci.tar")
	if output != "Loaded image: docker.io/library/dummy:latest\n" {
		t.Fatalf("Unexpected output: %s", output)
	}

	output = run(t, "load", "-i", "testdata/img-load/docker.tar")
	expected := `The image docker.io/library/dummy:latest already exists, leaving the old one with ID sha256:e08488191147b6fc575452dfac3238721aa5b86d2545a9edaa1c1d88632b2233 orphaned
Loaded image: docker.io/library/dummy:latest
`
	if output != expected {
		t.Fatalf("Unexpected output: %s", output)
	}

	output = run(t, "load", "-i", "testdata/img-load/docker.tar")
	if output != "Loaded image: docker.io/library/dummy:latest\n" {
		t.Fatalf("Unexpected output: %s", output)
	}
}
