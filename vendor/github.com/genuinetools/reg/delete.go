package main

import (
	"fmt"

	"github.com/genuinetools/reg/repoutils"
	"github.com/urfave/cli"
)

var deleteCommand = cli.Command{
	Name:    "delete",
	Aliases: []string{"rm"},
	Usage:   "delete a specific reference of a repository",
	Action: func(c *cli.Context) error {
		if len(c.Args()) < 1 {
			return fmt.Errorf("pass the name of the repository")
		}

		repo, ref, err := repoutils.GetRepoAndRef(c.Args()[0])
		if err != nil {
			return err
		}

		if err := r.Delete(repo, ref); err != nil {
			return fmt.Errorf("Delete %s@%s failed: %v", repo, ref, err)
		}
		fmt.Printf("Deleted %s@%s\n", repo, ref)

		return nil
	},
}
