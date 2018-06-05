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

func TestTagImageAndPush(t *testing.T) {
	runBuild(t, "tagthingandpush", withDockerfile(`
    FROM busybox
    RUN echo tagtestandpush
    `))

	run(t, "tag", "tagthingandpush", "jess/tagtestandpush")

	out := run(t, "ls")

	if !strings.Contains(out, "tagthingandpush:latest") || !strings.Contains(out, "jess/tagtestandpush:latest") {
		t.Fatalf("expected ls output to have tagthingandpush:latest and jess/tagtestandpush:latest but got: %s", out)
	}

	out, err := doRun([]string{"push", "jess/tagtestandpush"}, nil)
	if !strings.Contains(err.Error(), "insufficient_scope: authorization failed") {
		t.Fatalf("expected push to fail with 'insufficient_scope: authorization failed' got: %s %v", out, err)
	}
}

func TestTagPullAndPush(t *testing.T) {
	// Test an official image,
	run(t, "pull", "busybox")

	run(t, "tag", "busybox", "jess/tagpullandpush")

	out := run(t, "ls")

	if !strings.Contains(out, "busybox:latest") || !strings.Contains(out, "jess/tagpullandpush:latest") {
		t.Fatalf("expected ls output to have busybox:latest and jess/tagpullandpush:latest but got: %s", out)
	}

	out, err := doRun([]string{"push", "jess/tagpullandpush"}, nil)
	if !strings.Contains(err.Error(), "insufficient_scope: authorization failed") {
		t.Fatalf("expected push to fail with 'insufficient_scope: authorization failed' got: %s %v", out, err)
	}
}
