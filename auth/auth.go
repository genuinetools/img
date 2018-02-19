package auth

import (
	"io/ioutil"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/moby/buildkit/session/auth"
)

func NewDockerAuthProvider() *authProvider {
	return &authProvider{
		config: config.LoadDefaultConfigFile(ioutil.Discard),
	}
}

type authProvider struct {
	config *configfile.ConfigFile
}

func (ap *authProvider) Credentials(host string) (*auth.CredentialsResponse, error) {
	if host == "registry-1.docker.io" {
		host = "https://index.docker.io/v1/"
	}
	ac, err := ap.config.GetAuthConfig(host)
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
