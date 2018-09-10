package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/containerd/containerd/namespaces"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
)

const tagHelp = `Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE.`

func (cmd *tagCommand) Name() string       { return "tag" }
func (cmd *tagCommand) Args() string       { return "SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]" }
func (cmd *tagCommand) ShortHelp() string  { return tagHelp }
func (cmd *tagCommand) LongHelp() string   { return tagHelp }
func (cmd *tagCommand) Hidden() bool       { return false }
func (cmd *tagCommand) DoReexec() bool     { return true }
func (cmd *tagCommand) RequiresRunc() bool { return false }

func (cmd *tagCommand) Register(fs *flag.FlagSet) {}

type tagCommand struct {
	image  string
	target string
}

func (cmd *tagCommand) Run(ctx context.Context, args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("must pass an image or repository and target to tag")
	}

	reexec()

	// Get the specified image and target.
	cmd.image = args[0]
	cmd.target = args[1]

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

	if err := c.TagImage(ctx, cmd.image, cmd.target); err != nil {
		return err
	}

	fmt.Printf("Successfully tagged %s as %s\n", cmd.image, cmd.target)

	return nil
}
