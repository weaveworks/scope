package report

import (
	"compress/gzip"
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

// ReadBinary reads into a Report from a gzipped msgpack.
func (rep *Report) ReadBinary(r io.Reader) error {
	reader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	if err := codec.NewDecoder(reader, &codec.MsgpackHandle{}).Decode(&rep); err != nil {
		return err
	}
	return nil
}

// MakeFromBinary constructs a Report from a gzipped msgpack.
func MakeFromBinary(r io.Reader) (*Report, error) {
	rep := MakeReport()
	if err := rep.ReadBinary(r); err != nil {
		return nil, err
	}
	return &rep, nil
}
