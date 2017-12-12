package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver/filesystem"
	"github.com/jessfraz/reg/utils"
)

const buildShortHelp = `Build a Dockerfile or OCI campatible spec into an image.`
const buildLongHelp = `
`

func (cmd *buildCommand) Name() string      { return "build" }
func (cmd *buildCommand) Args() string      { return "[img] [mountPath]" }
func (cmd *buildCommand) ShortHelp() string { return buildShortHelp }
func (cmd *buildCommand) LongHelp() string  { return buildLongHelp }
func (cmd *buildCommand) Hidden() bool      { return false }

func (cmd *buildCommand) Register(fs *flag.FlagSet) {}

type buildCommand struct{}

func (cmd *buildCommand) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image to pull")
	}

	if len(args) < 2 {
		return fmt.Errorf("must pass a mount path")
	}

	// Name the mount path.
	mountPath := args[1]
	fi, err := os.Stat(mountPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Create the directory if it does not exist.
	if os.IsNotExist(err) {
		if err := os.MkdirAll(mountPath, 0755); err != nil {
			return err
		}
	}

	// Ensure the mount path is a directory.
	if !fi.IsDir() {
		return fmt.Errorf("mount path %q should be a directory", mountPath)
	}

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
	name, err := reference.ParseNormalizedNamed(args[0])
	if err != nil {
		return fmt.Errorf("not a valid image %q: %v", args[0], err)
	}
	// Add latest to the image name if it is empty.
	name = reference.TagNameOnly(name)

	// Get the tag for the repo.
	_, tag, err := utils.GetRepoAndRef(args[0])
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
	if err != nil {
		// Check if we got an unknown error, that means the tag does not exist.
		if _, ok := err.(*distribution.ErrTagUnknown); ok {
			// TODO: pull the image
			return fmt.Errorf("image not found locally, first pull the image %s", name.String())
		}

		return fmt.Errorf("getting local repository tag %q failed: %v", tag, err)
	}

	// Get the specific manifest for the tag.
	manifest, err := ms.Get(ctx, td.Digest)
	if err != nil {
		return fmt.Errorf("getting local manifest for digest %q failed: %v", td.Digest.String(), err)
	}

	blobStore := repo.Blobs(ctx)
	for _, ref := range manifest.References() {
		fmt.Printf("unpacking %v\n", ref.Digest.String())
		layer, err := blobStore.Open(ctx, ref.Digest)
		if err != nil {
			return fmt.Errorf("getting blob %q failed: %v", ref.Digest.String(), err)
		}

		// Unpack the tarfile to the mount path.
		if err := extractTarFile(mountPath, layer); err != nil {
			return fmt.Errorf("error extracting tar for %q: %v", ref.Digest.String(), err)
		}
	}

	fmt.Printf("%s mounted at %s\n", name.Name(), mountPath)
	return nil
}
