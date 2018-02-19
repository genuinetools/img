package fsutils

import (
	"os"
	"strings"

	"github.com/tonistiigi/fsutil"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

type walkerFn func(ctx context.Context, pathC chan<- *currentPath) error

type currentPath struct {
	path string
	f    os.FileInfo
	//	fullPath string
}

// doubleWalkDiff walks both directories to create a diff
func doubleWalkDiff(ctx context.Context, changeFn fsutil.ChangeFunc, a, b walkerFn) (err error) {
	g, ctx := errgroup.WithContext(ctx)

	var (
		c1 = make(chan *currentPath, 128)
		c2 = make(chan *currentPath, 128)

		f1, f2 *currentPath
		rmdir  string
	)
	g.Go(func() error {
		defer close(c1)
		err := a(ctx, c1)
		return err
	})
	g.Go(func() error {
		defer close(c2)
		err := b(ctx, c2)
		return err
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
			case fsutil.ChangeKindAdd:
				if rmdir != "" {
					rmdir = ""
				}
				f = f2.f
				f2 = nil
			case fsutil.ChangeKindDelete:
				// Check if this file is already removed by being
				// under of a removed directory
				if rmdir != "" && strings.HasPrefix(f1.path, rmdir) {
					f1 = nil
					continue
				} else if rmdir == "" && f1.f.IsDir() {
					rmdir = f1.path + string(os.PathSeparator)
				} else if rmdir != "" {
					rmdir = ""
				}
				f1 = nil
			case fsutil.ChangeKindModify:
				same, err := sameFile(f1, f2)
				if err != nil {
					return err
				}
				if f1.f.IsDir() && !f2.f.IsDir() {
					rmdir = f1.path + string(os.PathSeparator)
				} else if rmdir != "" {
					rmdir = ""
				}
				f = f2.f
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

func pathChange(lower, upper *currentPath) (fsutil.ChangeKind, string) {
	if lower == nil {
		if upper == nil {
			panic("cannot compare nil paths")
		}
		return fsutil.ChangeKindAdd, upper.path
	}
	if upper == nil {
		return fsutil.ChangeKindDelete, lower.path
	}

	switch i := fsutil.ComparePath(lower.path, upper.path); {
	case i < 0:
		// File in lower that is not in upper
		return fsutil.ChangeKindDelete, lower.path
	case i > 0:
		// File in upper that is not in lower
		return fsutil.ChangeKindAdd, upper.path
	default:
		return fsutil.ChangeKindModify, upper.path
	}
}

func sameFile(f1, f2 *currentPath) (same bool, retErr error) {
	// If not a directory also check size, modtime, and content
	if !f1.f.IsDir() {
		if f1.f.Size() != f2.f.Size() {
			return false, nil
		}

		t1 := f1.f.ModTime()
		t2 := f2.f.ModTime()
		if t1.UnixNano() != t2.UnixNano() {
			return false, nil
		}
	}

	ls1, ok := f1.f.Sys().(*fsutil.Stat)
	if !ok {
		return false, nil
	}
	ls2, ok := f1.f.Sys().(*fsutil.Stat)
	if !ok {
		return false, nil
	}

	return compareStat(ls1, ls2)
}

// compareStat returns whether the stats are equivalent,
// whether the files are considered the same file, and
// an error
func compareStat(ls1, ls2 *fsutil.Stat) (bool, error) {
	return ls1.Mode == ls2.Mode && ls1.Uid == ls2.Uid && ls1.Gid == ls2.Gid && ls1.Devmajor == ls2.Devmajor && ls1.Devminor == ls2.Devminor && ls1.Linkname == ls2.Linkname, nil
}

func nextPath(ctx context.Context, pathC <-chan *currentPath) (*currentPath, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-pathC:
		return p, nil
	}
}
