// +build !linux !amd64

package fs

// ReadDirCount, unoptimized version
func (realFS) ReadDirCount(path string) (int, error) {
	names, err := ReadDirNames(path)
	return len(names), err
}
