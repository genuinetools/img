package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestManifestV2(t *testing.T) {
	out, err := run("manifest", fmt.Sprintf("%s/busybox", domain))
	if err != nil {
		t.Fatalf("output: %s, error: %v", out, err)
	}

	expected := `"schemaVersion": 2,`
	if !strings.Contains(out, expected) {
		t.Fatalf("expected: %s\ngot: %s", expected, out)
	}
}

func TestManifestV1(t *testing.T) {
	out, err := run("manifest", "--v1", fmt.Sprintf("%s/busybox", domain))
	if err != nil {
		t.Fatalf("output: %s, error: %v", out, err)
	}

	expected := `"schemaVersion": 1,`
	if !strings.Contains(out, expected) {
		t.Fatalf("expected: %s\ngot: %s", expected, out)
	}
}
