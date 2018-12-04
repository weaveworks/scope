// +build linux

package tracer

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/iovisor/gobpf/elf"
)

/*
#include "../../tcptracer-bpf.h"
*/
import "C"

type tcpTracerStatus C.struct_tcptracer_status_t

const (
	// When reading kernel structs at different offsets, don't go over that
	// limit. This is an arbitrary choice to avoid infinite loops.
	threshold = 400

	// The source port is much further away in the inet sock.
	thresholdInetSock = 2000
)

// These constants should be in sync with the equivalent definitions in the ebpf program.
const (
	stateUninitialized C.__u64 = 0
	stateChecking              = 1 // status set by userspace, waiting for eBPF
	stateChecked               = 2 // status set by eBPF, waiting for userspace
	stateReady                 = 3 // fully initialized, all offset known
)

var stateString = map[C.__u64]string{
	stateUninitialized: "uninitialized",
	stateChecking:      "checking",
	stateChecked:       "checked",
	stateReady:         "ready",
}

// These constants should be in sync with the equivalent definitions in the ebpf program.
const (
	guessSaddr     C.__u64 = 0
	guessDaddr             = 1
	guessFamily            = 2
	guessSport             = 3
	guessDport             = 4
	guessNetns             = 5
	guessDaddrIPv6         = 6
)

var whatString = map[C.__u64]string{
	guessSaddr:     "source address",
	guessDaddr:     "destination address",
	guessFamily:    "family",
	guessSport:     "source port",
	guessDport:     "destination port",
	guessNetns:     "network namespace",
	guessDaddrIPv6: "destination address IPv6",
}

const listenIP = "127.0.0.2"

var zero uint64

type freePort struct {
	port uint16
	err  error
}

type fieldValues struct {
	saddr     uint32
	daddr     uint32
	sport     uint16
	dport     uint16
	netns     uint32
	family    uint16
	daddrIPv6 [4]uint32
}

func startServer() (chan struct{}, uint16, error) {
	// port 0 means we let the kernel choose a free port
	addr := fmt.Sprintf("%s:0", listenIP)
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return nil, 0, err
	}
	lport, err := strconv.Atoi(strings.Split(l.Addr().String(), ":")[1])
	if err != nil {
		return nil, 0, err
	}

	stop := make(chan struct{})
	go acceptV4(l, stop)

	return stop, uint16(lport), nil
}

func acceptV4(l net.Listener, stop chan struct{}) {
	for {
		_, ok := <-stop
		if ok {
			conn, err := l.Accept()
			if err != nil {
				l.Close()
				return
			}
			conn.Close()
		} else {
			// the main thread closed the channel, which signals there
			// won't be any more connections
			l.Close()
			return
		}
	}
}

func compareIPv6(a [4]C.__u32, b [4]uint32) bool {
	for i := 0; i < 4; i++ {
		if a[i] != C.__u32(b[i]) {
			return false
		}
	}
	return true
}

func ownNetNS() (uint64, error) {
	var s syscall.Stat_t
	if err := syscall.Stat("/proc/self/ns/net", &s); err != nil {
		return 0, err
	}
	return s.Ino, nil
}

func ipv6FromUint32Arr(ipv6Addr [4]uint32) net.IP {
	buf := make([]byte, 16)
	for i := 0; i < 16; i++ {
		buf[i] = *(*byte)(unsafe.Pointer((uintptr(unsafe.Pointer(&ipv6Addr[0])) + uintptr(i))))
	}
	return net.IP(buf)
}

func htons(a uint16) uint16 {
	var arr [2]byte
	binary.BigEndian.PutUint16(arr[:], a)
	return nativeEndian.Uint16(arr[:])
}

func generateRandomIPv6Address() (addr [4]uint32) {
	// multicast (ff00::/8) or link-local (fe80::/10) addresses don't work for
	// our purposes so let's choose a "random number" for the first 32 bits.
	//
	// chosen by fair dice roll.
	// guaranteed to be random.
	// https://xkcd.com/221/
	addr[0] = 0x87586031
	addr[1] = rand.Uint32()
	addr[2] = rand.Uint32()
	addr[3] = rand.Uint32()

	return
}

