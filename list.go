package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	units "github.com/docker/go-units"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/util/appcontext"
)

const listShortHelp = `List images and digests.`

// TODO: make the long help actually useful
const listLongHelp = `List images and digests.`

func (cmd *listCommand) Name() string      { return "ls" }
func (cmd *listCommand) Args() string      { return "[OPTIONS]" }
func (cmd *listCommand) ShortHelp() string { return listShortHelp }
func (cmd *listCommand) LongHelp() string  { return listLongHelp }
func (cmd *listCommand) Hidden() bool      { return false }

func (cmd *listCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.filter, "f", "", "Filter output based on conditions provided (snapshot ID supported)")
}

type listCommand struct {
	filter string
}

func (cmd *listCommand) Run(args []string) (err error) {
	// Create the context.
	ctx := appcontext.Context()
	ref := identity.NewID()

	// Create the controller.
	c, err := createController(cmd, ref)
	if err != nil {
		return err
	}

	resp, err := c.DiskUsage(ctx, &controlapi.DiskUsageRequest{Filter: cmd.filter})
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 1, 8, 1, '\t', 0)

	if debug {
		printDebug(tw, resp.Record)
	} else {
		printTable(tw, resp.Record)
	}

	if cmd.filter == "" {
		printSummary(tw, resp.Record)
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

func printTable(tw *tabwriter.Writer, du []*controlapi.UsageRecord) {
	fmt.Fprintln(tw, "ID\tRECLAIMABLE\tSIZE\tLAST ACCESSED")

	for _, di := range du {
		id := di.ID
		if di.Mutable {
			id += "*"
		}
		fmt.Fprintf(tw, "%-71s\t%-11v\t%s\t\n", id, !di.InUse, units.BytesSize(float64(di.Size_)))
	}

	tw.Flush()
}

func printSummary(tw *tabwriter.Writer, du []*controlapi.UsageRecord) {
	total := int64(0)
	reclaimable := int64(0)

	for _, di := range du {
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
