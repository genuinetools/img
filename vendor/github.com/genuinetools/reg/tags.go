package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

var tagsCommand = cli.Command{
	Name:  "tags",
	Usage: "get the tags for a repository",
	Action: func(c *cli.Context) error {
		if len(c.Args()) < 1 {
			return fmt.Errorf("pass the name of the repository")
		}

		tags, err := r.Tags(c.Args()[0])
		if err != nil {
			return err
		}

		// Print the tags.
		fmt.Println(strings.Join(tags, "\n"))

		return nil
	},
}
