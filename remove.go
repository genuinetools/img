package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"

	"github.com/containerd/containerd/namespaces"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
)

const removeUsageShortHelp = `Remove one or more images.`
const removeUsageLongHelp = `Remove one or more images.`

func newRemoveCommand() *cobra.Command {

	remove := &removeCommand{}

	cmd := &cobra.Command{
		Use:                   "rm [OPTIONS] IMAGE [IMAGE...]",
		DisableFlagsInUseLine: true,
		Short:                 removeUsageShortHelp,
		Long:                  removeUsageLongHelp,
		Args:                  remove.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return remove.Run(args)
		},
	}

	return cmd
}

type removeCommand struct{}

func (cmd *removeCommand) ValidateArgs(c *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image to remove")
	}

	return nil
}

func (cmd *removeCommand) Run(args []string) (err error) {
	reexec()

	// Create the context.
	id := identity.NewID()
	ctx := session.NewContext(context.Background(), id)
	ctx = namespaces.WithNamespace(ctx, "buildkit")

	// Create the client.
	c, err := client.New(stateDir, backend, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// Loop over the arguments as images and run remove.
	for _, image := range args {
		fmt.Printf("Removing %s...\n", image)

		err = c.RemoveImage(ctx, image)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully removed %s\n", image)
	}

	return nil
}
