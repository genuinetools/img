package main

import (
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	out, err := run("ls")
	if err != nil {
		t.Fatalf("output: %s, error: %v", string(out), err)
	}
	expected := []string{"alpine              latest", "busybox             glibc, musl"}
	for _, e := range expected {
		if !strings.Contains(out, e) {
			t.Logf("expected to contain: %s\ngot: %s", e, out)
		}
	}
}
