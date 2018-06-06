package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestUnpackFromBuild(t *testing.T) {
	name := "testbuildunpack"

	runBuild(t, name, withDockerfile(`
    FROM busybox
	RUN echo unpack
    `))

	tmpd, err := ioutil.TempDir("", "img-unpack")
	if err != nil {
		t.Fatalf("creating temporary directory for unpack failed: %v", err)
	}
	defer os.RemoveAll(tmpd)

	rootfs := filepath.Join(tmpd, "rootfs")

	run(t, "unpack", "-o", rootfs, name)

	// Make sure the image actually is unpacked in the directory.
	etc := filepath.Join(rootfs, "etc")
	if _, err := os.Stat(etc); os.IsNotExist(err) {
		t.Fatalf("expected etc directory at %q to exist but it did not", etc)
	}
}

func TestUnpackFromPull(t *testing.T) {
	run(t, "pull", "r.j3ss.co/stress")

	tmpd, err := ioutil.TempDir("", "img-unpack")
	if err != nil {
		t.Fatalf("creating temporary directory for unpack failed: %v", err)
	}
	defer os.RemoveAll(tmpd)

	rootfs := filepath.Join(tmpd, "rootfs")

	run(t, "unpack", "-o", rootfs, "r.j3ss.co/stress")

	// Make sure the image actually is unpacked in the directory.
	etc := filepath.Join(rootfs, "etc")
	if _, err := os.Stat(etc); os.IsNotExist(err) {
		t.Fatalf("expected etc directory at %q to exist but it did not", etc)
	}
}
