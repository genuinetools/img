package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"runtime"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver/filesystem"
	"github.com/jessfraz/reg/registry"
	"github.com/jessfraz/reg/utils"
	digest "github.com/opencontainers/go-digest"
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

	// Get the auth config.
	auth, err := utils.GetAuthConfig("", "", reference.Domain(name))
	if err != nil {
		return err
	}

	// TODO: add flag to flip switch for turning off SSL verification
	// Create a new registry client.
	r, err := registry.New(auth, *verbose)
	if err != nil {
		return fmt.Errorf("creating new registry api client failed: %v", err)
	}

	// Pull the image.
	if err := pull(ctx, local, r, name, tag); err != nil {
		return fmt.Errorf("pulling failed: %v", err)
	}

	return nil
}

func pull(ctx context.Context, dst distribution.Namespace, src *registry.Registry, name reference.Named, tag string) error {
	fmt.Println("pulling", name.String())

	imgPath := reference.Path(name)

	// Get the manifest.
	manifest, err := src.Manifest(imgPath, tag)
	if err != nil {
		return fmt.Errorf("getting manifest for '%s:%s' failed: %v", imgPath, tag, err)
	}

	switch v := manifest.(type) {
	case *schema1.SignedManifest:
		return pullV1()
	case *schema2.DeserializedManifest:
		return pullV2(ctx, dst, src, v, name, imgPath, tag)
	case *manifestlist.DeserializedManifestList:
		return pullManifestList(ctx, dst, src, v, name, imgPath, tag)
	}

	return errors.New("unsupported manifest format")
}

func pullV1() error {
	return errors.New("schema1 manifest not supported")
}

func pullV2(ctx context.Context, dst distribution.Namespace, src *registry.Registry, manifest *schema2.DeserializedManifest, name reference.Named, imgPath, tag string) error {
	dstRepo, err := dst.Repository(ctx, name)
	if err != nil {
		return fmt.Errorf("creating the destination repository failed: %v", err)
	}

	dstBlobStore := dstRepo.Blobs(ctx)
	for _, ref := range manifest.References() {
		blob, err := src.DownloadLayer(imgPath, ref.Digest)
		if err != nil {
			return fmt.Errorf("getting remote blob %q failed failed: %v", ref.Digest.String(), err)
		}

		upload, err := dstBlobStore.Create(ctx)
		if err != nil {
			return fmt.Errorf("creating the local blob writer failed: %v", err)
		}

		if _, err := io.Copy(upload, blob); err != nil {
			return fmt.Errorf("writing to the local blob failed: %v", err)
		}

		if _, err := upload.Commit(ctx, ref); err != nil {
			return fmt.Errorf("commiting %q locally failed: %v", ref.Digest.String(), err)
		}

		upload.Close()
	}

	// Create the manifest service locally.
	dms, err := dstRepo.Manifests(ctx)
	if err != nil {
		return fmt.Errorf("creating manifest service locally failed: %v", err)
	}

	// Put the manifest locally.
	manDst, err := dms.Put(ctx, manifest, distribution.WithTag(tag))
	if err != nil {
		return fmt.Errorf("putting the manifest with tag %q locally failed: %v", tag, err)
	}

	// TODO: find a better way to get the manifest descriptor locally.
	// Get the manifest descriptor.
	mf, err := dms.Get(ctx, manDst)
	if err != nil {
		return fmt.Errorf("getting the manifest with digest %q locally failed: %v", manDst.String(), err)
	}
	mediatype, pl, err := mf.Payload()
	if err != nil {
		return fmt.Errorf("payload failed: %v", err)
	}
	_, desc, err := distribution.UnmarshalManifest(mediatype, pl)
	if err != nil {
		return fmt.Errorf("umarshal failed: %v", err)
	}

	// Put the tag locally.
	if err := dstRepo.Tags(ctx).Tag(ctx, tag, desc); err != nil {
		return fmt.Errorf("establishing a relationship between the tag %q and digest %q locally failed: %v", tag, manDst.String(), err)
	}

	return nil
}

func pullManifestList(ctx context.Context, dst distribution.Namespace, src *registry.Registry, mfstList *manifestlist.DeserializedManifestList, name reference.Named, imgPath, tag string) error {
	if _, err := schema2ManifestDigest(name, mfstList); err != nil {
		return err
	}

	logf("%s resolved to a manifestList object with %d entries; looking for a %s/%s match", name, len(mfstList.Manifests), runtime.GOOS, runtime.GOARCH)

	manifestMatches := filterManifests(mfstList.Manifests, runtime.GOOS)

	if len(manifestMatches) == 0 {
		return fmt.Errorf("no matching manifest for %s/%s in the manifest list entries", runtime.GOOS, runtime.GOARCH)
	}

	if len(manifestMatches) > 1 {
		logf("found multiple matches in manifest list, choosing best match %s", manifestMatches[0].Digest.String())

	}
	manifestDigest := manifestMatches[0].Digest

	// Get the manifest.
	manifest, err := src.Manifest(imgPath, manifestDigest.String())
	if err != nil {
		return fmt.Errorf("getting manifest for %s@%s failed: %v", imgPath, manifestDigest.String(), err)
	}

	switch v := manifest.(type) {
	case *schema1.SignedManifest:
		return pullV1()
	case *schema2.DeserializedManifest:
		return pullV2(ctx, dst, src, v, name, imgPath, tag)
	}

	return errors.New("unsupported manifest format")
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
			return "", fmt.Errorf("manifest verification failed for digest %s", digested.Digest())
		}
		return digested.Digest(), nil
	}

	return digest.FromBytes(canonical), nil
}

func filterManifests(manifests []manifestlist.ManifestDescriptor, os string) []manifestlist.ManifestDescriptor {
	var matches []manifestlist.ManifestDescriptor
	for _, manifestDescriptor := range manifests {
		if manifestDescriptor.Platform.Architecture == runtime.GOARCH && manifestDescriptor.Platform.OS == os {
			matches = append(matches, manifestDescriptor)

			logf("found match for %s/%s with media type %s, digest %s", os, runtime.GOARCH, manifestDescriptor.MediaType, manifestDescriptor.Digest.String())
		}
	}
	return matches
}
