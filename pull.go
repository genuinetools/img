package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver/filesystem"
	"github.com/jessfraz/distribution/reference"
	"github.com/jessfraz/reg/registry"
	"github.com/jessfraz/reg/utils"
	digest "github.com/opencontainers/go-digest"
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

	_, tag, err := utils.GetRepoAndRef(args[0])
	if err != nil {
		return err
	}

	auth, err := utils.GetAuthConfig("", "", reference.Domain(name))
	if err != nil {
		return err
	}

	// TODO: add flag to flip switch for turning off SSL verification
	r, err := registry.New(auth, *verbose)
	if err != nil {
		return fmt.Errorf("creating new registry api client failed: %v", err)
	}

	fmt.Println("pulling", name.String(), "from", auth.ServerAddress)
	if err := pull(ctx, local, r, name, tag); err != nil {
		return fmt.Errorf("pulling failed: %v", err)
	}

	return nil
}

func pull(ctx context.Context, dst distribution.Namespace, src *registry.Registry, name reference.Named, tag string) error {
	imgPath := reference.Path(name)

	// Get the manifest.
	manifest, err := src.Manifest(imgPath, tag)
	if err != nil {
		return fmt.Errorf("getting manifest for %s:%s failed: %v", imgPath, tag, err)
	}

	dstRepo, err := dst.Repository(ctx, name)
	if err != nil {
		return fmt.Errorf("creating the destination repository failed: %v", err)
	}

	dstBlobStore := dstRepo.Blobs(ctx)
	for _, ref := range manifest.References() {
		blob, err := src.DownloadLayer(imgPath, ref.Digest)
		if err != nil {
			return fmt.Errorf("getting remote blob failed failed: %v", err)
		}

		upload, err := dstBlobStore.Create(ctx)
		if err != nil {
			return fmt.Errorf("creating the local blob writer failed: %v", err)
		}

		if _, err := io.Copy(upload, blob); err != nil {
			return fmt.Errorf("writing to the local blob failed: %v", err)
		}

		if _, err := upload.Commit(ctx, ref); err != nil {
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

// schema2ManifestDigest computes the manifest digest, and, if pulling by
// digest, ensures that it matches the requested digest.
func schema2ManifestDigest(ref reference.Named, mfst distribution.Manifest) (digest.Digest, error) {
	_, canonical, err := mfst.Payload()
	if err != nil {
		return "", err
	}

	// If pull by digest, then verify the manifest digest.
	if digested, isDigested := ref.(reference.Canonical); isDigested {
		verifier := digested.Digest().Verifier()
		if _, err := verifier.Write(canonical); err != nil {
			return "", err
		}
		if !verifier.Verified() {
			err := fmt.Errorf("manifest verification failed for digest %s", digested.Digest())
			//logrus.Error(err)
			return "", err
		}
		return digested.Digest(), nil
	}

	return digest.FromBytes(canonical), nil
}
