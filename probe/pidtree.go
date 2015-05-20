package main

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
			// This can happen as listing proc is not a consistent snapshot
			continue
		}
		child.parent = parent
		parent.children = append(parent.children, child)
	}

	return &pt, nil
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
