package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver/filesystem"
	"github.com/docker/docker/pkg/archive"
	"github.com/jessfraz/reg/utils"
)

// createRootFS creates the base filesystem for a docker image.
// It will pull the base image if it does not exist locally.
// This function takes in a image name and the directory where the
// rootfs should be created.
func createRootFS(image, dir string) error {
	// Create the context.
	ctx := context.Background()

	// Create the new local registry storage.
	local, err := storage.NewRegistry(ctx, filesystem.New(filesystem.DriverParameters{
		RootDirectory: defaultLocalRegistry,
		MaxThreads:    100,
	}))
	if err != nil {
		return fmt.Errorf("creating new registry storage failed: %v", err)
	}

	// Parse the repository name.
	name, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return fmt.Errorf("not a valid image %q: %v", image, err)
	}
	// Add latest to the image name if it is empty.
	name = reference.TagNameOnly(name)

	// Get the tag for the repo.
	_, tag, err := utils.GetRepoAndRef(image)
	if err != nil {
		return err
	}

	// Create the local repository.
	repo, err := local.Repository(ctx, name)
	if err != nil {
		return fmt.Errorf("creating local repository for %q failed: %v", reference.Path(name), err)
	}

	// Create the manifest service.
	ms, err := repo.Manifests(ctx)
	if err != nil {
		return fmt.Errorf("creating manifest service failed: %v", err)
	}

	// Get the specific tag.
	td, err := repo.Tags(ctx).Get(ctx, tag)
	// Check if we got an unknown error, that means the tag does not exist.
	if err != nil && strings.Contains(err.Error(), "unknown") {
		logf("image not found locally, pulling the image")

		// Pull the image.
		if err := pull(ctx, local, name, tag); err != nil {
			return fmt.Errorf("pulling failed: %v", err)
		}

		// Try to get the tag again.
		td, err = repo.Tags(ctx).Get(ctx, tag)
	}
	if err != nil {
		return fmt.Errorf("getting local repository tag %q failed: %v", tag, err)
	}

	// Get the specific manifest for the tag.
	manifest, err := ms.Get(ctx, td.Digest)
	if err != nil {
		return fmt.Errorf("getting local manifest for digest %q failed: %v", td.Digest.String(), err)
	}

	blobStore := repo.Blobs(ctx)
	for i, ref := range manifest.References() {
		if i == 0 {
			fmt.Printf("skipping config %v\n", ref.Digest.String())
			continue
		}
		fmt.Printf("unpacking %v\n", ref.Digest.String())
		layer, err := blobStore.Open(ctx, ref.Digest)
		if err != nil {
			return fmt.Errorf("getting blob %q failed: %v", ref.Digest.String(), err)
		}

		// Unpack the tarfile to the mount path.
		// TODO: figure out what options we really need.
		// FROM: https://godoc.org/github.com/moby/moby/pkg/archive#TarOptions
		if err := archive.Untar(layer, dir, &archive.TarOptions{
			NoLchown: true,
		}); err != nil {
			return fmt.Errorf("error extracting tar for %q: %v", ref.Digest.String(), err)
		}
	}

	fmt.Printf("%s rootfs created at %s\n", name.Name(), dir)
	return nil
}
