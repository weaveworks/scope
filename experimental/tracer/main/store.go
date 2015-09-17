package main

import (
	"math/rand"
	"sync"
	"log"

	"github.com/msackman/skiplist"

	"github.com/weaveworks/scope/experimental/tracer/ptrace"
)

const epsilon = int64(5)

// Traces are indexed by from addr, from port, and start time.
type key struct {
	fromAddr  uint32
	fromPort  uint16
	startTime int64
}

type trace struct {
	pid      int
	root     *ptrace.Fd
	children []*trace
}

type store struct {
	sync.RWMutex
	traces *skiplist.SkipList
}

func newKey(fd *ptrace.Fd) key {
	var fromAddr uint32
	for _, b := range fd.FromAddr.To4() {
		fromAddr <<= 8
		fromAddr |= uint32(b)
	}

	return key{fromAddr, fd.FromPort, fd.Start}
}

func (l key) LessThan(other skiplist.Comparable) bool {
	r := other.(key)
	return l.fromAddr < r.fromAddr && l.fromPort < r.fromPort && l.startTime < r.startTime
}

func (l key) Equal(other skiplist.Comparable) bool {
	r := other.(key)
	if l.fromAddr != r.fromAddr || l.fromPort != r.fromPort {
		return false
	}

	diff := l.startTime - r.startTime
	return -epsilon < diff && diff < epsilon
}

func newStore() *store {
	return &store{traces: skiplist.New(rand.New(rand.NewSource(0)))}
}

func (s *store) RecordConnection(pid int, connection *ptrace.Fd) {
	s.Lock()
	defer s.Unlock()

	newTrace := &trace{pid: pid, root: connection}
	newTraceKey := newKey(connection)

	log.Printf("Recording trace: %+v", newTrace)

	// First, see if this new conneciton is a child of an existing connection.
	// This indicates we have a parent connection to attach to.
	// If not, insert this connection.
	if parentNode := s.traces.Get(newTraceKey); parentNode != nil {
		parentNode.Remove()
		parentTrace := parentNode.Value.(*trace)
		log.Printf(" Found parent trace: %+v", parentTrace)
		parentTrace.children = append(parentTrace.children, newTrace)
	} else {
		s.traces.Insert(newTraceKey, newTrace)
	}

	// Next, see if we already know about the child connections
	// If not, insert each of our children.
	for _, childConnection := range connection.Children {
		childTraceKey := newKey(childConnection)

		if childNode := s.traces.Get(childTraceKey); childNode != nil {
			childNode.Remove()
			childTrace := childNode.Value.(*trace)
			log.Printf(" Found child trace: %+v", childTrace)
			newTrace.children = append(newTrace.children, childTrace)
		} else {
			s.traces.Insert(childTraceKey, newTrace)
		}
	}
}

func (s *store) Traces() []*trace {
	s.RLock()
	defer s.RUnlock()

	var traces []*trace
	var cur = s.traces.First()
	for {
		traces = append(traces, cur.Value.(*trace))
		cur = cur.Next()
		if cur == nil {
			break
		}
	}
	return traces
}
