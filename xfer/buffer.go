package xfer

import (
	"bytes"
	"io"
	"sync"
	"sync/atomic"
)

// A Buffer is a reference counted bytes.Buffer, which belongs
// to a sync.Pool
type Buffer interface {
	io.Reader
	io.Writer

	// Get returns a new buffer sharing the contents of this
	// buffer, but with its own cursor.  Get also increases
	// the reference count.  It is safe for concurrent calls.
	Get() Buffer

	// Put decreases the reference count, and when it hits zero, puts the
	// buffer back in the pool.
	Put()
}

type baseBuffer struct {
	bytes.Buffer
	pool *sync.Pool
	refs int32
}

type dependentBuffer struct {
	*bytes.Buffer
	buf *baseBuffer
}

// NewBuffer creates a new buffer
func NewBuffer(pool *sync.Pool) Buffer {
	return &baseBuffer{
		pool: pool,
		refs: 0,
	}
}

// Get implements Buffer
func (b *baseBuffer) Get() Buffer {
	atomic.AddInt32(&b.refs, 1)
	return &dependentBuffer{
		Buffer: bytes.NewBuffer(b.Bytes()),
		buf:    b,
	}
}

// Put implements Buffer
func (b *baseBuffer) Put() {
	if atomic.AddInt32(&b.refs, -1) == 0 {
		b.Reset()
		b.pool.Put(b)
	}
}

// Get implements Buffer
func (b *dependentBuffer) Get() Buffer {
	return b.buf.Get()
}

// Put implements Buffer
func (b *dependentBuffer) Put() {
	b.buf.Put()
}

// NewBufferPool creates a new buffer pool.
func NewBufferPool() *sync.Pool {
	result := &sync.Pool{}
	result.New = func() interface{} {
		return NewBuffer(result)
	}
	return result
}
