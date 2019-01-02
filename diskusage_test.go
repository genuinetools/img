package main

import (
	"strings"
	"testing"
)

func TestDiskUsage(t *testing.T) {
	// Test an official image,
	run(t, "pull", "alpine")

	out := run(t, "du")
	if !strings.Contains(out, "pulled from docker.io") {
		t.Fatalf(`expected "pulled from docker.io" in du output, got: %s`, out)
	}
}
