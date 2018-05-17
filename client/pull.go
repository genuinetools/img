package client

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/source"
	"github.com/moby/buildkit/util/pull"
)

// Pull retrieves an image from a remote registry.
func (c *Client) Pull(ctx context.Context, image string) (*ListedImage, error) {
	// Parse the image name and tag.
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return nil, fmt.Errorf("parsing image name %q failed: %v", image, err)
	}
	// Add the latest lag if they did not provide one.
	named = reference.TagNameOnly(named)
	image = named.String()

	// Get the identifier for the image.
	identifier, err := source.NewImageIdentifier(image)
	if err != nil {
		return nil, err
	}

	// Create the worker opts.
	opt, err := c.createWorkerOpt()
	if err != nil {
		return nil, fmt.Errorf("creating worker opt failed: %v", err)
	}

	puller := &pull.Puller{
		Snapshotter:  opt.Snapshotter,
		ContentStore: opt.ContentStore,
		Applier:      opt.Applier,
		Src:          identifier.Reference,
		Resolver:     pull.NewResolver(ctx, opt.SessionManager, opt.ImageStore),
	}
	pulled, err := puller.Pull(ctx)
	if err != nil {
		return nil, err
	}
	// Update the target image. Create it if it does not exist.
	img := images.Image{
		Name:      image,
		Target:    pulled.Descriptor,
		CreatedAt: time.Now(),
	}
	if _, err := opt.ImageStore.Update(ctx, img); err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, fmt.Errorf("updating image store for %s failed: %v", image, err)
		}

		// Create it if we didn't find it.
		if _, err := opt.ImageStore.Create(ctx, img); err != nil {
			return nil, fmt.Errorf("creating image in image store for %s failed: %v", image, err)
		}
	}
	size, err := img.Size(ctx, opt.ContentStore, platforms.Default())
	if err != nil {
		return nil, fmt.Errorf("calculating size of image %s failed: %v", image, err)
	}
	return &ListedImage{Image: img, ContentSize: size}, nil
}
