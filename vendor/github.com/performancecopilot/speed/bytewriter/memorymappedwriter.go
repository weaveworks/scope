package bytewriter

import (
	"os"
	"path/filepath"

	mmap "github.com/edsrzf/mmap-go"
	"github.com/pkg/errors"
)

// MemoryMappedWriter is a ByteBuffer that is also mapped into memory
type MemoryMappedWriter struct {
	*ByteWriter
	handle *os.File // file handle
	loc    string   // location of the memory mapped file
	size   int      // size in bytes
}

// NewMemoryMappedWriter will create and return a new instance of a MemoryMappedWriter
func NewMemoryMappedWriter(loc string, size int) (*MemoryMappedWriter, error) {
	if _, err := os.Stat(loc); err == nil {
		err = os.Remove(loc)
		if err != nil {
			return nil, err
		}
	}

	// ensure destination directory exists
	dir := filepath.Dir(loc)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(loc, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0644)
	if err != nil {
		return nil, err
	}

	l, err := f.Write(make([]byte, size))
	if err != nil {
		return nil, err
	}
	if l < size {
		return nil, errors.Errorf("Could not initialize %d bytes", size)
	}

	b, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		return nil, err
	}

	return &MemoryMappedWriter{
		NewByteWriterSlice(b),
		f,
		loc,
		size,
	}, nil
}

// Unmap will manually delete the memory mapping of a mapped buffer
func (b *MemoryMappedWriter) Unmap(removefile bool) error {
	m := mmap.MMap(b.buffer)
	if err := m.Unmap(); err != nil {
		return err
	}

	if err := b.handle.Close(); err != nil {
		return err
	}

	if removefile {
		if err := os.Remove(b.loc); err != nil {
			return err
		}
	}

	return nil
}
