package main

import (
	"strings"
	"testing"
)

func TestDelete(t *testing.T) {
	// Make sure we have busybox in list.
	out, err := run("ls")
	if err != nil {
		t.Fatalf("output: %s, error: %v", string(out), err)
	}
	expected := []string{"alpine              latest", "busybox             glibc, musl, latest"}
	for _, e := range expected {
		if !strings.Contains(out, e) {
			t.Logf("expected to contain: %s\ngot: %s", e, out)
		}
	}

	// Remove busybox image.
	if out, err := run("rm", "busybox"); err != nil {
		t.Fatalf("output: %s, error: %v", string(out), err)
	}

	// Make sure there is no busybox in list.
	out, err = run("ls")
	if err != nil {
		t.Fatalf("output: %s, error: %v", string(out), err)
	}
	expected = []string{"alpine              latest", "busybox             glibc, musl\n"}
	for _, e := range expected {
		if !strings.Contains(out, e) {
			t.Logf("expected to contain: %s\ngot: %s", e, out)
		}
	}
}
