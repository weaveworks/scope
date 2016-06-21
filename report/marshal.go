package report

import (
	"compress/gzip"
	"encoding/gob"
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

// ReadBinary reads bytes into a Report.
//
// Will decompress the binary if gzipped is true, and will use the given
// codecHandle to decode it. If codecHandle is nil, will decode as a gob.
func (rep *Report) ReadBinary(r io.Reader, gzipped bool, codecHandle codec.Handle) error {
	var err error
	var compressedSize, uncompressedSize uint64

	// We have historically had trouble with reports being too large. We are
	// keeping this instrumentation around to help us implement
	// weaveworks/scope#985.
	if log.GetLevel() == log.DebugLevel {
		r = byteCounter{next: r, count: &compressedSize}
	}
	if gzipped {
		r, err = gzip.NewReader(r)
		if err != nil {
			return err
		}
	}
	if log.GetLevel() == log.DebugLevel {
		r = byteCounter{next: r, count: &uncompressedSize}
	}
	var decoder func(interface{}) error
	if codecHandle != nil {
		decoder = codec.NewDecoder(r, codecHandle).Decode
	} else {
		decoder = gob.NewDecoder(r).Decode
	}
	if err := decoder(&rep); err != nil {
		return err
	}
	log.Debugf(
		"Received report sizes: compressed %d bytes, uncompressed %d bytes (%.2f%%)",
		compressedSize,
		uncompressedSize,
		float32(compressedSize)/float32(uncompressedSize)*100,
	)
	return nil
}

// MakeFromBinary constructs a Report from a gzipped msgpack.
func MakeFromBinary(r io.Reader) (*Report, error) {
	rep := MakeReport()
	if err := rep.ReadBinary(r, true, &codec.MsgpackHandle{}); err != nil {
		return nil, err
	}
	return &rep, nil
}
