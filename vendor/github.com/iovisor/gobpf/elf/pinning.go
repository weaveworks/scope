// +build linux

package elf

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/iovisor/gobpf/bpffs"
)

/*
#include <linux/unistd.h>
#include <linux/bpf.h>
#include <stdlib.h>
#include <unistd.h>

extern __u64 ptr_to_u64(void *);

int bpf_pin_object(int fd, const char *pathname)
{
	union bpf_attr attr = {};

	attr.pathname = ptr_to_u64((void *)pathname);
	attr.bpf_fd = fd;

	return syscall(__NR_bpf, BPF_OBJ_PIN, &attr, sizeof(attr));
}
*/
import "C"

const (
	BPFDirGlobals = "globals" // as in iproute2's BPF_DIR_GLOBALS
	BPFFSPath     = "/sys/fs/bpf/"
)

func pinObject(fd int, namespace, object, name string) error {
	mounted, err := bpffs.IsMounted()
	if err != nil {
		return fmt.Errorf("error checking if %q is mounted: %v", BPFFSPath, err)
	}
	if !mounted {
		return fmt.Errorf("bpf fs not mounted at %q", BPFFSPath)
	}
	mapPath := filepath.Join(BPFFSPath, namespace, object, name)
	err = os.MkdirAll(filepath.Dir(mapPath), 0755)
	if err != nil {
		return fmt.Errorf("error creating map directory %q: %v", filepath.Dir(mapPath), err)
	}
	err = os.RemoveAll(mapPath)
	if err != nil {
		return fmt.Errorf("error removing old map file %q: %v", mapPath, err)
	}

	mapPathC := C.CString(mapPath)
	defer C.free(unsafe.Pointer(mapPathC))

	ret, err := C.bpf_pin_object(C.int(fd), mapPathC)
	if ret != 0 {
		return fmt.Errorf("error pinning object to %q: %v", mapPath, err)
	}
	return nil
}

func PinObjectGlobal(fd int, namespace, name string) error {
	return pinObject(fd, namespace, BPFDirGlobals, name)
}
