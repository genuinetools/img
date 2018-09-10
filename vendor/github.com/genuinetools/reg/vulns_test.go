package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestVulns(t *testing.T) {
	out, err := run("vulns", "--clair", "http://localhost:6060", fmt.Sprintf("%s/alpine:3.5", domain))
	if err != nil {
		t.Fatalf("output: %s, error: %v", out, err)
	}

	expected := `clair.clair resp.Status=200 OK`
	if !strings.HasSuffix(strings.TrimSpace(out), expected) {
		t.Fatalf("expected: %s\ngot: %s", expected, out)
	}
}
