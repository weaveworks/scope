package bpffs

import (
	"fmt"
	"syscall"
	"unsafe"
)

const BPFFSPath = "/sys/fs/bpf"

var FsMagicBPFFS int32

func init() {
	// https://github.com/coreutils/coreutils/blob/v8.27/src/stat.c#L275
	// https://github.com/torvalds/linux/blob/v4.8/include/uapi/linux/magic.h#L80
	magic := uint32(0xCAFE4A11)
	// 0xCAFE4A11 overflows an int32, which is what's expected by Statfs_t.Type in 32bit platforms.
	// To avoid conditional compilation for all 32bit/64bit platforms, we use an unsafe cast
	FsMagicBPFFS = *(*int32)(unsafe.Pointer(&magic))
}

// IsMounted checks if the BPF fs is mounted already
func IsMounted() (bool, error) {
	var data syscall.Statfs_t
	if err := syscall.Statfs(BPFFSPath, &data); err != nil {
		return false, fmt.Errorf("cannot statfs %q: %v", BPFFSPath, err)
	}
	return int32(data.Type) == FsMagicBPFFS, nil
}

// Mount mounts the BPF fs if not already mounted
func Mount() error {
	mounted, err := IsMounted()
	if err != nil {
		return err
	}
	if mounted {
		return nil
	}
	if err := syscall.Mount(BPFFSPath, BPFFSPath, "bpf", 0, ""); err != nil {
		return fmt.Errorf("error mounting %q: %v", BPFFSPath, err)
	}
	return nil
}
