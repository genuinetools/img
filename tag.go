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

const tagUsageShortHelp = `Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE.`
const tagUsageLongHelp = `Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE.`

func newTagCommand() *cobra.Command {

	tag := &tagCommand{}

	cmd := &cobra.Command{
		Use:                   "tag SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Short:                 tagUsageShortHelp,
		Long:                  tagUsageLongHelp,
		Args:                  tag.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return tag.Run(args)
		},
	}

	return cmd
}

type tagCommand struct {
	image  string
	target string
}

func (cmd *tagCommand) ValidateArgs(c *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("must pass an image or repository and target to tag")
	}

	return nil
}

func (cmd *tagCommand) Run(args []string) (err error) {
	reexec()

	// Get the specified image and target.
	cmd.image = args[0]
	cmd.target = args[1]

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

	if err := c.TagImage(ctx, cmd.image, cmd.target); err != nil {
		return err
	}

	fmt.Printf("Successfully tagged %s as %s\n", cmd.image, cmd.target)

	return nil
}