// tryCurrentOffset creates a IPv4 or IPv6 connection so the corresponding
// tcp_v{4,6}_connect kprobes get triggered and save the value at the current
// offset in the eBPF map
func tryCurrentOffset(module *elf.Module, mp *elf.Map, status *tcpTracerStatus, expected *fieldValues, stop chan struct{}) error {
	// for ipv6, we don't need the source port because we already guessed
	// it doing ipv4 connections so we use a random destination address and
	// try to connect to it
	expected.daddrIPv6 = generateRandomIPv6Address()

	ip := ipv6FromUint32Arr(expected.daddrIPv6)

	bindAddress := fmt.Sprintf("%s:%d", listenIP, expected.dport)
	if status.what != guessDaddrIPv6 {
		// signal the server that we're about to connect, this will block until
		// the channel is free so we don't overload the server
		stop <- struct{}{}
		conn, err := net.Dial("tcp4", bindAddress)
		if err != nil {
			return fmt.Errorf("error dialing %q: %v", bindAddress, err)
		}

		// get the source port assigned by the kernel
		sport, err := strconv.Atoi(strings.Split(conn.LocalAddr().String(), ":")[1])
		if err != nil {
			return fmt.Errorf("error converting source port: %v", err)
		}

		expected.sport = uint16(sport)

		// set SO_LINGER to 0 so the connection state after closing is
		// CLOSE instead of TIME_WAIT. In this way, they will disappear
		// from the conntrack table after around 10 seconds instead of 2
		// minutes
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetLinger(0)
		} else {
			return fmt.Errorf("not a tcp connection unexpectedly")
		}

		conn.Close()
	} else {
		conn, err := net.DialTimeout("tcp6", fmt.Sprintf("[%s]:9092", ip), 10*time.Millisecond)
		// Since we connect to a random IP, this will most likely fail.
		// In the unlikely case where it connects successfully, we close
		// the connection to avoid a leak.
		if err == nil {
			conn.Close()
		}
	}

	return nil
}

// checkAndUpdateCurrentOffset checks the value for the current offset stored
// in the eBPF map against the expected value, incrementing the offset if it
// doesn't match, or going to the next field to guess if it does
func checkAndUpdateCurrentOffset(module *elf.Module, mp *elf.Map, status *tcpTracerStatus, expected *fieldValues, maxRetries *int) error {
	// get the updated map value so we can check if the current offset is
	// the right one
	if err := module.LookupElement(mp, unsafe.Pointer(&zero), unsafe.Pointer(status)); err != nil {
		return fmt.Errorf("error reading tcptracer_status: %v", err)
	}

	if status.state != stateChecked {
		if *maxRetries == 0 {
			return fmt.Errorf("invalid guessing state while guessing %v, got %v expected %v",
				whatString[status.what], stateString[status.state], stateString[stateChecked])
		} else {
			*maxRetries--
			time.Sleep(10 * time.Millisecond)
			return nil
		}
	}

	switch status.what {
	case guessSaddr:
		if status.saddr == C.__u32(expected.saddr) {
			status.what = guessDaddr
		} else {
			status.offset_saddr++
			status.saddr = C.__u32(expected.saddr)
		}
		status.state = stateChecking
	case guessDaddr:
		if status.daddr == C.__u32(expected.daddr) {
			status.what = guessFamily
		} else {
			status.offset_daddr++
			status.daddr = C.__u32(expected.daddr)
		}
		status.state = stateChecking
	case guessFamily:
		if status.family == C.__u16(expected.family) {
			status.what = guessSport
			// we know the sport ((struct inet_sock)->inet_sport) is
			// after the family field, so we start from there
			status.offset_sport = status.offset_family
		} else {
			status.offset_family++
		}
		status.state = stateChecking
	case guessSport:
		if status.sport == C.__u16(htons(expected.sport)) {
			status.what = guessDport
		} else {
			status.offset_sport++
		}
		status.state = stateChecking
	case guessDport:
		if status.dport == C.__u16(htons(expected.dport)) {
			status.what = guessNetns
		} else {
			status.offset_dport++
		}
		status.state = stateChecking
	case guessNetns:
		if status.netns == C.__u32(expected.netns) {
			status.what = guessDaddrIPv6
		} else {
			status.offset_ino++
			// go to the next offset_netns if we get an error
			if status.err != 0 || status.offset_ino >= threshold {
				status.offset_ino = 0
				status.offset_netns++
			}
		}
		status.state = stateChecking
	case guessDaddrIPv6:
		if compareIPv6(status.daddr_ipv6, expected.daddrIPv6) {
			// at this point, we've guessed all the offsets we need,
			// set the status to "stateReady"
			status.state = stateReady
		} else {
			status.offset_daddr_ipv6++
			status.state = stateChecking
		}
	default:
		return fmt.Errorf("unexpected field to guess: %v", whatString[status.what])
	}

	// update the map with the new offset/field to check
	if err := module.UpdateElement(mp, unsafe.Pointer(&zero), unsafe.Pointer(status), 0); err != nil {
		return fmt.Errorf("error updating tcptracer_status: %v", err)
	}

	return nil
}

