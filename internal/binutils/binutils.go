package binutils

import (
	"context"
	"os/exec"

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
