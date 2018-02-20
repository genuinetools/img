// Copyright 2016 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/internal/testutil"
)

func TestDefaultNodeGetAttr(t *testing.T) {
	dir := testutil.TempDir()
	defer os.RemoveAll(dir)

	opts := &nodefs.Options{
		// Note: defaultNode.GetAttr() calling file.GetAttr() is only useful if
		// AttrTimeout is zero.
		// See https://github.com/JonathonReinhart/gitlab-fuse/issues/2
		Owner: fuse.CurrentOwner(),
		Debug: testutil.VerboseTest(),
	}

	root := nodefs.NewDefaultNode()
	s, _, err := nodefs.MountRoot(dir, root, opts)
	if err != nil {
		t.Fatalf("MountRoot: %v", err)
	}
	go s.Serve()
	if err := s.WaitMount(); err != nil {
		t.Fatal("WaitMount", err)
	}
	defer s.Unmount()

	// Attach another custom node type
	root.Inode().NewChild("foo", false, &myNode{
		Node:    nodefs.NewDefaultNode(),
		content: []byte("success"),
	})

	filepath := path.Join(dir, "foo")

	// NewDefaultNode() should provide for stat that indicates 0-byte regular file
	fi, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if mode := (fi.Mode() & os.ModeType); mode != 0 {
		// Mode() & ModeType should be zero for regular files
		t.Fatalf("Unexpected mode: %#o", mode)
	}
	if size := fi.Size(); size != 0 {
		t.Fatalf("Unexpected size: %d", size)
	}

	// But when we open the file, we should get the content
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(content) != "success" {
		t.Fatalf("Unexpected content: %v", content)
	}
}

type myNode struct {
	nodefs.Node
	content []byte
}

func (n *myNode) Open(flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	return nodefs.NewDataFile(n.content), fuse.OK
}
