package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/containerd/containerd/namespaces"
	"github.com/docker/docker/pkg/term"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
)

// TODO(AkihiroSuda): support saving multiple images
// TODO(AkihiroSuda): support OCI archive
const saveHelp = `Save an image to a tar archive (streamed to STDOUT by default).`

func (cmd *saveCommand) Name() string      { return "save" }
func (cmd *saveCommand) Args() string      { return "[OPTIONS] IMAGE [IMAGE...]" }
func (cmd *saveCommand) ShortHelp() string { return saveHelp }
func (cmd *saveCommand) LongHelp() string  { return saveHelp }
func (cmd *saveCommand) Hidden() bool      { return false }
func (cmd *saveCommand) DoReexec() bool    { return true }

func (cmd *saveCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.output, "o", "", "Write to a file, instead of STDOUT")
}

type saveCommand struct {
	output string
}

func (cmd *saveCommand) Run(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image to save")
	}

	// Create the context.
	ctx := appcontext.Context()
	id := identity.NewID()
	ctx = session.NewContext(ctx, id)
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

	// Loop over the arguments as images and run save.
	for _, image := range args {
		if err := c.SaveImage(ctx, image, writer); err != nil {
			return err
		}
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
