package main

import (
	"fmt"
	"strings"

	"github.com/docker/docker-ce/components/cli/cli/config"
	"github.com/docker/docker/api/types"
)

// getAuthConfig returns the docker registry AuthConfig.
func getAuthConfig(domain string) (types.AuthConfig, error) {
	// Load the docker config locally.
	dcfg, err := config.Load(config.Dir())
	if err != nil {
		return types.AuthConfig{}, fmt.Errorf("Loading config file failed: %v", err)
	}

	// return error early if there are no auths saved
	if !dcfg.ContainsAuth() {
		return types.AuthConfig{}, fmt.Errorf("No docker registry auth was present in %s", config.Dir())
	}

	// try with the user input
	if creds, ok := dcfg.AuthConfigs[domain]; ok {
		return creds, nil
	}
	// add https:// to user input and try again
	// see https://github.com/jessfraz/reg/issues/32
	if !strings.HasPrefix(domain, "https://") && !strings.HasPrefix(domain, "http://") {
		if creds, ok := dcfg.AuthConfigs["https://"+domain]; ok {
			return creds, nil
		}
	}

	// Just try it anyways with no auth.
	return types.AuthConfig{
		ServerAddress: domain,
	}, nil
}
