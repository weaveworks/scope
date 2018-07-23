package bytewriter

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
)

// assumes Little Endian, use _arch.go to set it to BigEndian for those archs
var byteOrder = binary.LittleEndian

// ByteWriter is a simple wrapper over a byte slice that supports writing anywhere
type ByteWriter struct {
	buffer []byte
}

// NewByteWriter creates a new ByteWriter of the specified size
func NewByteWriter(n int) *ByteWriter {
	return &ByteWriter{make([]byte, n)}
}

// NewByteWriterSlice creates a new ByteWriter using the passed slice
func NewByteWriterSlice(buffer []byte) *ByteWriter {
	return &ByteWriter{buffer}
}

// Len returns the maximum size of the ByteWriter
func (w *ByteWriter) Len() int { return len(w.buffer) }

// Bytes returns the internal byte array of the ByteWriter
func (w *ByteWriter) Bytes() []byte { return w.buffer }

func (w *ByteWriter) Write(data []byte, offset int) (int, error) {
	l := len(data)

	if offset+l > w.Len() {
		return -1, errors.Errorf("cannot write %v bytes at offset %v", l, offset)
	}

	for i := 0; i < l; i++ {
		w.buffer[offset+i] = data[i]
	}

	return offset + l, nil
}

// MustWrite is a write that will panic if Write returns an error
func (w *ByteWriter) MustWrite(data []byte, offset int) int {
	off, err := w.Write(data, offset)
	if err != nil {
		panic(err)
	}
	return off
}

// WriteVal writes an arbitrary value to the buffer
func (w *ByteWriter) WriteVal(val interface{}, offset int) (int, error) {
	if s, isString := val.(string); isString {
		return w.WriteString(s, offset)
	}

	buf := bytes.NewBuffer(make([]byte, 0))

	err := binary.Write(buf, byteOrder, val)
	if err != nil {
		return 0, err
	}

	return w.Write(buf.Bytes(), offset)
}

// MustWriteVal panics if WriteVal fails
func (w *ByteWriter) MustWriteVal(val interface{}, offset int) int {
	if off, err := w.WriteVal(val, offset); err != nil {
		panic(err)
	} else {
		return off
	}
}

// WriteString writes a string to the buffer
func (w *ByteWriter) WriteString(val string, offset int) (int, error) {
	_, err := w.Write([]byte(val), offset)
	return offset + len(val), err
}

// MustWriteString panics if WriteString fails
func (w *ByteWriter) MustWriteString(val string, offset int) int {
	if off, err := w.WriteString(val, offset); err != nil {
		panic(err)
	} else {
		return off
	}
}

// WriteInt32 writes an int32 to the buffer
func (w *ByteWriter) WriteInt32(val int32, offset int) (int, error) {
	return w.WriteVal(val, offset)
}

// MustWriteInt32 panics if WriteInt32 fails
func (w *ByteWriter) MustWriteInt32(val int32, offset int) int {
	return w.MustWriteVal(val, offset)
}

// WriteInt64 writes an int64 to the buffer
func (w *ByteWriter) WriteInt64(val int64, offset int) (int, error) {
	return w.WriteVal(val, offset)
}

// MustWriteInt64 panics if WriteInt64 fails
func (w *ByteWriter) MustWriteInt64(val int64, offset int) int {
	return w.MustWriteVal(val, offset)
}

// WriteUint32 writes an uint32 to the buffer
func (w *ByteWriter) WriteUint32(val uint32, offset int) (int, error) {
	return w.WriteVal(val, offset)
}

// MustWriteUint32 panics if WriteInt32 fails
func (w *ByteWriter) MustWriteUint32(val uint32, offset int) int {
	return w.MustWriteVal(val, offset)
}

// WriteUint64 writes an uint64 to the buffer
func (w *ByteWriter) WriteUint64(val uint64, offset int) (int, error) {
	return w.WriteVal(val, offset)
}

// MustWriteUint64 panics if WriteUint64 fails
func (w *ByteWriter) MustWriteUint64(val uint64, offset int) int {
	return w.MustWriteVal(val, offset)
}

// WriteFloat32 writes an float32 to the buffer
func (w *ByteWriter) WriteFloat32(val float32, offset int) (int, error) {
	return w.WriteVal(val, offset)
}

// MustWriteFloat32 panics if WriteFloat32 fails
func (w *ByteWriter) MustWriteFloat32(val float32, offset int) int {
	return w.MustWriteVal(val, offset)
}

// WriteFloat64 writes an float64 to the buffer
func (w *ByteWriter) WriteFloat64(val float64, offset int) (int, error) {
	return w.WriteVal(val, offset)
}

// MustWriteFloat64 panics if WriteFloat64 fails
func (w *ByteWriter) MustWriteFloat64(val float64, offset int) int {
	return w.MustWriteVal(val, offset)
}
