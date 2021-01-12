package main

import (
	"encoding/json"
	"reflect"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestInspectImage(t *testing.T) {
	runBuild(t, "inspectthing", withDockerfile(`
    FROM busybox
	ENTRYPOINT ["echo"]
    `))

	out := run(t, "inspect", "inspectthing")

	var image ocispec.Image
	if err := json.Unmarshal([]byte(out), &image); err != nil {
		t.Fatalf("error decoding JSON: %s", err)
	}

	if !reflect.DeepEqual(image.Config.Entrypoint, []string{"echo"}) {
		t.Fatalf("expected entrypoint to be set: %#v", image)
	}
}
