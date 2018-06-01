// +build noembed

package binutils

import "errors"

// InstallRuncBinary when non-embedded errors out telling the user to install runc.
func InstallRuncBinary() (string, error) {
	return "", errors.New("please install `runc`")
}
