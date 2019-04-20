package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
	"time"

	"github.com/containerd/containerd/namespaces"
	"github.com/docker/go-units"
	"github.com/genuinetools/img/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
)

const listUsageShortHelp = `List images and digests.`
const listUsageLongHelp = `List images and digests.`

func newListCommand() *cobra.Command {

	list := &listCommand{
		filters: newListValue(),
	}

	cmd := &cobra.Command{
		Use:                   "ls [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 listUsageShortHelp,
		Long:                  listUsageLongHelp,
		Args:                  validateHasNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return list.Run(args)
		},
	}

	fs := cmd.Flags()

	fs.VarP(list.filters, "filter", "f", "Filter output based on conditions provided")

	return cmd
}

type listCommand struct {
	filters *listValue
}

func (cmd *listCommand) Run(args []string) (err error) {
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

	images, err := c.ListImages(ctx, cmd.filters.GetAll()...)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 1, 8, 1, '\t', 0)

	fmt.Fprintln(tw, "NAME\tSIZE\tCREATED AT\tUPDATED AT\tDIGEST")

	for _, image := range images {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			image.Name,
			units.BytesSize(float64(image.ContentSize)),
			units.HumanDuration(time.Now().UTC().Sub(image.CreatedAt))+" ago",
			units.HumanDuration(time.Now().UTC().Sub(image.UpdatedAt))+" ago",
			image.Target.Digest,
		)
	}

	tw.Flush()

	return nil
}
