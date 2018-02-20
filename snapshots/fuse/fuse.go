package fuse

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/snapshots"
	"github.com/containerd/containerd/snapshots/storage"
	"github.com/containerd/continuity/fs"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/hanwen/go-fuse/unionfs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func init() {
	plugin.Register(&plugin.Registration{
		Type: plugin.SnapshotPlugin,
		ID:   "fuse",
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			ic.Meta.Platforms = append(ic.Meta.Platforms, platforms.DefaultSpec())
			ic.Meta.Exports = map[string]string{"root": ic.Root}
			return NewSnapshotter(ic.Root)
		},
	})
}

type snapshotter struct {
	device string // device of the root
	root   string // root provides paths for external storage
	ms     *storage.MetaStore
	fs     *pathfs.PathNodeFs
	server *fuse.Server
}

type snapshotFs struct {
	pathfs.FileSystem
}

// NewSnapshotter returns a Snapshotter using fuse, which copies layers on the underlying
// file system. A metadata file is stored under the root.
// Root needs to be a mount point of fuse.
func NewSnapshotter(root string) (snapshots.Snapshotter, error) {
	// Check if the directory exists.
	if err := newSnapshotDir(root); err != nil {
		return nil, err
	}
	var (
		ro = filepath.Join(os.TempDir(), "img", "fuse-ro")
		rw = filepath.Join(os.TempDir(), "img", "fuse-rw")
	)

	for _, path := range []string{
		ro,
		rw,
	} {
		if err := os.MkdirAll(path, 0755); err != nil && !os.IsExist(err) {
			return nil, err
		}
	}

	// TODO: turn off debug output
	ufsOptions := unionfs.UnionFsOptions{
		DeletionCacheTTL: time.Duration(5 * time.Second),
		BranchCacheTTL:   time.Duration(5 * time.Second),
		DeletionDirName:  "GOUNIONFS_DELETIONS",
	}
	ufs, err := unionfs.NewUnionFsFromRoots([]string{rw, ro}, &ufsOptions, true)
	if err != nil {
		return nil, fmt.Errorf("Create FUSE UnionFS failed: %v", err)
	}
	nodeFs := pathfs.NewPathNodeFs(ufs, &pathfs.PathNodeFsOptions{ClientInodes: true})
	mOpts := nodefs.Options{
		EntryTimeout:    time.Duration(1 * time.Second),
		AttrTimeout:     time.Duration(1 * time.Second),
		NegativeTimeout: time.Duration(1 * time.Second),
		PortableInodes:  false, // Use sequential 32-bit inode numbers.
		Debug:           true,
	}
	state, _, err := nodefs.MountRoot(root, nodeFs.Root(), &mOpts)
	if err != nil {
		return nil, fmt.Errorf("FUSE mount to root %s failed: %v", root, err)
	}

	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			state.Unmount()
			logrus.Infof("Received %s, Unmounting FUSE Server", sig.String())
			os.Exit(1)
		}
	}()

	go func() {
		logrus.Infof("Starting FUSE server...")
		state.Serve()
	}()

	mnt, err := mount.Lookup(root)
	if err != nil {
		return nil, err
	}
	if mnt.FSType != "fuse.pathfs.pathInode" {
		return nil, fmt.Errorf("path %s must be a fuse filesystem to be used with the fuse snapshotter, got %s", root, mnt.FSType)
	}

	if err := os.Mkdir(filepath.Join(root, "snapshots"), 0755); err != nil && !os.IsExist(err) {
		return nil, err
	}

	ms, err := storage.NewMetaStore(filepath.Join(root, "metadata.db"))
	if err != nil {
		return nil, err
	}

	return &snapshotter{
		device: mnt.Source,
		root:   root,
		ms:     ms,
		fs:     nodeFs,
		server: state,
	}, nil
}

// Stat returns the info for an active or committed snapshot by name or
// key.
//
// Should be used for parent resolution, existence checks and to discern
// the kind of snapshot.
func (o *snapshotter) Stat(ctx context.Context, key string) (snapshots.Info, error) {
	ctx, t, err := o.ms.TransactionContext(ctx, false)
	if err != nil {
		return snapshots.Info{}, err
	}
	defer t.Rollback()
	_, info, _, err := storage.GetInfo(ctx, key)
	if err != nil {
		return snapshots.Info{}, err
	}

	return info, nil
}

