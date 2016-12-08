//+build !linux

package bpf

import (
	"fmt"
)

// not supported; dummy struct
type BPFKProbePerf struct{}

func NewBpfPerfEvent(fileName string) *BPFKProbePerf {
	// not supported
	return nil
}

func (b *BPFKProbePerf) Load() error {
	return fmt.Errorf("not supported")
}

func (b *BPFKProbePerf) PollStart(mapName string, receiverChan chan []byte) {
	// not supported
	return
}

func (b *BPFKProbePerf) PollStop(mapName string) {
	// not supported
	return
}
