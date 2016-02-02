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
	if proc, ok := c.procs[n.inode]; ok {
		n.Proc = *proc
	}
	return n
}

// cbConnections sets Connections()
var cbConnections = func(processes bool, _ process.Walker) (ConnIter, error) {
	// buffer for contents of /proc/<pid>/net/tcp
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	var procs map[uint64]*Proc
	if processes {
		br, err := getBackgroundReader()
		if err != nil {
			return nil, err
		}
		procs = br.getWalkedProcPid(buf)
	}

	if buf.Len() == 0 {
		readFile(procRoot+"/net/tcp", buf)
		readFile(procRoot+"/net/tcp6", buf)
	}

	return &pnConnIter{
		pn:    NewProcNet(buf.Bytes(), tcpEstablished),
		buf:   buf,
		procs: procs,
	}, nil
}
