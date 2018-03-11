package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/urfave/cli"
)

var listCommand = cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "list all repositories",
	Action: func(c *cli.Context) error {
		// Get the repositories via catalog.
		repos, err := r.Catalog("")
		if err != nil {
			return err
		}

		fmt.Printf("Repositories for %s\n", auth.ServerAddress)

		// Setup the tab writer.
		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)

		// Print header.
		fmt.Fprintln(w, "REPO\tTAGS")

		var (
			l  sync.Mutex
			wg sync.WaitGroup
		)

		wg.Add(len(repos))
		for _, repo := range repos {
			go func(repo string) {
				// Get the tags and print to stdout.
				tags, err := r.Tags(repo)
				if err != nil {
					fmt.Printf("Get tags of [%s] error: %s", repo, err)
				}
				out := fmt.Sprintf("%s\t%s\n", repo, strings.Join(tags, ", "))

				// Lock around the tabwriter to prevent garbled output.
				// See: https://github.com/genuinetools/reg/issues/54
				l.Lock()
				w.Write([]byte(out))
				l.Unlock()

				wg.Done()
			}(repo)
		}
		wg.Wait()

		w.Flush()

		return nil
	},
}
