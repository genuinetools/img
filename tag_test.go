package main

import (
	"strings"
	"testing"
)

func TestTagImage(t *testing.T) {
	runBuild(t, "thing", withDockerfile(`
    FROM busybox
    ENTRYPOINT ["echo"]
    CMD echo test
    `))

	run(t, "tag", "thing", "jess/tagtest")

	out := run(t, "ls")

	if !strings.Contains(out, "thing:latest") || !strings.Contains(out, "jess/tagtest:latest") {
		t.Fatalf("expected ls output to havething:latest and jess/tagtest:latest but got: %s", out)
	}
}
