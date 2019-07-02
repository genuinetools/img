// +build !noembed

package binutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

func (r *BinaryAvailabilityCheck) updatePathForBinaries() error {
	if r.DisableEmbeddedRunc {
		// if the embedded binary is disabled, only use the system path
		return nil
	}

	// Set the environment variable for PATH to the current plus the path to the
	// embedded binary.
	path := os.Getenv("PATH")
	dir := r.embeddedBinaryDirectory()

	err := os.Setenv("PATH", dir+string(os.PathListSeparator)+path)
	if err != nil {
		return fmt.Errorf("failed to update PATH for embedded bin dir: %s", err)
	}

	return nil
}

// runcBinaryExists checks if the embedded runc binary exists.
func (r *BinaryAvailabilityCheck) runcBinaryExists() error {
	runcPath, err := exec.LookPath("runc")
	if err != nil {
		return fmt.Errorf("unable to locate embedded `runc`")
	}

	// if the user has requested to disable embedded runc, then don't check that it
	// matches the embedded one.
	if r.DisableEmbeddedRunc {
		return nil
	}

	if runcPath != r.embeddedRuncPath() {
		return fmt.Errorf("runc was found on PATH but was not the embedded runc expected")
	}

	return nil
}

// installRuncBinary installs the embedded runc binary to the host.
// It installs the binary to a temporary file and then updates the PATH so it
// can be sourced throughout the execution of the program.
func (r *BinaryAvailabilityCheck) installRuncBinary() error {

	data, err := Asset("runc")
	if err != nil {
		return fmt.Errorf("embedded runc failure: retrieving runc binary asset data failed: %v", err)
	}

	err = r.ensureBinaryDirectoryExists()
	if err != nil {
		return err
	}

	f := r.embeddedRuncPath()
	if err := ioutil.WriteFile(f, data, 0755); err != nil {
		return fmt.Errorf("embedded runc failure: writing to path for runc binary failed: %v", err)
	}

	return nil
}

// the embeddedBinaryDirectory is the directory we will be placing our runc binary
func (r *BinaryAvailabilityCheck) embeddedBinaryDirectory() string {
	return path.Join(r.StateDir, "bin")
}

// embeddedRuncPath is the path to the version of runc currently embedded in this version
func (r *BinaryAvailabilityCheck) embeddedRuncPath() string {
	return path.Join(r.embeddedBinaryDirectory(), "runc")
}

func (r *BinaryAvailabilityCheck) ensureBinaryDirectoryExists() error {
	dir := r.embeddedBinaryDirectory()

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error creating dir for embedded binaries: %s", err)
	}

	return nil
}