func (o *snapshotter) Update(ctx context.Context, info snapshots.Info, fieldpaths ...string) (snapshots.Info, error) {
	ctx, t, err := o.ms.TransactionContext(ctx, true)
	if err != nil {
		return snapshots.Info{}, err
	}

	info, err = storage.UpdateInfo(ctx, info, fieldpaths...)
	if err != nil {
		t.Rollback()
		return snapshots.Info{}, err
	}

	if err := t.Commit(); err != nil {
		return snapshots.Info{}, err
	}

	return info, nil
}

func (o *snapshotter) Usage(ctx context.Context, key string) (snapshots.Usage, error) {
	ctx, t, err := o.ms.TransactionContext(ctx, false)
	if err != nil {
		return snapshots.Usage{}, err
	}
	defer t.Rollback()

	id, info, usage, err := storage.GetInfo(ctx, key)
	if err != nil {
		return snapshots.Usage{}, err
	}

	if info.Kind == snapshots.KindActive {
		du, err := fs.DiskUsage(o.getSnapshotDir(id))
		if err != nil {
			return snapshots.Usage{}, err
		}
		usage = snapshots.Usage(du)
	}

	return usage, nil
}

func (o *snapshotter) Prepare(ctx context.Context, key, parent string, opts ...snapshots.Opt) ([]mount.Mount, error) {
	return o.createSnapshot(ctx, snapshots.KindActive, key, parent, opts)
}

func (o *snapshotter) View(ctx context.Context, key, parent string, opts ...snapshots.Opt) ([]mount.Mount, error) {
	return o.createSnapshot(ctx, snapshots.KindView, key, parent, opts)
}

// Mounts returns the mounts for the transaction identified by key. Can be
// called on an read-write or readonly transaction.
//
// This can be used to recover mounts after calling View or Prepare.
func (o *snapshotter) Mounts(ctx context.Context, key string) ([]mount.Mount, error) {
	ctx, t, err := o.ms.TransactionContext(ctx, false)
	if err != nil {
		return nil, err
	}
	s, err := storage.GetSnapshot(ctx, key)
	t.Rollback()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get snapshot mount")
	}
	return o.mounts(s), nil
}

func (o *snapshotter) Commit(ctx context.Context, name, key string, opts ...snapshots.Opt) error {
	ctx, t, err := o.ms.TransactionContext(ctx, true)
	if err != nil {
		return err
	}

	id, _, _, err := storage.GetInfo(ctx, key)
	if err != nil {
		return err
	}

	usage, err := fs.DiskUsage(o.getSnapshotDir(id))
	if err != nil {
		return err
	}

	if _, err := storage.CommitActive(ctx, key, name, snapshots.Usage(usage), opts...); err != nil {
		if rerr := t.Rollback(); rerr != nil {
			log.G(ctx).WithError(rerr).Warn("failed to rollback transaction")
		}
		return errors.Wrap(err, "failed to commit snapshot")
	}
	return t.Commit()
}

// Remove abandons the transaction identified by key. All resources
// associated with the key will be removed.
func (o *snapshotter) Remove(ctx context.Context, key string) (err error) {
	ctx, t, err := o.ms.TransactionContext(ctx, true)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil && t != nil {
			if rerr := t.Rollback(); rerr != nil {
				log.G(ctx).WithError(rerr).Warn("failed to rollback transaction")
			}
		}
	}()

	id, _, err := storage.Remove(ctx, key)
	if err != nil {
		return errors.Wrap(err, "failed to remove")
	}

	path := o.getSnapshotDir(id)
	renamed := filepath.Join(o.root, "snapshots", "rm-"+id)
	if err := os.Rename(path, renamed); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to rename")
		}
		renamed = ""
	}

	err = t.Commit()
	t = nil
	if err != nil {
		if renamed != "" {
			if err1 := os.Rename(renamed, path); err1 != nil {
				// May cause inconsistent data on disk
				log.G(ctx).WithError(err1).WithField("path", renamed).Errorf("failed to rename after failed commit")
			}
		}
		return errors.Wrap(err, "failed to commit")
	}
	if renamed != "" {
		if err := os.RemoveAll(renamed); err != nil {
			// Must be cleaned up, any "rm-*" could be removed if no active transactions
			log.G(ctx).WithError(err).WithField("path", renamed).Warnf("failed to remove root filesystem")
		}
	}

	return nil
}

