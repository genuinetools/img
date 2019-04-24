package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"
	dockerapitypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
	registryapi "github.com/genuinetools/reg/registry"
	"github.com/sirupsen/logrus"
)

const loginUsageShortHelp = `Log in to a Docker registry.`

var loginUsageLongHelp = loginUsageShortHelp + fmt.Sprintf("\n\nIf no server is specified, the default (%s) is used.", defaultDockerRegistry)

func newLoginCommand() *cobra.Command {

	login := &loginCommand{}

	cmd := &cobra.Command{
		Use:                   "login [OPTIONS] [SERVER]",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Short:                 loginUsageShortHelp,
		Long:                  loginUsageLongHelp,
		Args:                  login.ValidateArgs(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return login.Run(args)
		},
	}

	fs := cmd.Flags()

	fs.StringVarP(&login.user, "user", "u", "", "Username")
	fs.StringVarP(&login.password, "password", "p", "", "Password")
	fs.BoolVar(&login.passwordStdin, "password-stdin", false, "Take the password from stdin")

	return cmd
}

type loginCommand struct {
	user          string
	password      string
	passwordStdin bool

	serverAddress string
}

func (cmd *loginCommand) ValidateArgs() cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if cmd.password != "" {
			logrus.Warnf("WARNING! Using --password via the CLI is insecure. Use --password-stdin.")
			if cmd.passwordStdin {
				return errors.New("--password and --password-stdin are mutually exclusive")
			}
		}

		// Handle when the password is coming over stdin.
		if cmd.passwordStdin {
			if cmd.user == "" {
				return errors.New("must provide --username with --password-stdin")
			}

			// Read from stadin.
			contents, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			cmd.password = strings.TrimSuffix(string(contents), "\n")
			cmd.password = strings.TrimSuffix(cmd.password, "\r")
		}

		if len(args) > 0 {
			cmd.serverAddress = args[0]
		}

		// Set the default registry server address.
		if cmd.serverAddress == "" {
			cmd.serverAddress = defaultDockerRegistry
		}

		return nil
	}
}

func (cmd *loginCommand) Run(args []string) error {

	// Get the auth config.
	dcfg, authConfig, err := configureAuth(cmd.user, cmd.password, cmd.serverAddress)
	if err != nil {
		return err
	}

	// Attempt to login to the registry.
	r, err := registryapi.New(cliconfigtypes2dockerapitypes(authConfig), registryapi.Opt{Debug: debug})
	if err != nil {
		return fmt.Errorf("creating registry client failed: %v", err)
	}
	token, err := r.Token(r.URL)
	if err != nil && err != registryapi.ErrBasicAuth {
		return fmt.Errorf("getting registry token failed: %v", err)
	}

	// Configure the token.
	if token != "" {
		authConfig.Password = ""
		authConfig.IdentityToken = token
	}

	// Save the config value.
	if err := dcfg.GetCredentialsStore(authConfig.ServerAddress).Store(authConfig); err != nil {
		return fmt.Errorf("saving credentials failed: %v", err)
	}

	fmt.Println("Login succeeded.")

	return nil
}

// configureAuth returns an types.AuthConfig from the specified user, password and server.
func configureAuth(flUser, flPassword, serverAddress string) (*configfile.ConfigFile, types.AuthConfig, error) {
	if serverAddress != defaultDockerRegistry {
		serverAddress = registry.ConvertToHostname(serverAddress)
	}

	dcfg, err := config.Load(config.Dir())
	if err != nil {
		return dcfg, types.AuthConfig{}, fmt.Errorf("loading config file failed: %v", err)
	}
	authConfig, err := dcfg.GetAuthConfig(serverAddress)
	if err != nil {
		return dcfg, authConfig, fmt.Errorf("getting auth config for %s failed: %v", serverAddress, err)
	}

	_, isTerminal := term.GetFdInfo(os.Stdin)
	if flPassword == "" && !isTerminal {
		return dcfg, authConfig, errors.New("cannot perform an interactive login from a non TTY device")
	}

	authConfig.Username = strings.TrimSpace(authConfig.Username)

	if flUser = strings.TrimSpace(flUser); flUser == "" {
		if serverAddress == defaultDockerRegistry {
			// if this is a default registry (docker hub), then display the following message.
			fmt.Printf("Login with your Docker ID to push and pull images from Docker Hub. If you don't have a Docker ID, head over to https://hub.docker.com to create one.\n")
		}
		promptWithDefault(os.Stdout, "Username", authConfig.Username)
		flUser = readInput(os.Stdin)
		flUser = strings.TrimSpace(flUser)
		if flUser == "" {
			flUser = authConfig.Username
		}
	}
	if flUser == "" {
		return dcfg, authConfig, fmt.Errorf("username cannot be empty")
	}
	if flPassword == "" {
		oldState, err := term.SaveState(os.Stdin.Fd())
		if err != nil {
			return dcfg, authConfig, err
		}
		fmt.Fprintf(os.Stdout, "Password: ")
		term.DisableEcho(os.Stdin.Fd(), oldState)

		flPassword = readInput(os.Stdin)
		fmt.Fprint(os.Stdout, "\n")

		if err := term.RestoreTerminal(os.Stdin.Fd(), oldState); err != nil {
			return dcfg, authConfig, fmt.Errorf("restoring old terminal failed: %v", err)
		}
		if flPassword == "" {
			return dcfg, authConfig, fmt.Errorf("password is required")
		}
	}

	authConfig.Username = flUser
	authConfig.Password = flPassword
	authConfig.ServerAddress = serverAddress
	authConfig.IdentityToken = ""

	return dcfg, authConfig, nil
}

func readInput(in io.Reader) string {
	reader := bufio.NewReader(in)
	line, _, err := reader.ReadLine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading input failed: %v", err)
		os.Exit(1)
	}

	return string(line)
}

func promptWithDefault(out io.Writer, prompt string, configDefault string) {
	if configDefault == "" {
		fmt.Fprintf(out, "%s: ", prompt)
		return
	}

	fmt.Fprintf(out, "%s (%s): ", prompt, configDefault)
}

func cliconfigtypes2dockerapitypes(in types.AuthConfig) dockerapitypes.AuthConfig {
	return dockerapitypes.AuthConfig{
		Username:      in.Username,
		Password:      in.Password,
		Auth:          in.Auth,
		Email:         in.Email,
		ServerAddress: in.ServerAddress,
		IdentityToken: in.IdentityToken,
		RegistryToken: in.RegistryToken,
	}
}
