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
type PIDTree struct {
	processes map[int]*Process
}

// Process represents a single process.
type Process struct {
	PID, PPID int
	Comm      string
	parent    *Process
	children  []*Process
}

// Hooks for mocking
var (
	readDir  = ioutil.ReadDir
	readFile = ioutil.ReadFile
)

// NewPIDTree returns a new PIDTree that can be polled.
func NewPIDTree(procRoot string) (*PIDTree, error) {
	dirEntries, err := readDir(procRoot)
	if err != nil {
		return nil, err
	}

	pt := PIDTree{processes: map[int]*Process{}}
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

		pt.processes[pid] = &Process{
			PID:  pid,
			PPID: ppid,
			Comm: comm,
		}
	}

	for _, child := range pt.processes {
		parent, ok := pt.processes[child.PPID]
		if !ok {
			// This can happen as listing proc is not a consistent snapshot
			continue
		}
		child.parent = parent
		parent.children = append(parent.children, child)
	}

	return &pt, nil
}

func (pt *PIDTree) getParent(pid int) (int, error) {
	proc, ok := pt.processes[pid]
	if !ok {
		return -1, fmt.Errorf("PID %d not found", pid)
	}

	return proc.PPID, nil
}

// allChildren returns a flattened list of child pids including the given pid
func (pt *PIDTree) allChildren(pid int) ([]int, error) {
	proc, ok := pt.processes[pid]
	if !ok {
		return []int{}, fmt.Errorf("PID %d not found", pid)
	}

	var result []int

	var f func(*Process)
	f = func(p *Process) {
		result = append(result, p.PID)
		for _, child := range p.children {
			f(child)
		}
	}

	f(proc)
	return result, nil
}

// ProcessTopology returns a process topology based on the current state of the PIDTree.
func (pt *PIDTree) ProcessTopology(hostID string) report.Topology {
	t := report.NewTopology()
	for pid, proc := range pt.processes {
		pidstr := strconv.Itoa(pid)
		nodeID := report.MakeProcessNodeID(hostID, pidstr)
		t.NodeMetadatas[nodeID] = report.NodeMetadata{
			"pid":  pidstr,
			"comm": proc.Comm,
		}
		if proc.PPID > 0 {
			t.NodeMetadatas[nodeID]["ppid"] = strconv.Itoa(proc.PPID)
		}
	}
	return t
}
