package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
	registryapi "github.com/genuinetools/reg/registry"
	"github.com/sirupsen/logrus"
)

const loginShortHelp = `Log in to a Docker registry.`

var loginLongHelp = loginShortHelp + fmt.Sprintf("\nIf no server is specified, the default (%s) is used.", defaultDockerRegistry)

func (cmd *loginCommand) Name() string       { return "login" }
func (cmd *loginCommand) Args() string       { return "[OPTIONS] [SERVER]" }
func (cmd *loginCommand) ShortHelp() string  { return loginShortHelp }
func (cmd *loginCommand) LongHelp() string   { return loginLongHelp }
func (cmd *loginCommand) Hidden() bool       { return false }
func (cmd *loginCommand) DoReexec() bool     { return false }
func (cmd *loginCommand) RequiresRunc() bool { return false }

func (cmd *loginCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.user, "u", "", "Username")
	fs.StringVar(&cmd.password, "p", "", "Password")
	fs.BoolVar(&cmd.passwordStdin, "password-stdin", false, "Take the password from stdin")
}

type loginCommand struct {
	user          string
	password      string
	passwordStdin bool

	serverAddress string
}

func (cmd *loginCommand) Run(ctx context.Context, args []string) error {
	if cmd.password != "" {
		logrus.Warnf("WARNING! Using --password via the CLI is insecure. Use --password-stdin.")
		if cmd.passwordStdin {
			return errors.New("--password and --password-stdin are mutually exclusive")
		}
	}

	// Handle when the password is coming over stdin.
	if cmd.passwordStdin {
		if cmd.user == "" {
			return errors.New("Must provide --username with --password-stdin")
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

	// Get the auth config.
	dcfg, authConfig, err := configureAuth(cmd.user, cmd.password, cmd.serverAddress)
	if err != nil {
		return err
	}

	// Attempt to login to the registry.
	r, err := registryapi.New(authConfig, registryapi.Opt{Debug: debug})
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
		return dcfg, authConfig, fmt.Errorf("Username cannot be empty")
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
			return dcfg, authConfig, fmt.Errorf("Password is required")
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
