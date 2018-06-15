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

func TestSaveImageOCI(t *testing.T) {
	runBuild(t, "savethingoci", withDockerfile(`
    FROM busybox
	RUN echo savetest
    `))

	tmpf := filepath.Join(os.TempDir(), "save-oci-test.tar")
	defer os.RemoveAll(tmpf)

	run(t, "save", "--format", "oci", "-o", tmpf, "savethingoci")

	// Make sure the file exists
	if _, err := os.Stat(tmpf); os.IsNotExist(err) {
		t.Fatalf("%s should exist after saving the image but it didn't", tmpf)
	}
}

func TestSaveImageInvalid(t *testing.T) {
	runBuild(t, "savethinginvalid", withDockerfile(`
    FROM busybox
	RUN echo savetest
    `))

	tmpf := filepath.Join(os.TempDir(), "save-invalid.tar")
	defer os.RemoveAll(tmpf)

	out, err := doRun([]string{"save", "--format", "blah", "-o", tmpf, "savethinginvalid"}, nil)
	if err == nil {
		t.Fatalf("expected invalid format to fail but did not: %s", string(out))
	}
}
