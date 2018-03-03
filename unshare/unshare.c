#define _GNU_SOURCE
#include <endian.h>
#include <errno.h>
#include <fcntl.h>
#include <getopt.h>
#include <grp.h>
#include <inttypes.h>
#include <pwd.h>
#include <sched.h>
#include <setjmp.h>
#include <signal.h>
#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include <sys/ioctl.h>
#include <sys/mount.h>
#include <sys/prctl.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/wait.h>

#include <linux/limits.h>
#include <linux/netlink.h>
#include <linux/types.h>

/* Get all of the CLONE_NEW* flags. */
#include "namespace.h"

/* Synchronisation values. */
enum sync_t {
	SYNC_USERMAP_PLS = 0x40,	/* Request parent to map our users. */
	SYNC_USERMAP_ACK = 0x41,	/* Mapping finished by the parent. */
	SYNC_CHILD_READY = 0x45,	/* The child is ready to return. */

	/* XXX: This doesn't help with segfaults and other such issues. */
	SYNC_ERR = 0xFF,	/* Fatal error, no turning back. The error code follows. */
};

/* longjmp() arguments. */
#define JUMP_PARENT 0x00
#define JUMP_CHILD  0xA0

/* Assume the stack grows down, so arguments should be above it. */
struct clone_t {
	/*
	 * Reserve some space for clone() to locate arguments
	 * and retcode in this place
	 */
	char stack[4096] __attribute__ ((aligned(16)));
	char stack_ptr[0];

	/* There's two children. This is used to execute the different code. */
	jmp_buf *env;
	int jmpval;
};

/*
 * Use the raw syscall for versions of glibc which don't include a function for
 * it, namely (glibc 2.12).
 */
#if __GLIBC__ == 2 && __GLIBC_MINOR__ < 14
#	define _GNU_SOURCE
#	include "syscall.h"
#	if !defined(SYS_setns) && defined(__NR_setns)
#		define SYS_setns __NR_setns
#	endif

#ifndef SYS_setns
#	error "setns(2) syscall not supported by glibc version"
#endif

int setns(int fd, int nstype)
{
	return syscall(SYS_setns, fd, nstype);
}
#endif

static int syncfd = -1;

