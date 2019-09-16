package report

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	log "github.com/sirupsen/logrus"
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
	gzwriter := gzipWriterPool.Get().(*gzip.Writer)
	gzwriter.Reset(w)
	defer gzipWriterPool.Put(gzwriter)
	if err := codec.NewEncoder(gzwriter, &codec.MsgpackHandle{}).Encode(&rep); err != nil {
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

// buffer pools to reduce garbage-collection
var bufferPool = &sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}
var gzipWriterPool = &sync.Pool{
	// NewWriterLevel() only errors if the compression level is invalid, which can't happen here
	New: func() interface{} { w, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression); return w },
}

// ReadBinary reads bytes into a Report.
//
// Will decompress the binary if gzipped is true, and decode as
// msgpack if true, otherwise JSON
func (rep *Report) ReadBinary(ctx context.Context, r io.Reader, gzipped bool, msgpack bool) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "report.ReadBinary")
	defer span.Finish()
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
	if err := codec.NewDecoderBytes(buf.Bytes(), codecHandle(msgpack)).Decode(&rep); err != nil {
		return err
	}
	log.Debugf(
		"Received report sizes: compressed %d bytes, uncompressed %d bytes (%.2f%%)",
		compressedSize,
		uncompressedSize,
		float32(compressedSize)/float32(uncompressedSize)*100,
	)
	span.LogFields(otlog.Uint64("compressedSize", compressedSize), otlog.Int64("uncompressedSize", uncompressedSize))
	return nil
}

// MakeFromBinary constructs a Report from a gzipped msgpack.
func MakeFromBinary(ctx context.Context, r io.Reader) (*Report, error) {
	rep := MakeReport()
	if err := rep.ReadBinary(ctx, r, true, true); err != nil {
		return nil, err
	}
	return &rep, nil
}

// MakeFromFile construct a Report from a file, with the encoding
// determined by the extension (".msgpack" or ".json", with an
// optional ".gz").
func MakeFromFile(ctx context.Context, path string) (rpt Report, _ error) {
	f, err := os.Open(path)
	if err != nil {
		return rpt, err
	}
	defer f.Close()

	msgpack, gzipped, err := fileType(path)
	if err != nil {
		return rpt, err
	}

	err = rpt.ReadBinary(ctx, f, gzipped, msgpack)
	return rpt, err
}

// WriteToFile writes a Report to a file. The encoding is determined
// by the file extension (".msgpack" or ".json", with an optional
// ".gz").
func (rep *Report) WriteToFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	msgpack, gzipped, err := fileType(path)
	if err != nil {
		return err
	}

	var w io.Writer
	bufwriter := bufio.NewWriter(f)
	defer bufwriter.Flush()
	w = bufwriter
	if gzipped {
		gzwriter := gzipWriterPool.Get().(*gzip.Writer)
		gzwriter.Reset(w)
		defer gzipWriterPool.Put(gzwriter)
		defer gzwriter.Close()
		w = gzwriter
	}

	return codec.NewEncoder(w, codecHandle(msgpack)).Encode(rep)
}

func fileType(path string) (msgpack bool, gzipped bool, err error) {
	fileType := filepath.Ext(path)
	gzipped = false
	if fileType == ".gz" {
		gzipped = true
		fileType = filepath.Ext(strings.TrimSuffix(path, fileType))
	}
	switch fileType {
	case ".json":
		return false, gzipped, nil
	case ".msgpack":
		return true, gzipped, nil
	default:
		return false, false, fmt.Errorf("Unsupported file extension: %v", fileType)
	}
}

func codecHandle(msgpack bool) codec.Handle {
	if msgpack {
		return &codec.MsgpackHandle{}
	}
	return &codec.JsonHandle{}
}
