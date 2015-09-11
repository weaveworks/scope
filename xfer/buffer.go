package xfer

import (
	"bytes"
	"sync"
	"sync/atomic"
)

// A Buffer is a reference counted bytes.Buffer, which belongs
// to a sync.Pool
type Buffer struct {
	bytes.Buffer
	pool *sync.Pool
	refs int32
}

// NewBuffer creates a new buffer
func NewBuffer(pool *sync.Pool) *Buffer {
	return &Buffer{
		pool: pool,
		refs: 0,
	}
}

// Get increases the reference count.  It is safe for concurrent calls.
func (b *Buffer) Get() {
	atomic.AddInt32(&b.refs, 1)
}

// Put decreases the reference count, and when it hits zero, puts the
// buffer back in the pool.
func (b *Buffer) Put() {
	if atomic.AddInt32(&b.refs, -1) == 0 {
		b.Reset()
		b.pool.Put(b)
	}
}

// NewBufferPool creates a new buffer pool.
func NewBufferPool() *sync.Pool {
	result := &sync.Pool{}
	result.New = func() interface{} {
		return NewBuffer(result)
	}
	return result
}
