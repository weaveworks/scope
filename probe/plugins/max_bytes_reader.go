package plugins

import (
	"io"
)

// MaxBytesReader is similar to net/http.MaxBytesReader, but lets us choose how
// to handle an overflow by providing an error. net/http.MaxBytesReader uses
// net/http internals to render a naff error message. There are other
// discrepancies with how this detects overflows. Not sure if that will cause
// issues. If you want to use it as part of an HTTP server, it's probably best
// to change it so you can provide a callback func, which renders your error
// message, as returning an error into the middle of the net/http server will
// not be useful.
func MaxBytesReader(r io.ReadCloser, maxBytes int64, err error) io.ReadCloser {
	if r == nil {
		return nil
	}

	return &maxBytesReader{
		ReadCloser:     r,
		bytesRemaining: maxBytes,
		err:            err,
	}
}

type maxBytesReader struct {
	io.ReadCloser
	bytesRemaining int64
	err            error // Callback when overflowing
}

func (r *maxBytesReader) Read(p []byte) (int, error) {
	if r.bytesRemaining <= 0 {
		return 0, r.err
	}
	if int64(len(p)) > r.bytesRemaining {
		p = p[0:r.bytesRemaining]
	}
	n, err := r.ReadCloser.Read(p)
	r.bytesRemaining -= int64(n)
	return n, err
}
