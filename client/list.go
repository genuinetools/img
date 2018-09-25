package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/images"
	ctdmetadata "github.com/containerd/containerd/metadata"
	"github.com/containerd/containerd/platforms"
	bolt "go.etcd.io/bbolt"
)

// ListedImage represents an image structure returuned from ListImages.
// It extends containerd/images.Image with extra fields.
type ListedImage struct {
	images.Image
	ContentSize int64
}

// ListImages returns the images from the image store.
func (c *Client) ListImages(ctx context.Context, filters ...string) ([]ListedImage, error) {
	dbPath := filepath.Join(c.root, "containerdmeta.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// The metadata database does not exist so we should just return as if there
		// were no results.
		return nil, nil
	}

	// Open the bolt database for metadata.
	// Since we are only listing we can open it as read-only.
	db, err := bolt.Open(dbPath, 0644, &bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("opening boltdb failed: %v", err)
	}

	// Create the content store locally.
	contentStore, err := local.NewStore(filepath.Join(c.root, "content"))
	if err != nil {
		return nil, fmt.Errorf("creating content store failed: %v", err)
	}

	// Create the database for metadata.
	mdb := ctdmetadata.NewDB(db, contentStore, nil)

	// Create the image store.
	imageStore := ctdmetadata.NewImageStore(mdb)

	// List the images in the image store.
	i, err := imageStore.List(ctx, filters...)
	if err != nil {
		return nil, fmt.Errorf("listing images with filters (%s) failed: %v", strings.Join(filters, ", "), err)
	}

	listedImages := []ListedImage{}
	for _, image := range i {
		size, err := image.Size(ctx, contentStore, platforms.Default())
		if err != nil {
			return nil, fmt.Errorf("calculating size of image %s failed: %v", image.Name, err)
		}
		listedImages = append(listedImages, ListedImage{Image: image, ContentSize: size})
	}
	return listedImages, nil
}
