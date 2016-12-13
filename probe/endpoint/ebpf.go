package endpoint

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"
	"sync"
	"unsafe"

	log "github.com/Sirupsen/logrus"
	bpflib "github.com/kinvolk/gobpf-elf-loader/bpf"
)

var byteOrder binary.ByteOrder

type eventType uint32

func init() {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	if b == 0x04 {
		byteOrder = binary.LittleEndian
	} else {
		byteOrder = binary.BigEndian
	}
}

// These constants should be in sync with the equivalent definitions in the ebpf program.
const (
	_ eventType = iota
	EventConnect
	EventAccept
	EventClose
)

func (e eventType) String() string {
	switch e {
	case EventConnect:
		return "connect"
	case EventAccept:
		return "accept"
	case EventClose:
		return "close"
	default:
		return "unknown"
	}
}

// tcpEvent should be in sync with the struct in the ebpf maps.
type tcpEvent struct {
	// Timestamp must be the first field, the sorting depends on it
	Timestamp uint64

	CPU   uint64
	Type  uint32
	Pid   uint32
	Comm  [16]byte
	SAddr uint32
	DAddr uint32
	SPort uint16
	DPort uint16
	NetNS uint32
}

// An ebpfConnection represents a TCP connection
type ebpfConnection struct {
	tuple            fourTuple
	networkNamespace string
	incoming         bool
	pid              int
}

type eventTracker interface {
	handleConnection(eventType string, tuple fourTuple, pid int, networkNamespace string)
	hasDied() bool
	run()
	walkConnections(f func(ebpfConnection))
	initialize()
	isInitialized() bool
	stop()
}

var ebpfTracker *EbpfTracker

// nilTracker is a tracker that does nothing, and it implements the eventTracker interface.
// It is returned when the useEbpfConn flag is false.
type nilTracker struct{}

func (n nilTracker) handleConnection(_ string, _ fourTuple, _ int, _ string) {}
func (n nilTracker) hasDied() bool                                           { return true }
func (n nilTracker) run()                                                    {}
func (n nilTracker) walkConnections(f func(ebpfConnection))                  {}
func (n nilTracker) initialize()                                             {}
func (n nilTracker) isInitialized() bool                                     { return false }
func (n nilTracker) stop()                                                   {}

// EbpfTracker contains the sets of open and closed TCP connections.
// Closed connections are kept in the `closedConnections` slice for one iteration of `walkConnections`.
type EbpfTracker struct {
	sync.Mutex
	reader      *bpflib.BPFKProbePerf
	initialized bool
	dead        bool

	openConnections   map[string]ebpfConnection
	closedConnections []ebpfConnection
}

func newEbpfTracker(useEbpfConn bool) eventTracker {
	if !useEbpfConn {
		return &nilTracker{}
	}

	bpfObjectFile, err := findBpfObjectFile()
	if err != nil {
		log.Infof("cannot find BPF object file: %v", err)
		return &nilTracker{}
	}

	bpfPerfEvent := bpflib.NewBpfPerfEvent(bpfObjectFile)
	if bpfPerfEvent == nil {
		return &nilTracker{}
	}
	err = bpfPerfEvent.Load()
	if err != nil {
		log.Errorf("Error loading BPF program: %v", err)
		return &nilTracker{}
	}

	tracker := &EbpfTracker{
		openConnections: map[string]ebpfConnection{},
		reader:          bpfPerfEvent,
	}
	tracker.run()

	ebpfTracker = tracker
	return tracker
}

func (t *EbpfTracker) handleConnection(eventType string, tuple fourTuple, pid int, networkNamespace string) {
	t.Lock()
	defer t.Unlock()

	switch eventType {
	case "connect":
		conn := ebpfConnection{
			incoming:         false,
			tuple:            tuple,
			pid:              pid,
			networkNamespace: networkNamespace,
		}
		t.openConnections[tuple.String()] = conn
	case "accept":
		conn := ebpfConnection{
			incoming:         true,
			tuple:            tuple,
			pid:              pid,
			networkNamespace: networkNamespace,
		}
		t.openConnections[tuple.String()] = conn
	case "close":
		if deadConn, ok := t.openConnections[tuple.String()]; ok {
			delete(t.openConnections, tuple.String())
			t.closedConnections = append(t.closedConnections, deadConn)
		} else {
			log.Errorf("EbpfTracker error: unmatched close event: %s pid=%d netns=%s", tuple.String(), pid, networkNamespace)
		}
	}
}

func tcpEventCallback(event tcpEvent) {
	typ := eventType(event.Type)
	pid := event.Pid & 0xffffffff

	saddrbuf := make([]byte, 4)
	daddrbuf := make([]byte, 4)

	binary.LittleEndian.PutUint32(saddrbuf, uint32(event.SAddr))
	binary.LittleEndian.PutUint32(daddrbuf, uint32(event.DAddr))

	sIP := net.IPv4(saddrbuf[0], saddrbuf[1], saddrbuf[2], saddrbuf[3])
	dIP := net.IPv4(daddrbuf[0], daddrbuf[1], daddrbuf[2], daddrbuf[3])

	sport := event.SPort
	dport := event.DPort

	tuple := fourTuple{sIP.String(), dIP.String(), uint16(sport), uint16(dport)}

	log.Debugf("tcpEventCallback(%v, [%v:%v --> %v:%v], pid=%v, netNS=%v, cpu=%v, ts=%v)",
		typ.String(), tuple.fromAddr, tuple.fromPort, tuple.toAddr, tuple.toPort, pid, event.NetNS, event.CPU, event.Timestamp)
	ebpfTracker.handleConnection(typ.String(), tuple, int(pid), strconv.FormatUint(uint64(event.NetNS), 10))
}

// walkConnections calls f with all open connections and connections that have come and gone
// since the last call to walkConnections
func (t *EbpfTracker) walkConnections(f func(ebpfConnection)) {
	t.Lock()
	defer t.Unlock()

	for _, connection := range t.openConnections {
		f(connection)
	}
	for _, connection := range t.closedConnections {
		f(connection)
	}
	t.closedConnections = t.closedConnections[:0]
}

func (t *EbpfTracker) run() {
	channel := make(chan []byte)

	go func() {
		var event tcpEvent
		for {
			data := <-channel
			err := binary.Read(bytes.NewBuffer(data), byteOrder, &event)
			if err != nil {
				log.Errorf("failed to decode received data: %s\n", err)
				continue
			}
			tcpEventCallback(event)
		}
	}()

	t.reader.PollStart("tcp_event_ipv4", channel)
}

func (t *EbpfTracker) hasDied() bool {
	t.Lock()
	defer t.Unlock()

	return t.dead
}

func (t *EbpfTracker) initialize() {
	t.initialized = true
}

func (t *EbpfTracker) isInitialized() bool {
	return t.initialized
}

func (t *EbpfTracker) stop() {
	// TODO: stop the go routine in run()
}
