package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLayer(t *testing.T) {
	// Get the digest.
	out, err := run("digest", fmt.Sprintf("%s/busybox", domain))
	if err != nil {
		t.Fatalf("output: %s, error: %v", out, err)
	}

	tmpf := filepath.Join(os.TempDir(), "download-layer.tar")
	defer os.RemoveAll(tmpf)

	// Download the layer.
	lines := strings.Split(strings.TrimSpace(out), "\n")
	layer := fmt.Sprintf("%s/busybox@%s", domain, strings.TrimSpace(lines[len(lines)-1]))
	out, err = run("layer", "-o", tmpf, layer)
	if err != nil {
		t.Fatalf("output: %s, error: %v", out, err)
	}

	// Make sure the file exists
	if _, err := os.Stat(tmpf); os.IsNotExist(err) {
		t.Fatalf("%s should exist after downloading the layer but it didn't", tmpf)
	}
}
