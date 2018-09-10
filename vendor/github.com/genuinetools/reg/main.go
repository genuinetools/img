package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/genuinetools/pkg/cli"
	"github.com/genuinetools/reg/registry"
	"github.com/genuinetools/reg/repoutils"
	"github.com/genuinetools/reg/version"
	"github.com/sirupsen/logrus"
)

var (
	insecure    bool
	forceNonSSL bool
	skipPing    bool

	timeout time.Duration

	username string
	password string

	debug bool
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "reg"
	p.Description = "Docker registry v2 client"
	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Build the list of available commands.
	p.Commands = []cli.Command{
		&digestCommand{},
		&layerCommand{},
		&listCommand{},
		&manifestCommand{},
		&removeCommand{},
		&serverCommand{},
		&tagsCommand{},
		&vulnsCommand{},
	}

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.BoolVar(&insecure, "insecure", false, "do not verify tls certificates")
	p.FlagSet.BoolVar(&insecure, "k", false, "do not verify tls certificates")

	p.FlagSet.BoolVar(&forceNonSSL, "force-non-ssl", false, "force allow use of non-ssl")
	p.FlagSet.BoolVar(&forceNonSSL, "f", false, "force allow use of non-ssl")

	p.FlagSet.BoolVar(&skipPing, "skip-ping", false, "skip pinging the registry while establishing connection")

	p.FlagSet.DurationVar(&timeout, "timeout", time.Minute, "timeout for HTTP requests")

	p.FlagSet.StringVar(&username, "username", "", "username for the registry")
	p.FlagSet.StringVar(&username, "u", "", "username for the registry")

	p.FlagSet.StringVar(&password, "password", "", "password for the registry")
	p.FlagSet.StringVar(&password, "p", "", "password for the registry")

	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// On ^C, or SIGTERM handle exit.
		signals := make(chan os.Signal, 0)
		signal.Notify(signals, os.Interrupt)
		signal.Notify(signals, syscall.SIGTERM)
		_, cancel := context.WithCancel(ctx)
		go func() {
			for sig := range signals {
				cancel()
				logrus.Infof("Received %s, exiting.", sig.String())
				os.Exit(0)
			}
		}()

		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		return nil
	}

	// Run our program.
	p.Run()
}

func createRegistryClient(domain string) (*registry.Registry, error) {
	auth, err := repoutils.GetAuthConfig(username, password, domain)
	if err != nil {
		return nil, err
	}

	// Prevent non-ssl unless explicitly forced
	if !forceNonSSL && strings.HasPrefix(auth.ServerAddress, "http:") {
		return nil, fmt.Errorf("Attempted to use insecure protocol! Use force-non-ssl option to force")
	}

	// Create the registry client.
	return registry.New(auth, registry.Opt{
		Insecure: insecure,
		Debug:    debug,
		SkipPing: skipPing,
		Timeout:  timeout,
	})
}
