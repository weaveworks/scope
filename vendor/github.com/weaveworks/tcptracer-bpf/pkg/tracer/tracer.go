// +build linux

package tracer

import (
	"bytes"
	"fmt"

	bpflib "github.com/iovisor/gobpf/elf"
)

type Tracer struct {
	m           *bpflib.Module
	perfMapIPV4 *bpflib.PerfMap
	perfMapIPV6 *bpflib.PerfMap
	stopChan    chan struct{}
}

func TracerAsset() ([]byte, error) {
	buf, err := Asset("tcptracer-ebpf.o")
	if err != nil {
		return nil, fmt.Errorf("couldn't find asset: %s", err)
	}
	return buf, nil
}

func NewTracer(tcpEventCbV4 func(TcpV4), tcpEventCbV6 func(TcpV6)) (*Tracer, error) {
	buf, err := Asset("tcptracer-ebpf.o")
	if err != nil {
		return nil, fmt.Errorf("couldn't find asset: %s", err)
	}
	reader := bytes.NewReader(buf)

	m := bpflib.NewModuleFromReader(reader)
	if m == nil {
		return nil, fmt.Errorf("BPF not supported")
	}

	err = m.Load()
	if err != nil {
		return nil, err
	}

	err = m.EnableKprobes()
	if err != nil {
		return nil, err
	}

	channelV4 := make(chan []byte)
	channelV6 := make(chan []byte)

	perfMapIPV4, err := initializeIPv4(m, channelV4)
	if err != nil {
		return nil, fmt.Errorf("failed to init perf map for IPv4 events: %s", err)
	}

	perfMapIPV6, err := initializeIPv6(m, channelV6)
	if err != nil {
		return nil, fmt.Errorf("failed to init perf map for IPv6 events: %s", err)
	}

	perfMapIPV4.SetTimestampFunc(tcpV4Timestamp)
	perfMapIPV6.SetTimestampFunc(tcpV6Timestamp)

	stopChan := make(chan struct{})

	go func() {
		for {
			select {
			case <-stopChan:
				return
			case data := <-channelV4:
				tcpEventCbV4(tcpV4ToGo(&data))
			}
		}
	}()

	go func() {
		for {
			select {
			case <-stopChan:
				return
			case data := <-channelV6:
				tcpEventCbV6(tcpV6ToGo(&data))
			}
		}
	}()

	perfMapIPV4.PollStart()
	perfMapIPV6.PollStart()

	return &Tracer{
		m:           m,
		perfMapIPV4: perfMapIPV4,
		perfMapIPV6: perfMapIPV6,
		stopChan:    stopChan,
	}, nil
}

func (t *Tracer) Stop() {
	close(t.stopChan)
	t.perfMapIPV4.PollStop()
	t.perfMapIPV6.PollStop()
}

func initialize(module *bpflib.Module, eventMapName string, eventChan chan []byte) (*bpflib.PerfMap, error) {
	if err := guess(module); err != nil {
		return nil, fmt.Errorf("error guessing offsets: %v", err)
	}

	pm, err := bpflib.InitPerfMap(module, eventMapName, eventChan)
	if err != nil {
		return nil, fmt.Errorf("error initializing perf map for %q: %v", eventMapName, err)
	}

	return pm, nil

}

func initializeIPv4(module *bpflib.Module, eventChan chan []byte) (*bpflib.PerfMap, error) {
	return initialize(module, "tcp_event_ipv4", eventChan)
}

func initializeIPv6(module *bpflib.Module, eventChan chan []byte) (*bpflib.PerfMap, error) {
	return initialize(module, "tcp_event_ipv6", eventChan)
}
