// Package bytewriter implements writers that support concurrent writing within a fixed length block
//
// initially tried to use bytes.Buffer but the main restriction with that is that
// it does not allow the freedom to move around in the buffer. Further, it always
// writes at the end of the buffer
//
// another attempt was to maintain a position identifier that was always passed
// to any function that wrote anything and the function *had* to return the
// next writable location, which resulted in calls like
//
// ```go
// pos = writeString(buffer, pos, "mmv")
// ```
//
// which became unmaintainable after a while, and along with all the side
// maintainance looked extremely ugly
//
// then implemented a minimal buffer wrapper that gives the freedom
// to move around and write anywhere you want, which worked if you only need
// one position identifier, i.e. only one write operation happens at a time
//
// this implements a writer that supports multiple concurrent writes within a fixed length block
package bytewriter

// Writer defines an abstraction for an object that allows writing of binary
// values anywhere within a fixed range
type Writer interface {
	Bytes() []byte
	Len() int

	Write([]byte, int) (int, error)
	WriteVal(interface{}, int) (int, error)
	WriteString(string, int) (int, error)
	WriteInt32(int32, int) (int, error)
	WriteInt64(int64, int) (int, error)
	WriteUint32(uint32, int) (int, error)
	WriteUint64(uint64, int) (int, error)
	WriteFloat32(float32, int) (int, error)
	WriteFloat64(float64, int) (int, error)

	MustWrite([]byte, int) int
	MustWriteVal(interface{}, int) int
	MustWriteString(string, int) int
	MustWriteInt32(int32, int) int
	MustWriteInt64(int64, int) int
	MustWriteUint32(uint32, int) int
	MustWriteUint64(uint64, int) int
	MustWriteFloat32(float32, int) int
	MustWriteFloat64(float64, int) int
}
