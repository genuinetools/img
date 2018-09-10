package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/namespaces"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
)

const unpackHelp = `Unpack an image to a rootfs directory.`

func (cmd *unpackCommand) Name() string       { return "unpack" }
func (cmd *unpackCommand) Args() string       { return "[OPTIONS] IMAGE" }
func (cmd *unpackCommand) ShortHelp() string  { return unpackHelp }
func (cmd *unpackCommand) LongHelp() string   { return unpackHelp }
func (cmd *unpackCommand) Hidden() bool       { return false }
func (cmd *unpackCommand) DoReexec() bool     { return true }
func (cmd *unpackCommand) RequiresRunc() bool { return false }

func (cmd *unpackCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.output, "output", "", "Directory to unpack the rootfs to. (defaults to rootfs/ in the current working directory)")
	fs.StringVar(&cmd.output, "o", "", "Directory to unpack the rootfs to. (defaults to rootfs/ in the current working directory)")
}

type unpackCommand struct {
	image  string
	output string
}

func (cmd *unpackCommand) Run(ctx context.Context, args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("must pass an image to unpack as a rootfs")
	}

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
	ctx = session.NewContext(ctx, id)
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
