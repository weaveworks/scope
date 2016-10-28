package endpoint

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"sync"
	"unsafe"

	log "github.com/Sirupsen/logrus"
	"github.com/iovisor/gobpf"
)

/*
#cgo CFLAGS: -I/usr/include/bcc/compat
#cgo LDFLAGS: -lbcc
#include <bcc/bpf_common.h>
#include <bcc/libbpf.h>
#include <bcc/perf_reader.h>
#include <stdio.h>

void *bpf_open_perf_buffer(perf_reader_raw_cb raw_cb, void *cb_cookie, int pid, int cpu);

extern void tcpEventCb();

#define TASK_COMM_LEN 16 // linux/sched.h

struct tcp_event_t {
        char ev_type[12];
        uint32_t pid;
        char comm[TASK_COMM_LEN];
        uint32_t saddr;
        uint32_t daddr;
        uint16_t sport;
        uint16_t dport;
        uint32_t netns;
};

*/
import "C"

const source string = `
#include <uapi/linux/ptrace.h>
#include <net/sock.h>
#include <net/inet_sock.h>
#include <net/net_namespace.h>
#include <bcc/proto.h>

#define TCP_EVENT_TYPE_CONNECT 1
#define TCP_EVENT_TYPE_ACCEPT  2
#define TCP_EVENT_TYPE_CLOSE   3

struct tcp_event_t {
	char ev_type[12];
	u32 pid;
	char comm[TASK_COMM_LEN];
	u32 saddr;
	u32 daddr;
	u16 sport;
	u16 dport;
	u32 netns;
};

BPF_PERF_OUTPUT(tcp_event);
BPF_HASH(connectsock, u64, struct sock *);
BPF_HASH(closesock, u64, struct sock *);

int kprobe__tcp_v4_connect(struct pt_regs *ctx, struct sock *sk)
{
	u64 pid = bpf_get_current_pid_tgid();

	// stash the sock ptr for lookup on return
	connectsock.update(&pid, &sk);

	return 0;
};

int kretprobe__tcp_v4_connect(struct pt_regs *ctx)
{
	int ret = PT_REGS_RC(ctx);
	u64 pid = bpf_get_current_pid_tgid();

	struct sock **skpp;
	skpp = connectsock.lookup(&pid);
	if (skpp == 0) {
		return 0;	// missed entry
	}

	if (ret != 0) {
		// failed to send SYNC packet, may not have populated
		// socket __sk_common.{skc_rcv_saddr, ...}
		connectsock.delete(&pid);
		return 0;
	}


	// pull in details
	struct sock *skp = *skpp;
	struct ns_common *ns;
	u32 saddr = 0, daddr = 0, net_ns_inum = 0;
	u16 sport = 0, dport = 0;
	bpf_probe_read(&sport, sizeof(sport), &((struct inet_sock *)skp)->inet_sport);
	bpf_probe_read(&saddr, sizeof(saddr), &skp->__sk_common.skc_rcv_saddr);
	bpf_probe_read(&daddr, sizeof(daddr), &skp->__sk_common.skc_daddr);
	bpf_probe_read(&dport, sizeof(dport), &skp->__sk_common.skc_dport);

// Get network namespace id, if kernel supports it
#ifdef CONFIG_NET_NS
	possible_net_t skc_net;
	bpf_probe_read(&skc_net, sizeof(skc_net), &skp->__sk_common.skc_net);
	bpf_probe_read(&net_ns_inum, sizeof(net_ns_inum), &skc_net.net->ns.inum);
#else
	net_ns_inum = 0;
#endif

	// output
	struct tcp_event_t evt = {
		.ev_type = "connect",
		.pid = pid >> 32,
		.saddr = saddr,
		.daddr = daddr,
		.sport = ntohs(sport),
		.dport = ntohs(dport),
		.netns = net_ns_inum,
	};

	bpf_get_current_comm(&evt.comm, sizeof(evt.comm));

	// do not send event if IP address is 0.0.0.0 or port is 0
	if (evt.saddr != 0 && evt.daddr != 0 && evt.sport != 0 && evt.dport != 0) {
		tcp_event.perf_submit(ctx, &evt, sizeof(evt));
	}

	connectsock.delete(&pid);

	return 0;
}

int kprobe__tcp_close(struct pt_regs *ctx, struct sock *sk)
{
	u64 pid = bpf_get_current_pid_tgid();

	// stash the sock ptr for lookup on return
	closesock.update(&pid, &sk);

	return 0;
};

int kretprobe__tcp_close(struct pt_regs *ctx)
{
	u64 pid = bpf_get_current_pid_tgid();

	struct sock **skpp;
	skpp = closesock.lookup(&pid);
	if (skpp == 0) {
		return 0;	// missed entry
	}

	closesock.delete(&pid);

	// pull in details
	struct sock *skp = *skpp;
	u16 family = 0;
	bpf_probe_read(&family, sizeof(family), &skp->__sk_common.skc_family);
	if (family != AF_INET) {
		return 0;
	}

	u32 saddr = 0, daddr = 0, net_ns_inum = 0;
	u16 sport = 0, dport = 0;
	bpf_probe_read(&saddr, sizeof(saddr), &skp->__sk_common.skc_rcv_saddr);
	bpf_probe_read(&daddr, sizeof(daddr), &skp->__sk_common.skc_daddr);
	bpf_probe_read(&sport, sizeof(sport), &((struct inet_sock *)skp)->inet_sport);
	bpf_probe_read(&dport, sizeof(dport), &skp->__sk_common.skc_dport);

// Get network namespace id, if kernel supports it
#ifdef CONFIG_NET_NS
	possible_net_t skc_net;
	bpf_probe_read(&skc_net, sizeof(skc_net), &skp->__sk_common.skc_net);
	bpf_probe_read(&net_ns_inum, sizeof(net_ns_inum), &skc_net.net->ns.inum);
#else
	net_ns_inum = 0;
#endif

	// output
	struct tcp_event_t evt = {
		.ev_type = "close",
		.pid = pid >> 32,
		.saddr = saddr,
		.daddr = daddr,
		.sport = ntohs(sport),
		.dport = ntohs(dport),
		.netns = net_ns_inum,
	};

	bpf_get_current_comm(&evt.comm, sizeof(evt.comm));

	// do not send event if IP address is 0.0.0.0 or port is 0
	if (evt.saddr != 0 && evt.daddr != 0 && evt.sport != 0 && evt.dport != 0) {
		tcp_event.perf_submit(ctx, &evt, sizeof(evt));
	}

	return 0;
}

int kretprobe__inet_csk_accept(struct pt_regs *ctx)
{
	struct sock *newsk = (struct sock *)PT_REGS_RC(ctx);
	u64 pid = bpf_get_current_pid_tgid();

	if (newsk == NULL)
		return 0;

	// check this is TCP
	u8 protocol = 0;
	// workaround for reading the sk_protocol bitfield:
	bpf_probe_read(&protocol, 1, (void *)((long)&newsk->sk_wmem_queued) - 3);
	if (protocol != IPPROTO_TCP)
		return 0;

	// pull in details
	u16 family = 0, lport = 0, dport = 0;
	u32 net_ns_inum = 0;
	bpf_probe_read(&family, sizeof(family), &newsk->__sk_common.skc_family);
	bpf_probe_read(&lport, sizeof(lport), &newsk->__sk_common.skc_num);
	bpf_probe_read(&dport, sizeof(dport), &newsk->__sk_common.skc_dport);

// Get network namespace id, if kernel supports it
#ifdef CONFIG_NET_NS
	possible_net_t skc_net;
	bpf_probe_read(&skc_net, sizeof(skc_net), &newsk->__sk_common.skc_net);
	bpf_probe_read(&net_ns_inum, sizeof(net_ns_inum), &skc_net.net->ns.inum);
#else
	net_ns_inum = 0;
#endif

	if (family == AF_INET) {
		struct tcp_event_t evt = {.ev_type = "accept", .netns = net_ns_inum};
		evt.pid = pid >> 32;
		bpf_probe_read(&evt.saddr, sizeof(u32),
			&newsk->__sk_common.skc_rcv_saddr);
		bpf_probe_read(&evt.daddr, sizeof(u32),
			&newsk->__sk_common.skc_daddr);
			evt.sport = lport;
		evt.dport = ntohs(dport);
		bpf_get_current_comm(&evt.comm, sizeof(evt.comm));
		tcp_event.perf_submit(ctx, &evt, sizeof(evt));
	}
	// else drop

	return 0;
}
`

