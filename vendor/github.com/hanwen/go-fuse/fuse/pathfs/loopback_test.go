// Copyright 2018 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathfs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/internal/testutil"
)

// Check that loopbackFileSystem.Utimens() works as expected
func TestLoopbackFileSystemUtimens(t *testing.T) {
	fs := NewLoopbackFileSystem(os.TempDir())
	f, err := ioutil.TempFile("", "TestLoopbackFileSystemUtimens")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	name := filepath.Base(path)
	f.Close()
	defer syscall.Unlink(path)

	utimensFn := func(atime *time.Time, mtime *time.Time) fuse.Status {
		return fs.Utimens(name, atime, mtime, nil)
	}
	testutil.TestLoopbackUtimens(t, path, utimensFn)
}
