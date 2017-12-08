package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver/filesystem"
	"github.com/jessfraz/distribution/reference"
)

const (
	defaultDockerRegistry = "https://registry-1.docker.io"
	// TODO: change this from tmpfs
	defaultLocalRegistry = "/tmp/img-local-registry"

	simultaneousLayerPullWindow = 4
)

const pullShortHelp = `Pull and verify an image from a registry.`
const pullLongHelp = `
`

func (cmd *pullCommand) Name() string      { return "pull" }
func (cmd *pullCommand) Args() string      { return "name[:tag|@digest]" }
func (cmd *pullCommand) ShortHelp() string { return pullShortHelp }
func (cmd *pullCommand) LongHelp() string  { return pullLongHelp }
func (cmd *pullCommand) Hidden() bool      { return false }

func (cmd *pullCommand) Register(fs *flag.FlagSet) {}

type pullCommand struct{}

func (cmd *pullCommand) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image to pull")
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

	name, err := reference.ParseNormalizedNamed(args[0])
	if err != nil {
		return fmt.Errorf("not a valid image %q: %v", args[0], err)
	}
	name = reference.TagNameOnly(name)

	_, tag, _, err := splitNameTag(args[0])
	if err != nil {
		return err
	}

	auth, err := getAuthConfig(reference.Domain(name))
	if err != nil {
		return err
	}

	domain := defaultDockerRegistry
	if reference.Domain(name) != "docker.io" && reference.Domain(name) != "" {
		domain = auth.ServerAddress
	}

	// TODO: add flag to flip switch for turning off SSL verification
	transport := newRegistryTransport(auth, false)

	// Create the new Registry client.
	remote, err := client.NewRepository(name, domain, transport)
	if err != nil {
		return fmt.Errorf("creating new registry repository failed: %v", err)
	}

	fmt.Println("pulling", name.String(), "from", domain)
	if err := pull(ctx, local, remote, name, tag); err != nil {
		return fmt.Errorf("pulling failed: %v", err)
	}

	return nil
}

func pull(ctx context.Context, dst distribution.Namespace, src distribution.Repository, name reference.Named, tag string) error {
	// Create the manifest service
	ms, err := src.Manifests(ctx)
	if err != nil {
		return fmt.Errorf("getting manifest service failed: %v", err)
	}

	// Create the tags service
	ts := src.Tags(ctx)

	// Get the tag descriptor for the digest
	descriptor, err := ts.Get(ctx, tag)
	if err != nil {
		return fmt.Errorf("getting the tag descriptor failed: %v", err)
	}

	// Get the manifest
	manifest, err := ms.Get(ctx, descriptor.Digest)
	if err != nil {
		return fmt.Errorf("getting the manifest failed: %v", err)
	}

	dstRepo, err := dst.Repository(ctx, name)
	if err != nil {
		return fmt.Errorf("creating the destination repository failed: %v", err)
	}

	srcBlobStore := src.Blobs(ctx)
	dstBlobStore := dstRepo.Blobs(ctx)
	for _, ref := range manifest.References() {
		blob, err := srcBlobStore.Get(ctx, ref.Digest)
		if err != nil {
			return fmt.Errorf("getting remote blob failed failed: %v", err)
		}

		upload, err := dstBlobStore.Create(ctx)
		if err != nil {
			return fmt.Errorf("creating the local blob writer failed: %v", err)
		}

		if _, err := upload.Write(blob); err != nil {
			return fmt.Errorf("writing to the local blob failed: %v", err)
		}

		descriptor = ref
		if _, err := upload.Commit(ctx, descriptor); err != nil {
			return fmt.Errorf("commiting locally failed: %v", err)
		}

		upload.Close()
	}

	dms, err := dstRepo.Manifests(ctx)
	if err != nil {
		return fmt.Errorf("creating manifest service locally failed: %v", err)
	}

	if _, err := dms.Put(ctx, manifest); err != nil {
		return fmt.Errorf("putting the manifest locally failed: %v", err)
	}

	return nil
}

func splitNameTag(raw string) (name, tag, revision string, err error) {
	name = raw
	if strings.Contains(name, "@") {
		parts := strings.Split(name, "@")
		if len(parts) != 2 {
			return "", "", "", fmt.Errorf("not a valid name %q", raw)
		}
		name = parts[0]
		revision = parts[1]
	}

	if strings.Contains(name, ":") {
		parts := strings.Split(name, ":")

		if len(parts) != 2 {
			return "", "", "", fmt.Errorf("not a valid name %q", raw)
		}

		name = parts[0]
		tag = parts[1]
	}

	// set the default tag to latest
	if tag == "" {
		tag = "latest"
	}

	return
}
