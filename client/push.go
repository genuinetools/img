package client

import (
	"context"
	"fmt"

	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/util/push"
)

// Push sends an image to a remote registry.
func (c *Client) Push(ctx context.Context, image string, insecure bool) error {
	// Parse the image name and tag.
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return fmt.Errorf("parsing image name %q failed: %v", image, err)
	}
	// Add the latest lag if they did not provide one.
	named = reference.TagNameOnly(named)
	image = named.String()

	// Create the worker opts.
	opt, err := c.createWorkerOpt(false)
	if err != nil {
		return fmt.Errorf("creating worker opt failed: %v", err)
	}

	imgObj, err := opt.ImageStore.Get(ctx, image)
	if err != nil {
		return fmt.Errorf("getting image %q failed: %v", image, err)
	}

	sm, err := c.getSessionManager()
	if err != nil {
		return err
	}
	return push.Push(ctx, sm, opt.ContentStore, imgObj.Target.Digest, image, insecure, opt.RegistryHosts, false)
}
