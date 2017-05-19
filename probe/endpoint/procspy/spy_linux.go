package procspy

import (
	"bytes"
	"sync"

	"github.com/weaveworks/scope/probe/process"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 5000))
	},
}

type pnConnIter struct {
	pn    *ProcNet
	buf   *bytes.Buffer
	procs map[uint64]*Proc
}

func (c *pnConnIter) Next() *Connection {
	n := c.pn.Next()
	if n == nil {
		// Done!
		bufPool.Put(c.buf)
		return nil
	}
	if proc, ok := c.procs[n.Inode]; ok {
		n.Proc = *proc
	}
	return n
}

// NewConnectionScanner creates a new Linux ConnectionScanner
func NewConnectionScanner(walker process.Walker) ConnectionScanner {
	br := newBackgroundReader(walker)
	return &linuxScanner{br}
}

// NewSyncConnectionScanner creates a new synchronous Linux ConnectionScanner
func NewSyncConnectionScanner(walker process.Walker) ConnectionScanner {
	fr := newForegroundReader(walker)
	return &linuxScanner{fr}
}

type linuxScanner struct {
	r reader
}

func (s *linuxScanner) Connections(processes bool) (ConnIter, error) {
	// buffer for contents of /proc/<pid>/net/tcp
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	var procs map[uint64]*Proc
	if processes {
		var err error
		if procs, err = s.r.getWalkedProcPid(buf); err != nil {
			return nil, err
		}
	}

	if buf.Len() == 0 {
		readFile(procRoot+"/net/tcp", buf)
		readFile(procRoot+"/net/tcp6", buf)
	}

	return &pnConnIter{
		pn:    NewProcNet(buf.Bytes()),
		buf:   buf,
		procs: procs,
	}, nil
}

func (s *linuxScanner) Stop() {
	s.r.stop()
}
