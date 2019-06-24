// +build noembed

package binutils

import (
	"errors"
	"fmt"
	"os/exec"
)

// updatePathForBinaries does nothing when not embedding. the system path is used.
func (r *BinaryAvailabilityCheck) updatePathForBinaries() error {
	return nil
}

// runcBinaryExists checks if the runc binary exists.
func (r *BinaryAvailabilityCheck) runcBinaryExists() error {
	_, err := exec.LookPath("runc")
	if err != nil {
		return fmt.Errorf("please install `runc`")
	}

	return nil
}

// installRuncBinary when non-embedded errors out telling the user to install runc.
func (r *BinaryAvailabilityCheck) installRuncBinary() error {
	return errors.New("please install `runc`")
}
