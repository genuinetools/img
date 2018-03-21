// +build seccomp

package runc

import (
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	libseccomp "github.com/seccomp/libseccomp-golang"
)

func arches() []specs.Arch {
	var native, err = libseccomp.GetNativeArch()
	if err != nil {
		return []specs.Arch{}
	}
	var a = native.String()
	switch a {
	case "amd64":
		return []specs.Arch{specs.ArchX86_64, specs.ArchX86, specs.ArchX32}
	case "arm64":
		return []specs.Arch{specs.ArchARM, specs.ArchAARCH64}
	case "mips64":
		return []specs.Arch{specs.ArchMIPS, specs.ArchMIPS64, specs.ArchMIPS64N32}
	case "mips64n32":
		return []specs.Arch{specs.ArchMIPS, specs.ArchMIPS64, specs.ArchMIPS64N32}
	case "mipsel64":
		return []specs.Arch{specs.ArchMIPSEL, specs.ArchMIPSEL64, specs.ArchMIPSEL64N32}
	case "mipsel64n32":
		return []specs.Arch{specs.ArchMIPSEL, specs.ArchMIPSEL64, specs.ArchMIPSEL64N32}
	default:
		return []specs.Arch{}
	}
}

// DefaultSeccompProfile defines the whitelist for the default seccomp profile.
var DefaultSeccompProfile = &specs.LinuxSeccomp{
	DefaultAction: specs.ActErrno,
	Architectures: arches(),
	Syscalls: []specs.LinuxSyscall{
		{
			Names:  []string{"accept"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"accept4"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"access"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"alarm"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"arch_prctl"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"bind"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"brk"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"capget"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"capset"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"chdir"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"chmod"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"chown"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"chown32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"chroot"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"clock_getres"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"clock_gettime"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"clock_nanosleep"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"clone"},
			Action: specs.ActAllow,
			Args: []specs.LinuxSeccompArg{
				{
					Index:    0,
					Value:    syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWNET,
					ValueTwo: 0,
					Op:       specs.OpMaskedEqual,
				},
			},
		},
		{
			Names:  []string{"close"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"connect"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"copy_file_range"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"creat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"dup"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"dup2"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"dup3"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"epoll_create"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"epoll_create1"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"epoll_ctl"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"epoll_ctl_old"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"epoll_pwait"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"epoll_wait"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"epoll_wait_old"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"eventfd"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"eventfd2"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"execve"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"execveat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"exit"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"exit_group"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"faccessat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fadvise64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fadvise64_64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fallocate"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fanotify_init"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fanotify_mark"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fchdir"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fchmod"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fchmodat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fchown"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fchown32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fchownat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fcntl"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fcntl64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fdatasync"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fgetxattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"flistxattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"flock"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fork"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fremovexattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fsetxattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fstat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fstat64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fstatat64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fstatfs"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fstatfs64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"fsync"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"ftruncate"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"ftruncate64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"futex"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"futimesat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getcpu"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getcwd"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getdents"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getdents64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getegid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getegid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"geteuid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"geteuid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getgid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getgid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getgroups"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getgroups32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getitimer"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getpeername"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getpgid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getpgrp"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getpid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getppid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getpriority"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getrandom"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getresgid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getresgid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getresuid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getresuid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getrlimit"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"get_robust_list"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getrusage"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getsid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getsockname"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getsockopt"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"get_thread_area"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"gettid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"gettimeofday"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getuid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getuid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"getxattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"inotify_add_watch"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"inotify_init"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"inotify_init1"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"inotify_rm_watch"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"io_cancel"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"ioctl"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"io_destroy"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"io_getevents"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"ioprio_get"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"ioprio_set"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"io_setup"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"io_submit"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"ipc"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"kill"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"lchown"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"lchown32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"lgetxattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"link"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"linkat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"listen"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"listxattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"llistxattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"_llseek"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"lremovexattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"lseek"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"lsetxattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"lstat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"lstat64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"madvise"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"memfd_create"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mincore"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mkdir"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mkdirat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mknod"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mknodat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mlock"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mlock2"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mlockall"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mmap"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mmap2"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mprotect"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mq_getsetattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mq_notify"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mq_open"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mq_timedreceive"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mq_timedsend"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mq_unlink"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"mremap"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"msgctl"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"msgget"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"msgrcv"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"msgsnd"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"msync"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"munlock"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"munlockall"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"munmap"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"nanosleep"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"newfstatat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"_newselect"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"open"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"openat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"pause"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"personality"},
			Action: specs.ActAllow,
			Args: []specs.LinuxSeccompArg{
				{
					Index: 0,
					Value: 0x0,
					Op:    specs.OpEqualTo,
				},
			},
		},
		{
			Names:  []string{"personality"},
			Action: specs.ActAllow,
			Args: []specs.LinuxSeccompArg{
				{
					Index: 0,
					Value: 0x0008,
					Op:    specs.OpEqualTo,
				},
			},
		},
		{
			Names:  []string{"personality"},
			Action: specs.ActAllow,
			Args: []specs.LinuxSeccompArg{
				{
					Index: 0,
					Value: 0xffffffff,
					Op:    specs.OpEqualTo,
				},
			},
		},
		{
			Names:  []string{"pipe"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"pipe2"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"poll"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"ppoll"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"prctl"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"pread64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"preadv"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"prlimit64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"pselect6"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"pwrite64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"pwritev"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"read"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"readahead"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"readlink"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"readlinkat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"readv"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"recv"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"recvfrom"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"recvmmsg"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"recvmsg"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"remap_file_pages"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"removexattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"rename"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"renameat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"renameat2"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"restart_syscall"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"rmdir"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"rt_sigaction"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"rt_sigpending"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"rt_sigprocmask"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"rt_sigqueueinfo"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"rt_sigreturn"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"rt_sigsuspend"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"rt_sigtimedwait"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"rt_tgsigqueueinfo"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_getaffinity"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_getattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_getparam"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_get_priority_max"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_get_priority_min"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_getscheduler"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_rr_get_interval"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_setaffinity"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_setattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_setparam"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_setscheduler"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sched_yield"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"seccomp"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"select"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"semctl"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"semget"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"semop"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"semtimedop"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"send"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sendfile"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sendfile64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sendmmsg"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sendmsg"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sendto"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setdomainname"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setfsgid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setfsgid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setfsuid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setfsuid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setgid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setgid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setgroups"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setgroups32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sethostname"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setitimer"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setpgid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setpriority"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setregid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setregid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setresgid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setresgid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setresuid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setresuid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setreuid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setreuid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setrlimit"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"set_robust_list"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setsid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setsockopt"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"set_thread_area"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"set_tid_address"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setuid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setuid32"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"setxattr"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"shmat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"shmctl"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"shmdt"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"shmget"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"shutdown"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sigaltstack"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"signalfd"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"signalfd4"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sigreturn"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"socket"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"socketpair"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"splice"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"stat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"stat64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"statfs"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"statfs64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"symlink"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"symlinkat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sync"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sync_file_range"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"syncfs"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"sysinfo"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"syslog"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"tee"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"tgkill"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"time"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"timer_create"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"timer_delete"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"timerfd_create"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"timerfd_gettime"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"timerfd_settime"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"timer_getoverrun"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"timer_gettime"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"timer_settime"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"times"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"tkill"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"truncate"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"truncate64"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"ugetrlimit"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"umask"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"uname"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"unlink"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"unlinkat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"utime"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"utimensat"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"utimes"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"vfork"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"vhangup"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"vmsplice"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"wait4"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"waitid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"waitpid"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"write"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"writev"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		// i386 specific syscalls
		{
			Names:  []string{"modify_ldt"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		// arm specific syscalls
		{
			Names:  []string{"breakpoint"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"cacheflush"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
		{
			Names:  []string{"set_tls"},
			Action: specs.ActAllow,
			Args:   []specs.LinuxSeccompArg{},
		},
	},
}
