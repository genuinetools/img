package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker/api/types"
	registrytypes "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
	"github.com/sirupsen/logrus"
)

const loginShortHelp = `Log in to a Docker registry.`

var loginLongHelp = loginShortHelp + fmt.Sprintf("\nIf no server is specified, the default (%s) is used.", defaultDockerRegistry)

func (cmd *loginCommand) Name() string      { return "login" }
func (cmd *loginCommand) Args() string      { return "[OPTIONS] [SERVER]" }
func (cmd *loginCommand) ShortHelp() string { return loginShortHelp }
func (cmd *loginCommand) LongHelp() string  { return loginLongHelp }
func (cmd *loginCommand) Hidden() bool      { return false }

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

func (cmd *loginCommand) Run(args []string) error {
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
	resp, err := registryLogin(authConfig)
	if err != nil {
		return err
	}

	// Configure the token.
	if resp.IdentityToken != "" {
		authConfig.Password = ""
		authConfig.IdentityToken = resp.IdentityToken
	}

	// Save the config value.
	if err := dcfg.GetCredentialsStore(authConfig.ServerAddress).Store(authConfig); err != nil {
		return fmt.Errorf("saving credentials failed: %v", err)
	}

	if resp.Status != "" {
		logrus.Infof("Registry login status: %s", resp.Status)
	}
	return nil
}

func registryLogin(auth types.AuthConfig) (authResponse registrytypes.AuthenticateOKBody, err error) {
	// Encode the body.
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(auth); err != nil {
		return authResponse, err
	}

	// Create the request.
	client := http.DefaultClient
	url := strings.TrimSuffix(auth.ServerAddress, "/") + "/auth"
	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		return authResponse, fmt.Errorf("creating POST request to %s failed: %v", url, err)
	}

	// Add the headers.
	req.Header.Add("Content-Type", "application/json")

	// Do the request.
	resp, err := client.Do(req)
	if err != nil {
		return authResponse, fmt.Errorf("doing POST request to %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&authResponse); err != nil {
		return authResponse, err
	}

	return authResponse, nil
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
		return dcfg, authConfig, fmt.Errorf("getting auth config for %s failed: %v", err)
	}

	fd, isTerminal := term.GetFdInfo(os.Stdin)
	if flPassword == "" && !isTerminal {
		return dcfg, authConfig, errors.New("cannot perform an interactive login from a non TTY device")
	}

	authConfig.Username = strings.TrimSpace(authConfig.Username)

	if flUser = strings.TrimSpace(flUser); flUser == "" {
		if serverAddress != defaultDockerRegistry {
			// if this is a default registry (docker hub), then display the following message.
			fmt.Printf("Login with your Docker ID to push and pull images from Docker Hub. If you don't have a Docker ID, head over to https://hub.docker.com to create one.")
		}
		promptWithDefault(os.Stdout, "Username", authConfig.Username)
		flUser = readInput(os.Stdin, os.Stdout)
		flUser = strings.TrimSpace(flUser)
		if flUser == "" {
			flUser = authConfig.Username
		}
	}
	if flUser == "" {
		return dcfg, authConfig, fmt.Errorf("Username cannot be empty")
	}
	if flPassword == "" {
		oldState, err := term.SaveState(fd)
		if err != nil {
			return dcfg, authConfig, err
		}
		fmt.Fprintf(os.Stdout, "Password: ")
		term.DisableEcho(fd, oldState)

		flPassword = readInput(os.Stdin, os.Stdout)
		fmt.Fprint(os.Stdout, "\n")

		term.RestoreTerminal(fd, oldState)
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

func readInput(in io.Reader, out io.Writer) string {
	reader := bufio.NewReader(in)
	line, _, err := reader.ReadLine()
	if err != nil {
		fmt.Fprintln(out, err.Error())
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
