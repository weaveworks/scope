package procspy

import (
	"bytes"
	"fmt"
)

// ReadTCPFiles reads the proc files tcp and tcp6 for a pid
func ReadTCPFiles(pid int, buf *bytes.Buffer) (int64, error) {
	return 0, fmt.Errorf("not supported on non-Linux systems")
}

// ReadNetnsFromPID gets the netns inode of the specified pid
func ReadNetnsFromPID(pid int) (uint64, error) {
	return 0, fmt.Errorf("not supported on non-Linux systems")
}
