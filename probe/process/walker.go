package process

import (
	"bytes"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

// Process represents a single process.
type Process struct {
	PID, PPID int
	Comm      string
	Cmdline   string
	Threads   int
}

// Hooks exposed for mocking
var (
	ReadDir  = ioutil.ReadDir
	ReadFile = ioutil.ReadFile
)

// Walk walks the supplied directory (expecting it to look like /proc)
// and marshalls the files into instances of Process, which it then
// passes one-by-one to the supplied function.  Walk is only made public
// so that is can be tested.
var Walk = func(procRoot string, f func(*Process)) error {
	dirEntries, err := ReadDir(procRoot)
	if err != nil {
		return err
	}

	for _, dirEntry := range dirEntries {
		filename := dirEntry.Name()
		pid, err := strconv.Atoi(filename)
		if err != nil {
			continue
		}

		stat, err := ReadFile(path.Join(procRoot, filename, "stat"))
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
		if cmdlineBuf, err := ReadFile(path.Join(procRoot, filename, "cmdline")); err == nil {
			cmdlineBuf = bytes.Replace(cmdlineBuf, []byte{'\000'}, []byte{' '}, -1)
			cmdline = string(cmdlineBuf)
		}

		comm := "(unknown)"
		if commBuf, err := ReadFile(path.Join(procRoot, filename, "comm")); err == nil {
			comm = string(commBuf)
		}

		f(&Process{
			PID:     pid,
			PPID:    ppid,
			Comm:    comm,
			Cmdline: cmdline,
			Threads: threads,
		})
	}

	return nil
}
