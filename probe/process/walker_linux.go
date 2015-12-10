package process

import (
	"bytes"
	"path"
	"strconv"
	"strings"

	"github.com/weaveworks/scope/common/fs"
)

type walker struct {
	procRoot string
}

// NewWalker creates a new process Walker.
func NewWalker(procRoot string) Walker {
	return &walker{procRoot: procRoot}
}

// Walk walks the supplied directory (expecting it to look like /proc)
// and marshalls the files into instances of Process, which it then
// passes one-by-one to the supplied function. Walk is only made public
// so that is can be tested.
func (w *walker) Walk(f func(Process)) error {
	dirEntries, err := fs.ReadDir(w.procRoot)
	if err != nil {
		return err
	}

	for _, dirEntry := range dirEntries {
		filename := dirEntry.Name()
		pid, err := strconv.Atoi(filename)
		if err != nil {
			continue
		}

		stat, err := fs.ReadFile(path.Join(w.procRoot, filename, "stat"))
		if err != nil {
			continue
		}
		splits := strings.Fields(string(stat))
		ppid, err := strconv.Atoi(splits[3])
		if err != nil {
			return err
		}

		threads, err := strconv.Atoi(splits[19])
		if err != nil {
			return err
		}

		cmdline := ""
		if cmdlineBuf, err := fs.ReadFile(path.Join(w.procRoot, filename, "cmdline")); err == nil {
			cmdlineBuf = bytes.Replace(cmdlineBuf, []byte{'\000'}, []byte{' '}, -1)
			cmdline = string(cmdlineBuf)
		}

		comm := "(unknown)"
		if commBuf, err := fs.ReadFile(path.Join(w.procRoot, filename, "comm")); err == nil {
			comm = strings.TrimSpace(string(commBuf))
		}

		f(Process{
			PID:     pid,
			PPID:    ppid,
			Comm:    comm,
			Cmdline: cmdline,
			Threads: threads,
		})
	}

	return nil
}
