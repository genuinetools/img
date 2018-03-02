package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/containerd/containerd/namespaces"
	"github.com/docker/docker/pkg/term"
	"github.com/jessfraz/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
)

// TODO(AkihiroSuda): support saving multiple images
// TODO(AkihiroSuda): support OCI archive
const saveHelp = `Save an image to a tar archive (streamed to STDOUT by default)`

func (cmd *saveCommand) Name() string      { return "save" }
func (cmd *saveCommand) Args() string      { return "[OPTIONS] NAME" }
func (cmd *saveCommand) ShortHelp() string { return saveHelp }
func (cmd *saveCommand) LongHelp() string  { return saveHelp }
func (cmd *saveCommand) Hidden() bool      { return false }

func (cmd *saveCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.output, "o", "", "Write to a file, instead of STDOUT")
}

type saveCommand struct {
	image  string
	output string
}

func (cmd *saveCommand) Run(args []string) (err error) {
	if len(args) != 1 {
		return fmt.Errorf("must pass an image")
	}

	// Get the specified image.
	cmd.image = args[0]
	// Add the latest lag if they did not provide one.
	cmd.image = addLatestTagSuffix(cmd.image)

	// Create the context.
	ctx := appcontext.Context()
	id := identity.NewID()
	ctx = session.NewContext(ctx, id)
	ctx = namespaces.WithNamespace(ctx, namespaces.Default)

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

	return c.SaveImage(ctx, cmd.image, writer)
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
