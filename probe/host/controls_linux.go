package host

import (
	"bytes"
	"os/exec"
	"syscall"

	"github.com/willdonnelly/passwd"
)

var hostShellCmd []string

func init() {
	if isProbeContainerized() {
		// Escape the container namespaces and jump into the ones from
		// the host's init process.
		// Note: There should be no need to enter into the host network
		// and PID namespace because we should already already be there
		// but it doesn't hurt.
		readPasswdCmd := []string{"/usr/bin/nsenter", "-t1", "-m", "--no-fork", "cat", "/etc/passwd"}
		uid, gid, shell := getRootUserDetails(readPasswdCmd)
		hostShellCmd = []string{
			"/usr/bin/nsenter", "-t1", "-m", "-i", "-n", "-p", "--no-fork",
			"--setuid", uid,
			"--setgid", gid,
			shell,
		}
		return
	}

	_, _, shell := getRootUserDetails([]string{"cat", "/etc/passwd"})
	hostShellCmd = []string{shell}
}

func getRootUserDetails(readPasswdCmd []string) (uid, gid, shell string) {
	uid = "0"
	gid = "0"
	shell = "/bin/sh"

	cmd := exec.Command(readPasswdCmd[0], readPasswdCmd[1:]...)
	cmdBuffer := &bytes.Buffer{}
	cmd.Stdout = cmdBuffer
	if err := cmd.Run(); err != nil {
		return
	}

	entries, err := passwd.ParseReader(cmdBuffer)
	if err != nil {
		return
	}

	entry, ok := entries["root"]
	if !ok {
		return
	}

	return entry.Uid, entry.Gid, entry.Shell
}

func isProbeContainerized() bool {
	// Figure out whether we are running in a container by checking if our
	// mount namespace matches the one from init process. This works
	// because, when containerized, the Scope probes run in the host's PID
	// namespace (and if they weren't due to a configuration problem, we
	// wouldn't have a way to escape the container anyhow).
	var statT syscall.Stat_t

	if err := syscall.Stat("/proc/self/ns/mnt", &statT); err != nil {
		return false
	}
	selfMountNamespaceID := statT.Ino

	if err := syscall.Stat("/proc/1/ns/mnt", &statT); err != nil {
		return false
	}

	return selfMountNamespaceID != statT.Ino
}
