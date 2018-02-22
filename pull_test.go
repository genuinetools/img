package main

import "testing"

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
