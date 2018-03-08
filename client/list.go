package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
)

// ListedImage represents an image structure returuned from ListImages.
// It extends containerd/images.Image with extra fields.
type ListedImage struct {
	images.Image
	ContentSize int64
}

// ListImages returns the images from the image store.
func (c *Client) ListImages(ctx context.Context, filters ...string) ([]ListedImage, error) {
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

	listedImages := []ListedImage{}
	for _, image := range i {
		size, err := image.Size(ctx, opt.ContentStore, platforms.Default())
		if err != nil {
			return nil, fmt.Errorf("calculating size of image %s failed: %v", image.Name, err)
		}
		listedImages = append(listedImages, ListedImage{Image: image, ContentSize: size})
	}
	return listedImages, nil
}
