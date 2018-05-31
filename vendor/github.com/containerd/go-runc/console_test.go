/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package runc

import (
	"errors"
	"os"
	"testing"
)

func TestTempConsole(t *testing.T) {
	c, path := testSocketWithCorrectStickyBitMode(t, 0)
	ensureSocketCleanup(t, c, path)
}

func TestTempConsoleWithXdgRuntimeDir(t *testing.T) {
	tmpDir := "/tmp/foo"
	if err := os.Setenv("XDG_RUNTIME_DIR", tmpDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}

	c, path := testSocketWithCorrectStickyBitMode(t, os.ModeSticky)
	ensureSocketCleanup(t, c, path)

	if err := os.RemoveAll(tmpDir); err != nil {
		t.Fatal(err)
	}
}

func testSocketWithCorrectStickyBitMode(t *testing.T, expectedMode os.FileMode) (*Socket, string) {
	c, err := NewTempConsoleSocket()
	if err != nil {
		t.Fatal(err)
	}
	path := c.Path()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if (info.Mode() & os.ModeSticky) != expectedMode {
		t.Fatal(errors.New("socket has incorrect mode"))
	}
	return c, path
}

func ensureSocketCleanup(t *testing.T, c *Socket, path string) {
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("path still exists")
	}
}
