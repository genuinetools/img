package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/images/oci"
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/util/dockerexporter"
)

// SaveImage exports an image as a tarball which can then be imported by docker.
func (c *Client) SaveImage(ctx context.Context, image, format string, writer io.WriteCloser) error {
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

	if opt.ImageStore == nil {
		return errors.New("image store is nil")
	}

	img, err := opt.ImageStore.Get(ctx, image)
	if err != nil {
		return fmt.Errorf("getting image %s from image store failed: %v", image, err)
	}

	var exporter images.Exporter
	switch format {
	case "docker":
		exporter = &dockerexporter.DockerExporter{
			Names: []string{img.Name},
		}
	case "oci":
		exporter = &oci.V1Exporter{}
	default:
		return fmt.Errorf("%q is not a valid format", format)
	}

	if err := exporter.Export(ctx, opt.ContentStore, img.Target, writer); err != nil {
		return fmt.Errorf("exporting image %s failed: %v", image, err)
	}

	return writer.Close()
}
