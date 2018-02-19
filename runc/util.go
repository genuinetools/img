package runc

import (
	"os/exec"
)

// BinaryExists checks if the runc binary exists.
func BinaryExists() bool {
	_, err := exec.LookPath("runc")
	if err != nil {
		return false
	}
	return true
}
