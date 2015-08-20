package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
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

func getLocalAddr(pid, fd int) (addr uint32, port uint16, err error) {
	var (
		socket    string
		match     []string
		inode     int
		tcpFile   *os.File
		scanner   *bufio.Scanner
		candidate int
		port64    int64
		addr64    int64
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

		addr64, err = strconv.ParseInt(match[2], 16, 32)
		if err != nil {
			return
		}
		addr = uint32(addr64)

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
