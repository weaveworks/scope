package tag

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/weaveworks/scope/report"
)

type pidTree struct {
	processes map[int]*process
}

type process struct {
	pid, ppid int
	parent    *process
	children  []*process
	comm      string
}

// Hooks for mocking
var (
	readDir  = ioutil.ReadDir
	readFile = ioutil.ReadFile
)

func newPIDTree(procRoot string) (*pidTree, error) {
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

func (pt *pidTree) getParent(pid int) (int, error) {
	proc, ok := pt.processes[pid]
	if !ok {
		return -1, fmt.Errorf("PID %d not found", pid)
	}

	return proc.ppid, nil
}

// allChildren returns a flattened list of child pids including the given pid
func (pt *pidTree) allChildren(pid int) ([]int, error) {
	proc, ok := pt.processes[pid]
	if !ok {
		return []int{}, fmt.Errorf("PID %d not found", pid)
	}

	var result []int

	var f func(*process)
	f = func(p *process) {
		result = append(result, p.pid)
		for _, child := range p.children {
			f(child)
		}
	}

	f(proc)
	return result, nil
}

func (pt *pidTree) processTopology(hostID string) report.Topology {
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
