package main

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/namespaces"
	"github.com/jessfraz/img/worker/runc"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/sirupsen/logrus"
)

const tagHelp = `Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE.`

func (cmd *tagCommand) Name() string      { return "tag" }
func (cmd *tagCommand) Args() string      { return "SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]" }
func (cmd *tagCommand) ShortHelp() string { return tagHelp }
func (cmd *tagCommand) LongHelp() string  { return tagHelp }
func (cmd *tagCommand) Hidden() bool      { return false }

func (cmd *tagCommand) Register(fs *flag.FlagSet) {}

type tagCommand struct {
	image  string
	target string
}

func (cmd *tagCommand) Run(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("must pass an image or repository and target to tag")
	}

	// Get the specified image and target.
	cmd.image = args[0]
	cmd.target = args[1]
	// Add the latest lag if they did not provide one.
	cmd.image = addLatestTagSuffix(cmd.image)
	cmd.target = addLatestTagSuffix(cmd.target)

	// Create the context.
	ctx := appcontext.Context()
	id := identity.NewID()
	ctx = session.NewContext(ctx, id)
	ctx = namespaces.WithNamespace(ctx, namespaces.Default)

	// Create the runc worker.
	opt, fuseserver, err := runc.NewWorkerOpt(stateDir, backend)
	defer unmount(fuseserver)
	if err != nil {
		return fmt.Errorf("creating runc worker opt failed: %v", err)
	}
	handleSignals(fuseserver)

	if opt.ImageStore == nil {
		return errors.New("image store is nil")
	}

	// Get the source image.
	image, err := opt.ImageStore.Get(ctx, cmd.image)
	if err != nil {
		return fmt.Errorf("getting image %s from image store failed: %v", cmd.image, err)
	}

	// Update the target image. Create it if it does not exist.
	img := images.Image{
		Name:      cmd.target,
		Target:    image.Target,
		CreatedAt: time.Now(),
	}
	if _, err := opt.ImageStore.Update(ctx, img); err != nil {
		if !errdefs.IsNotFound(err) {
			return err
		}

		logrus.Debugf("Creating new tag: %s", cmd.target)
		if _, err := opt.ImageStore.Create(ctx, img); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully tagged %s as %s", cmd.image, cmd.target)

	return nil
}
