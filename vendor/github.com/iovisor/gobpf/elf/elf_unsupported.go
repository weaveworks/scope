// +build !linux

package elf

import (
	"fmt"
)

// not supported; dummy struct
type BPFKProbePerf struct{}
type SectionParams struct{}

func (b *Module) Load(parameters map[string]SectionParams) error {
	return fmt.Errorf("not supported")
}

func NewBpfPerfEvent(fileName string) *BPFKProbePerf {
	// not supported
	return nil
}

func (b *BPFKProbePerf) Load() error {
	return fmt.Errorf("not supported")
}

func (b *BPFKProbePerf) PollStart(mapName string, receiverChan chan []byte, lostChan chan uint64) {
	// not supported
	return
}

func (b *BPFKProbePerf) PollStop(mapName string) {
	// not supported
	return
}
