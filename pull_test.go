package main

import (
	"strings"
	"testing"
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
