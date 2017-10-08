package host

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/willdonnelly/passwd"
)

func getHostShellCmd() []string {
	if isProbeContainerized() {
		// Escape the container namespaces and jump into the ones from
		// the host's init process.
		// Note: There should be no need to enter into the host network
		// and PID namespace because we should already already be there
		// but it doesn't hurt.
		readPasswdCmd := []string{"/usr/bin/nsenter", "-t1", "-m", "--no-fork", "cat", "/etc/passwd"}
		uid, gid, shell := getRootUserDetails(readPasswdCmd)
		return []string{
			"/usr/bin/nsenter", "-t1", "-m", "-i", "-n", "-p", "--no-fork",
			"--setuid", uid,
			"--setgid", gid,
			shell, "-l",
		}
	}

	_, _, shell := getRootUserDetails([]string{"cat", "/etc/passwd"})
	return []string{shell, "-l"}
}

func getRootUserDetails(readPasswdCmd []string) (uid, gid, shell string) {
	uid = "0"
	gid = "0"
	shell = "/bin/sh"

	cmd := exec.Command(readPasswdCmd[0], readPasswdCmd[1:]...)
	cmdBuffer := &bytes.Buffer{}
	cmd.Stdout = cmdBuffer
	if err := cmd.Run(); err != nil {
		log.Warnf(
			"getRootUserDetails(): error running read passwd command %q: %s",
			strings.Join(readPasswdCmd, " "),
			err,
		)
		return
	}

	entries, err := passwd.ParseReader(cmdBuffer)
	if err != nil {
		log.Warnf("getRootUserDetails(): error parsing passwd: %s", err)
		return
	}

	entry, ok := entries["root"]
	if !ok {
		log.Warnf("getRootUserDetails(): no root entry in passwd")
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

	path := "/proc/self/ns/mnt"
	if err := syscall.Stat(path, &statT); err != nil {
		log.Warnf("isProbeContainerized(): stat() error on %q: %s", path, err)
		return false
	}
	selfMountNamespaceID := statT.Ino

	path = "/proc/1/ns/mnt"
	if err := syscall.Stat(path, &statT); err != nil {
		log.Warnf("isProbeContainerized(): stat() error on %q: %s", path, err)
		return false
	}

	return selfMountNamespaceID != statT.Ino
}
