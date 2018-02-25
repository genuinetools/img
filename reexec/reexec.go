package reexec

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/docker/docker/pkg/idtools"
	"github.com/jessfraz/img/types"
	"github.com/opencontainers/runc/libcontainer/user"
	"golang.org/x/sys/unix"
)

// Command returns *exec.Cmd which has Path as current binary. Also it setting
// SysProcAttr.Pdeathsig to SIGTERM.
// This will use the in-memory version (/proc/self/exe) of the current binary,
// it is thus safe to delete or replace the on-disk binary (os.Args[0]).
// This also sets the unshare flags and the uid and gid mappings.
func Command(args ...string) (*exec.Cmd, error) {
	// get the current user
	u, err := user.CurrentUser()
	if err != nil {
		return nil, fmt.Errorf("getting current user failed: %v", err)
	}

	// get the current group
	g, err := user.CurrentGroup()
	if err != nil {
		return nil, fmt.Errorf("getting current group failed: %v", err)
	}

	// create the id mappings
	idmaps, err := idtools.NewIDMappings(u.Name, g.Name)
	if err != nil {
		return nil, fmt.Errorf("creating id mappings failed: %v", err)
	}

	// get the uid maps
	umaps := []syscall.SysProcIDMap{}
	for _, umap := range idmaps.UIDs() {
		umaps = append(umaps, syscall.SysProcIDMap{
			ContainerID: umap.ContainerID,
			HostID:      umap.HostID,
			Size:        umap.Size,
		})
	}

	// get the gid maps
	gmaps := []syscall.SysProcIDMap{}
	for _, gmap := range idmaps.GIDs() {
		gmaps = append(gmaps, syscall.SysProcIDMap{
			ContainerID: gmap.ContainerID,
			HostID:      gmap.HostID,
			Size:        gmap.Size,
		})
	}

	return &exec.Cmd{
		Path: "/proc/self/exe",
		Args: args,
		SysProcAttr: &syscall.SysProcAttr{
			Pdeathsig:                  unix.SIGTERM,
			UidMappings:                umaps,
			GidMappings:                gmaps,
			GidMappingsEnableSetgroups: false,
			Unshareflags:               syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		},
		Env: []string{fmt.Sprintf("%s=1", types.InUnshareEnv)},
	}, nil
}
