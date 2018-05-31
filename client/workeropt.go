package client

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/diff/apply"
	"github.com/containerd/containerd/diff/walking"
	ctdmetadata "github.com/containerd/containerd/metadata"
	ctdsnapshot "github.com/containerd/containerd/snapshots"
	"github.com/containerd/containerd/snapshots/native"
	"github.com/containerd/containerd/snapshots/overlay"
	"github.com/genuinetools/img/internal/executor/runc"
	"github.com/genuinetools/img/types"
	"github.com/moby/buildkit/cache/metadata"
	containerdsnapshot "github.com/moby/buildkit/snapshot/containerd"
	"github.com/moby/buildkit/util/throttle"
	"github.com/moby/buildkit/worker/base"
	"github.com/opencontainers/runc/libcontainer/system"
	"github.com/sirupsen/logrus"
)

// createWorkerOpt creates a base.WorkerOpt to be used for a new worker.
func (c *Client) createWorkerOpt() (opt base.WorkerOpt, err error) {
	sm, err := c.getSessionManager()
	if err != nil {
		return opt, err
	}
	// Create the metadata store.
	md, err := metadata.NewStore(filepath.Join(c.root, "metadata.db"))
	if err != nil {
		return opt, err
	}

	snapshotRoot := filepath.Join(c.root, "snapshots")
	unprivileged := system.GetParentNSeuid() != 0
	noMount := unprivileged

	// Create the snapshotter.
	var (
		s ctdsnapshot.Snapshotter
	)
	switch c.backend {
	case types.NativeBackend:
		s, err = native.NewSnapshotter(snapshotRoot)
	case types.OverlayFSBackend:
		// On some distros such as Ubuntu overlayfs can be mounted without privileges
		noMount = false
		s, err = overlay.NewSnapshotter(snapshotRoot)
	default:
		// "auto" backend needs to be already resolved on Client instantiation
		return opt, fmt.Errorf("%s is not a valid snapshots backend", c.backend)
	}
	if err != nil {
		return opt, fmt.Errorf("creating %s snapshotter failed: %v", c.backend, err)
	}

	exe, err := runc.New(filepath.Join(c.root, "executor"), unprivileged, noMount)
	if err != nil {
		return opt, err
	}

	// Create the content store locally.
	contentStore, err := local.NewStore(filepath.Join(c.root, "content"))
	if err != nil {
		return opt, err
	}

	// Open the bolt database for metadata.
	db, err := bolt.Open(filepath.Join(c.root, "containerdmeta.db"), 0644, nil)
	if err != nil {
		return opt, err
	}

	// Create the new database for metadata.
	mdb := ctdmetadata.NewDB(db, contentStore, map[string]ctdsnapshot.Snapshotter{
		c.backend: s,
	})
	if err := mdb.Init(context.TODO()); err != nil {
		return opt, err
	}

	// Create the image store.
	imageStore := ctdmetadata.NewImageStore(mdb)

	// Create the garbage collector.
	throttledGC := throttle.Throttle(time.Second, func() {
		if _, err := mdb.GarbageCollect(context.TODO()); err != nil {
			logrus.Errorf("GC error: %+v", err)
		}
	})

	gc := func(ctx context.Context) error {
		throttledGC()
		return nil
	}

	contentStore = containerdsnapshot.NewContentStore(mdb.ContentStore(), "buildkit", gc)

	id, err := base.ID(c.root)
	if err != nil {
		return opt, err
	}

	xlabels := base.Labels("oci", c.backend)

	opt = base.WorkerOpt{
		ID:             id,
		Labels:         xlabels,
		SessionManager: sm,
		MetadataStore:  md,
		Executor:       exe,
		Snapshotter:    containerdsnapshot.NewSnapshotter(mdb.Snapshotter(c.backend), contentStore, md, "buildkit", gc),
		ContentStore:   contentStore,
		Applier:        apply.NewFileSystemApplier(contentStore),
		Differ:         walking.NewWalkingDiff(contentStore),
		ImageStore:     imageStore,
	}

	return opt, err
}
