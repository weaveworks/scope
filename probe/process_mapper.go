package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type processMapper interface {
	Key() string
	Map(pid uint) (string, error)
}

type identityMapper struct{}

func (m identityMapper) Key() string                  { return "identity" }
func (m identityMapper) Map(pid uint) (string, error) { return strconv.FormatUint(uint64(pid), 10), nil }

// cgroupMapper is a cgroup task mapper.
type cgroupMapper struct {
	sync.RWMutex
	root string
	d    map[uint]string
}

func newCgroupMapper(root string, interval time.Duration) *cgroupMapper {
	m := cgroupMapper{
		root: root,
		d:    map[uint]string{},
	}
	m.update()
	go m.loop(interval)
	return &m
}

func (m *cgroupMapper) Key() string { return "cgroup" }

// Map uses the cache to find the process name for pid. It is safe for
// concurrent use.
func (m *cgroupMapper) Map(pid uint) (string, error) {
	m.RLock()
	p, ok := m.d[pid]
	m.RUnlock()

	if !ok {
		return "", fmt.Errorf("no cgroup for PID %d", pid)
	}

	return p, nil
}

func (m *cgroupMapper) loop(d time.Duration) {
	for _ = range time.Tick(d) {
		m.update()
	}
}

func (m *cgroupMapper) update() {
	// We want to read "<root>/<processname>/tasks" files.
	fh, err := os.Open(m.root)
	if err != nil {
		log.Printf("cgroup mapper: %s", err)
		return
	}

	dirNames, err := fh.Readdirnames(-1)
	fh.Close()
	if err != nil {
		log.Printf("cgroup mapper: %s", err)
		return
	}

	pmap := map[uint]string{}
	for _, d := range dirNames {
		cg := normalizeCgroup(d)
		dirFilename := filepath.Join(m.root, d)

		s, err := os.Stat(dirFilename)
		if err != nil || !s.IsDir() {
			continue
		}

		taskFilename := filepath.Join(dirFilename, "tasks")

		f, err := os.Open(taskFilename)
		if err != nil {
			continue
		}

		r := bufio.NewReader(f)
		for {
			line, _, err := r.ReadLine()
			if err != nil {
				break // we expect an EOF
			}

			pid, err := strconv.ParseUint(string(line), 10, 64)
			if err != nil {
				log.Printf("continue mapper: %s", err)
				continue
			}

			pmap[uint(pid)] = cg
		}

		f.Close()
	}

	m.Lock()
	m.d = pmap
	m.Unlock()
}

var lxcRe = regexp.MustCompile(`^([^-]+)-([^-]+)-([A-Fa-f0-9]+)-([0-9]+)$`)

func normalizeCgroup(s string) string {
	// Format is currently "primarykey-secondarykey-revision-instance". We
	// want to collapse all instances (and maybe all revisions, in the future)
	// to the same node. So we remove the instance.
	if m := lxcRe.FindStringSubmatch(s); len(m) > 0 {
		return strings.Join([]string{m[1], m[2], m[3]}, "-")
	}
	return s
}
