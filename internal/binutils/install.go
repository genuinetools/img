// +build !noembed

package binutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// InstallRuncBinary installs the embedded runc binary to the host.
// It installs the binary to a temporary file and then updates the PATH so it
// can be sourced throughout the execution of the program.
func InstallRuncBinary() (string, error) {
	data, err := Asset("runc")
	if err != nil {
		return "", fmt.Errorf("retrieving runc binary asset data failed: %v", err)
	}

	// Create a new temporary directory to house the binary.
	dir, err := ioutil.TempDir("", "img-runc")
	if err != nil {
		return dir, fmt.Errorf("creating a temporary directory failed: %v", err)
	}

	f := filepath.Join(dir, "runc")
	if err := ioutil.WriteFile(f, data, 0755); err != nil {
		return dir, fmt.Errorf("writing to temporary file for runc binary failed: %v", err)
	}

	// Set the environment variable for PATH to the current plus the path to the
	// embedded binary.
	path := os.Getenv("PATH")
	return dir, os.Setenv("PATH", dir+":"+path)
}
