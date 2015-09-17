package ptrace

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"time"
)

const (
	listening = iota
	incoming
	outgoing
)

const (
	socketPattern = `^socket:\[(\d+)\]$`
	tcpPattern    = `^\s*(?P<fd>\d+): (?P<localaddr>[A-F0-9]{8}):(?P<localport>[A-F0-9]{4}) ` +
		`(?P<remoteaddr>[A-F0-9]{8}):(?P<remoteport>[A-F0-9]{4}) (?:[A-F0-9]{2}) (?:[A-F0-9]{8}):(?:[A-F0-9]{8}) ` +
		`(?:[A-F0-9]{2}):(?:[A-F0-9]{8}) (?:[A-F0-9]{8}) \s+(?:\d+) \s+(?:\d+) (?P<inode>\d+)`
)

var (
	socketRegex = regexp.MustCompile(socketPattern)
	tcpRegexp   = regexp.MustCompile(tcpPattern)
)

type ConnectionDetails struct {
	direction int

	Start    int64
	Stop     int64
	sent     int64
	received int64

	FromAddr net.IP
	FromPort uint16
	ToAddr   net.IP
	ToPort   uint16
}

// Fd represents a connect and subsequent connections caused by it.
type Fd struct {
	fd     int
	closed bool

	ConnectionDetails

	// Fds are connections, and can have a causal-link to other Fds
	Children []*Fd
}

func getLocalAddr(pid, fd int) (addr net.IP, port uint16, err error) {
	var (
		socket    string
		match     []string
		inode     int
		tcpFile   *os.File
		scanner   *bufio.Scanner
		candidate int
		port64    int64
	)

	socket, err = os.Readlink(fmt.Sprintf("/proc/%d/fd/%d", pid, fd))
	if err != nil {
		return
	}

	match = socketRegex.FindStringSubmatch(socket)
	if match == nil {
		err = fmt.Errorf("Fd %d not a socket", fd)
		return
	}

	inode, err = strconv.Atoi(match[1])
	if err != nil {
		return
	}

	tcpFile, err = os.Open(fmt.Sprintf("/proc/%d/net/tcp", pid))
	if err != nil {
		return
	}
	defer tcpFile.Close()

	scanner = bufio.NewScanner(tcpFile)
	for scanner.Scan() {
		match = tcpRegexp.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}

		candidate, err = strconv.Atoi(match[6])
		if err != nil {
			return
		}
		if candidate != inode {
			continue
		}

		addr = make([]byte, 4)
		if _, err = hex.Decode(addr, []byte(match[2])); err != nil {
			return
		}
		addr[0], addr[1], addr[2], addr[3] = addr[3], addr[2], addr[1], addr[0]

		// use a 32 bit int for target, at the result is a uint16
		port64, err = strconv.ParseInt(match[3], 16, 32)
		if err != nil {
			return
		}
		port = uint16(port64)

		return
	}

	if err = scanner.Err(); err != nil {
		return
	}

	err = fmt.Errorf("Fd %d not found for proc %d", fd, pid)
	return
}

// in milliseconds
func now() int64 {
	return time.Now().UnixNano() / 1000000
}

// We want to get the listening address from /proc
func newListeningFd(pid, fd int) (*Fd, error) {
	localAddr, localPort, err := getLocalAddr(pid, fd)
	if err != nil {
		return nil, err
	}

	return &Fd{
		fd: fd,

		ConnectionDetails: ConnectionDetails{
			direction: listening,
			Start:     now(),
			ToAddr:    localAddr,
			ToPort:    uint16(localPort),
		},
	}, nil
}

// We intercepted a connect syscall
func newConnectionFd(pid, fd int, remoteAddr net.IP, remotePort uint16) (*Fd, error) {
	localAddr, localPort, err := getLocalAddr(pid, fd)
	if err != nil {
		return nil, err
	}

	return &Fd{
		fd: fd,

		ConnectionDetails: ConnectionDetails{
			direction: outgoing,
			Start:     now(),
			FromAddr:  localAddr,
			FromPort:  uint16(localPort),
			ToAddr:    remoteAddr,
			ToPort:    remotePort,
		},
	}, nil
}

// We got a new connection on a listening socket
func (fd *Fd) newConnection(addr net.IP, port uint16, newFd int) (*Fd, error) {
	if fd.direction != listening {
		return nil, fmt.Errorf("New connection on non-listening fd!")
	}

	return &Fd{
		fd: newFd,

		ConnectionDetails: ConnectionDetails{
			direction: incoming,
			Start:     now(),
			ToAddr:    fd.ToAddr,
			ToPort:    fd.ToPort,
			FromAddr:  addr,
			FromPort:  port,
		},
	}, nil
}

func (fd *Fd) close() {
	fd.closed = true
	fd.Stop = now()
}