var byteOrder binary.ByteOrder

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

// EbpfTracker contains the sets of open and closed TCP connections.
// Closed connections are kept in the `closedConnections` slice for one iteration of `walkConnections`.
type EbpfTracker struct {
	sync.Mutex
	readers     []*C.struct_perf_reader
	initialized bool
	dead        bool

	openConnections   map[string]ebpfConnection
	closedConnections []ebpfConnection
}

func newEbpfTracker(useEbpfConn bool) eventTracker {
	if !useEbpfConn {
		return &nilTracker{}
	}
	m := bpf.NewBpfModule(source, []string{})

	connectKprobe, err := m.LoadKprobe("kprobe__tcp_v4_connect")
	if err != nil {
		return &nilTracker{}
	}

	err = m.AttachKprobe("tcp_v4_connect", connectKprobe)
	if err != nil {
		return &nilTracker{}
	}

	connectKretprobe, err := m.LoadKprobe("kretprobe__tcp_v4_connect")
	if err != nil {
		return &nilTracker{}
	}

	err = m.AttachKretprobe("tcp_v4_connect", connectKretprobe)
	if err != nil {
		return &nilTracker{}
	}

	closeKprobe, err := m.LoadKprobe("kprobe__tcp_close")
	if err != nil {
		return &nilTracker{}
	}

	err = m.AttachKprobe("tcp_close", closeKprobe)
	if err != nil {
		return &nilTracker{}
	}

	closeKretprobe, err := m.LoadKprobe("kretprobe__tcp_close")
	if err != nil {
		return &nilTracker{}
	}

	err = m.AttachKretprobe("tcp_close", closeKretprobe)
	if err != nil {
		return &nilTracker{}
	}

	acceptKretprobe, err := m.LoadKprobe("kretprobe__inet_csk_accept")
	if err != nil {
		return &nilTracker{}
	}

	err = m.AttachKretprobe("inet_csk_accept", acceptKretprobe)
	if err != nil {
		return &nilTracker{}
	}

	t := bpf.NewBpfTable(0, m)
	readers, err := initPerfMap(t)
	if err != nil {
		return &nilTracker{}
	}
	tracker := &EbpfTracker{
		openConnections: map[string]ebpfConnection{},
		readers:         readers,
	}
	go tracker.run()

	ebpfTracker = tracker
	return tracker
}

