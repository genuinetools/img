package fsutils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/moby/buildkit/session/filesync"
	"github.com/moby/buildkit/source"
	"github.com/sirupsen/logrus"
	"github.com/tonistiigi/fsutil"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// The source directory must exist, destination directory should *not* exist, unless we are updating the cache.
func CopyDir(src, dest string, li source.LocalIdentifier, cu filesync.CacheUpdater) error {
	st := time.Now()
	defer func() {
		logrus.Debugf("copydir took: %v", time.Since(st))
	}()

	// Setup the context.
	g, ctx := errgroup.WithContext(context.Background())

	// Get the properties of the src directory.
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.New("CopyDir: src is not a directory")
	}

	if _, err = os.Open(dest); !os.IsNotExist(err) && !DirIsEmpty(dest) {
		logrus.Debugf("destination already exists, using cache: %s", dest)
	}

	// Create the destination directory
	if err = os.MkdirAll(dest, fi.Mode()); err != nil {
		return err
	}

	var cf fsutil.ChangeFunc
	var ch fsutil.ContentHasher
	if cu != nil {
		cu.MarkSupported(true)
		cf = cu.HandleChange
		ch = cu.ContentHasher()
	}

	syncDataFunc := func(ctx context.Context, p string, wc io.WriteCloser) error {
		dfp := filepath.Join(dest, p)
		sfp := filepath.Join(src, p)

		r, err := os.Open(sfp)
		if err != nil {
			return err
		}

		// perform copy
		if _, err := io.Copy(wc, r); err != nil {
			return fmt.Errorf("copy file %s -> %s failed: %v", sfp, dfp, err)
		}

		return wc.Close()
	}

	dw, err := fsutil.NewDiskWriter(ctx, dest, fsutil.DiskWriterOpt{
		SyncDataCb:    syncDataFunc,
		NotifyCb:      cf,
		ContentHasher: ch,
	})
	if err != nil {
		return err
	}

	w := newDynamicWalker()

	g.Go(func() (retErr error) {
		defer func() {
			if retErr != nil {
				logrus.Errorf("fsutils doubleWalkDiff return error: %v", retErr)
			}
		}()

		destWalker := getWalkerFn(dest)
		return doubleWalkDiff(ctx, dw.HandleChange, destWalker, w.fill)
	})

	err = fsutil.Walk(ctx, src, &fsutil.WalkOpt{IncludePatterns: li.IncludePatterns, ExcludePatterns: li.ExcludePatterns}, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return w.update(nil)
		}

		cp := &currentPath{path: path, f: info}
		return w.update(cp)
	})
	if err != nil {
		return err
	}

	// Close the channel or we will wait here for eternity.
	close(w.walkChan)

	return g.Wait()
}

// DirIsEmpty checks if the directory is empty.
func DirIsEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err = f.Readdir(1); err == io.EOF {
		return true
	}

	return false
}
