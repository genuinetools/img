package repoutils

import (
	"fmt"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker-ce/components/cli/cli/config"
	"github.com/docker/docker/api/types"
)

const (
	// DefaultDockerRegistry is the default docker registry address.
	DefaultDockerRegistry = "https://registry-1.docker.io"

	latestTagSuffix = ":latest"
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

	authConfigs, err := dcfg.GetAllCredentials()
	if err != nil {
		return types.AuthConfig{}, fmt.Errorf("Getting credentials failed: %v", err)
	}

	// if they passed a specific registry, return those creds _if_ they exist
	if registry != "" {
		// try with the user input
		if creds, ok := authConfigs[registry]; ok {
			return creds, nil
		}

		// remove https:// from user input and try again
		if strings.HasPrefix(registry, "https://") {
			if creds, ok := authConfigs[strings.TrimPrefix(registry, "https://")]; ok {
				return creds, nil
			}
		}

		// remove http:// from user input and try again
		if strings.HasPrefix(registry, "http://") {
			if creds, ok := authConfigs[strings.TrimPrefix(registry, "http://")]; ok {
				return creds, nil
			}
		}

		// add https:// to user input and try again
		// see https://github.com/genuinetools/reg/issues/32
		if !strings.HasPrefix(registry, "https://") && !strings.HasPrefix(registry, "http://") {
			if creds, ok := authConfigs["https://"+registry]; ok {
				return creds, nil
			}
		}

		fmt.Printf("Using registry %q with no authentication\n", registry)

		// Otherwise just use the registry with no auth.
		return setDefaultRegistry(types.AuthConfig{
			ServerAddress: registry,
		}), nil
	}

	// Just set the auth config as the first registryURL, username and password
	// found in the auth config.
	for _, creds := range authConfigs {
		fmt.Printf("No registry passed. Using registry %q\n", creds.ServerAddress)
		return creds, nil
	}

	// Don't use any authentication.
	// We should never get here.
	fmt.Println("Not using any authentication")
	return types.AuthConfig{}, nil
}

// GetRepoAndRef parses the repo name and reference.
func GetRepoAndRef(image string) (repo, ref string, err error) {
	if image == "" {
		return "", "", reference.ErrNameEmpty
	}

	image = addLatestTagSuffix(image)

	var parts []string
	if strings.Contains(image, "@") {
		parts = strings.Split(image, "@")
	} else if strings.Contains(image, ":") {
		parts = strings.Split(image, ":")
	}

	repo = parts[0]
	if len(parts) > 1 {
		ref = parts[1]
	}

	return
}

// addLatestTagSuffix adds :latest to the image if it does not have a tag
func addLatestTagSuffix(image string) string {
	if !strings.Contains(image, ":") {
		return image + latestTagSuffix
	}
	return image
}

func setDefaultRegistry(auth types.AuthConfig) types.AuthConfig {
	if auth.ServerAddress == "docker.io" {
		auth.ServerAddress = DefaultDockerRegistry
	}

	return auth
}
