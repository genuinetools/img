package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/genuinetools/reg/registry"
	"github.com/genuinetools/reg/repoutils"
	"github.com/genuinetools/reg/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	auth types.AuthConfig
	r    *registry.Registry
)

func main() {
	app := cli.NewApp()
	app.Name = "reg"
	app.Version = fmt.Sprintf("version %s, build %s", version.VERSION, version.GITCOMMIT)
	app.Author = "The Genuinetools Authors"
	app.Email = "no-reply@butts.com"
	app.Usage = "Docker registry v2 client."

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "run in debug mode",
		},
		cli.BoolFlag{
			Name:  "insecure, k",
			Usage: "do not verify tls certificates",
		},
		cli.BoolFlag{
			Name:  "force-non-ssl, f",
			Usage: "force allow use of non-ssl",
		},
		cli.StringFlag{
			Name:  "username, u",
			Usage: "username for the registry",
		},
		cli.StringFlag{
			Name:  "password, p",
			Usage: "password for the registry",
		},
		cli.StringFlag{
			Name:   "registry, r",
			Usage:  "URL to the private registry (ex. r.j3ss.co)",
			Value:  repoutils.DefaultDockerRegistry,
			EnvVar: "REG_REGISTRY",
		},
	}

	app.Commands = []cli.Command{
		deleteCommand,
		layerCommand,
		listCommand,
		manifestCommand,
		tagsCommand,
		vulnsCommand,
	}

	app.Before = func(c *cli.Context) (err error) {
		// Preload initializes any global options and configuration
		// before the main or sub commands are run.
		if c.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if len(c.Args()) == 0 {
			return
		}

		if c.Args()[0] == "help" {
			return
		}

		auth, err = repoutils.GetAuthConfig(c.GlobalString("username"), c.GlobalString("password"), c.GlobalString("registry"))
		if err != nil {
			return err
		}

		// Prevent non-ssl unless explicitly forced
		if !c.GlobalBool("force-non-ssl") && strings.HasPrefix(auth.ServerAddress, "http:") {
			return fmt.Errorf("Attempt to use insecure protocol! Use non-ssl option to force")
		}

		// create the registry client
		if c.GlobalBool("insecure") {
			r, err = registry.NewInsecure(auth, c.GlobalBool("debug"))
			if err != nil {
				return err
			}

			return
		}

		r, err = registry.New(auth, c.GlobalBool("debug"))
		return err
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
