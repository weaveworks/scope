package fs

import (
	"io"
	"os"
)

// Open is a mockable version of os.Open
var Open = func(path string) (io.ReadWriteCloser, error) {
	return os.Open(path)
}
