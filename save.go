package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/containerd/containerd/namespaces"
	"github.com/docker/docker/pkg/term"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
)

// TODO(AkihiroSuda): support OCI archive
const saveUsageShortHelp = `Save one or more images to a tar archive (streamed to STDOUT by default).`
const saveUsageLongHelp = `Save one or more images to a tar archive (streamed to STDOUT by default).`

func newSaveCommand() *cobra.Command {

	save := &saveCommand{}

	cmd := &cobra.Command{
		Use:                   "save [OPTIONS] IMAGE [IMAGE...]",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Short:                 saveUsageShortHelp,
		Long:                  saveUsageLongHelp,
		Args:                  save.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return save.Run(args)
		},
	}

	fs := cmd.Flags()

	fs.StringVarP(&save.output, "output", "o", "", "write to a file, instead of STDOUT")
	fs.StringVar(&save.format, "format", "docker", "image output format (docker|oci)")

	return cmd
}

type saveCommand struct {
	output string
	format string
}

func (cmd *saveCommand) ValidateArgs(c *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image to save")
	}

	return nil
}

func (cmd *saveCommand) Run(args []string) (err error) {
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

	// Create the writer.
	writer, err := cmd.writer()
	if err != nil {
		return err
	}

	// Assume that the arguments are all image references
	if err := c.SaveImages(ctx, args, cmd.format, writer); err != nil {
		return err
	}

	return nil
}

func (cmd *saveCommand) writer() (io.WriteCloser, error) {
	if cmd.output != "" {
		return os.Create(cmd.output)
	}

	if term.IsTerminal(os.Stdout.Fd()) {
		return nil, fmt.Errorf("cowardly refusing to save to a terminal. Use the -o flag or redirect")
	}

	return os.Stdout, nil
}
