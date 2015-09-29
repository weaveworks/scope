package process

import (
	"bytes"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// processes that we will skip just based on the name
// note: we must be sure about these names, so keep it for kernel tasks,
//       system processes, command line tools, etc...
var skippedComms = []string{
	"(sd-pam)",
	"ModemManager",
	"NetworkManager",
	"Xorg",
	"accounts-daemon",
	"acpi_.*",
	"acpid",
	"agetty",
	"akonadi.*",
	"ata_sff",
	"baloo_file",
	"bash",
	"bioset",
	"bluetoothd",
	"cat",
	"cfg80211",
	"charger_manager",
	"cron",
	"crypto",
	"dbus",
	"dbus-daemon",
	"dbus-launch",
	"dconf-service",
	"deferwq",
	"devfreq_wq",
	"ecryptfs-kthrea",
	"ext4-rsv-conver",
	"fsnotifier64",
	"fsnotify_mark",
	"gconfd-.*",
	"gnome-keyring-.*",
	"gpg-agent",
	"gvfsd",
	"ipv6_addrconf",
	"irq/.*",
	"irqbalance",
	"jbd[[:digit:]]/.*",
	"kaccess",
	"kactivitymanage",
	"kauditd",
	"kblockd",
	"kdeconnectd",
	"kded.*",
	"kdeinit.*",
	"kdevtmpfs",
	"kerneloops",
	"kglobalaccel5",
	"khelper",
	"khugepaged",
	"khungtaskd",
	"kintegrityd",
	"krunner",
	"kscreen_backend",
	"ksmd",
	"ksmserver",
	"ksoftirqd/0",
	"ksuperkey",
	"init",
	"irq/.*",
	"jbd2/.*",
	"kswapd.*",
	"kwalletd.*",
	"kwin_.*",
	"kworker.*",
	"migration/[[:digit:]]",
	"rcu_bh",
	"rcu_sched",
	"rcuob/[[:digit:]]",
	"rcuos/[[:digit:]]",
	"rsyslogd",
	"rtkit-daemon",
	"scsi_eh_[[:digit:]]",
	"scsi_tmf_[[:digit:]]",
	"sddm",
	"sddm-helper",
	"ssh-agent",
	"start_kdeinit",
	"startkde",
	"systemd",
	"systemd-.*",
	"thermald",
	"udisksd",
	"uniq",
	"update-apt-.*",
	"upowerd",
	"watchdog/[[:digit:]]",
	"whoopsie",
	"wpa_supplicant",
	"writeback",
}

var skippedCommsRegexps = []*regexp.Regexp{}

func init() {
	for _, s := range skippedComms {
		skippedCommsRegexps = append(skippedCommsRegexps, regexp.MustCompile(s))
	}
}

func isSkippedComm(c *bytes.Buffer) bool {
	for _, regexp := range skippedCommsRegexps {
		if regexp.Match(c.Bytes()) {
			return true
		}
	}
	return false
}

// ProcReader is the Linux "/proc" reader
type ProcReader struct {
	processes   []*Process
	connections []*Connection

	proc      Dir
	handles   *filesCache
	withProcs bool

	sync.RWMutex
}

// NewReader creates a new /proc reader.
func NewReader(proc Dir, withProcs bool) *ProcReader {
	return &ProcReader{
		proc:      proc,
		handles:   newFilesCache(proc),
		withProcs: withProcs,
	}
}

// Read reads the processes and connections
func (r *ProcReader) Read() error {
	procsByInode := make(ProcsInodes)
	newProcesses := []*Process{}
	newConnections := []*Connection{}
	var fdStat syscall.Stat_t
	var namespaces = map[uint64]struct{}{}
	buf := bytes.Buffer{}
	tcpBuf := bytes.Buffer{}

	dirEntries, err := r.proc.ReadDirNames("")
	if err != nil {
		return err
	}

	appendConnectionsFrom := func(subFiles []string) error {
		for _, tcpFile := range subFiles {
			if err := r.handles.ReadInto(tcpFile, &tcpBuf, false); err != nil {
				return err
			}
		}
		return nil
	}

	for _, subdir := range dirEntries {
		readIntoBuffer := func(filename string, pinned bool) error {
			buf.Reset()
			return r.handles.ReadInto(path.Join(subdir, filename), &buf, pinned)
		}

		pid, err := strconv.Atoi(subdir)
		if err != nil {
			continue // Not a number, so not a PID subdir
		}

		comm := "(unknown)"
		if readIntoBuffer("comm", true) == nil {
			if isSkippedComm(&buf) {
				continue
			}
			comm = strings.TrimSpace(buf.String())
		}

		if readIntoBuffer("stat", false) != nil {
			continue
		}
		splits := strings.Fields(buf.String())
		ppid, err := strconv.Atoi(splits[3])
		if err != nil {
			return err
		}

		threads, err := strconv.Atoi(splits[19])
		if err != nil {
			return err
		}

		cmdline := ""
		if readIntoBuffer("cmdline", true) == nil {
			cmdlineBuf := bytes.Replace(buf.Bytes(), []byte{'\000'}, []byte{' '}, -1)
			cmdline = string(cmdlineBuf)
		}

		fdsSubdir := path.Join(subdir, "fd")
		fdNames, err := r.proc.ReadDirNames(fdsSubdir)
		if err != nil {
			continue // skip this process: it has probably disappeared...
		}

		inodes := []uint64{}
		for _, fdName := range fdNames {
			// Direct use of syscall.Stat() to save garbage.
			fdPath := path.Join(fdsSubdir, fdName)
			err = r.proc.Stat(fdPath, false, &fdStat)
			if err == nil && (fdStat.Mode&syscall.S_IFMT == syscall.S_IFSOCK) { // We want sockets only.
				inodes = append(inodes, fdStat.Ino)
			}
		}
		p := Process{
			PID:     pid,
			PPID:    ppid,
			Comm:    comm,
			Cmdline: cmdline,
			Threads: threads,
			Inodes:  inodes,
		}
		newProcesses = append(newProcesses, &p)

		if r.withProcs {
			for _, inode := range inodes {
				procsByInode[inode] = &p
			}
		}

		// Read network namespace, and if we haven't seen it before,
		// read /proc/<pid>/net/tcp
		if err := r.proc.Stat(path.Join(subdir, "/ns/net"), true, &fdStat); err != nil {
			continue
		}

		if _, ok := namespaces[fdStat.Ino]; !ok {
			namespaces[fdStat.Ino] = struct{}{}
			appendConnectionsFrom([]string{
				path.Join(subdir, "net", "tcp"),
				path.Join(subdir, "net", "tcp6")})
		}
	}

	if tcpBuf.Len() > 0 {
		pn := newNetReader(tcpBuf.Bytes(), tcpEstablished)
		for {
			conn := pn.Next()
			if conn == nil {
				break // Done!
			}
			newConn := conn.Copy()
			newConnections = append(newConnections, &newConn)
		}

		// perform the connection<->process matching with the inodes
		if r.withProcs {
			for _, conn := range newConnections {
				if proc, ok := procsByInode[conn.inode]; ok {
					conn.Process = *proc
				}
			}
		}
	}

	r.Lock()
	defer r.Unlock()
	r.processes = newProcesses
	r.connections = newConnections

	return nil
}

// Tick updates the processes and connections lists
func (r *ProcReader) Tick() error {
	return r.Read()
}

// Close closes the Linux "/proc" reader
func (r *ProcReader) Close() error {
	r.RLock()
	defer r.RUnlock()

	return r.handles.Close()
}

// Processes walks the /proc directory and marshalls the files into
// instances of Process, which it then passes one-by-one to the4
// supplied function. Processes() is only made public so that is
// can be tested.
func (r *ProcReader) Processes(f func(Process)) error {
	r.RLock()
	defer r.RUnlock()

	for _, p := range r.processes {
		f(*p)
	}
	return nil
}

// Connections walks through all the connections in the "/proc"
// Note: we never perform the process<->connection matching in Linux:
//       use the CachingReader instead
func (r *ProcReader) Connections(f func(Connection)) error {
	r.RLock()
	defer r.RUnlock()

	for _, c := range r.connections {
		f(*c)
	}
	return nil
}
