package report

import (
	"compress/gzip"
	"encoding/gob"
	"io"

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

// ReadBinary reads bytes into a Report.
//
// Will decompress the binary if gzipped is true, and will use the given
// codecHandle to decode it. If codecHandle is nil, will decode as a gob.
func (rep *Report) ReadBinary(r io.Reader, gzipped bool, codecHandle codec.Handle) error {
	var err error
	if gzipped {
		r, err = gzip.NewReader(r)
		if err != nil {
			return err
		}
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
