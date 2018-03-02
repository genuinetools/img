package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveImage(t *testing.T) {
	runBuild(t, "savething", withDockerfile(`
    FROM busybox
	RUN echo savetest
    `))

	tmpf := filepath.Join(os.TempDir(), "save-image-test.tar")
	defer os.RemoveAll(tmpf)

	run(t, "save", "-o", tmpf, "savething")

	// Make sure the file exists
	if _, err := os.Stat(tmpf); os.IsNotExist(err) {
		t.Fatalf("%s should exist after saving the image but it didn't", tmpf)
	}
}
