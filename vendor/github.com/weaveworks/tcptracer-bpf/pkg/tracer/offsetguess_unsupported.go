// +build !linux

package tracer

import (
	"fmt"

	"github.com/iovisor/gobpf/elf"
)

func guess(b *elf.Module) error {
	return fmt.Errorf("not supported on non-Linux systems")
}
