package auth

import (
	"io/ioutil"

	"github.com/docker/cli/cli/config"
	"github.com/moby/buildkit/session/auth"
)

// DockerAuthCredentials returns the authentication credentials for the registry host passed..
func DockerAuthCredentials(host string) (*auth.CredentialsResponse, error) {
	if host == "registry-1.docker.io" {
		host = "https://index.docker.io/v1/"
	}

	c := config.LoadDefaultConfigFile(ioutil.Discard)

	ac, err := c.GetAuthConfig(host)
	if err != nil {
		return nil, err
	}

	res := &auth.CredentialsResponse{}
	if ac.IdentityToken != "" {
		res.Secret = ac.IdentityToken
	} else {
		res.Username = ac.Username
		res.Secret = ac.Password
	}

	return res, nil
}