#define bail(fmt, ...)								\
	do {									\
		int ret = __COUNTER__ + 1;					\
		fprintf(stderr, "nsenter: " fmt ": %m\n", ##__VA_ARGS__);	\
		if (syncfd >= 0) {						\
			enum sync_t s = SYNC_ERR;				\
			if (write(syncfd, &s, sizeof(s)) != sizeof(s))		\
			fprintf(stderr, "nsenter: failed: write(s)");	\
			if (write(syncfd, &ret, sizeof(ret)) != sizeof(ret))	\
			fprintf(stderr, "nsenter: failed: write(ret)");	\
		}								\
		exit(ret);							\
	} while(0)

/* A dummy function that just jumps to the given jumpval. */
static int child_func(void *arg) __attribute__ ((noinline));
static int child_func(void *arg)
{
	struct clone_t *ca = (struct clone_t *)arg;
	longjmp(*ca->env, ca->jmpval);
}

static int clone_parent(jmp_buf *env, int jmpval) __attribute__ ((noinline));
static int clone_parent(jmp_buf *env, int jmpval)
{
	struct clone_t ca = {
		.env = env,
		.jmpval = jmpval,
	};

	return clone(child_func, ca.stack_ptr, CLONE_PARENT | SIGCHLD, &ca);
}

/* uid gid arguments */
#define UID 1
#define GID 0

#define _PATH_PROC_SETGROUPS	"/proc/self/setgroups"

/* synchronize parent and child by pipe */
#define PIPE_SYNC_BYTE	0x06

/* 'private' is kernel default */
#define UNSHARE_PROPAGATION_DEFAULT	(MS_REC | MS_PRIVATE)

#define getid(type) ((unsigned) ((type) == GID ? getgid() : getuid()))
#define idfile(type) ((type) == GID ? "gid_map" : "uid_map")
#define idtool(type) ((type) == GID ? "/usr/bin/newgidmap" : "/usr/bin/newuidmap")
#define subpath(type) ((type) == GID ? "/etc/subgid" : "/etc/subuid")
#define mappath(type) ((type) == GID ? "/proc/self/gid_map" : "/proc/self/uid_map")

char *append(char **destination, const char *format, ...) {
	char *extra, *result;
	va_list args;

	va_start(args, format);
	if (vasprintf(&extra, format, args) < 0)
		bail("asprintf");
	va_end(args);

	if (*destination == NULL) {
		*destination = extra;
		return extra;
	}

	if (asprintf(&result, "%s%s", *destination, extra) < 0)
		bail("asprintf");

	free(*destination);
	free(extra);
	*destination = result;
	return result;
}

char *string(const char *format, ...) {
	char *result;
	va_list args;

	va_start(args, format);
	if (vasprintf(&result, format, args) < 0)
		bail("asprintf");
	va_end(args);
	return result;
}

static char *range_item(char *range, unsigned *start, unsigned *length) {
	ssize_t skip;

	while (range && *range && strchr(",;", *range))
		range++;
	if (range == NULL || *range == '\0')
		return NULL;
	if (sscanf(range, "%u:%u%zn", start, length, &skip) < 2)
		bail("Invalid ID range '%s'", range);
	return range + skip;
}


static int try_mapping_tool(const char *app, int pid, char *range, char *id)
{
	int child;

	/*
	 * If @app is NULL, execve will segfault. Just check it here and bail (if
	 * we're in this path, the caller is already getting desparate and there
	 * isn't a backup to this failing). This usually would be a configuration
	 * or programming issue.
	 */
	if (!app)
		bail("mapping tool not present");

	child = fork();
	if (child < 0)
		bail("failed to fork");

	if (!child) {
#define MAX_ARGV 9
		char *argv[MAX_ARGV];
		char *envp[] = { NULL };
		char pid_fmt[16];
		int argc = 0;
		unsigned start, length;

		snprintf(pid_fmt, 16, "%d", pid);

		argv[argc++] = (char *)app;
		argv[argc++] = pid_fmt;
		argv[argc++] = "0";
		argv[argc++] = (char *)id;
		argv[argc++] = "1";
		argv[argc++] = "1";
		/*
		 * Convert the map string into a list of argument that
		 * newuidmap/newgidmap can understand.
		 */

		while ((range = range_item(range, &start, &length))) {
			char startstr[16];
			char lengthstr[16];
			sprintf(startstr, "%u", start);
			sprintf(lengthstr, "%u", length);
			argv[6] = startstr;
			argv[7] = lengthstr;
			argv[8]= (char*)0;
		}

		execve(app, argv, envp);
		fflush(stdout);
		fflush(stderr);
		bail("failed to execv");
	} else {
		int status;

		while (1) {
			if (waitpid(child, &status, 0) < 0) {
				if (errno == EINTR)
					continue;
				bail("failed to waitpid");
			}
			if (WIFEXITED(status) || WIFSIGNALED(status))
				return WEXITSTATUS(status);
		}
	}

	return -1;
}

static char *read_ranges(int type) {
	char *line = NULL, *entry, *range, *user;
	size_t end, size;
	//struct passwd *passwd;
	uid_t uid;
	unsigned int length, start;
	FILE *file;

	range = string("%u:1", getid(type));
	if (!(file = fopen(subpath(type), "r")))
		return range;

	uid = getuid();
	user = getenv("USER");
	user = user ? user : getlogin();

	while (getline(&line, &size, file) >= 0) {
		if (strtol(line, &entry, 10) != uid || entry == line) {
			if (strncmp(line, user, strlen(user)))
				continue;
			entry = line + strlen(user);
		}
		if (sscanf(entry, ":%u:%u%zn", &start, &length, &end) < 2)
			continue;
		if (strchr(":\n", entry[end + 1]))
			append(&range, ",%u:%u", start, length);
	}

	free(line);
	fclose(file);

	return range;
}

static void set_propagation(unsigned long flags)
{
	if (flags == 0)
		return;

	if (mount("none", "/", NULL, flags, NULL) != 0)
		bail("cannot change root filesystem propagation");
}

void nsexec(void)
{
	/*
	 * Return early if we are just running the tests.
	 */
	const char* running_tests = getenv("IMG_RUNNING_TESTS");
	if (running_tests){
		return;
	}

	unsigned long propagation = UNSHARE_PROPAGATION_DEFAULT;
	jmp_buf env;
	int sync_child_pipe[2];
	char euid_fmt[16];
	char egid_fmt[16];

	/*
	 * Get our current euid and egid.
	 */
	uid_t real_euid = geteuid();
	gid_t real_egid = getegid();
	snprintf(euid_fmt, 16, "%u",real_euid);
	snprintf(egid_fmt, 16, "%u",real_egid);

	/*
	 * Read our uid and gid map ranges.
	 */
	char *uid_map;
	uid_map = read_ranges(UID);
	char *gid_map;
	gid_map = read_ranges(GID);

	/*
	 * Make the process non-dumpable, to avoid various race conditions that
	 * could cause processes in namespaces we're joining to access host
	 * resources (or potentially execute code).
	 */
	if (prctl(PR_SET_DUMPABLE, 0, 0, 0, 0) < 0){
		bail("failed to set process as non-dumpable");
	}

	/* Pipe so we can tell the child when we've finished setting up. */
	if (socketpair(AF_LOCAL, SOCK_STREAM, 0, sync_child_pipe) < 0)
		bail("failed to setup sync pipe between parent and child");

	/*
	 * Setup the stages.
	 * See: https://github.com/opencontainers/runc/blob/master/libcontainer/nsenter/nsexec.c#L631
	 */
	switch (setjmp(env)) {
		/*
		 * Stage 0: We're in the parent. Our job is just to create a new child
		 *          (stage 1: JUMP_CHILD) process and write its uid_map and
		 *          gid_map. That process will go on to create a new process, then
		 *          it will send us its PID which we will send to the bootstrap
		 *          process.
		 */
	case JUMP_PARENT:{
		pid_t child = -1;
		bool ready = false;

		/* For debugging. */
		prctl(PR_SET_NAME, (unsigned long)"runc:[0:PARENT]", 0, 0, 0);

		/* Start the process of getting a container. */
		child = clone_parent(&env, JUMP_CHILD);
		if (child < 0)
			bail("unable to fork: child_func");

		/*
		 * State machine for synchronisation with the children.
		 *
		 * Father only return when both child is
		 * ready, so we can receive all possible error codes
		 * generated by children.
		 */
		while (!ready) {
			enum sync_t s;
			int ret;

			syncfd = sync_child_pipe[1];
			close(sync_child_pipe[0]);

			if (read(syncfd, &s, sizeof(s)) != sizeof(s))
				bail("[parent]: failed to sync with child: next state");

			switch (s) {
			case SYNC_ERR:
				/* We have to mirror the error code of the child. */
				if (read(syncfd, &ret, sizeof(ret)) != sizeof(ret))
					bail("failed to sync with child: read(error code)");

				exit(ret);
			case SYNC_USERMAP_PLS:
				/* Set up mappings. */
				if (try_mapping_tool(idtool(UID), child, uid_map, euid_fmt))
					bail("failed to use newuidmap");

				if (try_mapping_tool(idtool(GID), child, gid_map, egid_fmt))
					bail("failed to use newgidmap");

				s = SYNC_USERMAP_ACK;
				if (write(syncfd, &s, sizeof(s)) != sizeof(s)) {
					kill(child, SIGKILL);
					bail("failed to sync with child: write(SYNC_USERMAP_ACK)");
				}
				break;
			case SYNC_CHILD_READY:
				ready = true;
				break;
			default:
				bail("unexpected sync value: %u", s);
			}
		}

		exit(0);
	}

		/*
		 * Stage 1: We're in the first child process. Our job is to join any
		 *          provided namespaces in the netlink payload and unshare all
		 *          of the requested namespaces. If we've been asked to
		 *          CLONE_NEWUSER, we will ask our parent (stage 0) to set up
		 *          our user mappings for us.
		 */
	case JUMP_CHILD:{
		enum sync_t s;

		/* We're in a child and thus need to tell the parent if we die. */
		syncfd = sync_child_pipe[0];
		close(sync_child_pipe[1]);

		/* For debugging. */
		prctl(PR_SET_NAME, (unsigned long)"runc:[1:CHILD]", 0, 0, 0);

		/*
		 * Unshare all of the namespaces. Now, it should be noted that this
		 * ordering might break in the future (especially with rootless
		 * containers). But for now, it's not possible to split this into
		 * CLONE_NEWUSER + [the rest] because of some RHEL SELinux issues.
		 *
		 * Note that we don't merge this with clone() because there were
		 * some old kernel versions where clone(CLONE_PARENT | CLONE_NEWPID)
		 * was broken, so we'll just do it the long way anyway.
		 */
		if (unshare(CLONE_NEWNS | CLONE_NEWUSER) < 0)
			bail("failed to unshare namespaces");

		/* Set the mount propogation */
		set_propagation(propagation);

		/*
		 * Deal with user namespaces first. They are quite special, as they
		 * affect our ability to unshare other namespaces and are used as
		 * context for privilege checks.
		 */
		/*
		 * We don't have the privileges to do any mapping here (see the
		 * clone_parent rant). So signal our parent to hook us up.
		 */


		/* Switching is only necessary if we joined namespaces. */
		if (prctl(PR_SET_DUMPABLE, 1, 0, 0, 0) < 0)
			bail("failed to set process as dumpable");
		s = SYNC_USERMAP_PLS;
		if (write(syncfd, &s, sizeof(s)) != sizeof(s))
			bail("failed to sync with parent: write(SYNC_USERMAP_PLS)");

		/* ... wait for mapping ... */

		if (read(syncfd, &s, sizeof(s)) != sizeof(s))
			bail("failed to sync with parent: read(SYNC_USERMAP_ACK)");
		if (s != SYNC_USERMAP_ACK)
			bail("failed to sync with parent: SYNC_USERMAP_ACK: got %u", s);
		/* Switching is only necessary if we joined namespaces. */
		if (prctl(PR_SET_DUMPABLE, 0, 0, 0, 0) < 0)
			bail("failed to set process as dumpable");

		s = SYNC_CHILD_READY;
		if (write(syncfd, &s, sizeof(s)) != sizeof(s)) {
			bail("failed to sync with parent: write(SYNC_CHILD_READY)");
		}

		/* Close sync pipes. */
		close(sync_child_pipe[0]);

		/* Finish executing, let the Go runtime take over. */
		return;
	}

	default:
		bail("unexpected jump value");
	}

	/* Should never be reached. */
	bail("should never be reached");
}
