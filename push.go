package main

import (
	"flag"
	"fmt"

	"github.com/containerd/containerd/namespaces"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/jessfraz/img/exporter/containerimage"
	"github.com/jessfraz/img/exporter/imagepush"
	"github.com/jessfraz/img/worker/runc"
	"github.com/moby/buildkit/exporter"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
)

const pushHelp = `Push an image or a repository to a registry.`

func (cmd *pushCommand) Name() string      { return "push" }
func (cmd *pushCommand) Args() string      { return "[OPTIONS] NAME[:TAG]" }
func (cmd *pushCommand) ShortHelp() string { return pushHelp }
func (cmd *pushCommand) LongHelp() string  { return pushHelp }
func (cmd *pushCommand) Hidden() bool      { return false }

func (cmd *pushCommand) Register(fs *flag.FlagSet) {}

type pushCommand struct {
	image string
}

func (cmd *pushCommand) Run(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image or repository to push")
	}

	// Get the specified image.
	cmd.image = args[0]
	// Add the latest lag if they did not provide one.
	addLatestTagSuffix(cmd.image)

	// Create the context.
	ctx := appcontext.Context()
	id := identity.NewID()
	ctx = session.NewContext(ctx, id)
	ctx = namespaces.WithNamespace(ctx, namespaces.Default)

	// Create the source manager.
	imgpush, fuseserver, err := createImagePusher()
	defer unmount(fuseserver)
	if err != nil {
		return err
	}
	handleSignals(fuseserver)

	// Resolve (ie. push) the image.
	ip, err := imgpush.Resolve(ctx, map[string]string{
		"name": cmd.image,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Pushing %s...\n", cmd.image)

	// Snapshot the image.
	if err := ip.Export(ctx, nil, nil); err != nil {
		return err
	}

	fmt.Printf("Successfully pushed %s", cmd.image)

	return nil
}

func createImagePusher() (exporter.Exporter, *fuse.Server, error) {
	// Create the runc worker.
	opt, fuseserver, err := runc.NewWorkerOpt(stateDir, backend)
	if err != nil {
		return nil, fuseserver, fmt.Errorf("creating runc worker opt failed: %v", err)
	}

	iw, err := containerimage.NewImageWriter(containerimage.WriterOpt{
		Snapshotter:  opt.Snapshotter,
		ContentStore: opt.ContentStore,
		Differ:       opt.Differ,
	})
	if err != nil {
		return nil, fuseserver, err
	}

	imagePusher, err := imagepush.New(imagepush.Opt{
		Images:      opt.ImageStore,
		ImageWriter: iw,
	})
	if err != nil {
		return nil, fuseserver, err
	}

	return imagePusher, fuseserver, nil
}
