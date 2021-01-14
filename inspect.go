package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"

	"github.com/containerd/containerd/namespaces"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
)

const inspectUsageShortHelp = `Return the JSON-encoded OCI image config. The output format is not compatible with "docker inspect".`
const inspectUsageLongHelp = `Return the JSON-encoded OCI image config. The output format is not compatible with "docker inspect".`

func newInspectCommand() *cobra.Command {
	inspect := &inspectCommand{}

	cmd := &cobra.Command{
		Use:                   "inspect NAME[:TAG]",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Short:                 inspectUsageShortHelp,
		Long:                  inspectUsageLongHelp,
		Args:                  inspect.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return inspect.Run(args)
		},
	}

	return cmd
}

type inspectCommand struct {
	image string
}

func (cmd *inspectCommand) ValidateArgs(c *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image to inspect")
	}

	return nil
}

func (cmd *inspectCommand) Run(args []string) (err error) {
	reexec()

	// Get the specified image and target.
	cmd.image = args[0]

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

	image, err := c.InspectImage(ctx, cmd.image)
	if err != nil {
		return err
	}

	fmted, err := json.MarshalIndent(image, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(fmted))

	return nil
}
