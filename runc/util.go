package runc

import (
	"os/exec"
)

// BinaryExists checks if the runc binary exists.
func BinaryExists() bool {
	_, err := exec.LookPath("runc")
	// Return true when there is no error.
	return err == nil
}