// guess expects elf.Module to hold a tcptracer-bpf object and initializes the
// tracer by guessing the right struct sock kernel struct offsets. Results are
// stored in the `tcptracer_status` map as used by the module.
//
// To guess the offsets, we create connections from localhost (127.0.0.1) to
// 127.0.0.2:$PORT, where we have a server listening. We store the current
// possible offset and expected value of each field in a eBPF map. Each
// connection will trigger the eBPF program attached to tcp_v{4,6}_connect
// where, for each field to guess, we store the value of
//     (struct sock *)skp + possible_offset
// in the eBPF map. Then, back in userspace (checkAndUpdateCurrentOffset()), we
// check that value against the expected value of the field, advancing the
// offset and repeating the process until we find the value we expect. Then, we
// guess the next field.
func guess(b *elf.Module) error {
	currentNetns, err := ownNetNS()
	if err != nil {
		return fmt.Errorf("error getting current netns: %v", err)
	}

	mp := b.Map("tcptracer_status")

	// pid & tid must not change during the guessing work: the communication
	// between ebpf and userspace relies on it
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	pidTgid := uint64(os.Getpid())<<32 | uint64(syscall.Gettid())

	status := &tcpTracerStatus{
		state:    stateChecking,
		pid_tgid: C.__u64(pidTgid),
	}

	// if we already have the offsets, just return
	err = b.LookupElement(mp, unsafe.Pointer(&zero), unsafe.Pointer(status))
	if err == nil && status.state == stateReady {
		return nil
	}

	stop, listenPort, err := startServer()
	if err != nil {
		return err
	}
	defer close(stop)

	// initialize map
	if err := b.UpdateElement(mp, unsafe.Pointer(&zero), unsafe.Pointer(status), 0); err != nil {
		return fmt.Errorf("error initializing tcptracer_status map: %v", err)
	}

	expected := &fieldValues{
		// 127.0.0.1
		saddr: 0x0100007F,
		// 127.0.0.2
		daddr: 0x0200007F,
		// will be set later
		sport:  0,
		dport:  listenPort,
		netns:  uint32(currentNetns),
		family: syscall.AF_INET,
	}

	// if the kretprobe for tcp_v4_connect() is configured with a too-low
	// maxactive, some kretprobe might be missing. In this case, we detect
	// it and try again.
	// See https://github.com/weaveworks/tcptracer-bpf/issues/24
	var maxRetries int = 100

	for status.state != stateReady {
		if err := tryCurrentOffset(b, mp, status, expected, stop); err != nil {
			return err
		}

		if err := checkAndUpdateCurrentOffset(b, mp, status, expected, &maxRetries); err != nil {
			return err
		}

		// Stop at a reasonable offset so we don't run forever.
		// Reading too far away in kernel memory is not a big deal:
		// probe_kernel_read() handles faults gracefully.
		if status.offset_saddr >= threshold || status.offset_daddr >= threshold ||
			status.offset_sport >= thresholdInetSock || status.offset_dport >= threshold ||
			status.offset_netns >= threshold || status.offset_family >= threshold ||
			status.offset_daddr_ipv6 >= threshold {
			return fmt.Errorf("overflow while guessing %v, bailing out", whatString[status.what])
		}
	}

	return nil
}
