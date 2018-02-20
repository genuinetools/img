package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/containerd/containerd/namespaces"
	units "github.com/docker/go-units"
	"github.com/jessfraz/img/worker/runc"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
)

const listHelp = `List images and digests.`

func (cmd *listCommand) Name() string      { return "ls" }
func (cmd *listCommand) Args() string      { return "[OPTIONS]" }
func (cmd *listCommand) ShortHelp() string { return listHelp }
func (cmd *listCommand) LongHelp() string  { return listHelp }
func (cmd *listCommand) Hidden() bool      { return false }

func (cmd *listCommand) Register(fs *flag.FlagSet) {
	fs.Var(&cmd.filters, "f", "Filter output based on conditions provided")
}

type listCommand struct {
	filters stringSlice
}

func (cmd *listCommand) Run(args []string) (err error) {
	// Create the context.
	ctx := appcontext.Context()
	id := identity.NewID()
	ctx = session.NewContext(ctx, id)
	ctx = namespaces.WithNamespace(ctx, namespaces.Default)

	// Create the runc worker options.
	opt, fuseserver, err := runc.NewWorkerOpt(defaultStateDirectory, backend)
	defer unmount(fuseserver)
	if err != nil {
		return fmt.Errorf("creating runc worker opt failed: %v", err)
	}
	handleSignals(fuseserver)

	images, err := opt.ImageStore.List(ctx, cmd.filters...)
	if err != nil {
		return fmt.Errorf("listing images with filters (%s) failed: %v", strings.Join(cmd.filters, ", "), err)
	}

	tw := tabwriter.NewWriter(os.Stdout, 1, 8, 1, '\t', 0)

	fmt.Fprintln(tw, "NAME\tSIZE\tCREATED AT\tUPDATED AT\tDIGEST")

	for _, image := range images {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			image.Name,
			units.BytesSize(float64(image.Target.Size)),
			units.HumanDuration(time.Now().UTC().Sub(image.CreatedAt))+" ago",
			units.HumanDuration(time.Now().UTC().Sub(image.UpdatedAt))+" ago",
			image.Target.Digest,
		)
	}

	tw.Flush()

	return nil
}