func initPerfMap(table *bpf.BpfTable) ([]*C.struct_perf_reader, error) {
	fd := table.Config()["fd"].(int)
	keySize := table.Config()["key_size"].(uint64)
	leafSize := table.Config()["leaf_size"].(uint64)

	if keySize != 4 || leafSize != 4 {
		return nil, fmt.Errorf("wrong size")
	}

	key := make([]byte, keySize)
	leaf := make([]byte, leafSize)
	keyP := unsafe.Pointer(&key[0])
	leafP := unsafe.Pointer(&leaf[0])

	readers := []*C.struct_perf_reader{}

	cpu := 0
	res := 0
	for res == 0 {
		reader := C.bpf_open_perf_buffer((*[0]byte)(C.tcpEventCb), nil, -1, C.int(cpu))
		if reader == nil {
			return nil, fmt.Errorf("failed to get reader")
		}

		perfFd := C.perf_reader_fd(reader)

		readers = append(readers, (*C.struct_perf_reader)(reader))

		// copy perfFd into leaf, respecting the host endienness
		byteOrder.PutUint32(leaf, uint32(perfFd))

		r, err := C.bpf_update_elem(C.int(fd), keyP, leafP, 0)
		if r != 0 {
			return nil, fmt.Errorf("unable to initialize perf map: %v", err)
		}

		res = int(C.bpf_get_next_key(C.int(fd), keyP, keyP))
		cpu++
	}
	return readers, nil
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

func handleConnection(eventType string, tuple fourTuple, pid int, networkNamespace string) {
	ebpfTracker.handleConnection(eventType, tuple, pid, networkNamespace)
}

func (t *EbpfTracker) run() {
	for {
		C.perf_reader_poll(C.int(len(t.readers)), &t.readers[0], -1)
	}
}

func tcpEventCallback(cpu int, tcpEvent *C.struct_tcp_event_t) {
	typ := C.GoString(&tcpEvent.ev_type[0])
	pid := tcpEvent.pid & 0xffffffff

	saddrbuf := make([]byte, 4)
	daddrbuf := make([]byte, 4)

	binary.LittleEndian.PutUint32(saddrbuf, uint32(tcpEvent.saddr))
	binary.LittleEndian.PutUint32(daddrbuf, uint32(tcpEvent.daddr))

	sIP := net.IPv4(saddrbuf[0], saddrbuf[1], saddrbuf[2], saddrbuf[3])
	dIP := net.IPv4(daddrbuf[0], daddrbuf[1], daddrbuf[2], daddrbuf[3])

	sport := tcpEvent.sport
	dport := tcpEvent.dport
	netns := tcpEvent.netns

	tuple := fourTuple{sIP.String(), dIP.String(), uint16(sport), uint16(dport)}
	handleConnection(typ, tuple, int(pid), strconv.Itoa(int(netns)))
}

//export tcpEventCb
func tcpEventCb(cbCookie unsafe.Pointer, raw unsafe.Pointer, rawSize C.int) {
	// See src/cc/perf_reader.c:parse_sw()
	// struct {
	//     uint32_t size;
	//     char data[0];
	// };

	var tcpEvent C.struct_tcp_event_t

	if int(rawSize) != 4+int(unsafe.Sizeof(tcpEvent)) {
		fmt.Printf("invalid perf event: rawSize=%d != %d + %d\n", rawSize, 4, unsafe.Sizeof(tcpEvent))
		return
	}

	tcpEventCallback(0, (*C.struct_tcp_event_t)(raw))
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
