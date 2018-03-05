package client

import (
	"context"
	"fmt"

	"github.com/docker/distribution/reference"
	"github.com/jessfraz/img/source/containerimage"
	"github.com/moby/buildkit/cache"
	"github.com/moby/buildkit/source"
)

// Pull retrieves an image from a remote registry.
func (c *Client) Pull(ctx context.Context, i string) (cache.ImmutableRef, error) {
	// Parse the repository name.
	image, err := reference.ParseNormalizedNamed(i)
	if err != nil {
		return nil, fmt.Errorf("not a valid image %q: %v", i, err)
	}

	// Add latest to the image name if it is empty.
	image = reference.TagNameOnly(image)

	// Get the identifier for the image.
	identifier, err := source.NewImageIdentifier(image.String())
	if err != nil {
		return nil, err
	}

	// Create the worker opts.
	opt, err := c.createWorkerOpt()
	if err != nil {
		return nil, fmt.Errorf("creating worker opt failed: %v", err)
	}

	// Create the cache manager.
	cm, err := cache.NewManager(cache.ManagerOpt{
		Snapshotter:   opt.Snapshotter,
		MetadataStore: opt.MetadataStore,
	})
	if err != nil {
		return nil, fmt.Errorf("creating new cache manager failed: %v", err)
	}

	// Create the source manager.
	sm, err := source.NewManager()
	if err != nil {
		return nil, fmt.Errorf("creating new source manager failed: %v", err)
	}

	// Add container image as a new source.
	is, err := containerimage.NewSource(containerimage.SourceOpt{
		Snapshotter:   opt.Snapshotter,
		ContentStore:  opt.ContentStore,
		Applier:       opt.Applier,
		CacheAccessor: cm,
		Images:        opt.ImageStore,
	})
	if err != nil {
		return nil, fmt.Errorf("creating new container image source failed: %v", err)
	}

	// Register container image as a source.
	sm.Register(is)

	// Resolve (ie. pull) the image.
	si, err := sm.Resolve(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("resolving the image %s failed: %v", image, err)
	}

	// Snapshot the image.
	ref, err := si.Snapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("snapshotting the image failed: %v", err)
	}

	return ref, nil
}
