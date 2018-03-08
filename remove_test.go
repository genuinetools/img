package main

import (
	"strings"
	"testing"
)

func TestRemoveBuiltImage(t *testing.T) {
	name := "testremoveimage"

	runBuild(t, name, withDockerfile(`
    FROM busybox
    CMD echo test
    `))

	// make sure our new image is there
	out := run(t, "ls")
	if !strings.Contains(out, name) {
		t.Fatalf("expected %s in ls output, got: %s", name, out)
	}

	// remove the image
	run(t, "rm", name)

	// make sure the image is not in ls output
	out = run(t, "ls")
	if strings.Contains(out, name) {
		t.Fatalf("expected %s to not be in ls output after removal, got: %s", name, out)
	}
}

func TestRemovePulledImage(t *testing.T) {
	image := "debian:buster"
	run(t, "pull", image)

	// make sure our image is there
	out := run(t, "ls")
	if !strings.Contains(out, image) {
		t.Fatalf("expected %s in ls output, got: %s", image, out)
	}

	// remove the image
	run(t, "rm", image)

	// make sure the image is not in ls output
	out = run(t, "ls")
	if strings.Contains(out, image) {
		t.Fatalf("expected %s to not be in ls output after removal, got: %s", image, out)
	}
}
