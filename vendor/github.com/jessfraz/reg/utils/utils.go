package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker-ce/components/cli/cli/config"
	"github.com/docker/docker/api/types"
)

const (
	// DefaultDockerRegistry is the default docker registry address.
	DefaultDockerRegistry = "https://registry-1.docker.io"
)

// GetAuthConfig returns the docker registry AuthConfig.
// Optionally takes in the authentication values, otherwise pulls them from the
// docker config file.
func GetAuthConfig(username, password, registry string) (types.AuthConfig, error) {
	if username != "" && password != "" && registry != "" {
		return types.AuthConfig{
			Username:      username,
			Password:      password,
			ServerAddress: registry,
		}, nil
	}

	dcfg, err := config.Load(config.Dir())
	if err != nil {
		return types.AuthConfig{}, fmt.Errorf("Loading config file failed: %v", err)
	}

	// return error early if there are no auths saved
	if !dcfg.ContainsAuth() {
		// If we were passed a registry, just use that.
		if registry != "" {
			return setDefaultRegistry(types.AuthConfig{
				ServerAddress: registry,
			}), nil
		}

		// Otherwise, just use an empty auth config.
		return types.AuthConfig{}, nil
	}

	// if they passed a specific registry, return those creds _if_ they exist
	if registry != "" {
		// try with the user input
		if creds, ok := dcfg.AuthConfigs[registry]; ok {
			return creds, nil
		}
		// add https:// to user input and try again
		// see https://github.com/jessfraz/reg/issues/32
		if !strings.HasPrefix(registry, "https://") && !strings.HasPrefix(registry, "http://") {
			if creds, ok := dcfg.AuthConfigs["https://"+registry]; ok {
				return creds, nil
			}
		}

		// Otherwise just use the registry with no auth.
		return setDefaultRegistry(types.AuthConfig{
			ServerAddress: registry,
		}), nil
	}

	// Just set the auth config as the first registryURL, username and password
	// found in the auth config.
	for _, creds := range dcfg.AuthConfigs {
		return creds, nil
	}

	// Don't use any authentication.
	return types.AuthConfig{}, nil
}

// GetRepoAndRef parses the repo name and reference.
func GetRepoAndRef(arg string) (repo, ref string, err error) {
	if arg == "" {
		return "", "", errors.New("pass the name of the repository")
	}

	var parts []string
	if strings.Contains(arg, "@") {
		parts = strings.Split(arg, "@")
	} else if strings.Contains(arg, ":") {
		parts = strings.Split(arg, ":")
	} else {
		parts = []string{arg}
	}

	repo = parts[0]
	ref = "latest"
	if len(parts) > 1 {
		ref = parts[1]
	}

	return
}

func setDefaultRegistry(auth types.AuthConfig) types.AuthConfig {
	if auth.ServerAddress == "docker.io" {
		auth.ServerAddress = DefaultDockerRegistry
	}

	return auth
}
