package mount

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/mount"
)

// WithTempMount mounts the provided mounts to a temp dir, and pass the temp dir to f.
// The mounts are valid during the call to the f.
// Finally we will unmount and remove the temp dir regardless of the result of f.
func WithTempMount(ctx context.Context, mounts []mount.Mount, f func(root string) error) error {
	/*root, uerr := ioutil.TempDir(tempMountLocation, "containerd-mount")
	if uerr != nil {
		return errors.Wrapf(uerr, "failed to create temp dir")
	}

	defer func() {
		if uerr = os.RemoveAll(root); uerr != nil {
			log.G(ctx).WithError(uerr).WithField("dir", root).Errorf("failed to remove mount temp dir")
		}
	}()*/

	var root string
	if len(mounts) > 1 {
		return fmt.Errorf("mounts holds more than 1 mount, got %d: %#v", len(mounts), mounts)
	}
	if len(mounts) == 1 {
		root = mounts[0].Source
	}

	err := f(root)
	if err != nil {
		return fmt.Errorf("mount callback failed on %s: %v", root, err)
	}

	return nil
}
