package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/containerd/containerd/namespaces"
	units "github.com/docker/go-units"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
)

const pruneHelp = `Prune and clean up the build cache.`

func (cmd *pruneCommand) Name() string       { return "prune" }
func (cmd *pruneCommand) Args() string       { return "[OPTIONS]" }
func (cmd *pruneCommand) ShortHelp() string  { return pruneHelp }
func (cmd *pruneCommand) LongHelp() string   { return pruneHelp }
func (cmd *pruneCommand) Hidden() bool       { return false }
func (cmd *pruneCommand) DoReexec() bool     { return true }
func (cmd *pruneCommand) RequiresRunc() bool { return false }

func (cmd *pruneCommand) Register(fs *flag.FlagSet) {}

type pruneCommand struct{}

func (cmd *pruneCommand) Run(args []string) (err error) {
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

	usage, err := c.Prune(ctx)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 1, 8, 1, '\t', 0)

	if debug {
		printDebug(tw, usage)
	} else {
		fmt.Fprintln(tw, "ID\tRECLAIMABLE\tSIZE\tDESCRIPTION")

		for _, di := range usage {
			id := di.ID
			if di.Mutable {
				id += "*"
			}
			desc := di.Description
			if len(desc) > 50 {
				desc = desc[0:50] + "..."
			}
			fmt.Fprintf(tw, "%s\t%t\t%s\t%s\n", id, !di.InUse, units.BytesSize(float64(di.Size_)), desc)
		}

		tw.Flush()
	}

	total := int64(0)
	reclaimable := int64(0)

	for _, di := range usage {
		if di.Size_ > 0 {
			total += di.Size_
			if !di.InUse {
				reclaimable += di.Size_
			}
		}
	}

	tw = tabwriter.NewWriter(os.Stdout, 1, 8, 1, '\t', 0)
	fmt.Fprintf(tw, "Reclaimed:\t%s\n", units.BytesSize(float64(reclaimable)))
	fmt.Fprintf(tw, "Total:\t%s\n", units.BytesSize(float64(total)))
	tw.Flush()

	return nil
}
