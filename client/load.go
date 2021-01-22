package client

import (
	"context"
	"encoding/json"
	"io"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/images/archive"
	"github.com/docker/distribution/reference"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// LoadProgress receives informational updates about the progress of image loading.
type LoadProgress interface {
	// LoadImageCreated indicates that the given image has been created.
	LoadImageCreated(images.Image) error

	// LoadImageUpdated indicates that the given image has been updated.
	LoadImageUpdated(images.Image) error

	// LoadImageReplaced indicates that an image has been replaced, leaving the old image without a name.
	LoadImageReplaced(before, after images.Image) error
}

// LoadImage imports an image into the image store.
func (c *Client) LoadImage(ctx context.Context, reader io.Reader, mon LoadProgress) error {
	// Create the worker opts.
	opt, err := c.createWorkerOpt(false)
	if err != nil {
		return errors.Wrap(err, "creating worker opt failed")
	}

	if opt.ImageStore == nil {
		return errors.New("image store is nil")
	}

	logrus.Debug("Importing index")
	index, err := importIndex(ctx, opt.ContentStore, reader)
	if err != nil {
		return err
	}

	logrus.Debug("Extracting manifests")
	manifests := extractManifests(index)

	for _, imageSkel := range manifests {
		err := load(ctx, opt.ImageStore, imageSkel, mon)
		if err != nil {
			return nil
		}
	}

	return nil
}

func importIndex(ctx context.Context, store content.Store, reader io.Reader) (*ocispec.Index, error) {
	d, err := archive.ImportIndex(ctx, store, reader)
	if err != nil {
		return nil, err
	}

	indexBytes, err := content.ReadBlob(ctx, store, d)
	if err != nil {
		return nil, err
	}

	var index ocispec.Index
	err = json.Unmarshal(indexBytes, &index)
	if err != nil {
		return nil, err
	}

	return &index, nil
}

func extractManifests(index *ocispec.Index) []images.Image {
	var result []images.Image
	for _, m := range index.Manifests {
		switch m.MediaType {
		case images.MediaTypeDockerSchema2Manifest:
			if name, ok := m.Annotations[images.AnnotationImageName]; ok {
				if ref, ok := m.Annotations[ocispec.AnnotationRefName]; ok {
					if normalized, err := reference.ParseNormalizedNamed(name); err == nil {
						if normalizedWithTag, err := reference.WithTag(normalized, ref); err == nil {
							if normalized == normalizedWithTag {
								result = append(result, images.Image{
									Name:   normalized.String(),
									Target: m,
								})
								continue
							}
						}
					}
				}
			}
			logrus.Debugf("Failed to extract image info from manifest: %v", m)
		}
	}

	return result
}

func load(ctx context.Context, store images.Store, imageSkel images.Image, mon LoadProgress) error {
	image, err := store.Get(ctx, imageSkel.Name)

	if errors.Cause(err) == errdefs.ErrNotFound {
		image, err = store.Create(ctx, imageSkel)
		if err != nil {
			return err
		}
		return mon.LoadImageCreated(image)
	}

	if err != nil {
		return err
	}

	updated, err := store.Update(ctx, imageSkel)
	if err != nil {
		return err
	}

	if image.Target.Digest == updated.Target.Digest {
		return mon.LoadImageUpdated(updated)
	}

	image.Name = ""
	return mon.LoadImageReplaced(image, updated)
}
