package tag

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

type pidTree struct {
	processes map[int]*process
}

type process struct {
	pid, ppid int
	parent    *process
	children  []*process
}

var (
	readDir  = ioutil.ReadDir
	readFile = ioutil.ReadFile
)

func newRealPIDTree(procRoot string) (*pidTree, error) {
	dirEntries, err := readDir(procRoot)
	if err != nil {
		return nil, err
	}

	pt := pidTree{processes: map[int]*process{}}
	for _, dirEntry := range dirEntries {
		pid, err := strconv.Atoi(dirEntry.Name())
		if err != nil {
			continue
		}

		stat, err := readFile(path.Join(procRoot, dirEntry.Name(), "stat"))
		if err != nil {
			continue
		}

		splits := strings.Split(string(stat), " ")
		ppid, err := strconv.Atoi(splits[3])
		if err != nil {
			return nil, err
		}

		pt.processes[pid] = &process{pid: pid, ppid: ppid}
	}

	for _, child := range pt.processes {
		parent, ok := pt.processes[child.ppid]
		if !ok {
			continue // can happen: listing proc is not a consistent snapshot
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

// allChildren returns a flat list of child PIDs, including the given PID.
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
