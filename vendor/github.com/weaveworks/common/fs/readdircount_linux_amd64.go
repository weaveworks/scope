// +build linux,amd64

package fs

import (
	"fmt"
	"os"
	"unsafe"

	"syscall"
)

func countDirEntries(buf []byte, n int) int {
	count := 0
	buf = buf[:n]
	for len(buf) > 0 {
		// see man page getdents(2) for struct linux_dirent64
		reclenOffset := unsafe.Offsetof(syscall.Dirent{}.Reclen)
		reclen := *(*uint16)(unsafe.Pointer(&buf[reclenOffset]))

		inoOffset := unsafe.Offsetof(syscall.Dirent{}.Ino)
		ino := *(*uint64)(unsafe.Pointer(&buf[inoOffset]))

		if int(reclen) > len(buf) {
			return count
		}
		buf = buf[reclen:]
		if ino == 0 {
			continue
		}
		count++
	}
	return count
}

// ReadDirCount is similar to ReadDirNames() and then counting with len() but
// it is optimized to avoid parsing the entries
func (realFS) ReadDirCount(dir string) (int, error) {
	buf := make([]byte, 4096)
	fh, err := os.Open(dir)
	if err != nil {
		return 0, err
	}
	defer fh.Close()

	openFilesCount := 0
	for {
		n, err := syscall.ReadDirent(int(fh.Fd()), buf)
		if err != nil {
			return 0, fmt.Errorf("ReadDirent() failed: %v", err)
		}
		if n == 0 {
			break
		}

		openFilesCount += countDirEntries(buf, n)
	}

	// "." and ".." don't count as files to be counted
	nDotFiles := 2
	return openFilesCount - nDotFiles, err
}
