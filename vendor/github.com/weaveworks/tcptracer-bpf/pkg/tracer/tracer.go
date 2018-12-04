// +build linux

package tracer

import (
	"bytes"
	"fmt"
	"unsafe"

	bpflib "github.com/iovisor/gobpf/elf"
)

type Tracer struct {
	m           *bpflib.Module
	perfMapIPV4 *bpflib.PerfMap
	perfMapIPV6 *bpflib.PerfMap
	stopChan    chan struct{}
}

// maxActive configures the maximum number of instances of the probed functions
// that can be handled simultaneously.
// This value should be enough to handle typical workloads (for example, some
// amount of processes blocked on the accept syscall).
const maxActive = 128

func TracerAsset() ([]byte, error) {
	buf, err := Asset("tcptracer-ebpf.o")
	if err != nil {
		return nil, fmt.Errorf("couldn't find asset: %s", err)
	}
	return buf, nil
}

func NewTracer(cb Callback) (*Tracer, error) {
	buf, err := Asset("tcptracer-ebpf.o")
	if err != nil {
		return nil, fmt.Errorf("couldn't find asset: %s", err)
	}
	reader := bytes.NewReader(buf)

	m := bpflib.NewModuleFromReader(reader)
	if m == nil {
		return nil, fmt.Errorf("BPF not supported")
	}

	sectionParams := make(map[string]bpflib.SectionParams)
	sectionParams["maps/tcp_event_ipv4"] = bpflib.SectionParams{PerfRingBufferPageCount: 256}
	err = m.Load(sectionParams)
	if err != nil {
		return nil, err
	}

	err = m.EnableKprobes(maxActive)
	if err != nil {
		return nil, err
	}

	channelV4 := make(chan []byte)
	channelV6 := make(chan []byte)
	lostChanV4 := make(chan uint64)
	lostChanV6 := make(chan uint64)

	perfMapIPV4, err := initializeIPv4(m, channelV4, lostChanV4)
	if err != nil {
		return nil, fmt.Errorf("failed to init perf map for IPv4 events: %s", err)
	}

	perfMapIPV6, err := initializeIPv6(m, channelV6, lostChanV6)
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
				// On stop, stopChan will be closed but the other channels will
				// also be closed shortly after. The select{} has no priorities,
				// therefore, the "ok" value must be checked below.
				return
			case data, ok := <-channelV4:
				if !ok {
					return // see explanation above
				}
				cb.TCPEventV4(tcpV4ToGo(&data))
			case lost, ok := <-lostChanV4:
				if !ok {
					return // see explanation above
				}
				cb.LostV4(lost)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-stopChan:
				return
			case data, ok := <-channelV6:
				if !ok {
					return // see explanation above
				}
				cb.TCPEventV6(tcpV6ToGo(&data))
			case lost, ok := <-lostChanV6:
				if !ok {
					return // see explanation above
				}
				cb.LostV6(lost)
			}
		}
	}()

	return &Tracer{
		m:           m,
		perfMapIPV4: perfMapIPV4,
		perfMapIPV6: perfMapIPV6,
		stopChan:    stopChan,
	}, nil
}

func (t *Tracer) Start() {
	t.perfMapIPV4.PollStart()
	t.perfMapIPV6.PollStart()
}

func (t *Tracer) AddFdInstallWatcher(pid uint32) (err error) {
	var one uint32 = 1
	mapFdInstall := t.m.Map("fdinstall_pids")
	err = t.m.UpdateElement(mapFdInstall, unsafe.Pointer(&pid), unsafe.Pointer(&one), 0)
	return err
}

func (t *Tracer) RemoveFdInstallWatcher(pid uint32) (err error) {
	mapFdInstall := t.m.Map("fdinstall_pids")
	err = t.m.DeleteElement(mapFdInstall, unsafe.Pointer(&pid))
	return err
}

func (t *Tracer) Stop() {
	close(t.stopChan)
	t.perfMapIPV4.PollStop()
	t.perfMapIPV6.PollStop()
	t.m.Close()
}

func initialize(module *bpflib.Module, eventMapName string, eventChan chan []byte, lostChan chan uint64) (*bpflib.PerfMap, error) {
	if err := guess(module); err != nil {
		return nil, fmt.Errorf("error guessing offsets: %v", err)
	}

	pm, err := bpflib.InitPerfMap(module, eventMapName, eventChan, lostChan)
	if err != nil {
		return nil, fmt.Errorf("error initializing perf map for %q: %v", eventMapName, err)
	}

	return pm, nil

}

func initializeIPv4(module *bpflib.Module, eventChan chan []byte, lostChan chan uint64) (*bpflib.PerfMap, error) {
	return initialize(module, "tcp_event_ipv4", eventChan, lostChan)
}

func initializeIPv6(module *bpflib.Module, eventChan chan []byte, lostChan chan uint64) (*bpflib.PerfMap, error) {
	return initialize(module, "tcp_event_ipv6", eventChan, lostChan)
}
