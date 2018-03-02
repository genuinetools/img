package registry

import (
	"bytes"
	"fmt"
	"testing"

	"golang.org/x/net/context"

	"github.com/docker/cli/internal/test"
	"github.com/docker/docker/api/types"
	registrytypes "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

const userErr = "userunknownError"
const testAuthErrMsg = "UNKNOWN_ERR"

var testAuthErrors = map[string]error{
	userErr: fmt.Errorf(testAuthErrMsg),
}

var expiredPassword = "I_M_EXPIRED"

type fakeClient struct {
	client.Client
}

// nolint: unparam
func (c fakeClient) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registrytypes.AuthenticateOKBody, error) {
	if auth.Password == expiredPassword {
		return registrytypes.AuthenticateOKBody{}, fmt.Errorf("Invalid Username or Password")
	}
	err := testAuthErrors[auth.Username]
	return registrytypes.AuthenticateOKBody{}, err
}

func TestLoginWithCredStoreCreds(t *testing.T) {
	testCases := []struct {
		inputAuthConfig types.AuthConfig
		expectedMsg     string
		expectedErr     string
	}{
		{
			inputAuthConfig: types.AuthConfig{},
			expectedMsg:     "Authenticating with existing credentials...\n",
		},
		{
			inputAuthConfig: types.AuthConfig{
				Username: userErr,
			},
			expectedMsg: "Authenticating with existing credentials...\n",
			expectedErr: fmt.Sprintf("Login did not succeed, error: %s\n", testAuthErrMsg),
		},
		// can't easily test the 401 case because client.IsErrUnauthorized(err) involving
		// creating an error of a private type
	}
	ctx := context.Background()
	for _, tc := range testCases {
		cli := (*test.FakeCli)(test.NewFakeCli(&fakeClient{}))
		errBuf := new(bytes.Buffer)
		cli.SetErr(errBuf)
		loginWithCredStoreCreds(ctx, cli, &tc.inputAuthConfig)
		outputString := cli.OutBuffer().String()
		assert.Equal(t, tc.expectedMsg, outputString)
		errorString := errBuf.String()
		assert.Equal(t, tc.expectedErr, errorString)
	}
}

func TestRunLogin(t *testing.T) {
	const storedServerAddress = "reg1"
	const validUsername = "u1"
	const validPassword = "p1"
	const validPassword2 = "p2"

	validAuthConfig := types.AuthConfig{
		ServerAddress: storedServerAddress,
		Username:      validUsername,
		Password:      validPassword,
	}
	expiredAuthConfig := types.AuthConfig{
		ServerAddress: storedServerAddress,
		Username:      validUsername,
		Password:      expiredPassword,
	}
	testCases := []struct {
		inputLoginOption  loginOptions
		inputStoredCred   *types.AuthConfig
		expectedErr       string
		expectedSavedCred types.AuthConfig
	}{
		{
			inputLoginOption: loginOptions{
				serverAddress: storedServerAddress,
			},
			inputStoredCred:   &validAuthConfig,
			expectedErr:       "",
			expectedSavedCred: validAuthConfig,
		},
		{
			inputLoginOption: loginOptions{
				serverAddress: storedServerAddress,
			},
			inputStoredCred: &expiredAuthConfig,
			expectedErr:     "Error: Cannot perform an interactive login from a non TTY device",
		},
		{
			inputLoginOption: loginOptions{
				serverAddress: storedServerAddress,
				user:          validUsername,
				password:      validPassword2,
			},
			inputStoredCred: &validAuthConfig,
			expectedErr:     "",
			expectedSavedCred: types.AuthConfig{
				ServerAddress: storedServerAddress,
				Username:      validUsername,
				Password:      validPassword2,
			},
		},
		{
			inputLoginOption: loginOptions{
				serverAddress: storedServerAddress,
				user:          userErr,
				password:      validPassword,
			},
			inputStoredCred: &validAuthConfig,
			expectedErr:     testAuthErrMsg,
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{})
		errBuf := new(bytes.Buffer)
		cli.SetErr(errBuf)
		if tc.inputStoredCred != nil {
			cred := *tc.inputStoredCred
			cli.ConfigFile().GetCredentialsStore(cred.ServerAddress).Store(cred)
		}
		loginErr := runLogin(cli, tc.inputLoginOption)
		if tc.expectedErr != "" {
			assert.Equal(t, tc.expectedErr, loginErr.Error())
		} else {
			assert.Nil(t, loginErr)
			savedCred, credStoreErr := cli.ConfigFile().GetCredentialsStore(tc.inputStoredCred.ServerAddress).Get(tc.inputStoredCred.ServerAddress)
			assert.Nil(t, credStoreErr)
			assert.Equal(t, tc.expectedSavedCred, savedCred)
		}
	}
}
