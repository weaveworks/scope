package process

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	linuxproc "github.com/c9s/goprocinfo/linux"

	"github.com/weaveworks/common/fs"
	"github.com/weaveworks/scope/probe/host"
	procconnector "github.com/weaveworks/scope/probe/process/proc_connector"
)

type walker struct {
	procRoot string

	procConnector *procconnector.ProcConnector
}

// NewWalker creates a new process Walker.
func NewWalker(procRoot string) Walker {
	procConnector, err := procconnector.NewProcConnector()

	if err != nil {
		log.Infof("Process walker cannot use the proc connector, fall back on the old /proc: %s", err)
	}

	return &walker{
		procRoot:      procRoot,
		procConnector: procConnector,
	}
}

func readStats(path string) (ppid, threads int, jiffies, rss, rssLimit uint64, err error) {
	var (
		buf                               []byte
		userJiffies, sysJiffies, rssPages uint64
	)
	buf, err = fs.ReadFile(path)
	if err != nil {
		return
	}
	splits := strings.Fields(string(buf))
	if len(splits) < 25 {
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
	rssLimit, err = strconv.ParseUint(splits[24], 10, 64)
	return
}

func readLimits(path string) (openFilesLimit uint64, err error) {
	buf, err := cachedReadFile(path)
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(buf), "\n") {
		if strings.HasPrefix(line, "Max open files") {
			splits := strings.Fields(line)
			if len(splits) < 6 {
				return 0, fmt.Errorf("Invalid /proc/PID/limits")
			}
			openFilesLimit, err := strconv.Atoi(splits[3])
			return uint64(openFilesLimit), err
		}
	}
	return 0, nil
}

// Walk walks the supplied directory (expecting it to look like /proc)
// and marshalls the files into instances of Process, which it then
// passes one-by-one to the supplied function. Walk is only made public
// so that is can be tested.
func (w *walker) Walk(f func(Process, Process)) error {
	if w.procConnector.IsRunning() {
		w.procConnector.Walk(func(p procconnector.Process) {
			filename := strconv.Itoa(p.Pid)
			w.walkOne(p.Pid, filename, p.Cmdline, p.Name, f)
		})
	} else {
		dirEntries, err := fs.ReadDirNames(w.procRoot)
		if err != nil {
			return err
		}

		for _, filename := range dirEntries {
			pid, err := strconv.Atoi(filename)
			if err != nil {
				/* this is not an error: some files in /proc
				 * are not about processes (e.g. /proc/mounts)
				 */
				continue
			}

			name, cmdline := procconnector.GetCmdline(pid)

			w.walkOne(pid, filename, cmdline, name, f)
		}
	}

	return nil
}

func (w *walker) walkOne(pid int, filename string, cmdline string, name string, f func(Process, Process)) {
	pr := Process{
		PID:     pid,
		Name:    name,
		Cmdline: cmdline,
	}

	// Always call the callback with the process. If the process has
	// terminated, we might not be able to get the dynamic details from
	// /proc but at least we will have the basic details (pid, name,
	// cmdline).
	defer func() { f(pr, Process{}) }()

	var err error
	pr.PPID, pr.Threads, pr.Jiffies, pr.RSSBytes, pr.RSSBytesLimit, err = readStats(path.Join(w.procRoot, filename, "stat"))
	if err != nil {
		return
	}

	openFiles, err := fs.ReadDirNames(path.Join(w.procRoot, filename, "fd"))
	if err != nil {
		return
	}
	pr.OpenFilesCount = len(openFiles)

	pr.OpenFilesLimit, err = readLimits(path.Join(w.procRoot, filename, "limits"))
	if err != nil {
		return
	}
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
