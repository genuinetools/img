// Copyright 2016 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package test

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
	"testing"
)

// See https://github.com/hanwen/go-fuse/issues/170
func disabledTestFlock(t *testing.T) {
	cmd, err := exec.LookPath("flock")
	if err != nil {
		t.Skip("flock command not found.")
	}
	tc := NewTestCase(t)
	defer tc.Cleanup()

	contents := []byte{1, 2, 3}
	tc.WriteFile(tc.origFile, []byte(contents), 0700)

	f, err := os.OpenFile(tc.mountFile, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("OpenFile(%q): %v", tc.mountFile, err)
	}
	defer f.Close()

	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Errorf("Flock returned: %v", err)
		return
	}

	if out, err := runExternalFlock(cmd, tc.mountFile); !bytes.Contains(out, []byte("failed to get lock")) {
		t.Errorf("runExternalFlock(%q): %s (%v)", tc.mountFile, out, err)
	}
}

func runExternalFlock(flockPath, fname string) ([]byte, error) {
	f, err := os.OpenFile(fname, os.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cmd := exec.Command(flockPath, "--verbose", "--exclusive", "--nonblock", "3")
	cmd.Env = append(cmd.Env, "LC_ALL=C") // in case the user's shell language is different
	cmd.ExtraFiles = []*os.File{f}
	return cmd.CombinedOutput()
}
