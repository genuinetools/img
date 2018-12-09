package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/docker/cli/cli/config"
	"github.com/docker/docker/registry"
)

const logoutShortHelp = `Log out from a Docker registry.`

var logoutLongHelp = logoutShortHelp + fmt.Sprintf("\n\nIf no server is specified, the default (%s) is used.", defaultDockerRegistry)

func (cmd *logoutCommand) Name() string       { return "logout" }
func (cmd *logoutCommand) Args() string       { return "[SERVER]" }
func (cmd *logoutCommand) ShortHelp() string  { return logoutShortHelp }
func (cmd *logoutCommand) LongHelp() string   { return logoutLongHelp }
func (cmd *logoutCommand) Hidden() bool       { return false }
func (cmd *logoutCommand) DoReexec() bool     { return false }
func (cmd *logoutCommand) RequiresRunc() bool { return false }

func (cmd *logoutCommand) Register(fs *flag.FlagSet) {
}

type logoutCommand struct {
	serverAddress string
}

func (cmd *logoutCommand) Run(ctx context.Context, args []string) error {
	if len(args) > 0 {
		cmd.serverAddress = args[0]
	}

	// Set the default registry server address.
	if cmd.serverAddress == "" {
		cmd.serverAddress = defaultDockerRegistry
	} else {
		cmd.serverAddress = registry.ConvertToHostname(cmd.serverAddress)
	}

	dcfg, err := config.Load(config.Dir())
	if err != nil {
		return fmt.Errorf("loading config file failed: %v", err)
	}

	// check if we're logged in based on the records in the config file
	// which means it couldn't have user/pass cause they may be in the creds store
	if _, loggedIn := dcfg.AuthConfigs[cmd.serverAddress]; !loggedIn {
		fmt.Fprintf(os.Stdout, "Not logged in to %s\n", cmd.serverAddress)
		return nil
	}

	fmt.Fprintf(os.Stdout, "Removing login credentials for %s\n", cmd.serverAddress)
	if err := dcfg.GetCredentialsStore(cmd.serverAddress).Erase(cmd.serverAddress); err != nil {
		fmt.Fprintf(os.Stdout, "WARNING: could not erase credentials: %v\n", err)
	}

	fmt.Println("Logout succeeded.")

	return nil
}
