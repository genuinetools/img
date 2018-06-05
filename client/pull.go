package client

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/platforms"
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/cache"
	imageexporter "github.com/moby/buildkit/exporter/containerimage"
	"github.com/moby/buildkit/source"
	"github.com/moby/buildkit/source/containerimage"
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

	cm, err := cache.NewManager(cache.ManagerOpt{
		Snapshotter:   opt.Snapshotter,
		MetadataStore: opt.MetadataStore,
	})
	if err != nil {
		return nil, err
	}

	// Create the source for the pull.
	srcOpt := containerimage.SourceOpt{
		SessionManager: opt.SessionManager,
		Snapshotter:    opt.Snapshotter,
		ContentStore:   opt.ContentStore,
		Applier:        opt.Applier,
		CacheAccessor:  cm,
		ImageStore:     opt.ImageStore,
	}
	src, err := containerimage.NewSource(srcOpt)
	if err != nil {
		return nil, err
	}
	s, err := src.Resolve(ctx, identifier)
	if err != nil {
		return nil, err
	}
	ref, err := s.Snapshot(ctx)
	if err != nil {
		return nil, err
	}

	// Create the exporter for the pull.
	iw, err := imageexporter.NewImageWriter(imageexporter.WriterOpt{
		Snapshotter:  opt.Snapshotter,
		ContentStore: opt.ContentStore,
		Differ:       opt.Differ,
	})
	if err != nil {
		return nil, err
	}
	expOpt := imageexporter.Opt{
		SessionManager: opt.SessionManager,
		Images:         opt.ImageStore,
		ImageWriter:    iw,
	}
	exp, err := imageexporter.New(expOpt)
	if err != nil {
		return nil, err
	}
	e, err := exp.Resolve(ctx, map[string]string{"name": image})
	if err != nil {
		return nil, err
	}
	if _, err := e.Export(ctx, ref, nil); err != nil {
		return nil, err
	}

	// Get the image.
	img, err := opt.ImageStore.Get(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("getting image %s from image store failed: %v", image, err)
	}
	size, err := img.Size(ctx, opt.ContentStore, platforms.Default())
	if err != nil {
		return nil, fmt.Errorf("calculating size of image %s failed: %v", img.Name, err)
	}

	return &ListedImage{Image: img, ContentSize: size}, nil
}
