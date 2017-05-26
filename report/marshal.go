package report

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/ugorji/go/codec"
)

// Include this in a struct to be able to call CodecDecodeSelf() before code generation
type dummySelfer struct{}

func (s *dummySelfer) CodecDecodeSelf(decoder *codec.Decoder) {
	panic("This shouldn't happen: perhaps something has gone wrong in code generation?")
}

func (s *dummySelfer) CodecEncodeSelf(encoder *codec.Encoder) {
	panic("This shouldn't happen: perhaps something has gone wrong in code generation?")
}

// WriteBinary writes a Report as a gzipped msgpack.
func (rep Report) WriteBinary(w io.Writer, compressionLevel int) error {
	gzwriter, err := gzip.NewWriterLevel(w, compressionLevel)
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
// codecHandle to decode it.
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
	if err := codec.NewDecoder(r, codecHandle).Decode(&rep); err != nil {
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

// ReadBytes reads bytes into a Report, using a codecHandle.
func (rep *Report) ReadBytes(buf []byte, codecHandle codec.Handle) error {
	return codec.NewDecoderBytes(buf, codecHandle).Decode(&rep)
}

// MakeFromBytes constructs a Report from a gzipped msgpack.
func MakeFromBytes(buf []byte) (*Report, error) {
	compressedSize := len(buf)
	r, err := gzip.NewReader(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	buf, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	uncompressedSize := len(buf)
	log.Debugf(
		"Received report sizes: compressed %d bytes, uncompressed %d bytes (%.2f%%)",
		compressedSize,
		uncompressedSize,
		float32(compressedSize)/float32(uncompressedSize)*100,
	)
	rep := MakeReport()
	if err := rep.ReadBytes(buf, &codec.MsgpackHandle{}); err != nil {
		return nil, err
	}
	return &rep, nil
}

// MakeFromFile construct a Report from a file, with the encoding
// determined by the extension (".msgpack" or ".json", with an
// optional ".gz").
func MakeFromFile(path string) (rpt Report, _ error) {
	f, err := os.Open(path)
	if err != nil {
		return rpt, err
	}
	defer f.Close()

	var (
		handle  codec.Handle
		gzipped bool
	)
	fileType := filepath.Ext(path)
	if fileType == ".gz" {
		gzipped = true
		fileType = filepath.Ext(strings.TrimSuffix(path, fileType))
	}
	switch fileType {
	case ".json":
		handle = &codec.JsonHandle{}
	case ".msgpack":
		handle = &codec.MsgpackHandle{}
	default:
		return rpt, fmt.Errorf("Unsupported file extension: %v", fileType)
	}

	var buf []byte
	if gzipped {
		r, err := gzip.NewReader(f)
		if err != nil {
			return rpt, err
		}
		buf, err = ioutil.ReadAll(r)
	} else {
		buf, err = ioutil.ReadAll(f)
	}
	if err != nil {
		return rpt, err
	}
	err = rpt.ReadBytes(buf, handle)

	return rpt, err
}
