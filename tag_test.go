package main

import (
	"strings"
	"testing"
)

func TestTagImage(t *testing.T) {
	runBuild(t, "tagthing", withDockerfile(`
    FROM busybox
    RUN echo tagtest
    `))

	run(t, "tag", "tagthing", "jess/tagtest")

	out := run(t, "ls")

	if !strings.Contains(out, "tagthing:latest") || !strings.Contains(out, "jess/tagtest:latest") {
		t.Fatalf("expected ls output to have tagthing:latest and jess/tagtest:latest but got: %s", out)
	}
}
