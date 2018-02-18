package fsutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestWalkerSimple(t *testing.T) {
	d, err := tmpDir(changeStream([]string{
		"ADD foo file",
		"ADD foo2 file",
	}))
	assert.NoError(t, err)
	defer os.RemoveAll(d)
	b := &bytes.Buffer{}
	err = Walk(context.Background(), d, nil, bufWalk(b))
	assert.NoError(t, err)

	assert.Equal(t, string(b.Bytes()), `file foo
file foo2
`)

}

func TestWalkerInclude(t *testing.T) {
	d, err := tmpDir(changeStream([]string{
		"ADD bar dir",
		"ADD bar/foo file",
		"ADD foo2 file",
	}))
	assert.NoError(t, err)
	defer os.RemoveAll(d)
	b := &bytes.Buffer{}
	err = Walk(context.Background(), d, &WalkOpt{
		IncludePatterns: []string{"bar", "bar/foo"},
	}, bufWalk(b))
	assert.NoError(t, err)

	assert.Equal(t, `dir bar
file bar/foo
`, string(b.Bytes()))

}

func TestWalkerExclude(t *testing.T) {
	d, err := tmpDir(changeStream([]string{
		"ADD bar file",
		"ADD foo dir",
		"ADD foo2 file",
		"ADD foo/bar2 file",
	}))
	assert.NoError(t, err)
	defer os.RemoveAll(d)
	b := &bytes.Buffer{}
	err = Walk(context.Background(), d, &WalkOpt{
		ExcludePatterns: []string{"foo*", "!foo/bar2"},
	}, bufWalk(b))
	assert.NoError(t, err)

	assert.Equal(t, `file bar
dir foo
file foo/bar2
`, string(b.Bytes()))

}

func TestWalkerMap(t *testing.T) {
	d, err := tmpDir(changeStream([]string{
		"ADD bar file",
		"ADD foo dir",
		"ADD foo2 file",
		"ADD foo/bar2 file",
	}))
	assert.NoError(t, err)
	defer os.RemoveAll(d)
	b := &bytes.Buffer{}
	err = Walk(context.Background(), d, &WalkOpt{
		Map: func(s *Stat) bool {
			if strings.HasPrefix(s.Path, "foo") {
				s.Path = "_" + s.Path
				return true
			}
			return false
		},
	}, bufWalk(b))
	assert.NoError(t, err)

	assert.Equal(t, `dir _foo
file _foo/bar2
file _foo2
`, string(b.Bytes()))
}

func bufWalk(buf *bytes.Buffer) filepath.WalkFunc {
	return func(path string, fi os.FileInfo, err error) error {
		stat, ok := fi.Sys().(*Stat)
		if !ok {
			return errors.Errorf("invalid symlink %s", path)
		}
		t := "file"
		if fi.IsDir() {
			t = "dir"
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			t = "symlink:" + stat.Linkname
		}
		fmt.Fprintf(buf, "%s %s", t, path)
		if fi.Mode()&os.ModeSymlink == 0 && stat.Linkname != "" {
			fmt.Fprintf(buf, " >%s", stat.Linkname)
		}
		fmt.Fprintln(buf)
		return nil
	}
}

func tmpDir(inp []*change) (dir string, retErr error) {
	tmpdir, err := ioutil.TempDir("", "diff")
	if err != nil {
		return "", err
	}
	defer func() {
		if retErr != nil {
			os.RemoveAll(tmpdir)
		}
	}()
	for _, c := range inp {
		if c.kind == ChangeKindAdd {
			p := filepath.Join(tmpdir, c.path)
			stat, ok := c.fi.Sys().(*Stat)
			if !ok {
				return "", errors.Errorf("invalid symlink change %s", p)
			}
			if c.fi.IsDir() {
				if err := os.Mkdir(p, 0700); err != nil {
					return "", err
				}
			} else if c.fi.Mode()&os.ModeSymlink != 0 {
				if err := os.Symlink(stat.Linkname, p); err != nil {
					return "", err
				}
			} else if len(stat.Linkname) > 0 {
				if err := os.Link(filepath.Join(tmpdir, stat.Linkname), p); err != nil {
					return "", err
				}
			} else {
				f, err := os.Create(p)
				if err != nil {
					return "", err
				}
				if len(c.data) > 0 {
					if _, err := f.Write([]byte(c.data)); err != nil {
						return "", err
					}
				}
				f.Close()
			}
		}
	}
	return tmpdir, nil
}
