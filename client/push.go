package client

import (
	"context"
	"fmt"

	"github.com/docker/distribution/reference"
	"github.com/genuinetools/img/exporter/containerimage"
	"github.com/genuinetools/img/exporter/imagepush"
)

// Push sends an image to a remote registry.
func (c *Client) Push(ctx context.Context, image string) error {
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

	// Create the image writer.
	iw, err := containerimage.NewImageWriter(containerimage.WriterOpt{
		Snapshotter:  opt.Snapshotter,
		ContentStore: opt.ContentStore,
		Differ:       opt.Differ,
	})
	if err != nil {
		return fmt.Errorf("creating new container image writer failed: %v", err)
	}

	// Create the image pusher.
	imagePusher, err := imagepush.New(imagepush.Opt{
		Images:      opt.ImageStore,
		ImageWriter: iw,
	})
	if err != nil {
		return fmt.Errorf("creating new image pusher failed: %v", err)
	}

	// Resolve (ie. push) the image.
	ip, err := imagePusher.Resolve(ctx, map[string]string{
		"name": image,
	})
	if err != nil {
		return fmt.Errorf("resolving image %s failed: %v", image, err)
	}

	// Snapshot the image.
	if _, err := ip.Export(ctx, nil, nil); err != nil {
		return fmt.Errorf("exporting the image %s failed: %v", image, err)
	}

	return nil
}
