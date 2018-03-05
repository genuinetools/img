package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/containerd/images"
)

// ListImages returns the images from the image store.
func (c *Client) ListImages(ctx context.Context, filters ...string) ([]images.Image, error) {
	// Create the worker opts.
	opt, err := c.createWorkerOpt()
	if err != nil {
		return nil, fmt.Errorf("creating worker opt failed: %v", err)
	}

	// List the images in the image store.
	i, err := opt.ImageStore.List(ctx, filters...)
	if err != nil {
		return nil, fmt.Errorf("listing images with filters (%s) failed: %v", strings.Join(filters, ", "), err)
	}

	return i, nil
}
