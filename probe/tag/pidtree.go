package tag

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/weaveworks/scope/report"
)

// PIDTree represents all processes on the machine.
type PIDTree interface {
	GetParent(pid int) (int, error)
	ProcessTopology(hostID string) report.Topology
}

type pidTree struct {
	processes map[int]*process
}

// Process represents a single process.
type process struct {
	pid, ppid int
	comm      string
	parent    *process
	children  []*process
}

// Hooks for mocking
var (
	readDir  = ioutil.ReadDir
	readFile = ioutil.ReadFile
)

// NewPIDTree returns a new PIDTree that can be polled.
func NewPIDTree(procRoot string) (PIDTree, error) {
	dirEntries, err := readDir(procRoot)
	if err != nil {
		return nil, err
	}

	pt := pidTree{processes: map[int]*process{}}
	for _, dirEntry := range dirEntries {
		filename := dirEntry.Name()
		pid, err := strconv.Atoi(filename)
		if err != nil {
			continue
		}

		stat, err := readFile(path.Join(procRoot, filename, "stat"))
		if err != nil {
			continue
		}
		splits := strings.Split(string(stat), " ")
		ppid, err := strconv.Atoi(splits[3])
		if err != nil {
			return nil, err
		}

		comm := "(unknown)"
		if commBuf, err := readFile(path.Join(procRoot, filename, "comm")); err == nil {
			comm = string(commBuf)
		}

		pt.processes[pid] = &process{
			pid:  pid,
			ppid: ppid,
			comm: comm,
		}
	}

	for _, child := range pt.processes {
		parent, ok := pt.processes[child.ppid]
		if !ok {
			// This can happen as listing proc is not a consistent snapshot
			continue
		}
		child.parent = parent
		parent.children = append(parent.children, child)
	}

	return &pt, nil
}

// GetParent returns the pid of the parent process for a given pid
func (pt *pidTree) GetParent(pid int) (int, error) {
	proc, ok := pt.processes[pid]
	if !ok {
		return -1, fmt.Errorf("PID %d not found", pid)
	}

	return proc.ppid, nil
}

// ProcessTopology returns a process topology based on the current state of the PIDTree.
func (pt *pidTree) ProcessTopology(hostID string) report.Topology {
	t := report.NewTopology()
	for pid, proc := range pt.processes {
		pidstr := strconv.Itoa(pid)
		nodeID := report.MakeProcessNodeID(hostID, pidstr)
		t.NodeMetadatas[nodeID] = report.NodeMetadata{
			"pid":  pidstr,
			"comm": proc.comm,
		}
		if proc.ppid > 0 {
			t.NodeMetadatas[nodeID]["ppid"] = strconv.Itoa(proc.ppid)
		}
	}
	return t
}