// Walk the committed snapshots.
func (o *snapshotter) Walk(ctx context.Context, fn func(context.Context, snapshots.Info) error) error {
	ctx, t, err := o.ms.TransactionContext(ctx, false)
	if err != nil {
		return err
	}
	defer t.Rollback()
	return storage.WalkInfo(ctx, fn)
}

func (o *snapshotter) createSnapshot(ctx context.Context, kind snapshots.Kind, key, parent string, opts []snapshots.Opt) ([]mount.Mount, error) {
	var (
		err      error
		path, td string
	)

	if kind == snapshots.KindActive || parent == "" {
		td, err = ioutil.TempDir(filepath.Join(o.root, "snapshots"), "new-")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create temp dir")
		}
		defer func() {
			if err != nil {
				if td != "" {
					if err1 := os.RemoveAll(td); err1 != nil {
						err = errors.Wrapf(err, "remove failed: %v", err1)
					}
				}
				if path != "" {
					if err1 := os.RemoveAll(path); err1 != nil {
						err = errors.Wrapf(err, "failed to remove path: %v", err1)
					}
				}
			}
		}()
	}

	ctx, t, err := o.ms.TransactionContext(ctx, true)
	if err != nil {
		return nil, err
	}

	s, err := storage.CreateSnapshot(ctx, kind, key, parent, opts...)
	if err != nil {
		if rerr := t.Rollback(); rerr != nil {
			log.G(ctx).WithError(rerr).Warn("failed to rollback transaction")
		}
		return nil, errors.Wrap(err, "failed to create snapshot")
	}

	if td != "" {
		if len(s.ParentIDs) > 0 {
			parent := o.getSnapshotDir(s.ParentIDs[0])
			if err := fs.CopyDir(td, parent); err != nil {
				return nil, errors.Wrap(err, "copying of parent failed")
			}
		}

		path = o.getSnapshotDir(s.ID)
		if err := os.Rename(td, path); err != nil {
			if rerr := t.Rollback(); rerr != nil {
				log.G(ctx).WithError(rerr).Warn("failed to rollback transaction")
			}
			return nil, errors.Wrap(err, "failed to rename")
		}
		td = ""
	}

	if err := t.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit failed")
	}

	return o.mounts(s), nil
}

func (o *snapshotter) getSnapshotDir(id string) string {
	return filepath.Join(o.root, "snapshots", id)
}

func (o *snapshotter) mounts(s storage.Snapshot) []mount.Mount {
	var (
		roFlag string
		source string
	)

	if s.Kind == snapshots.KindView {
		roFlag = "ro"
	} else {
		roFlag = "rw"
	}

	if len(s.ParentIDs) == 0 || s.Kind == snapshots.KindActive {
		source = o.getSnapshotDir(s.ID)
	} else {
		source = o.getSnapshotDir(s.ParentIDs[0])
	}

	return []mount.Mount{
		{
			Source: source,
			Type:   "bind",
			Options: []string{
				roFlag,
				"rbind",
			},
		},
	}
}

// Close closes the snapshotter
func (o *snapshotter) Close() error {
	return o.ms.Close()
}

// If directory does not exist, create it.
func newSnapshotDir(root string) error {
	if _, err := os.Stat(root); err != nil {
		if !os.IsNotExist(err) {
			if strings.Contains(err.Error(), "transport endpoint is not connected") {
				// Unmount the mount
				unmount(root)
				return nil
			}
			return err
		}

		if err := os.Mkdir(root, 0755); err != nil {
			return err
		}

		return nil
	}

	// Check if it's already mounted (if we exited poorly) and unmount it.
	unmount(root)

	return nil
}

// unmount checks if root is already mounted (if we exited poorly) and unmounts it.
func unmount(root string) {
	_, err := mount.Lookup(root)
	logrus.Printf("mount err %#v", err)
	if err == nil {
		cmd := exec.Command("fusermount", "-u", root)
		out, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Warnf("Calling `fusermount -u %s` failed: %s %v", root, string(out), err)
		}
	}
}
