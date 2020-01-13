// +build !linux

package elf

import "fmt"

type PerfMap struct{}

func InitPerfMap(b *Module, mapName string, receiverChan chan []byte) (*PerfMap, error) {
	return nil, fmt.Errorf("not supported")
}

func (pm *PerfMap) SetTimestampFunc(timestamp func(*[]byte) uint64) {}

func (pm *PerfMap) PollStart() {}

func (pm *PerfMap) PollStop() {}
