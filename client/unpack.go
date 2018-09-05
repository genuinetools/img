package client

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/pkg/archive"
	"github.com/sirupsen/logrus"
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
	opt, err := c.createWorkerOpt(true)
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

	manifest, err := images.Manifest(ctx, opt.ContentStore, img.Target, platforms.Default())
	if err != nil {
		return fmt.Errorf("getting image manifest failed: %v", err)
	}

	for _, desc := range manifest.Layers {
		logrus.Debugf("Unpacking layer %s", desc.Digest.String())

		// Read the blob from the content store.
		layer, err := opt.ContentStore.ReaderAt(ctx, desc)
		if err != nil {
			return fmt.Errorf("getting reader for digest %s failed: %v", desc.Digest.String(), err)
		}

		// Unpack the tarfile to the rootfs path.
		// FROM: https://godoc.org/github.com/moby/moby/pkg/archive#TarOptions
		if err := archive.Untar(content.NewReader(layer), dest, &archive.TarOptions{
			NoLchown: true,
		}); err != nil {
			return fmt.Errorf("extracting tar for %s to directory %s failed: %v", desc.Digest.String(), dest, err)
		}
	}

	return nil
}
