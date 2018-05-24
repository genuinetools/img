package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/genuinetools/reg/repoutils"
	digest "github.com/opencontainers/go-digest"
	"github.com/urfave/cli"
)

var layerCommand = cli.Command{
	Name:    "layer",
	Aliases: []string{"download"},
	Usage:   "download a layer for the specific reference of a repository",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "output, o",
			Usage: "output file, default to stdout",
		},
	},
	Action: func(c *cli.Context) error {
		if len(c.Args()) < 1 {
			return fmt.Errorf("pass the name of the repository")
		}

		repo, ref, err := repoutils.GetRepoAndRef(c.Args()[0])
		if err != nil {
			return err
		}

		layer, err := r.DownloadLayer(repo, digest.FromString(ref))
		if err != nil {
			return err
		}
		defer layer.Close()

		b, err := ioutil.ReadAll(layer)
		if err != nil {
			return err
		}

		if c.String("output") != "" {
			return ioutil.WriteFile(c.String("output"), b, 0644)
		}

		fmt.Fprint(os.Stdout, string(b))

		return nil
	},
}
