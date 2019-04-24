package main

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/containerd/containerd/namespaces"
	"github.com/docker/go-units"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
	"golang.org/x/sync/errgroup"
)

const pullUsageShortHelp = `Pull an image or a repository from a registry.`
const pullUsageLongHelp = `Pull an image or a repository from a registry.`

func newPullCommand() *cobra.Command {

	pull := &pullCommand{}

	cmd := &cobra.Command{
		Use:                   "pull [OPTIONS] NAME[:TAG|@DIGEST]",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Short:                 pullUsageShortHelp,
		Long:                  pullUsageLongHelp,
		Args:                  validatePullImageArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pull.Run(args)
		},
	}

	return cmd
}

func validatePullImageArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image or repository to pull")
	}

	return nil
}

type pullCommand struct {
	image string
}

func (cmd *pullCommand) Run(args []string) (err error) {
	reexec()

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
