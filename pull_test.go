package main

import (
	"encoding/json"
	"strings"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestPullFromDefaultRegistry(t *testing.T) {
	// Test a user repo on docker hub.
	run(t, "pull", "jess/stress")
}

func TestPullFromSelfHostedRegistry(t *testing.T) {
	// Test a repo on a private registry.
	run(t, "pull", "r.j3ss.co/stress")
}

func TestPullOfficialImage(t *testing.T) {
	// Test an official image,
	run(t, "pull", "alpine")
}

func TestPullIsInListOutput(t *testing.T) {
	// Test an official image,
	run(t, "pull", "busybox")

	out := run(t, "ls")
	if !strings.Contains(out, "busybox:latest") {
		t.Fatalf("expected busybox:latest in ls output, got: %s", out)
	}
}

func TestPullRetainsConfig(t *testing.T) {
	// Test an official image,
	run(t, "pull", "alpine")

	out := run(t, "inspect", "alpine")

	var image ocispec.Image
	if err := json.Unmarshal([]byte(out), &image); err != nil {
		t.Fatalf("error decoding JSON: %s", err)
	}

	if len(image.Config.Cmd) == 0 {
		t.Fatal("image config should be populated")
	}
}
