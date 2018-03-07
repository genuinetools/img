package client

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/images"
	"github.com/docker/distribution/reference"
)

// RemoveImage removes image from the image store.
func (c *Client) RemoveImage(ctx context.Context, image string) error {
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return fmt.Errorf("parsing image name failed: %v", err)
	}
	image = named.String()

	// Create the worker opts.
	opt, err := c.createWorkerOpt()
	if err != nil {
		return fmt.Errorf("creating worker opt failed: %v", err)
	}

	// Remove the image from the image store.
	err = opt.ImageStore.Delete(ctx, image, images.SynchronousDelete())
	if err != nil {
		return fmt.Errorf("removing image failed: %v", err)
	}

	return nil
}
