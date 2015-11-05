package procspy

import (
	"bytes"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 5000))
	},
}

type pnConnIter struct {
	pn    *ProcNet
	buf   *bytes.Buffer
	procs map[uint64]Proc
}

func (c *pnConnIter) Next() *Connection {
	n := c.pn.Next()
	if n == nil {
		// Done!
		bufPool.Put(c.buf)
		return nil
	}
	if proc, ok := c.procs[n.Inode]; ok {
		n.Proc = proc
	}
	return n
}

// cbConnections sets Connections()
var cbConnections = func(processes bool) (ConnIter, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	readFile(procRoot+"/net/tcp", buf)
	readFile(procRoot+"/net/tcp6", buf)
	var procs map[uint64]Proc
	if processes {
		var err error
		if procs, err = walkProcPid(); err != nil {
			return nil, err
		}
	}
	return &pnConnIter{
		pn:    NewProcNet(buf.Bytes(), TCPEstablished),
		buf:   buf,
		procs: procs,
	}, nil
}
