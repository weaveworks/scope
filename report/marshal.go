package report

import (
	"compress/gzip"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/ugorji/go/codec"
)

// WriteBinary writes a Report as a gzipped msgpack.
func (rep Report) WriteBinary(w io.Writer) error {
	gzwriter, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return err
	}
	if err = codec.NewEncoder(gzwriter, &codec.MsgpackHandle{}).Encode(&rep); err != nil {
		return err
	}
	gzwriter.Close() // otherwise the content won't get flushed to the output stream
	return nil
}

type byteCounter struct {
	next  io.Reader
	count *uint64
}

func (c byteCounter) Read(p []byte) (n int, err error) {
	n, err = c.next.Read(p)
	*c.count += uint64(n)
	return n, err
}

// ReadStats includes statistics related to unmarshalling a report
type ReadStats struct {
	CompressedSize, UncompressedSize uint64
}

// ReadBinaryWithStats reads bytes into a Report and obtains statistics when running in debug mode.
//
// Will decompress the binary if gzipped is true, and will use the given
// codecHandle to decode it.
func (rep *Report) ReadBinaryWithStats(r io.Reader, gzipped bool, codecHandle codec.Handle) (ReadStats, error) {
	var err error
	var compressedSize, uncompressedSize uint64

	if log.GetLevel() == log.DebugLevel {
		r = byteCounter{next: r, count: &compressedSize}
	}
	if gzipped {
		r, err = gzip.NewReader(r)
		if err != nil {
			return ReadStats{}, err
		}
	}
	if log.GetLevel() == log.DebugLevel {
		r = byteCounter{next: r, count: &uncompressedSize}
	}
	if err := codec.NewDecoder(r, codecHandle).Decode(&rep); err != nil {
		return ReadStats{}, err
	}
	return ReadStats{compressedSize, uncompressedSize}, nil
}

// ReadBinary is identical to ReadBinaryWithStats without obtaining statistics
func (rep *Report) ReadBinary(r io.Reader, gzipped bool, codecHandle codec.Handle) error {
	_, err := rep.ReadBinaryWithStats(r, gzipped, codecHandle)
	return err
}

// MakeFromBinary constructs a Report from a gzipped msgpack.
func MakeFromBinary(r io.Reader) (*Report, error) {
	rep := MakeReport()
	if err := rep.ReadBinary(r, true, &codec.MsgpackHandle{}); err != nil {
		return nil, err
	}
	return &rep, nil
}

// MakeFromBinaryWithStats constructs a Report from a gzipped msgpack and provides
// statistics when running in debug mode.
func MakeFromBinaryWithStats(r io.Reader) (*Report, ReadStats, error) {
	rep := MakeReport()
	stats, err := rep.ReadBinaryWithStats(r, true, &codec.MsgpackHandle{})
	if err != nil {
		return nil, stats, err
	}
	return &rep, stats, nil
}
