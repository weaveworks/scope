package process

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	linuxproc "github.com/c9s/goprocinfo/linux"

	"github.com/weaveworks/scope/common/fs"
	"github.com/weaveworks/scope/probe/host"
)

type walker struct {
	procRoot string
}

// NewWalker creates a new process Walker.
func NewWalker(procRoot string) Walker {
	return &walker{procRoot: procRoot}
}

func readStats(path string) (ppid, threads int, jiffies, rss uint64, err error) {
	var (
		buf                               []byte
		userJiffies, sysJiffies, rssPages uint64
	)
	buf, err = fs.ReadFile(path)
	if err != nil {
		return
	}
	splits := strings.Fields(string(buf))
	if len(splits) < 24 {
		err = fmt.Errorf("Invalid /proc/PID/stat")
		return
	}
	ppid, err = strconv.Atoi(splits[3])
	if err != nil {
		return
	}
	threads, err = strconv.Atoi(splits[19])
	if err != nil {
		return
	}
	userJiffies, err = strconv.ParseUint(splits[13], 10, 64)
	if err != nil {
		return
	}
	sysJiffies, err = strconv.ParseUint(splits[14], 10, 64)
	if err != nil {
		return
	}
	jiffies = userJiffies + sysJiffies
	rssPages, err = strconv.ParseUint(splits[23], 10, 64)
	if err != nil {
		return
	}
	rss = rssPages * uint64(os.Getpagesize())
	return
}

// Walk walks the supplied directory (expecting it to look like /proc)
// and marshalls the files into instances of Process, which it then
// passes one-by-one to the supplied function. Walk is only made public
// so that is can be tested.
func (w *walker) Walk(f func(Process, Process)) error {
	dirEntries, err := fs.ReadDirNames(w.procRoot)
	if err != nil {
		return err
	}

	for _, filename := range dirEntries {
		pid, err := strconv.Atoi(filename)
		if err != nil {
			continue
		}

		ppid, threads, jiffies, rss, err := readStats(path.Join(w.procRoot, filename, "stat"))
		if err != nil {
			continue
		}

		cmdline := ""
		if cmdlineBuf, err := cachedReadFile(path.Join(w.procRoot, filename, "cmdline")); err == nil {
			cmdlineBuf = bytes.Replace(cmdlineBuf, []byte{'\000'}, []byte{' '}, -1)
			cmdline = string(cmdlineBuf)
		}

		comm := "(unknown)"
		if commBuf, err := cachedReadFile(path.Join(w.procRoot, filename, "comm")); err == nil {
			comm = strings.TrimSpace(string(commBuf))
		}

		f(Process{
			PID:      pid,
			PPID:     ppid,
			Comm:     comm,
			Cmdline:  cmdline,
			Threads:  threads,
			Jiffies:  jiffies,
			RSSBytes: rss,
		}, Process{})
	}

	return nil
}

var previousStat = linuxproc.CPUStat{}

// GetDeltaTotalJiffies returns the number of jiffies that have passed since it
// was last called.  In that respect, it is side-effect-ful.
func GetDeltaTotalJiffies() (uint64, float64, error) {
	stat, err := linuxproc.ReadStat(host.ProcStat)
	if err != nil {
		return 0, 0.0, err
	}

	var (
		currentStat = stat.CPUStatAll
		prevTotal   = (previousStat.Idle + previousStat.IOWait + previousStat.User +
			previousStat.Nice + previousStat.System + previousStat.IRQ +
			previousStat.SoftIRQ + previousStat.Steal)
		currentTotal = (currentStat.Idle + currentStat.IOWait + currentStat.User +
			currentStat.Nice + currentStat.System + currentStat.IRQ +
			currentStat.SoftIRQ + currentStat.Steal)
	)
	previousStat = currentStat
	return currentTotal - prevTotal, float64(len(stat.CPUStats)) * 100., nil
}
