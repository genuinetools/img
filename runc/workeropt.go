package runc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/diff/apply"
	"github.com/containerd/containerd/diff/walking"
	ctdmetadata "github.com/containerd/containerd/metadata"
	ctdsnapshot "github.com/containerd/containerd/snapshots"
	"github.com/containerd/containerd/snapshots/overlay"
	libfuse "github.com/hanwen/go-fuse/fuse"
	"github.com/jessfraz/img/snapshots/fuse"
	"github.com/moby/buildkit/cache/metadata"
	containerdsnapshot "github.com/moby/buildkit/snapshot/containerd"
	"github.com/moby/buildkit/worker/base"
)

// NewWorkerOpt creates a WorkerOpt.
func NewWorkerOpt(root, backend string) (opt base.WorkerOpt, fuseserver *libfuse.Server, err error) {
	name := "runc"

	// Create the root/
	root = filepath.Join(root, name, backend)
	if err := os.MkdirAll(root, 0700); err != nil {
		return opt, nil, err
	}

	// Create the metadata store.
	md, err := metadata.NewStore(filepath.Join(root, "metadata.db"))
	if err != nil {
		return opt, nil, err
	}

	// Create the runc executor.
	exe, err := newExecutor(filepath.Join(root, "executor"))
	if err != nil {
		return opt, nil, err
	}

	// Create the snapshotter.
	var (
		s ctdsnapshot.Snapshotter
	)
	switch backend {
	case "fuse":
		s, fuseserver, err = fuse.NewSnapshotter(filepath.Join(root, "snapshots"))
	case "overlayfs":
		s, err = overlay.NewSnapshotter(filepath.Join(root, "snapshots"))
	default:
		return opt, nil, fmt.Errorf("%s is not a valid snapshots backend", backend)
	}
	if err != nil {
		return opt, fuseserver, fmt.Errorf("creating %s snapshotter failed: %v", backend, err)
	}

	// Create the content store locally.
	c, err := local.NewStore(filepath.Join(root, "content"))
	if err != nil {
		return opt, fuseserver, err
	}

	// Open the bolt database for metadata.
	db, err := bolt.Open(filepath.Join(root, "containerdmeta.db"), 0644, nil)
	if err != nil {
		return opt, fuseserver, err
	}

	// Create the new database for metadata.
	mdb := ctdmetadata.NewDB(db, c, map[string]ctdsnapshot.Snapshotter{
		backend: s,
	})
	if err := mdb.Init(context.TODO()); err != nil {
		return opt, fuseserver, err
	}

	gc := func(ctx context.Context) error {
		_, err := mdb.GarbageCollect(ctx)
		return err
	}

	c = containerdsnapshot.NewContentStore(mdb.ContentStore(), "buildkit", gc)

	id, err := base.ID(root)
	if err != nil {
		return opt, fuseserver, err
	}

	xlabels := base.Labels("oci", backend)

	opt = base.WorkerOpt{
		ID:            id,
		Labels:        xlabels,
		MetadataStore: md,
		Executor:      exe,
		Snapshotter:   containerdsnapshot.NewSnapshotter(mdb.Snapshotter(backend), c, md, "buildkit", gc),
		ContentStore:  c,
		Applier:       apply.NewFileSystemApplier(c),
		Differ:        walking.NewWalkingDiff(c),
		ImageStore:    nil, // explicitly
	}
	return opt, fuseserver, nil
}
