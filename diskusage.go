package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/containerd/containerd/namespaces"
	units "github.com/docker/go-units"
	"github.com/genuinetools/img/client"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
)

const diskUsageShortHelp = `Show image disk usage.`

// TODO: make the long help actually useful
const diskUsageLongHelp = `Show image disk usage.`

func (cmd *diskUsageCommand) Name() string      { return "du" }
func (cmd *diskUsageCommand) Args() string      { return "[OPTIONS]" }
func (cmd *diskUsageCommand) ShortHelp() string { return diskUsageShortHelp }
func (cmd *diskUsageCommand) LongHelp() string  { return diskUsageLongHelp }
func (cmd *diskUsageCommand) Hidden() bool      { return false }

func (cmd *diskUsageCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.filter, "f", "", "Filter output based on conditions provided (snapshot ID supported)")
}

type diskUsageCommand struct {
	filter string
}

func (cmd *diskUsageCommand) Run(args []string) (err error) {
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

	resp, err := c.DiskUsage(ctx, &controlapi.DiskUsageRequest{Filter: cmd.filter})
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 1, 8, 1, '\t', 0)

	if debug {
		printDebug(tw, resp.Record)
	} else {
		fmt.Fprintln(tw, "ID\tRECLAIMABLE\tSIZE\tDESCRIPTION")

		for _, di := range resp.Record {
			id := di.ID
			if di.Mutable {
				id += "*"
			}
			fmt.Fprintf(tw, "%-71s\t%-11v\t%s\t%s\n", id, !di.InUse, units.BytesSize(float64(di.Size_)), di.Description)
		}

		tw.Flush()
	}

	if cmd.filter == "" {
		total := int64(0)
		reclaimable := int64(0)

		for _, di := range resp.Record {
			if di.Size_ > 0 {
				total += di.Size_
				if !di.InUse {
					reclaimable += di.Size_
				}
			}
		}

		tw = tabwriter.NewWriter(os.Stdout, 1, 8, 1, '\t', 0)
		fmt.Fprintf(tw, "Reclaimable:\t%s\n", units.BytesSize(float64(reclaimable)))
		fmt.Fprintf(tw, "Total:\t%s\n", units.BytesSize(float64(total)))
		tw.Flush()
	}

	return nil
}

func printDebug(tw *tabwriter.Writer, du []*controlapi.UsageRecord) {
	for _, di := range du {
		fmt.Fprintf(tw, "%s:\t%v\n", "ID", di.ID)
		if di.Parent != "" {
			fmt.Fprintf(tw, "%s:\t%v\n", "Parent", di.Parent)
		}
		fmt.Fprintf(tw, "%s:\t%v\n", "Created at", di.CreatedAt)
		fmt.Fprintf(tw, "%s:\t%v\n", "Mutable", di.Mutable)
		fmt.Fprintf(tw, "%s:\t%v\n", "Reclaimable", !di.InUse)
		fmt.Fprintf(tw, "%s:\t%s\n", "Size", units.BytesSize(float64(di.Size_)))
		if di.Description != "" {
			fmt.Fprintf(tw, "%s:\t%v\n", "Description", di.Description)
		}
		fmt.Fprintf(tw, "%s:\t%d\n", "Usage count", di.UsageCount)
		if di.LastUsedAt != nil {
			fmt.Fprintf(tw, "%s:\t%v\n", "Last used", di.LastUsedAt)
		}

		fmt.Fprintf(tw, "\n")
	}

	tw.Flush()
}
