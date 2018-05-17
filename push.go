package main

import (
	"flag"
	"fmt"

	"github.com/containerd/containerd/namespaces"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
	"golang.org/x/sync/errgroup"
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

	// Create the client.
	c, err := client.New(stateDir, backend, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	fmt.Printf("Pushing %s...\n", cmd.image)

	// Create the context.
	ctx := appcontext.Context()
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
		return c.Push(ctx, cmd.image)
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	fmt.Printf("Successfully pushed %s\n", cmd.image)

	return nil
}
