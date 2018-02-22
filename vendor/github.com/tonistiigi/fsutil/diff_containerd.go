package fsutil

import (
	"os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

// Everything below is copied from containerd/fs. TODO: remove duplication @dmcgowan

// Const redefined because containerd/fs doesn't build on !linux

// ChangeKind is the type of modification that
// a change is making.
type ChangeKind int

const (
	// ChangeKindAdd represents an addition of
	// a file
	ChangeKindAdd ChangeKind = iota

	// ChangeKindModify represents a change to
	// an existing file
	ChangeKindModify

	// ChangeKindDelete represents a delete of
	// a file
	ChangeKindDelete
)

// ChangeFunc is the type of function called for each change
// computed during a directory changes calculation.
type ChangeFunc func(ChangeKind, string, os.FileInfo, error) error

type CurrentPath struct {
	Path     string
	FileInfo os.FileInfo
	//	fullPath string
}

// DoubleWalkDiff walks both directories to create a diff
func DoubleWalkDiff(ctx context.Context, changeFn ChangeFunc, a, b walkerFn) (err error) {
	g, ctx := errgroup.WithContext(ctx)

	var (
		c1 = make(chan *CurrentPath, 128)
		c2 = make(chan *CurrentPath, 128)

		f1, f2 *CurrentPath
		rmdir  string
	)
	g.Go(func() error {
		defer close(c1)
		return a(ctx, c1)
	})
	g.Go(func() error {
		defer close(c2)
		return b(ctx, c2)
	})
	g.Go(func() error {
	loop0:
		for c1 != nil || c2 != nil {
			if f1 == nil && c1 != nil {
				f1, err = nextPath(ctx, c1)
				if err != nil {
					return err
				}
				if f1 == nil {
					c1 = nil
				}
			}

			if f2 == nil && c2 != nil {
				f2, err = nextPath(ctx, c2)
				if err != nil {
					return err
				}
				if f2 == nil {
					c2 = nil
				}
			}
			if f1 == nil && f2 == nil {
				continue
			}

			var f os.FileInfo
			k, p := pathChange(f1, f2)
			switch k {
			case ChangeKindAdd:
				if rmdir != "" {
					rmdir = ""
				}
				f = f2.FileInfo
				f2 = nil
			case ChangeKindDelete:
				// Check if this file is already removed by being
				// under of a removed directory
				if rmdir != "" && strings.HasPrefix(f1.Path, rmdir) {
					f1 = nil
					continue
				} else if rmdir == "" && f1.FileInfo.IsDir() {
					rmdir = f1.Path + string(os.PathSeparator)
				} else if rmdir != "" {
					rmdir = ""
				}
				f1 = nil
			case ChangeKindModify:
				same, err := sameFile(f1, f2)
				if err != nil {
					return err
				}
				if f1.FileInfo.IsDir() && !f2.FileInfo.IsDir() {
					rmdir = f1.Path + string(os.PathSeparator)
				} else if rmdir != "" {
					rmdir = ""
				}
				f = f2.FileInfo
				f1 = nil
				f2 = nil
				if same {
					continue loop0
				}
			}
			if err := changeFn(k, p, f, nil); err != nil {
				return err
			}
		}
		return nil
	})

	return g.Wait()
}

func pathChange(lower, upper *CurrentPath) (ChangeKind, string) {
	if lower == nil {
		if upper == nil {
			panic("cannot compare nil paths")
		}
		return ChangeKindAdd, upper.Path
	}
	if upper == nil {
		return ChangeKindDelete, lower.Path
	}

	switch i := ComparePath(lower.Path, upper.Path); {
	case i < 0:
		// File in lower that is not in upper
		return ChangeKindDelete, lower.Path
	case i > 0:
		// File in upper that is not in lower
		return ChangeKindAdd, upper.Path
	default:
		return ChangeKindModify, upper.Path
	}
}

func sameFile(f1, f2 *CurrentPath) (same bool, retErr error) {
	// If not a directory also check size, modtime, and content
	if !f1.FileInfo.IsDir() {
		if f1.FileInfo.Size() != f2.FileInfo.Size() {
			return false, nil
		}

		t1 := f1.FileInfo.ModTime()
		t2 := f2.FileInfo.ModTime()
		if t1.UnixNano() != t2.UnixNano() {
			return false, nil
		}
	}

	ls1, ok := f1.FileInfo.Sys().(*Stat)
	if !ok {
		return false, nil
	}
	ls2, ok := f1.FileInfo.Sys().(*Stat)
	if !ok {
		return false, nil
	}

	return compareStat(ls1, ls2)
}

// compareStat returns whether the stats are equivalent,
// whether the files are considered the same file, and
// an error
func compareStat(ls1, ls2 *Stat) (bool, error) {
	return ls1.Mode == ls2.Mode && ls1.Uid == ls2.Uid && ls1.Gid == ls2.Gid && ls1.Devmajor == ls2.Devmajor && ls1.Devminor == ls2.Devminor && ls1.Linkname == ls2.Linkname, nil
}

func nextPath(ctx context.Context, pathC <-chan *CurrentPath) (*CurrentPath, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-pathC:
		return p, nil
	}
}
