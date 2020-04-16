package client

import (
	"context"
	"fmt"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/moby/buildkit/util/leaseutil"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/diff/apply"
	"github.com/containerd/containerd/diff/walking"
	ctdmetadata "github.com/containerd/containerd/metadata"
	"github.com/containerd/containerd/platforms"
	ctdsnapshot "github.com/containerd/containerd/snapshots"
	"github.com/containerd/containerd/snapshots/native"
	"github.com/containerd/containerd/snapshots/overlay"
	"github.com/genuinetools/img/types"
	"github.com/moby/buildkit/cache/metadata"
	"github.com/moby/buildkit/executor"
	executoroci "github.com/moby/buildkit/executor/oci"
	"github.com/moby/buildkit/executor/runcexecutor"
	containerdsnapshot "github.com/moby/buildkit/snapshot/containerd"
	"github.com/moby/buildkit/util/binfmt_misc"
	"github.com/moby/buildkit/util/network/netproviders"
	"github.com/moby/buildkit/worker/base"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runc/libcontainer/system"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// createWorkerOpt creates a base.WorkerOpt to be used for a new worker.
func (c *Client) createWorkerOpt(withExecutor bool) (opt base.WorkerOpt, err error) {
	// Create the metadata store.
	md, err := metadata.NewStore(filepath.Join(c.root, "metadata.db"))
	if err != nil {
		return opt, err
	}

	snapshotRoot := filepath.Join(c.root, "snapshots")
	unprivileged := system.GetParentNSeuid() != 0

	// Create the snapshotter.
	var (
		s ctdsnapshot.Snapshotter
	)
	switch c.backend {
	case types.NativeBackend:
		s, err = native.NewSnapshotter(snapshotRoot)
	case types.OverlayFSBackend:
		// On some distros such as Ubuntu overlayfs can be mounted without privileges
		s, err = overlay.NewSnapshotter(snapshotRoot)
	default:
		// "auto" backend needs to be already resolved on Client instantiation
		return opt, fmt.Errorf("%s is not a valid snapshots backend", c.backend)
	}
	if err != nil {
		return opt, fmt.Errorf("creating %s snapshotter failed: %v", c.backend, err)
	}

	var exe executor.Executor
	if withExecutor {
		exeOpt := runcexecutor.Opt{
			Root:        filepath.Join(c.root, "executor"),
			Rootless:    unprivileged,
			ProcessMode: processMode(),
		}

		np, err := netproviders.Providers(netproviders.Opt{Mode: "auto"})
		if err != nil {
			return base.WorkerOpt{}, err
		}

		exe, err = runcexecutor.New(exeOpt, np)
		if err != nil {
			return opt, err
		}
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

	contentStore = containerdsnapshot.NewContentStore(mdb.ContentStore(), "buildkit")

	id, err := base.ID(c.root)
	if err != nil {
		return opt, err
	}

	xlabels := base.Labels("oci", c.backend)

	var supportedPlatforms []specs.Platform
	for _, p := range binfmt_misc.SupportedPlatforms(false) {
		parsed, err := platforms.Parse(p)
		if err != nil {
			return opt, err
		}
		supportedPlatforms = append(supportedPlatforms, platforms.Normalize(parsed))
	}

	opt = base.WorkerOpt{
		ID:             id,
		Labels:         xlabels,
		MetadataStore:  md,
		Executor:       exe,
		Snapshotter:    containerdsnapshot.NewSnapshotter(c.backend, mdb.Snapshotter(c.backend), "buildkit", nil),
		ContentStore:   contentStore,
		Applier:        apply.NewFileSystemApplier(contentStore),
		Differ:         walking.NewWalkingDiff(contentStore),
		ImageStore:     imageStore,
		Platforms:      supportedPlatforms,
		RegistryHosts:  docker.ConfigureDefaultRegistries(),
		LeaseManager:   leaseutil.WithNamespace(ctdmetadata.NewLeaseManager(mdb), "buildkit"),
		GarbageCollect: mdb.GarbageCollect,
	}

	return opt, err
}

func processMode() executoroci.ProcessMode {
	mountArgs := []string{"-t", "proc", "none", "/proc"}
	cmd := exec.Command("mount", mountArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig:    syscall.SIGKILL,
		Cloneflags:   syscall.CLONE_NEWPID,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	if b, err := cmd.CombinedOutput(); err != nil {
		logrus.Warnf("Process sandbox is not available, consider unmasking procfs: %v", string(b))
		return executoroci.NoProcessSandbox
	}
	return executoroci.ProcessSandbox
}
