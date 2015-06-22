package process

import (
	"fmt"
)

// Tree represents all processes on the machine.
type Tree interface {
	GetParent(pid int) (int, error)
}

type tree struct {
	processes map[int]*Process
}

// NewTree returns a new Tree that can be polled.
func NewTree(procRoot string) (Tree, error) {
	pt := tree{processes: map[int]*Process{}}
	err := Walk(procRoot, func(p *Process) {
		pt.processes[p.PID] = p
	})

	return &pt, err
}

// GetParent returns the pid of the parent process for a given pid
func (pt *tree) GetParent(pid int) (int, error) {
	proc, ok := pt.processes[pid]
	if !ok {
		return -1, fmt.Errorf("PID %d not found", pid)
	}

	return proc.PPID, nil
}
