// +build !seccomp

package runc

import (
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// DefaultSeccompProfile defines the whitelist for the default seccomp profile.
var DefaultSeccompProfile = &specs.LinuxSeccomp{}
