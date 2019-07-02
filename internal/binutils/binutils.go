package binutils

import (
	"context"
	"fmt"
	gorunc "github.com/containerd/go-runc"
	log "github.com/sirupsen/logrus"
)

// BinaryAvailabilityCheck provides info on the runc binary, whether embedded or non-embedded
type BinaryAvailabilityCheck struct {
	StateDir            string
	DisableEmbeddedRunc bool
}

// EnsureRuncIsAvailable makes sure that runc is available, by using the
// embedded runc or locating the system installed runc.
func (r *BinaryAvailabilityCheck) EnsureRuncIsAvailable() error {
	log.WithFields(log.Fields{"state": r.StateDir, "disableEmbeddedRunc": r.DisableEmbeddedRunc}).Debug("checking runc")

	err := r.updatePathForBinaries()
	if err != nil {
		return err
	}

	err = r.runcBinaryExists()
	if err != nil {
		if r.DisableEmbeddedRunc {
			// don't install- we fail because we couldn't locate a system binary
			return fmt.Errorf("no installed runc found, and embedded runc is disabled")
		}

		if err := r.installRuncBinary(); err != nil {
			return err
		}
	}

	_, err = GetRuncVersion()
	return err
}

// GetRuncVersion validates basic runc operation by checking the version
// TODO(jessfraz): check if it's the right version as well.
func GetRuncVersion() (gorunc.Version, error) {
	// Try to get the version.
	runcContext := &gorunc.Runc{}

	v, err := runcContext.Version(context.Background())
	if err != nil {
		return gorunc.Version{}, fmt.Errorf("unable to check runc version")
	}

	log.WithFields(log.Fields{"version": v.Runc, "commit": v.Commit, "spec": v.Spec}).Debug("runc found")

	return v, nil
}
