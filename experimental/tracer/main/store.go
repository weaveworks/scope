package main

import (
	"math/rand"
	"sync"
	"log"
	"fmt"
	"sort"

	"github.com/msackman/skiplist"

	"github.com/weaveworks/scope/experimental/tracer/ptrace"
)

const epsilon = int64(5) * 1000 // milliseconds

// Traces are indexed by from addr, from port, and start time.
type key struct {
	fromAddr  uint32
	fromPort  uint16
	startTime int64
}

func (k key) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%x.%x.%x\"", k.fromAddr, k.fromPort, k.startTime)), nil
}

type trace struct {
	PID      int
	Key      key
	ServerDetails *ptrace.ConnectionDetails
	ClientDetails *ptrace.ConnectionDetails
	Children []*trace
	Level    int
}

type byKey []*trace

func (a byKey) Len() int           { return len(a) }
func (a byKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byKey) Less(i, j int) bool { return a[i].Key.startTime < a[j].Key.startTime }

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

	newTrace := &trace{
		PID: pid,
		Key: newKey(connection),
		ServerDetails: &connection.ConnectionDetails,
	}
	for _, child := range connection.Children {
		newTrace.Children = append(newTrace.Children, &trace{
			Level: 1,
			ClientDetails: &child.ConnectionDetails,
		})
	}

	newTraceKey := newTrace.Key

	// First, see if this new conneciton is a child of an existing connection.
	// This indicates we have a parent connection to attach to.
	// If not, insert this connection.
	if parentNode := s.traces.Get(newTraceKey); parentNode != nil {
		parentNode.Remove()
		parentTrace := parentNode.Value.(*trace)
		log.Printf(" Found parent trace: %+v", parentTrace)
		newTrace.Level = parentTrace.Level + 1
		parentTrace.Children = append(parentTrace.Children, newTrace)
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
			IncrementLevel(childTrace, newTrace.Level)
			newTrace.Children = append(newTrace.Children, childTrace)
		} else {
			s.traces.Insert(childTraceKey, newTrace)
		}
	}
}

func IncrementLevel(trace *trace, increment int) {
	trace.Level += increment
	for _, child := range trace.Children {
		IncrementLevel(child, increment)
	}
}

func (s *store) Traces() []*trace {
	s.RLock()
	defer s.RUnlock()

	traces := map[key]*trace{}
	var cur = s.traces.First()
	for cur != nil {
		trace := cur.Value.(*trace)
		traces[trace.Key] = trace
		cur = cur.Next()
	}

	result := []*trace{}
	for _, trace := range traces {
		result = append(result, trace)
	}

	sort.Sort(byKey(result))

	return result
}
