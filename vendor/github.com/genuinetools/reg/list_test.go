package main

import (
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	out, err := run("ls", domain)
	if err != nil {
		t.Fatalf("output: %s, error: %v", out, err)
	}

	expected := `REPO                TAGS
alpine              3.5, latest
busybox             glibc, latest, musl`
	if !strings.HasSuffix(strings.TrimSpace(out), expected) {
		t.Fatalf("expected to contain: %s\ngot: %s", expected, out)
	}
}
