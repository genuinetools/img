package binutils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	librunc "github.com/containerd/go-runc"
)

// RuncBinaryExists checks if the runc binary exists.
// TODO(jessfraz): check if it's the right version as well.
func RuncBinaryExists() bool {
	_, err := exec.LookPath("runc")
	if err != nil {
		return false
	}

	// Try to get the version.
	r := &librunc.Runc{}
	_, err = r.Version(context.Background())

	// Return true when there is no error.
	return err == nil
}

// InstallRuncBinary installs the embedded runc binary to the host.
// It installs the binary to a temporary file and then updates the PATH so it
// can be sourced throughout the execution of the program.
func InstallRuncBinary() error {
	data, err := Asset("runc")
	if err != nil {
		return fmt.Errorf("retrieving runc binary asset data failed: %v", err)
	}

	// Create a new temporary directory to house the binary.
	dir, err := ioutil.TempDir("", "img-runc")
	if err != nil {
		return fmt.Errorf("creating a temporary directory failed: %v", err)
	}

	f := filepath.Join(dir, "runc")
	if err := ioutil.WriteFile(f, data, 0755); err != nil {
		return fmt.Errorf("writing to temporary file for runc binary failed: %v", err)
	}

	// Set the environment variable for PATH to the current plus the path to the
	// embedded binary.
	path := os.Getenv("PATH")
	return os.Setenv("PATH", dir+":"+path)
}
