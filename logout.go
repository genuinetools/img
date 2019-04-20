package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/docker/cli/cli/config"
	"github.com/docker/docker/registry"
)

const logoutShortHelp = `Log out from a Docker registry.`

var logoutLongHelp = logoutShortHelp + fmt.Sprintf("\n\nIf no server is specified, the default (%s) is used.", defaultDockerRegistry)

func newLogoutCommand() *cobra.Command {

	logout := &logoutCommand{}

	cmd := &cobra.Command{
		Use:                   "logout [SERVER]",
		DisableFlagsInUseLine: true,
		Short:                 logoutShortHelp,
		Long:                  logoutLongHelp,
		Args:                  logout.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return logout.Run(args)
		},
	}

	return cmd
}

type logoutCommand struct {
	serverAddress string
}

func (cmd *logoutCommand) ValidateArgs(c *cobra.Command, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("logout expects zero or one arguments, found %d", len(args))
	}

	return nil
}

func (cmd *logoutCommand) Run(args []string) error {
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
