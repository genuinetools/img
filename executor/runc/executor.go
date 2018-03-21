package runc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/containerd/containerd/mount"
	containerdoci "github.com/containerd/containerd/oci"
	"github.com/containerd/continuity/fs"
	runc "github.com/containerd/go-runc"
	"github.com/moby/buildkit/cache"
	"github.com/moby/buildkit/executor"
	"github.com/moby/buildkit/executor/oci"
	"github.com/moby/buildkit/identity"
	"github.com/opencontainers/runc/libcontainer/specconv"
)

// Executor is the definition of an executor.
type Executor struct {
	runc         *runc.Runc
	root         string
	unprivileged bool
}

// New creates a new runc executor.
func New(root string, unprivileged bool) (executor.Executor, error) {
	// Make sure the runc binary exists.
	if exists := BinaryExists(); !exists {
		return nil, errors.New("cannot find runc binary locally, please install runc")
	}

	// Make the root directory.
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, fmt.Errorf("failed to create root directory in %s: %v", root, err)
	}

	// Get the absolute path to the root directory,
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	// TODO: check that root is not symlink to fail early

	runtime := &runc.Runc{
		Log:          filepath.Join(root, "runc-executor-log.json"),
		LogFormat:    runc.JSON,
		PdeathSignal: syscall.SIGKILL,
		Setpgid:      true,
	}

	e := &Executor{
		runc:         runtime,
		root:         root,
		unprivileged: unprivileged,
	}

	return e, nil
}

// Exec executes arguments via runc.
func (w *Executor) Exec(ctx context.Context, meta executor.Meta, root cache.Mountable, mounts []executor.Mount, stdin io.ReadCloser, stdout, stderr io.WriteCloser) error {
	// Get the resolv.conf.
	resolvConf, err := oci.GetResolvConf(ctx, w.root)
	if err != nil {
		return fmt.Errorf("oci.GetResolvConf failed: %v", err)
	}

	// Get the hosts file.
	hostsFile, err := oci.GetHostsFile(ctx, w.root)
	if err != nil {
		return fmt.Errorf("oci.GetHostsFile failed: %v", err)
	}

	// Mount the cache.
	rootMount, err := root.Mount(ctx, false)
	if err != nil {
		return fmt.Errorf("root.Mount failed: %v", err)
	}

	// Create a new UUID.
	id := identity.NewID()
	bundle := filepath.Join(w.root, id)

	// Create the bundle directory.
	if err := os.Mkdir(bundle, 0700); err != nil {
		return fmt.Errorf("os.Mkdir failed: %v", err)
	}
	defer os.RemoveAll(bundle)

	// Create the rootfs path.
	var (
		rootFSPath string
	)
	if w.unprivileged {
		// If we are running as unprivileged do not do the mounts.
		rootFSPath = rootMount[0].Source

	} else {
		rootFSPath = filepath.Join(bundle, "rootfs")
		if err := os.Mkdir(rootFSPath, 0700); err != nil {
			return fmt.Errorf("failed to create directory at %s: %v", rootFSPath, err)
		}

		// Mount the root and rootfs path.
		if err := mount.All(rootMount, rootFSPath); err != nil {
			return fmt.Errorf("mount.All failed: %v", err)
		}
		defer mount.Unmount(rootFSPath, 0)
	}

	// Get the user in the bundle.
	uid, gid, err := oci.GetUser(ctx, rootFSPath, meta.User)
	if err != nil {
		return err
	}

	// Create the config file.
	f, err := os.Create(filepath.Join(bundle, "config.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	// Generate the spec.
	spec, cleanup, err := oci.GenerateSpec(ctx, meta, mounts, id, resolvConf, hostsFile, containerdoci.WithUIDGID(uid, gid))
	if err != nil {
		return err
	}
	defer cleanup()

	// Set the spec root to rootfs.
	spec.Root.Path = rootFSPath
	if _, ok := root.(cache.ImmutableRef); ok { // TODO: pass in with mount, not ref type
		spec.Root.Readonly = true
	}

	newp, err := fs.RootPath(rootFSPath, meta.Cwd)
	if err != nil {
		return fmt.Errorf("working dir %s points to invalid target: %v", newp, err)
	}

	if err := os.MkdirAll(newp, 0700); err != nil {
		return fmt.Errorf("failed to create directory at %s: %v", newp, err)
	}

	// if we are not running as root setup unprivileged.
	if w.unprivileged {
		// Make sure the spec is rootless.
		// Only if we are not running as root.
		specconv.ToRootless(spec, &specconv.RootlessOpts{
			// This comes from https://github.com/opencontainers/runc/pull/1692
			// MapSubUIDGID: true,
		})
		// Remove the cgroups path.
		spec.Linux.CgroupsPath = ""
	}

	// Set the default seccomp profile.
	spec.Linux.Seccomp = DefaultSeccompProfile

	if err := json.NewEncoder(f).Encode(spec); err != nil {
		return err
	}

	fmt.Printf("RUN %v\n", meta.Args)
	fmt.Println("--->")

	opts := &runc.CreateOpts{
		IO:           &forwardIO{stdin: stdin, stdout: stdout, stderr: stderr},
		NoNewKeyring: true,
	}
	//if w.unprivileged {
	// This comes from https://github.com/opencontainers/runc/pull/1693
	// opts.ForceMappingTool = true
	//}
	status, err := w.runc.Run(ctx, id, bundle, opts)

	fmt.Printf("<--- %s %v %v\n", id, status, err)

	if status != 0 {
		select {
		case <-ctx.Done():
			// runc can't report context.Cancelled directly
			return fmt.Errorf("exit code %d: %v", status, ctx.Err())
		default:
		}
		return fmt.Errorf("exit code %d", status)
	}

	return err
}

type forwardIO struct {
	stdin          io.ReadCloser
	stdout, stderr io.WriteCloser
}

func (s *forwardIO) Close() error {
	return nil
}

func (s *forwardIO) Set(cmd *exec.Cmd) {
	cmd.Stdin = s.stdin
	cmd.Stdout = s.stdout
	cmd.Stderr = s.stderr
}

func (s *forwardIO) Stdin() io.WriteCloser {
	return nil
}

func (s *forwardIO) Stdout() io.ReadCloser {
	return nil
}

func (s *forwardIO) Stderr() io.ReadCloser {
	return nil
}

// BinaryExists checks if the runc binary exists.
func BinaryExists() bool {
	_, err := exec.LookPath("runc")
	// Return true when there is no error.
	return err == nil
}
