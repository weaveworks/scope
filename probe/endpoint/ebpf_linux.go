//+build linux

package endpoint

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

import "C"

func findBpfObjectFile() (string, error) {
	var buf syscall.Utsname
	err := syscall.Uname(&buf)
	if err != nil {
		return "", err
	}

	// parse "ID=" in /etc/host-os-release (this is a bind mount of
	// /etc/os-release on the host)
	hostDistroFile, err := os.Open("/etc/host-os-release")
	if err != nil {
		return "", err
	}
	defer hostDistroFile.Close()

	scanner := bufio.NewScanner(hostDistroFile)
	var distro string
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "ID=") {
			distro = strings.TrimPrefix(scanner.Text(), "ID=")
			break
		}
	}
	if err = scanner.Err(); err != nil {
		return "", err
	}
	if distro == "" {
		return "", fmt.Errorf("distro ID not found")
	}

	arch := C.GoString((*C.char)(unsafe.Pointer(&buf.Machine[0])))
	release := C.GoString((*C.char)(unsafe.Pointer(&buf.Release[0])))
	fileName := fmt.Sprintf("/usr/libexec/scope/ebpf/%s/%s/%s/ebpf.o", distro, arch, release)

	return fileName, nil
}
