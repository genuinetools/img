package main

import (
	"strings"
	"testing"
)

func TestTags(t *testing.T) {
	out, err := run("tags", "busybox")
	if err != nil {
		t.Fatalf("output: %s, error: %v", string(out), err)
	}
	expected := `glibc
musl
`
	if !strings.HasSuffix(out, expected) {
		t.Logf("expected: %s\ngot: %s", expected, out)
	}
}
