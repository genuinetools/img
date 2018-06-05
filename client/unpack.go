package client

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/containerd/containerd/platforms"
	"github.com/docker/distribution/reference"
	"github.com/genuinetools/img/internal/ioutils"
	bkidentity "github.com/moby/buildkit/identity"
	"github.com/opencontainers/image-spec/identity"
)

// Unpack exports an image to a rootfs destination directory.
func (c *Client) Unpack(ctx context.Context, image, dest string) error {
	if len(dest) < 1 {
		return errors.New("destination directory for rootfs cannot be empty")
	}

	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("destination directory already exists: %s", dest)
	}

	// Parse the image name and tag.
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return fmt.Errorf("parsing image name %q failed: %v", image, err)
	}
	// Add the latest lag if they did not provide one.
	named = reference.TagNameOnly(named)
	image = named.String()

	// Create the worker opts.
	opt, err := c.createWorkerOpt()
	if err != nil {
		return fmt.Errorf("creating worker opt failed: %v", err)
	}

	if opt.ImageStore == nil {
		return errors.New("image store is nil")
	}

	img, err := opt.ImageStore.Get(ctx, image)
	if err != nil {
		return fmt.Errorf("getting image %s from image store failed: %v", image, err)
	}

	rootfs, err := img.RootFS(ctx, opt.ContentStore, platforms.Default())
	if err != nil {
		return fmt.Errorf("getting image rootfs digest failed: %v", err)
	}

	chainID := identity.ChainID(rootfs)

	// TODO(jessfraz): do a better containerKey.
	containerKey := bkidentity.NewID()
	// Get the snapshot mounts.
	mounts, err := opt.Snapshotter.View(ctx, containerKey, chainID.String())
	if err != nil {
		return fmt.Errorf("viewing snapshot for %s failed: %v", chainID.String(), err)
	}

	m, err := mounts.Mount()
	if err != nil {
		return err
	}

	// Make sure there is only one mount.
	if len(m) > 1 {
		return fmt.Errorf("expected 1 mount got: %d", len(m))
	}

	// Copy the snapshot to the destination directory.
	if err := ioutils.Copy(m[0].Source, dest); err != nil {
		return fmt.Errorf("copying %s to %s failed: %v", m[0].Source, dest, err)
	}

	return nil
}
