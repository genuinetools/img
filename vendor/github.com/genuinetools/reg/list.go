package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
)

const listHelp = `List all repositories.`

func (cmd *listCommand) Name() string      { return "ls" }
func (cmd *listCommand) Args() string      { return "[OPTIONS] REGISTRY_DOMAIN" }
func (cmd *listCommand) ShortHelp() string { return listHelp }
func (cmd *listCommand) LongHelp() string  { return listHelp }
func (cmd *listCommand) Hidden() bool      { return false }

func (cmd *listCommand) Register(fs *flag.FlagSet) {}

type listCommand struct{}

func (cmd *listCommand) Run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("pass the domain of the registry")
	}

	// Create the registry client.
	r, err := createRegistryClient(args[0])
	if err != nil {
		return err
	}
	// Get the repositories via catalog.
	repos, err := r.Catalog("")
	if err != nil {
		return err
	}
	sort.Strings(repos)

	fmt.Printf("Repositories for %s\n", r.Domain)

	var (
		l        sync.Mutex
		wg       sync.WaitGroup
		repoTags = map[string][]string{}
	)

	wg.Add(len(repos))
	for _, repo := range repos {
		go func(repo string) {
			// Get the tags.
			tags, err := r.Tags(repo)
			if err != nil {
				fmt.Printf("Get tags of [%s] error: %s", repo, err)
			}
			// Sort the tags
			sort.Strings(tags)

			// Lock on the write to the map.
			l.Lock()
			repoTags[repo] = tags
			l.Unlock()

			wg.Done()
		}(repo)
	}
	wg.Wait()

	// Setup the tab writer.
	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)

	// Print header.
	fmt.Fprintln(w, "REPO\tTAGS")

	// Sort the repos.
	for _, repo := range repos {
		w.Write([]byte(fmt.Sprintf("%s\t%s\n", repo, strings.Join(repoTags[repo], ", "))))
	}

	w.Flush()

	return nil
}
