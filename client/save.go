package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/containerd/containerd/images/archive"
	"github.com/docker/distribution/reference"
)

// SaveImages exports a list of images as a tarball which can then be imported by docker.
func (c *Client) SaveImages(ctx context.Context, images []string, format string, writer io.WriteCloser) error {

	exportOpts := []archive.ExportOpt{}

	// Create the worker opts.
	opt, err := c.createWorkerOpt(false)
	if err != nil {
		return fmt.Errorf("creating worker opt failed: %v", err)
	}

	if opt.ImageStore == nil {
		return errors.New("image store is nil")
	}

	for _, image := range images {
		// Parse the image name and tag.
		named, err := reference.ParseNormalizedNamed(image)
		if err != nil {
			return fmt.Errorf("parsing image name %q failed: %v", image, err)
		}

		// Add the latest lag if they did not provide one.
		named = reference.TagNameOnly(named)
		image = named.String()

		exportOpts = append(exportOpts, archive.WithImage(opt.ImageStore, image))
	}

	switch format {
	case "docker":

	case "oci":
		exportOpts = append(exportOpts, archive.WithSkipDockerManifest())

	default:
		return fmt.Errorf("%q is not a valid format", format)
	}

	if err := archive.Export(ctx, opt.ContentStore, writer, exportOpts...); err != nil {
		return fmt.Errorf("exporting images %v failed: %v", images, err)
	}

	return writer.Close()
}
