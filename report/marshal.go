package report

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

// StdoutPublisher is useful when debugging
type StdoutPublisher struct{}

// Publish implements probe.ReportPublisher
func (StdoutPublisher) Publish(rep Report) error {
	handle := &codec.JsonHandle{Indent: 2}
	handle.Canonical = true
	return codec.NewEncoder(os.Stdout, handle).Encode(rep)
}

// WriteBinary writes a Report as a gzipped msgpack into a bytes.Buffer
func (rep Report) WriteBinary() (*bytes.Buffer, error) {
	w := &bytes.Buffer{}
	gzwriter, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
	if err != nil {
		return nil, err
	}
	if err = codec.NewEncoder(gzwriter, &codec.MsgpackHandle{}).Encode(&rep); err != nil {
		return nil, err
	}
	gzwriter.Close() // otherwise the content won't get flushed to the output stream
	return w, nil
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

// buffer pool to reduce garbage-collection
var bufferPool = &sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

// ReadBinary reads bytes into a Report.
//
// Will decompress the binary if gzipped is true, and will use the given
// codecHandle to decode it.
func (rep *Report) ReadBinary(r io.Reader, gzipped bool, codecHandle codec.Handle) error {
	var err error
	var compressedSize uint64

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
	// Read everything into memory before decoding: it's faster
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	uncompressedSize, err := buf.ReadFrom(r)
	if err != nil {
		return err
	}
	if err := rep.ReadBytes(buf.Bytes(), codecHandle); err != nil {
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
	buffer := bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer bufferPool.Put(buffer)
	uncompressedSize, err := buffer.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	log.Debugf(
		"Received report sizes: compressed %d bytes, uncompressed %d bytes (%.2f%%)",
		compressedSize,
		uncompressedSize,
		float32(compressedSize)/float32(uncompressedSize)*100,
	)
	rep := MakeReport()
	if err := rep.ReadBytes(buffer.Bytes(), &codec.MsgpackHandle{}); err != nil {
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

	handle, gzipped, err := handlerFromFileType(path)
	if err != nil {
		return rpt, err
	}

	err = rpt.ReadBinary(f, gzipped, handle)
	return rpt, err
}

// WriteToFile writes a Report to a file. The encoding is determined
// by the file extension (".msgpack" or ".json", with an optional
// ".gz").
func (rep *Report) WriteToFile(path string, compressionLevel int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	handle, gzipped, err := handlerFromFileType(path)
	if err != nil {
		return err
	}

	var w io.Writer
	bufwriter := bufio.NewWriter(f)
	defer bufwriter.Flush()
	w = bufwriter
	if gzipped {
		gzwriter, err := gzip.NewWriterLevel(w, compressionLevel)
		if err != nil {
			return err
		}
		defer gzwriter.Close()
		w = gzwriter
	}

	return codec.NewEncoder(w, handle).Encode(rep)
}

func handlerFromFileType(path string) (codec.Handle, bool, error) {
	fileType := filepath.Ext(path)
	gzipped := false
	if fileType == ".gz" {
		gzipped = true
		fileType = filepath.Ext(strings.TrimSuffix(path, fileType))
	}
	switch fileType {
	case ".json":
		return &codec.JsonHandle{}, gzipped, nil
	case ".msgpack":
		return &codec.MsgpackHandle{}, gzipped, nil
	default:
		return nil, false, fmt.Errorf("Unsupported file extension: %v", fileType)
	}
}
