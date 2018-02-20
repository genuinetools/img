// Copyright 2016 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"io/ioutil"
	"os"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/hanwen/go-fuse/fuse"
)

func TestTouch(t *testing.T) {
	ts := NewTestCase(t)
	defer ts.Cleanup()

	contents := []byte{1, 2, 3}
	err := ioutil.WriteFile(ts.origFile, []byte(contents), 0700)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	err = os.Chtimes(ts.mountFile, time.Unix(42, 0), time.Unix(43, 0))
	if err != nil {
		t.Fatalf("Chtimes failed: %v", err)
	}

	var stat syscall.Stat_t
	err = syscall.Lstat(ts.mountFile, &stat)
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}
	if stat.Atim.Sec != 42 {
		t.Errorf("Got atime.sec %d, want 42. Stat_t was %#v", stat.Atim.Sec, stat)
	}
	if stat.Mtim.Sec != 43 {
		t.Errorf("Got mtime.sec %d, want 43. Stat_t was %#v", stat.Mtim.Sec, stat)
	}
}

func TestNegativeTime(t *testing.T) {
	ts := NewTestCase(t)
	defer ts.Cleanup()

	_, err := os.Create(ts.origFile)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var stat syscall.Stat_t

	// set negative nanosecond will occur errors on UtimesNano as invalid argument
	ut := time.Date(1960, time.January, 10, 23, 0, 0, 0, time.UTC)
	tim := []syscall.Timespec{
		syscall.NsecToTimespec(ut.UnixNano()),
		syscall.NsecToTimespec(ut.UnixNano()),
	}
	err = syscall.UtimesNano(ts.mountFile, tim)
	if err != nil {
		t.Fatalf("UtimesNano failed: %v", err)
	}
	err = syscall.Lstat(ts.mountFile, &stat)
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}

	if stat.Atim.Sec >= 0 || stat.Mtim.Sec >= 0 {
		t.Errorf("Got wrong timestamps %v", stat)
	}
}

// Setting nanoseconds should work for dates after 1970
func TestUtimesNano(t *testing.T) {
	tc := NewTestCase(t)
	defer tc.Cleanup()

	path := tc.mountFile
	err := ioutil.WriteFile(path, []byte("xyz"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	ts := make([]syscall.Timespec, 2)
	// atime
	ts[0].Sec = 1
	ts[0].Nsec = 2
	// mtime
	ts[1].Sec = 3
	ts[1].Nsec = 4
	err = syscall.UtimesNano(path, ts)
	if err != nil {
		t.Fatal(err)
	}

	var st syscall.Stat_t
	err = syscall.Stat(path, &st)
	if err != nil {
		t.Fatal(err)
	}
	if st.Atim != ts[0] {
		t.Errorf("Wrong atime: %v, want: %v", st.Atim, ts[0])
	}
	if st.Mtim != ts[1] {
		t.Errorf("Wrong mtime: %v, want: %v", st.Mtim, ts[1])
	}
}

func clearStatfs(s *syscall.Statfs_t) {
	empty := syscall.Statfs_t{}
	s.Type = 0
	s.Fsid = empty.Fsid
	s.Spare = empty.Spare
	// TODO - figure out what this is for.
	s.Flags = 0
}

func TestFallocate(t *testing.T) {
	ts := NewTestCase(t)
	defer ts.Cleanup()
	if ts.state.KernelSettings().Minor < 19 {
		t.Log("FUSE does not support Fallocate.")
		return
	}

	rwFile, err := os.OpenFile(ts.mnt+"/file", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer rwFile.Close()
	err = syscall.Fallocate(int(rwFile.Fd()), 0, 1024, 4096)
	if err != nil {
		t.Fatalf("FUSE Fallocate failed: %v", err)
	}
	fi, err := os.Lstat(ts.orig + "/file")
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}
	if fi.Size() < (1024 + 4096) {
		t.Fatalf("fallocate should have changed file size. Got %d bytes",
			fi.Size())
	}
}

// Check that "." and ".." exists. unix.Getdents is linux specific.
func TestSpecialEntries(t *testing.T) {
	tc := NewTestCase(t)
	defer tc.Cleanup()

	d, err := os.Open(tc.mnt)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer d.Close()
	buf := make([]byte, 100)
	n, err := unix.Getdents(int(d.Fd()), buf)
	if n == 0 {
		t.Errorf("directory is empty, entries '.' and '..' are missing")
	}
}

// Check that readdir(3) returns valid inode numbers in the directory entries
func TestReaddirInodes(t *testing.T) {
	tc := NewTestCase(t)
	defer tc.Cleanup()
	// create "hello.txt"
	filename := "hello.txt"
	path := tc.orig + "/" + filename
	err := ioutil.WriteFile(path, []byte("xyz"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	// open mountpoint dir
	d, err := os.Open(tc.mnt)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer d.Close()
	buf := make([]byte, 100)
	// readdir(3) use getdents64(2) internally which returns linux_dirent64
	// structures. We don't have readdir(3) so we call getdents64(2) directly.
	n, err := unix.Getdents(int(d.Fd()), buf)
	if n == 0 {
		t.Error("empty directory - we need at least one file")
	}
	buf = buf[:n]
	entries := parseDirents(buf)
	t.Logf("parseDirents returned %d entries", len(entries))
	// Find "hello.txt" and check inode number.
	for _, entry := range entries {
		if entry.name != filename {
			continue
		}
		if entry.ino != 0 && entry.ino != fuse.FUSE_UNKNOWN_INO {
			// Inode number looks good, we are done.
			return
		}
		t.Errorf("got invalid inode number: %d = 0x%x", entry.ino, entry.ino)
	}
	t.Errorf("%q not found in directory listing", filename)
}
