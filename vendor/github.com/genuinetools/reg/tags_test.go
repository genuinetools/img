package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestTags(t *testing.T) {
	out, err := run("tags", fmt.Sprintf("%s/busybox", domain))
	if err != nil {
		t.Fatalf("output: %s, error: %v", out, err)
	}
	expected := `glibc
latest
musl
`
	if !strings.HasSuffix(out, expected) {
		t.Fatalf("expected: %s\ngot: %s", expected, out)
	}
}
