// +build !linux

package elf

import (
	"fmt"
	"io"
)

type Module struct{}
type Kprobe struct{}
type CgroupProgram struct{}
type AttachType struct{}

func NewModule(fileName string) *Module {
	return nil
}

func NewModuleFromReader(fileReader io.ReaderAt) *Module {
	return nil
}

func (b *Module) EnableKprobe(secName string, maxactive int) error {
	return fmt.Errorf("not supported")
}

func (b *Module) IterKprobes() <-chan *Kprobe {
	return nil
}

func (b *Module) EnableKprobes(maxactive int) error {
	return fmt.Errorf("not supported")
}

func (b *Module) IterCgroupProgram() <-chan *CgroupProgram {
	return nil
}

func (b *Module) CgroupProgram(name string) *CgroupProgram {
	return nil
}

func (b *Module) AttachProgram(cgroupProg *CgroupProgram, cgroupPath string, attachType AttachType) error {
	return fmt.Errorf("not supported")
}
