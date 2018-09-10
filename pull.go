package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/containerd/containerd/namespaces"
	units "github.com/docker/go-units"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
	"golang.org/x/sync/errgroup"
)

const pullHelp = `Pull an image or a repository from a registry.`

func (cmd *pullCommand) Name() string       { return "pull" }
func (cmd *pullCommand) Args() string       { return "[OPTIONS] NAME[:TAG|@DIGEST]" }
func (cmd *pullCommand) ShortHelp() string  { return pullHelp }
func (cmd *pullCommand) LongHelp() string   { return pullHelp }
func (cmd *pullCommand) Hidden() bool       { return false }
func (cmd *pullCommand) DoReexec() bool     { return true }
func (cmd *pullCommand) RequiresRunc() bool { return false }

func (cmd *pullCommand) Register(fs *flag.FlagSet) {}

type pullCommand struct {
	image string
}

func (cmd *pullCommand) Run(ctx context.Context, args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image or repository to pull")
	}

	// Get the specified image.
	cmd.image = args[0]

	// Create the client.
	c, err := client.New(stateDir, backend, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	fmt.Printf("Pulling %s...\n", cmd.image)

	var listedImage *client.ListedImage
	// Create the context.
	ctx = appcontext.Context()
	sess, sessDialer, err := c.Session(ctx)
	if err != nil {
		return err
	}
	ctx = session.NewContext(ctx, sess.ID())
	ctx = namespaces.WithNamespace(ctx, "buildkit")
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return sess.Run(ctx, sessDialer)
	})
	eg.Go(func() error {
		defer sess.Close()
		var err error
		listedImage, err = c.Pull(ctx, cmd.image)
		return err
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	fmt.Printf("Pulled: %s\n", listedImage.Target.Digest)
	fmt.Printf("Size: %s\n", units.BytesSize(float64(listedImage.ContentSize)))

	return nil
}
