package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/namespaces"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
)

const unpackUsageShortHelp = `Unpack an image to a rootfs directory.`
const unpackUsageLongHelp = `Unpack an image to a rootfs directory.`

func newUnpackCommand() *cobra.Command {

	unpack := &unpackCommand{}

	cmd := &cobra.Command{
		Use:                   "unpack [OPTIONS] IMAGE",
		DisableFlagsInUseLine: true,
		Short:                 unpackUsageShortHelp,
		Long:                  unpackUsageLongHelp,
		Args:                  validateUnpackImageArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return unpack.Run(args)
		},
	}

	fs := cmd.Flags()

	fs.StringVarP(&unpack.output, "output", "o", "", "Directory to unpack the rootfs to. (defaults to rootfs/ in the current working directory)")

	return cmd
}

func validateUnpackImageArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image to unpack as a rootfs")
	}

	return nil
}

type unpackCommand struct {
	image  string
	output string
}

func (cmd *unpackCommand) Run(args []string) (err error) {
	reexec()

	cmd.image = args[0]

	if len(cmd.output) < 1 {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current working directory failed: %v", err)
		}
		cmd.output = filepath.Join(wd, "rootfs")
	}

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

	if err := c.Unpack(ctx, cmd.image, cmd.output); err != nil {
		return err
	}

	fmt.Printf("Successfully unpacked rootfs for %s to: %s\n", cmd.image, cmd.output)

	return nil
}
