package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/docker/distribution/reference"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// InspectImage returns the metadata about a given image.
func (c *Client) InspectImage(ctx context.Context, name string) (*ocispec.Image, error) {
	// Parse the image name and tag for the src image.
	named, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return nil, fmt.Errorf("parsing image name %q failed: %v", name, err)
	}

	// Add the latest lag if they did not provide one.
	named = reference.TagNameOnly(named)
	name = named.String()

	// Create the worker opts.
	opt, err := c.createWorkerOpt(false)
	if err != nil {
		return nil, fmt.Errorf("creating worker opt failed: %v", err)
	}

	if opt.ImageStore == nil {
		return nil, errors.New("image store is nil")
	}

	// Get the source image.
	image, err := opt.ImageStore.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("getting image %s from image store failed: %v", name, err)
	}

	var result ocispec.Image
	if err := images.Walk(ctx, images.Handlers(
		images.ChildrenHandler(opt.ContentStore),
		inspectHandler(opt.ContentStore, &result),
	), image.Target); err != nil {
		return nil, fmt.Errorf("error reading image %s: %v", name, err)
	}

	return &result, nil
}

func inspectHandler(provider content.Provider, result *ocispec.Image) images.Handler {
	return images.HandlerFunc(func(
		ctx context.Context,
		desc ocispec.Descriptor,
	) ([]ocispec.Descriptor, error) {
		// We only want the image config
		if desc.MediaType != images.MediaTypeDockerSchema2Config &&
			desc.MediaType != ocispec.MediaTypeImageConfig {
			return nil, nil
		}

		p, err := content.ReadBlob(ctx, provider, desc)
		if err != nil {
			return nil, err
		}

		return nil, json.Unmarshal(p, result)
	})
}
