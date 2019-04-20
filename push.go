package main

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/containerd/containerd/namespaces"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
	"golang.org/x/sync/errgroup"
)

const pushUsageShortHelp = `Push an image or a repository to a registry.`
const pushUsageLongHelp = `Push an image or a repository to a registry.`

func newPushCommand() *cobra.Command {

	push := &pushCommand{}

	cmd := &cobra.Command{
		Use:                   "push [OPTIONS] NAME[:TAG]",
		DisableFlagsInUseLine: true,
		Short:                 pushUsageShortHelp,
		Long:                  pushUsageLongHelp,
		Args:                  push.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return push.Run(args)
		},
	}

	fs := cmd.Flags()

	fs.BoolVar(&push.insecure, "insecure-registry", false, "Push to insecure registry")

	return cmd
}

type pushCommand struct {
	image    string
	insecure bool
}

func (cmd *pushCommand) ValidateArgs(c *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image or repository to push")
	}

	return nil
}

func (cmd *pushCommand) Run(args []string) (err error) {
	reexec()

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
		return c.Push(ctx, cmd.image, cmd.insecure)
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	fmt.Printf("Successfully pushed %s\n", cmd.image)

	return nil
}
