package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/namespaces"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
)

const loadUsageShortHelp = `Load an image from a tar archive or STDIN.`
const loadUsageLongHelp = `Load an image from a tar archive or STDIN.`

func newLoadCommand() *cobra.Command {

	load := &loadCommand{}

	cmd := &cobra.Command{
		Use:                   "load [OPTIONS]",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Short:                 loadUsageShortHelp,
		Long:                  loadUsageLongHelp,
		// Args:                  load.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return load.Run(args)
		},
	}

	fs := cmd.Flags()

	fs.StringVarP(&load.input, "input", "i", "", "Read from tar archive file, instead of STDIN")
	fs.BoolVarP(&load.quiet, "quiet", "q", false, "Suppress the load output")

	return cmd
}

type loadCommand struct {
	input string
	quiet bool
}

func (cmd *loadCommand) Run(args []string) (err error) {
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

	// Create the reader.
	reader, err := cmd.reader()
	if err != nil {
		return err
	}
	defer reader.Close()

	// Load images.
	return c.LoadImage(ctx, reader, cmd)
}

func (cmd *loadCommand) reader() (io.ReadCloser, error) {
	if cmd.input != "" {
		return os.Open(cmd.input)
	}

	return os.Stdin, nil
}

func (cmd *loadCommand) LoadImageCreated(image images.Image) error {
	if cmd.quiet {
		return nil
	}
	_, err := fmt.Printf("Loaded image: %s\n", image.Name)
	return err
}

func (cmd *loadCommand) LoadImageUpdated(image images.Image) error {
	return cmd.LoadImageCreated(image)
}

func (cmd *loadCommand) LoadImageReplaced(before, after images.Image) error {
	if !cmd.quiet {
		_, err := fmt.Printf(
			"The image %s already exists, leaving the old one with ID %s orphaned\n",
			after.Name, before.Target.Digest)
		if err != nil {
			return err
		}
	}

	return cmd.LoadImageCreated(after)
}
